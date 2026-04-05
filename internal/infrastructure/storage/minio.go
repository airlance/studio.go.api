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
	client        *minio.Client
	publicBaseURL string
}

func NewMinioStorage(cfg *config.Config) (domain.Storage, error) {
	endpoint := cfg.Storage.Endpoint
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		endpoint = u.Host
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKeyID, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	s := &minioStorage{
		client:        client,
		publicBaseURL: cfg.Storage.PublicBaseURL,
	}

	buckets := []string{"workspaces", "profiles"}
	ctx := context.Background()
	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}
		if !exists {
			if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}

		if err = client.SetBucketPolicy(ctx, bucket, publicReadPolicy(bucket)); err != nil {
			return nil, fmt.Errorf("failed to set bucket policy for %s: %w", bucket, err)
		}
	}

	return s, nil
}

func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Principal":"*",
			"Action":["s3:GetObject"],
			"Resource":["arn:aws:s3:::%s/*"]
		}]
	}`, bucket)
}

func (s *minioStorage) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioStorage) GetPresignedURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error) {
	base := s.publicBaseURL
	if base == "" {
		base = "http://" + s.client.EndpointURL().Host
	}
	return fmt.Sprintf("%s/%s/%s", base, bucketName, objectName), nil
}
