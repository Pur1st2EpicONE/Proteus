package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type Consumer interface {
	Run(ctx context.Context, strategy retry.Strategy)
}

func NewConsumer(logger logger.Logger, config config.Consumer, consumer *wbf.Consumer, imageStorage image_storage.ImageStorage) Consumer {
	return kafka.NewConsumer(logger, config, consumer, imageStorage)
}
