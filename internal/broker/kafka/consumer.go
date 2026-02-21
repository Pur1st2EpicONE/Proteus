package kafka

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"context"
	"encoding/json"
	"fmt"

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
			var task models.ImageProcessTask
			if err := json.Unmarshal(message.Value, &task); err != nil {
				fmt.Println("cannot unmarshal:", err)
				continue
			}
			if err := c.processImage(ctx, task); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (c *Consumer) processImage(ctx context.Context, image models.ImageProcessTask) error {
	return c.processFunc(ctx, image)
}
