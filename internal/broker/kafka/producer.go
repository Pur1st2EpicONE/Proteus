package kafka

import (
	"Proteus/internal/logger"
	"context"
	"fmt"

	kafka "github.com/wb-go/wbf/kafka"
)

type Producer struct {
	logger   logger.Logger
	producer *kafka.Producer
}

func NewProducer(logger logger.Logger, producer *kafka.Producer) *Producer {
	return &Producer{logger: logger, producer: producer}
}

func (p *Producer) Send(ctx context.Context, key []byte, value []byte) error {
	if err := p.producer.Send(ctx, key, value); err != nil {
		return fmt.Errorf("kafka producer send failed for key %s: %w", string(key), err)
	}
	return nil
}

func (p *Producer) Close() {
	if err := p.producer.Close(); err != nil {
		p.logger.LogError("producer — failed to close writer", err, "layer", "broker.kafka")
	} else {
		p.logger.LogInfo("producer — writer closed", "layer", "broker.kafka")
	}
}
