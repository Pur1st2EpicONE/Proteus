// Package minio contains the concrete implementation of the
// image_storage.ImageStorage interface using the official MinIO Go client.
package minio

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"

	"github.com/minio/minio-go/v7"
)

type ImageStorage struct {
	client     *minio.Client // client is the underlying MinIO client instance.
	bucketName string        // bucketName is the MinIO bucket where all images are stored.
	logger     logger.Logger // logger is used for structured logging of all storage operations.
}

// NewImageStorage constructs a new MinIO-backed ImageStorage
// with the provided logger, configuration and client.
func NewImageStorage(logger logger.Logger, config config.ImageStorage, imageDb *minio.Client) *ImageStorage {
	return &ImageStorage{client: imageDb, bucketName: config.MinIOBucket, logger: logger}
}

// Close shuts down idle connections in the MinIO client and logs the event.
func (s *ImageStorage) Close() {
	if s.client != nil && s.client.CredContext().Client != nil {
		s.client.CredContext().Client.CloseIdleConnections()
		s.logger.LogInfo("minio — idle connections closed", "layer", "repository.image_storage.minio")
	}
}

// Client returns the underlying MinIO client (exposed for tests only).
func (s *ImageStorage) Client() *minio.Client {
	return s.client
}

// BucketName returns the configured bucket name (exposed for tests only).
func (s *ImageStorage) BucketName() string {
	return s.bucketName
}
