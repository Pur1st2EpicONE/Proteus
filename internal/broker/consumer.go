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

type processFunc func(ctx context.Context, task models.ImageProcessTask) error

type Consumer interface {
	Run()
	Close()
}

func NewConsumer(ctx context.Context, l logger.Logger, cfg config.Consumer, cons *wbf.Consumer, pf processFunc, is image_storage.ImageStorage) Consumer {
	return kafka.NewConsumer(ctx, l, cfg, cons, pf, is)
}
