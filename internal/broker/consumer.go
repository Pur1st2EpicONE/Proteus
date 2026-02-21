package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type processFunc func(ctx context.Context, task models.ImageProcessTask) error

type Consumer interface {
	Run(ctx context.Context, strategy retry.Strategy)
}

func NewConsumer(l logger.Logger, cfg config.Consumer, cons *wbf.Consumer, processFunc processFunc, iStorage image_storage.ImageStorage) Consumer {
	return kafka.NewConsumer(l, cfg, cons, processFunc, iStorage)
}
