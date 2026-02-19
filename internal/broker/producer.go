package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
)

type Producer interface {
	Send(ctx context.Context, key []byte, value []byte) error
}

func NewProducer(logger logger.Logger, config config.Producer, producer *wbf.Producer) Producer {
	return kafka.NewProducer(logger, config, producer)
}
