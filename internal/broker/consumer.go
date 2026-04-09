package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
)

// processFunc is the signature of the callback that processes a single
// ImageProcessTask. It is injected into the consumer by the app layer.
type processFunc func(ctx context.Context, task models.ImageProcessTask) error

// Consumer defines the contract for consuming and processing messages
// from Kafka with support for graceful shutdown.
type Consumer interface {
	Run()   // Run starts the consumption loop and blocks until the context is cancelled.
	Close() // Close shuts down the consumer and releases all underlying resources.
}

// NewConsumer creates a new Kafka consumer implementation with all required
// dependencies (context, logger, config, underlying consumer, processing
// function and image storage).
func NewConsumer(ctx context.Context, l logger.Logger, cfg config.Consumer, cons *wbf.Consumer, pf processFunc, is image_storage.ImageStorage) Consumer {
	return kafka.NewConsumer(ctx, l, cfg, cons, pf, is)
}
