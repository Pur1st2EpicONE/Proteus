// Package broker defines high-level interfaces and factory functions for
// Kafka producers and consumers used by the Proteus service to handle
// asynchronous image processing tasks.
package broker

import (
	"Proteus/internal/broker/kafka"
	"Proteus/internal/logger"
	"context"

	wbf "github.com/wb-go/wbf/kafka"
)

// Producer defines the contract for sending messages to the Kafka broker.
type Producer interface {
	Send(ctx context.Context, key []byte, value []byte) error // Send publishes a key-value message to the configured topic.
	Close()                                                   // Close shuts down the underlying producer and flushes any buffered messages.
}

// NewProducer returns a concrete Kafka producer implementation wrapped with
// logging and error context.
func NewProducer(logger logger.Logger, producer *wbf.Producer) Producer {
	return kafka.NewProducer(logger, producer)
}
