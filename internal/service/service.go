package service

import (
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"Proteus/internal/service/impl"
)

type Service interface {
}

func NewService(logger logger.Logger, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) Service {
	return impl.NewService(logger, metaStorage, imageStorage)
}
