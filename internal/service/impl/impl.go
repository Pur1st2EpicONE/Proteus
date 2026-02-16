package impl

import (
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
)

type Service struct {
	logger       logger.Logger
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func NewService(logger logger.Logger, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) *Service {
	return &Service{logger: logger, metaStorage: metaStorage, imageStorage: imageStorage}
}
