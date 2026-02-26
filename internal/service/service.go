package service

import (
	"Proteus/internal/broker"
	"Proteus/internal/config"
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
	MarkAsDeleted(ctx context.Context, id string) error
	Cleaner(ctx context.Context)
}

func NewService(l logger.Logger, cfg config.Service, prod broker.Producer, ms meta_storage.MetaStorage, is image_storage.ImageStorage) Service {
	return impl.NewService(l, cfg, prod, ms, is)
}
