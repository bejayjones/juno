// Package storage defines the PhotoStorage interface for saving and retrieving
// inspection photos. Implementations are in sub-packages local and s3.
package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotImplemented is returned by storage drivers that are not yet available.
var ErrNotImplemented = errors.New("storage driver not implemented")

// AllowedMimeTypes contains the MIME types accepted for photo uploads.
var AllowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/heic": ".heic",
}

// MaxPhotoBytes is the maximum accepted photo size (20 MB).
const MaxPhotoBytes = 20 << 20 // 20 MiB

// PhotoStorage abstracts where photos are written and read.
// In local mode, photos live on disk. In cloud mode, photos live in S3.
type PhotoStorage interface {
	// Save writes photo data and returns the storage path (relative key) for
	// later retrieval. photoID should be a UUID; ext is the file extension
	// including the dot (e.g. ".jpg").
	Save(ctx context.Context, photoID, ext string, data io.Reader) (storagePath string, err error)

	// Get opens a photo by its storage path for streaming to the client.
	Get(ctx context.Context, storagePath string) (io.ReadCloser, error)

	// Delete removes a photo from storage. Returns nil if the photo does not exist.
	Delete(ctx context.Context, storagePath string) error
}
