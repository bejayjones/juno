// Package s3 provides an S3-backed PhotoStorage stub.
// Full implementation is deferred to Phase 9 (Sync Context).
package s3

import (
	"context"
	"io"

	"github.com/bejayjones/juno/pkg/storage"
)

// S3Storage is a placeholder that satisfies the PhotoStorage interface.
// Every method returns storage.ErrNotImplemented until Phase 9.
type S3Storage struct{}

// New returns an S3Storage stub.
func New() *S3Storage { return &S3Storage{} }

func (s *S3Storage) Save(_ context.Context, _, _ string, _ io.Reader) (string, error) {
	return "", storage.ErrNotImplemented
}

func (s *S3Storage) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, storage.ErrNotImplemented
}

func (s *S3Storage) Delete(_ context.Context, _ string) error {
	return storage.ErrNotImplemented
}
