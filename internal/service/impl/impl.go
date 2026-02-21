package impl

import (
	"Proteus/internal/broker"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
)

type Service struct {
	logger       logger.Logger
	producer     broker.Producer
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func NewService(logger logger.Logger, producer broker.Producer, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) *Service {
	return &Service{logger: logger, producer: producer, metaStorage: metaStorage, imageStorage: imageStorage}
}
