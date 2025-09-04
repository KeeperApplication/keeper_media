package service

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"keeper.media/internal/config"
)

type GcsService struct {
	client     *storage.Client
	projectID  string
	bucketName string
}

func NewGcsService(cfg *config.Config) (*GcsService, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.GCSServiceAccount))
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}

	return &GcsService{
		client:     client,
		projectID:  cfg.GCSProjectID,
		bucketName: cfg.GCSBucketName,
	}, nil
}

func (s *GcsService) GenerateV4UploadURL(objectName string, contentType string) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: contentType,
	}

	url, err := s.client.Bucket(s.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("Bucket.SignedURL: %w", err)
	}

	return url, nil
}

func (s *GcsService) ReadObject(ctx context.Context, objectName string) (*storage.Reader, error) {
	reader, err := s.client.Bucket(s.bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object.NewReader: %w", err)
	}
	return reader, nil
}
