package storage

import (
	"context"
	"io"
)

type PublicObjectStorage interface {
	PutPublicObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error)
}
