package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"Proteus/internal/config"
	"Proteus/internal/logger"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type ImageStorage struct {
	client     *minio.Client
	bucketName string
	logger     logger.Logger
}

func NewImageStorage(logger logger.Logger, config config.ImageStorage, imageDb *minio.Client) *ImageStorage {
	return &ImageStorage{client: imageDb, bucketName: config.MinIOBucket, logger: logger}
}

func (s *ImageStorage) Close() error {
	return nil
}

func (s *ImageStorage) UploadMultipart(ctx context.Context, file multipart.File, header *multipart.FileHeader, prefix string) (string, error) {

	objectName := fmt.Sprintf("%s%s%s", prefix, uuid.New().String(), filepath.Ext(header.Filename))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("minio PutObject failed: %w", err)
	}

	s.logger.LogInfo("file uploaded to MinIO", "bucket", s.bucketName, "object", objectName, "size_bytes", header.Size)

	return objectName, nil
}

func (s *ImageStorage) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	req, err := s.client.PresignedGetObject(ctx, s.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("cannot generate presigned URL: %w", err)
	}
	return req.String(), nil
}

func (s *ImageStorage) Delete(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio RemoveObject failed: %w", err)
	}

	s.logger.LogInfo("file deleted from MinIO", "object", objectKey)
	return nil
}
