package kafka

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"context"

	kafka "github.com/wb-go/wbf/kafka"
)

type Producer struct {
	config   config.Producer
	logger   logger.Logger
	producer *kafka.Producer
}

func NewProducer(logger logger.Logger, config config.Producer, producer *kafka.Producer) Producer {
	return Producer{logger: logger, config: config, producer: producer}
}

func (p Producer) Send(ctx context.Context, key []byte, value []byte) error {
	return p.producer.Send(ctx, key, value)
}
