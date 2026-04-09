// Package kafka contains the concrete implementations of the broker.Producer
// and broker.Consumer interfaces using the wb-go/wbf/kafka library.
package kafka

import (
	"Proteus/internal/logger"
	"context"
	"fmt"

	kafka "github.com/wb-go/wbf/kafka"
)

// Producer implements the broker.Producer interface by delegating to the
// wb-go/wbf Kafka producer and adding contextual error wrapping.
type Producer struct {
	logger   logger.Logger   // logger is used to log producer lifecycle events (close, errors).
	producer *kafka.Producer // producer is the underlying wb-go/wbf Kafka producer instance.
}

// NewProducer constructs a *Producer with the logger and underlying producer.
func NewProducer(logger logger.Logger, producer *kafka.Producer) *Producer {
	return &Producer{logger: logger, producer: producer}
}

// Send sends a key-value message through the underlying producer.
// Any error is wrapped with the message key for easier debugging.
func (p *Producer) Send(ctx context.Context, key []byte, value []byte) error {
	if err := p.producer.Send(ctx, key, value); err != nil {
		return fmt.Errorf("kafka producer send failed for key %s: %w", string(key), err)
	}
	return nil
}

// Close closes the underlying producer and logs success or failure.
func (p *Producer) Close() {
	if err := p.producer.Close(); err != nil {
		p.logger.LogError("producer — failed to close writer", err, "layer", "broker.kafka")
	} else {
		p.logger.LogInfo("producer — writer closed", "layer", "broker.kafka")
	}
}
