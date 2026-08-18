// Package storage defines a backend-agnostic file storage interface.
// Production uses S3; local development can use a filesystem-backed
// implementation so contributors are not required to have AWS credentials.
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, key string, contentType string, body io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
