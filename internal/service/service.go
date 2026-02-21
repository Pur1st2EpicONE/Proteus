package service

import (
	"Proteus/internal/broker"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"Proteus/internal/service/impl"
	"context"
)

type Service interface {
	UploadImage(ctx context.Context, image *models.Image) (string, error)
	ProcessImage(ctx context.Context, task models.ImageProcessTask) error
	GetImageMeta(ctx context.Context, id string) (string, string, error)
	DownloadImage(ctx context.Context, key string) ([]byte, string, error)
}

func NewService(logger logger.Logger, producer broker.Producer, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) Service {
	return impl.NewService(logger, producer, metaStorage, imageStorage)
}
