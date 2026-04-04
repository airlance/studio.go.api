package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/domain"
)

type minioStorage struct {
	client *minio.Client
	cfg    *config.Config
}

func NewMinioStorage(cfg *config.Config) (domain.Storage, error) {
	endpoint := cfg.Storage.Endpoint
	u, err := url.Parse(endpoint)
	if err == nil && u.Host != "" {
		endpoint = u.Host
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKeyID, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &minioStorage{
		client: client,
		cfg:    cfg,
	}

	// Ensure required buckets exist
	buckets := []string{"workspaces", "profiles"}
	ctx := context.Background()
	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}
		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
	}

	return storage, nil
}

func (s *minioStorage) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioStorage) GetPresignedURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
