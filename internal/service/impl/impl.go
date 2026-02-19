package impl

import (
	"Proteus/internal/broker"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/wb-go/wbf/helpers"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	logger       logger.Logger
	producer     broker.Producer
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func NewService(logger logger.Logger, producer broker.Producer, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) *Service {
	return &Service{logger: logger, producer: producer, metaStorage: metaStorage, imageStorage: imageStorage}
}

func (s *Service) UploadImage(ctx context.Context, image *models.Image) (string, error) {

	if err := validateImage(image); err != nil {
		return "", err
	}

	initialize(image)

	var g errgroup.Group

	g.Go(func() error {
		return s.metaStorage.SaveImageMeta(ctx, image)
	})

	g.Go(func() error {
		return s.imageStorage.UploadImage(ctx, image)
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	task := models.ImageProcessTask{
		ID:           image.ID,
		ObjectKey:    image.ObjectKey,
		OriginalName: image.FileHeader.Filename,
		MimeType:     image.FileHeader.Header.Get("Content-Type"),
		FileSize:     image.Size,
		Requested:    []string{"thumbnail", "medium", "watermarked"},
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return "", err
	}

	if err := s.producer.Send(ctx, []byte(image.ID), payload); err != nil {
		return "", err
	}

	s.logger.LogInfo("Image uploaded, metadata saved and processing queued", "id", image.ID)

	return image.ID, nil

}

func initialize(image *models.Image) {
	image.ID = helpers.CreateUUID()
	image.Size = int64(len(image.File))
	image.ObjectKey = datePrefix() + image.ID + filepath.Ext(image.FileHeader.Filename)
}

func datePrefix() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d/%02d/%02d/", now.Year(), int(now.Month()), now.Day())
}
