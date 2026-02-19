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
	imageStorage image_storage.ImageStorage
}

func NewConsumer(logger logger.Logger, config config.Consumer, consumer *kafka.Consumer) Consumer {
	return Consumer{logger: logger, config: config, consumer: consumer}
}

func (c Consumer) Run(ctx context.Context, out chan km.Message, strategy retry.Strategy) {
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
			fmt.Println(task)
		}
	}
}
