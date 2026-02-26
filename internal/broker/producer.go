package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/logger"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
)

type Producer interface {
	Send(ctx context.Context, key []byte, value []byte) error
	Close()
}

func NewProducer(logger logger.Logger, producer *wbf.Producer) Producer {
	return kafka.NewProducer(logger, producer)
}
