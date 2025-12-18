package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations
type Storage interface {
	Upload(ctx context.Context, opts UploadOptions) (*FileMetadata, error)
	Download(ctx context.Context, key string) (*File, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	List(ctx context.Context, prefix string) ([]*FileMetadata, error)
	Exists(ctx context.Context, key string) (bool, error)
}

// UploadOptions contains options for uploading a file
type UploadOptions struct {
	Key         string            // Storage key (path)
	Reader      io.Reader         // File content reader
	ContentType string            // MIME type
	Size        int64             // File size in bytes
	Metadata    map[string]string // Custom metadata
	ACL         string            // Access control (public-read, private)
}

// FileMetadata represents metadata about a stored file
type FileMetadata struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// File represents a downloaded file
type File struct {
	Metadata FileMetadata
	Reader   io.ReadCloser
}

// Config represents storage configuration
type Config struct {
	Provider   string            `yaml:"provider"`    // s3, minio, local
	S3         *S3Config         `yaml:"s3,omitempty"`
	Local      *LocalConfig      `yaml:"local,omitempty"`
}

// S3Config contains S3/MinIO configuration
type S3Config struct {
	Region      string `yaml:"region"`
	Bucket      string `yaml:"bucket"`
	Endpoint    string `yaml:"endpoint"`     // For MinIO
	AccessKey   string `yaml:"access_key"`
	SecretKey   string `yaml:"secret_key"`
	UseSSL      bool   `yaml:"use_ssl"`
	ForcePathStyle bool `yaml:"force_path_style"` // Required for MinIO
}

// LocalConfig contains local filesystem configuration
type LocalConfig struct {
	BasePath string `yaml:"base_path"`
}

// NewStorage creates a new storage instance based on configuration
func NewStorage(config *Config) (Storage, error) {
	switch config.Provider {
	case "s3", "minio":
		return NewS3Storage(config.S3)
	case "local":
		return NewLocalStorage(config.Local)
	default:
		return NewLocalStorage(&LocalConfig{BasePath: "./storage"})
	}
}
