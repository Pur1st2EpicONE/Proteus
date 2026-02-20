package minio

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"

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

func (s *ImageStorage) UploadImage(ctx context.Context, image *models.Image) error {

	_, err := s.client.PutObject(ctx, s.bucketName, image.ObjectKey, bytes.NewReader(image.File), image.Size,
		minio.PutObjectOptions{ContentType: image.FileHeader.Header.Get("Content-Type")})
	if err != nil {
		return fmt.Errorf("minio PutObject failed: %w", err)
	}

	s.logger.LogInfo("file uploaded to MinIO", "bucket", s.bucketName, "imageID", image.ID, "size_bytes", image.Size)

	return nil

}

func (s *ImageStorage) DownloadImage(ctx context.Context, objectKey string) ([]byte, error) {

	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio GetObject failed: %w", err)
	}
	defer object.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(object); err != nil {
		return nil, fmt.Errorf("reading object data failed: %w", err)
	}

	s.logger.LogInfo("file downloaded from MinIO", "bucket", s.bucketName, "object", objectKey, "size_bytes", buf.Len())

	return buf.Bytes(), nil

}

func (s *ImageStorage) UploadProcessed(ctx context.Context, objectKey string, file []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, bytes.NewReader(file),
		int64(len(file)), minio.PutObjectOptions{ContentType: contentType})
	return err
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
