// Package local provides a disk-backed PhotoStorage implementation.
package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalDiskStorage stores photos under a configurable base directory.
// Storage paths are relative to the base, e.g. "abc123.jpg".
type LocalDiskStorage struct {
	basePath string
}

// New creates a LocalDiskStorage rooted at basePath. The directory is created
// on first use if it does not exist.
func New(basePath string) *LocalDiskStorage {
	return &LocalDiskStorage{basePath: basePath}
}

// Save writes data to basePath/<photoID><ext> and returns the relative path.
func (s *LocalDiskStorage) Save(_ context.Context, photoID, ext string, data io.Reader) (string, error) {
	if err := os.MkdirAll(s.basePath, 0o755); err != nil {
		return "", fmt.Errorf("storage mkdir: %w", err)
	}

	filename := photoID + ext
	dest := filepath.Join(s.basePath, filename)

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("storage create: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, data); err != nil {
		_ = os.Remove(dest) // clean up partial write
		return "", fmt.Errorf("storage write: %w", err)
	}
	return filename, nil
}

// Get opens the file at basePath/<storagePath> for reading.
func (s *LocalDiskStorage) Get(_ context.Context, storagePath string) (io.ReadCloser, error) {
	path := filepath.Join(s.basePath, storagePath)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("photo not found on disk: %s", storagePath)
		}
		return nil, fmt.Errorf("storage open: %w", err)
	}
	return f, nil
}

// Delete removes the file at basePath/<storagePath>. A missing file is not an error.
func (s *LocalDiskStorage) Delete(_ context.Context, storagePath string) error {
	path := filepath.Join(s.basePath, storagePath)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("storage delete: %w", err)
	}
	return nil
}
