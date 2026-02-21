package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	appconfig "github.com/football.manager.api/internal/config"
)

type minioPublicStorage struct {
	httpClient    *http.Client
	endpoint      *url.URL
	region        string
	accessKeyID   string
	secretKey     string
	bucket        string
	publicBaseURL string
}

func NewMinIOPublicStorage(_ context.Context, cfg appconfig.StorageConfig) (PublicObjectStorage, error) {
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("storage bucket is required")
	}

	endpoint, err := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, err
	}

	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid storage endpoint: %w", err)
	}

	accessKeyID := strings.TrimSpace(cfg.AccessKeyID)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	if accessKeyID == "" || secretKey == "" {
		return nil, fmt.Errorf("storage credentials are required")
	}

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "us-east-1"
	}

	publicBaseURL := strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	if publicBaseURL == "" {
		publicBaseURL = strings.TrimRight(endpoint, "/") + "/" + bucket
	}

	return &minioPublicStorage{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		endpoint:      parsedEndpoint,
		region:        region,
		accessKeyID:   accessKeyID,
		secretKey:     secretKey,
		bucket:        bucket,
		publicBaseURL: publicBaseURL,
	}, nil
}

func (s *minioPublicStorage) PutPublicObject(ctx context.Context, key string, body io.Reader, _ int64, contentType string) (string, error) {
	objectKey := strings.Trim(strings.TrimSpace(key), "/")
	if objectKey == "" {
		return "", fmt.Errorf("storage object key is required")
	}

	payload, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload body: %w", err)
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	escapedKey := escapeObjectKey(objectKey)
	objectURL, canonicalURI := s.objectURL(escapedKey)

	amzTime := time.Now().UTC()
	amzDate := amzTime.Format("20060102T150405Z")
	shortDate := amzTime.Format("20060102")
	payloadHash := sha256Hex(payload)

	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := strings.Join([]string{
		"content-type:" + contentType,
		"host:" + s.endpoint.Host,
		"x-amz-content-sha256:" + payloadHash,
		"x-amz-date:" + amzDate,
		"",
	}, "\n")

	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		canonicalURI,
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", shortDate, s.region)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	signature := s.sign(shortDate, stringToSign)
	authorization := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKeyID,
		credentialScope,
		signedHeaders,
		signature,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to build upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("Authorization", authorization)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storage upload failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	return s.publicBaseURL + "/" + escapedKey, nil
}

func (s *minioPublicStorage) sign(shortDate, stringToSign string) string {
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), shortDate)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hmacSHA256(kSigning, stringToSign)
	return hex.EncodeToString(signature)
}

func (s *minioPublicStorage) objectURL(escapedObjectKey string) (string, string) {
	basePath := strings.TrimRight(s.endpoint.Path, "/")
	if basePath == "/" {
		basePath = ""
	}
	baseURL := strings.TrimRight(s.endpoint.Scheme+"://"+s.endpoint.Host+basePath, "/")
	fullPath := basePath + "/" + s.bucket + "/" + escapedObjectKey
	return baseURL + "/" + s.bucket + "/" + escapedObjectKey, fullPath
}

func normalizeEndpoint(raw string, useSSL bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("storage endpoint is required")
	}

	if !strings.Contains(value, "://") {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		value = scheme + "://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid storage endpoint: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid storage endpoint")
	}

	parsed.Path = path.Clean(parsed.Path)
	if parsed.Path == "." || parsed.Path == "/" {
		parsed.Path = ""
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}

func escapeObjectKey(key string) string {
	trimmed := strings.Trim(strings.TrimSpace(key), "/")
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
