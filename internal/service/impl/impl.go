// Package impl contains the concrete implementation of the
// service.Service interface. It handles image validation, upload
// with rollback logic, asynchronous processing using the imaging
// library, Kafka messaging and periodic cleanup of old images.
package impl

import (
	"Proteus/internal/broker"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
)

type Service struct {
	logger       logger.Logger              // logger is used for all structured logging from the service layer.
	config       config.Service             // config holds service-specific settings (cleaner toggle and cleanup interval).
	producer     broker.Producer            // producer is used to enqueue image processing tasks to Kafka.
	metaStorage  meta_storage.MetaStorage   // metaStorage persists image metadata and status transitions.
	imageStorage image_storage.ImageStorage // imageStorage handles raw and processed image files in MinIO.
}

// NewService constructs a new Service implementation wired with all required dependencies.
func NewService(l logger.Logger, cfg config.Service, prod broker.Producer, ms meta_storage.MetaStorage, is image_storage.ImageStorage) *Service {
	return &Service{logger: l, config: cfg, producer: prod, metaStorage: ms, imageStorage: is}
}
