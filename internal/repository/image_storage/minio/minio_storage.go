package minio

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"

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

func (s *ImageStorage) Close() {
	if s.client != nil && s.client.CredContext().Client != nil {
		s.client.CredContext().Client.CloseIdleConnections()
		s.logger.LogInfo("minio — idle connections closed", "layer", "repository.image_storage.minio")
	}
}
