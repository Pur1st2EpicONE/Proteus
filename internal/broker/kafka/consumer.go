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
	config       config.Consumer
	logger       logger.Logger
	consumer     *kafka.Consumer
	processFunc  func(ctx context.Context, task models.ImageProcessTask) error
	imageStorage image_storage.ImageStorage
}

func NewConsumer(l logger.Logger, cfg config.Consumer, cons *kafka.Consumer,
	processFunc func(ctx context.Context, task models.ImageProcessTask) error, iStorage image_storage.ImageStorage) *Consumer {
	return &Consumer{logger: l, config: cfg, consumer: cons, processFunc: processFunc, imageStorage: iStorage}
}

func (c Consumer) Run(ctx context.Context, strategy retry.Strategy) {

	out := make(chan km.Message)
	c.consumer.StartConsuming(ctx, out, strategy)

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-out:
			c.handleMessage(ctx, message)
		}
	}

}

func (c *Consumer) handleMessage(ctx context.Context, message km.Message) {

	c.logger.Debug("consumer — received new message", "layer", "broker.kafka")

	var task models.ImageProcessTask
	if err := json.Unmarshal(message.Value, &task); err != nil {
		c.logger.LogError("consumer — failed to unmarshal message", err, "layer", "broker.kafka")
		return
	}

	if err := c.processImage(ctx, task); err != nil {
		c.logger.LogError("consumer — failed to process image", err, "layer", "broker.kafka")
		return
	}

	if err := c.consumer.Commit(ctx, message); err != nil {
		c.logger.LogError("consumer — failed to commit message", err, "layer", "broker.kafka")
		return
	}

	c.logger.Debug("consumer — message processed", "layer", "broker.kafka")

}

func (c *Consumer) processImage(ctx context.Context, image models.ImageProcessTask) error {
	return c.processFunc(ctx, image)
}
