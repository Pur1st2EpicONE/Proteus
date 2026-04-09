package kafka

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"context"
	"encoding/json"

	km "github.com/segmentio/kafka-go"
	kafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

// Consumer implements the broker.Consumer interface.
// It receives messages from Kafka, unmarshals them into ImageProcessTask,
// delegates processing to the service and commits offsets on success.
type Consumer struct {
	ctx          context.Context                                               // ctx is the root context used to stop consumption on shutdown.
	config       config.Consumer                                               // config holds Kafka consumer settings (brokers, topic, group ID, retry strategy).
	logger       logger.Logger                                                 // logger is used for structured logging of all consumer events.
	consumer     *kafka.Consumer                                               // consumer is the underlying wb-go/wbf Kafka consumer instance.
	processFunc  func(ctx context.Context, task models.ImageProcessTask) error // processFunc is the injected callback that processes each task.
	imageStorage image_storage.ImageStorage                                    // imageStorage is used for any image-related operations during processing.
}

// NewConsumer constructs a *Consumer with the given dependencies.
func NewConsumer(ctx context.Context, l logger.Logger, cfg config.Consumer, cons *kafka.Consumer,
	processFunc func(ctx context.Context, task models.ImageProcessTask) error, iStorage image_storage.ImageStorage) *Consumer {
	return &Consumer{ctx: ctx, logger: l, config: cfg, consumer: cons, processFunc: processFunc, imageStorage: iStorage}
}

// Run starts the consumption loop. It creates a message channel, launches
// the underlying consumer with the configured retry strategy and dispatches
// every received message to handleMessage until the context is cancelled.
func (c Consumer) Run() {

	kafka := make(chan km.Message)
	c.consumer.StartConsuming(c.ctx, kafka, retry.Strategy(c.config.FetchRetryStrategy))

	for {
		select {
		case <-c.ctx.Done():
			return
		case message := <-kafka:
			c.handleMessage(message)
		}
	}

}

// handleMessage unmarshals the raw Kafka message into an ImageProcessTask,
// processes it and commits the offset if processing succeeds. Any error
// is logged and the message is not reprocessed.
func (c *Consumer) handleMessage(message km.Message) {

	c.logger.Debug("consumer — received new message", "layer", "broker.kafka")

	var image models.ImageProcessTask
	if err := json.Unmarshal(message.Value, &image); err != nil {
		c.logger.LogError("consumer — failed to unmarshal message", err, "layer", "broker.kafka")
		return
	}

	if err := c.processImage(image); err != nil {
		c.logger.LogError("consumer — failed to process image", err, "layer", "broker.kafka")
		return
	}

	if err := c.consumer.Commit(c.ctx, message); err != nil {
		c.logger.LogError("consumer — failed to commit message", err, "layer", "broker.kafka")
		return
	}

	c.logger.Debug("consumer — message processed", "layer", "broker.kafka")

}

// processImage calls the injected processing function (usually a closure
// around service.ProcessImage).
func (c *Consumer) processImage(image models.ImageProcessTask) error {
	return c.processFunc(c.ctx, image)
}

// Close shuts down the underlying Kafka consumer and logs the result.
func (c *Consumer) Close() {
	if err := c.consumer.Close(); err != nil {
		c.logger.LogError("consumer — failed to close reader", err, "layer", "broker.kafka")
	} else {
		c.logger.LogInfo("consumer — reader closed", "layer", "broker.kafka")
	}
}
