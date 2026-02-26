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

type Consumer struct {
	ctx          context.Context
	config       config.Consumer
	logger       logger.Logger
	consumer     *kafka.Consumer
	processFunc  func(ctx context.Context, task models.ImageProcessTask) error
	imageStorage image_storage.ImageStorage
}

func NewConsumer(ctx context.Context, l logger.Logger, cfg config.Consumer, cons *kafka.Consumer,
	processFunc func(ctx context.Context, task models.ImageProcessTask) error, iStorage image_storage.ImageStorage) *Consumer {
	return &Consumer{ctx: ctx, logger: l, config: cfg, consumer: cons, processFunc: processFunc, imageStorage: iStorage}
}

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

func (c *Consumer) processImage(image models.ImageProcessTask) error {
	return c.processFunc(c.ctx, image)
}

func (c *Consumer) Close() {
	if err := c.consumer.Close(); err != nil {
		c.logger.LogError("consumer — failed to close reader", err, "layer", "broker.kafka")
	} else {
		c.logger.LogInfo("consumer — reader closed", "layer", "broker.kafka")
	}
}
