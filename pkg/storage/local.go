package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(cfg *LocalConfig) (*LocalStorage, error) {
	if cfg == nil {
		cfg = &LocalConfig{BasePath: "./storage"}
	}

	if cfg.BasePath == "" {
		cfg.BasePath = "./storage"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: cfg.BasePath,
	}, nil
}

// Upload uploads a file to local filesystem
func (l *LocalStorage) Upload(ctx context.Context, opts UploadOptions) (*FileMetadata, error) {
	if opts.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	if opts.Reader == nil {
		return nil, fmt.Errorf("reader is required")
	}

	// Create full path
	fullPath := filepath.Join(l.basePath, opts.Key)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Calculate MD5 while copying
	hash := md5.New()
	multiWriter := io.MultiWriter(file, hash)

	// Copy data
	written, err := io.Copy(multiWriter, opts.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileMetadata{
		Key:          opts.Key,
		Size:         written,
		ContentType:  opts.ContentType,
		ETag:         fmt.Sprintf("%x", hash.Sum(nil)),
		LastModified: info.ModTime(),
		Metadata:     opts.Metadata,
	}, nil
}

// Download downloads a file from local filesystem
func (l *LocalStorage) Download(ctx context.Context, key string) (*File, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}

	fullPath := filepath.Join(l.basePath, key)

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &File{
		Metadata: FileMetadata{
			Key:          key,
			Size:         info.Size(),
			LastModified: info.ModTime(),
		},
		Reader: file,
	}, nil
}

// Delete deletes a file from local filesystem
func (l *LocalStorage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}

	fullPath := filepath.Join(l.basePath, key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns the local file path (not a real URL)
func (l *LocalStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is required")
	}

	fullPath := filepath.Join(l.basePath, key)

	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", key)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Return file:// URL
	return fmt.Sprintf("file://%s", fullPath), nil
}

// List lists files with a given prefix
func (l *LocalStorage) List(ctx context.Context, prefix string) ([]*FileMetadata, error) {
	searchPath := filepath.Join(l.basePath, prefix)

	var files []*FileMetadata

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(l.basePath, path)
		if err != nil {
			return err
		}

		files = append(files, &FileMetadata{
			Key:          relPath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})

		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return []*FileMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// Exists checks if a file exists
func (l *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key is required")
	}

	fullPath := filepath.Join(l.basePath, key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}
