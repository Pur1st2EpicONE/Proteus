package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"context"

	km "github.com/segmentio/kafka-go"
	wbf "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type Consumer interface {
	Run(ctx context.Context, out chan km.Message, strategy retry.Strategy)
}

func NewConsumer(logger logger.Logger, config config.Consumer, consumer *wbf.Consumer) Consumer {
	return kafka.NewConsumer(logger, config, consumer)
}
