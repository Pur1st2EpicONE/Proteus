package impl

import (
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/wb-go/wbf/helpers"
)

type Service struct {
	logger       logger.Logger
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func NewService(logger logger.Logger, metaStorage meta_storage.MetaStorage, imageStorage image_storage.ImageStorage) *Service {
	return &Service{logger: logger, metaStorage: metaStorage, imageStorage: imageStorage}
}

func (s *Service) UploadImage(ctx context.Context, image *models.Image) (string, error) {

	if err := validateImage(image); err != nil {
		return "", err
	}

	initialize(image)

	if err := s.imageStorage.UploadImage(ctx, image); err != nil {
		return "", err
	}

	if err := s.metaStorage.SaveImageMeta(ctx, image); err != nil {
		return "", err
	}

	s.logger.LogInfo("Image uploaded and metadata saved", "id", image.ID)

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
