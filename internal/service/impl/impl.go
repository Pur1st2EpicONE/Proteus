package impl

import (
	"Proteus/internal/broker"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
)

type Service struct {
	logger       logger.Logger
	config       config.Service
	producer     broker.Producer
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func NewService(l logger.Logger, cfg config.Service, prod broker.Producer, ms meta_storage.MetaStorage, is image_storage.ImageStorage) *Service {
	return &Service{logger: l, config: cfg, producer: prod, metaStorage: ms, imageStorage: is}
}
