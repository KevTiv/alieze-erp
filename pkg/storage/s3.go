package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements Storage interface for S3-compatible storage
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(cfg *S3Config) (*S3Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3 configuration is required")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	// Create custom endpoint resolver for MinIO
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		// Return default resolver
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Build AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
	})

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, opts UploadOptions) (*FileMetadata, error) {
	if opts.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	if opts.Reader == nil {
		return nil, fmt.Errorf("reader is required")
	}

	// Prepare metadata
	metadata := make(map[string]string)
	for k, v := range opts.Metadata {
		metadata[k] = v
	}

	// Determine ACL
	var acl types.ObjectCannedACL
	if opts.ACL == "public-read" {
		acl = types.ObjectCannedACLPublicRead
	} else {
		acl = types.ObjectCannedACLPrivate
	}

	// Upload object
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(opts.Key),
		Body:        opts.Reader,
		ContentType: aws.String(opts.ContentType),
		Metadata:    metadata,
		ACL:         acl,
	}

	if opts.Size > 0 {
		input.ContentLength = aws.Int64(opts.Size)
	}

	result, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Get metadata
	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(opts.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return &FileMetadata{
		Key:          opts.Key,
		Size:         aws.ToInt64(headOutput.ContentLength),
		ContentType:  aws.ToString(headOutput.ContentType),
		ETag:         aws.ToString(result.ETag),
		LastModified: aws.ToTime(headOutput.LastModified),
		Metadata:     headOutput.Metadata,
	}, nil
}

// Download downloads a file from S3
func (s *S3Storage) Download(ctx context.Context, key string) (*File, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Get object
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return &File{
		Metadata: FileMetadata{
			Key:          key,
			Size:         aws.ToInt64(result.ContentLength),
			ContentType:  aws.ToString(result.ContentType),
			ETag:         aws.ToString(result.ETag),
			LastModified: aws.ToTime(result.LastModified),
			Metadata:     result.Metadata,
		},
		Reader: result.Body,
	}, nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL generates a presigned URL for temporary access
func (s *S3Storage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is required")
	}

	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	// Generate presigned GET request
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// List lists files with a given prefix
func (s *S3Storage) List(ctx context.Context, prefix string) ([]*FileMetadata, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]*FileMetadata, 0, len(result.Contents))
	for _, obj := range result.Contents {
		files = append(files, &FileMetadata{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			ETag:         aws.ToString(obj.ETag),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}

	return files, nil
}

// Exists checks if a file exists
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key is required")
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if error is NoSuchKey
		return false, nil
	}

	return true, nil
}
