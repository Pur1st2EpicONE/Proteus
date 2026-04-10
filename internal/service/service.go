// Package service defines the high-level Service interface used by
// handlers, the Kafka consumer and the application root. It also
// provides the factory that returns the concrete implementation.
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

// Service orchestrates all business logic for image upload,
// asynchronous processing, retrieval, soft-delete and cleanup.
type Service interface {
	UploadImage(ctx context.Context, image *models.Image) (string, error)  // UploadImage validates the request, stores the image atomically in both storages and enqueues a processing task via Kafka.
	ProcessImage(ctx context.Context, task models.ImageProcessTask) error  // ProcessImage is the entry point called by the Kafka consumer; it downloads the original image, applies the requested transformations and saves the result.
	GetImageMeta(ctx context.Context, id string) (string, string, error)   // GetImageMeta returns the current object_key and status for the given image UUID (used by the GET /image/:id endpoint).
	DownloadImage(ctx context.Context, key string) ([]byte, string, error) // DownloadImage retrieves the processed image bytes and its correct Content-Type (used by the GET /image/:id endpoint when the image is ready).
	MarkAsDeleted(ctx context.Context, id string) error                    // MarkAsDeleted performs a soft-delete by updating the status in meta storage (used by the DELETE /image/:id endpoint).
	Cleaner(ctx context.Context)                                           // Cleaner runs the background goroutine that periodically deletes images marked as deleted or stuck in pending state.
}

// NewService returns the concrete implementation of the Service interface.
func NewService(l logger.Logger, cfg config.Service, prod broker.Producer, ms meta_storage.MetaStorage, is image_storage.ImageStorage) Service {
	return impl.NewService(l, cfg, prod, ms, is)
}
