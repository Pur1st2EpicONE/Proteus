package kafka

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
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

func NewConsumer(logger logger.Logger, config config.Consumer, consumer *kafka.Consumer, imageStorage image_storage.ImageStorage) Consumer {
	return Consumer{logger: logger, config: config, consumer: consumer, imageStorage: imageStorage}
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
			c.process(ctx, task)
		}
	}
}

func (c Consumer) process(ctx context.Context, task models.ImageProcessTask) {

	imageBytes, err := c.imageStorage.DownloadImage(ctx, task.ObjectKey)
	if err != nil {
		fmt.Println("download failed:", err)
		return
	}

	srcImg, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		fmt.Println("decode failed:", err)
		return
	}

	results := make(map[string][]byte)

	for _, r := range task.Action {

		switch r {

		case "thumbnail":
			img := imaging.Thumbnail(srcImg, 150, 150, imaging.Lanczos)
			buf, _ := encode(img, format)
			results["thumbnail"] = buf

		case "medium":
			img := imaging.Resize(srcImg, 800, 0, imaging.Lanczos)
			buf, _ := encode(img, format)
			results["medium"] = buf

		case "watermarked":
			img := addWatermark(srcImg)
			buf, _ := encode(img, format)
			results["watermarked"] = buf
		}
	}

	for name, file := range results {

		objectKey := fmt.Sprintf("processed/%s/%s.%s", task.ID, name, format)

		err := c.imageStorage.UploadProcessed(ctx, objectKey, file, task.MimeType)
		if err != nil {
			fmt.Println("upload processed failed:", err)
			return
		}
	}

	fmt.Println("processing completed:", task.ID)
}

func encode(img image.Image, format string) ([]byte, error) {

	var buf bytes.Buffer

	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		return buf.Bytes(), err
	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	case "gif":
		err := gif.Encode(&buf, img, nil)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func addWatermark(src image.Image) image.Image {

	watermark := imaging.New(200, 50, color.NRGBA{0, 0, 0, 100})

	result := imaging.Overlay(
		src,
		watermark,
		image.Pt(src.Bounds().Dx()-210, src.Bounds().Dy()-60),
		0.6,
	)

	return result
}
