package impl

import (
	"Proteus/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/wb-go/wbf/helpers"
	"golang.org/x/sync/errgroup"
)

const rbTimeout = 10 * time.Second

func (s *Service) UploadImage(ctx context.Context, image *models.Image) (string, error) {

	if err := validate(image); err != nil {
		return "", err
	}

	initialize(image)
	var g errgroup.Group

	g.Go(func() error { return s.metaStorage.SaveImageMeta(ctx, image) })
	g.Go(func() error { return s.imageStorage.UploadImage(ctx, image) })

	if err := g.Wait(); err != nil {
		go s.rollback(image, err)
		return "", err
	}

	payload, err := json.Marshal(models.ImageProcessTask{
		ID:           image.ID,
		ObjectKey:    image.ObjectKey,
		OriginalName: image.FileHeader.Filename,
		ContentType:  image.ContentType,
		FileSize:     image.Size,
		Action:       image.Request.Action,
		Watermark:    image.Request.Watermark,
		Height:       image.Request.Height,
		Width:        image.Request.Width,
		Quality:      image.Request.Quality,
	})
	if err != nil {
		s.logger.LogError("service — failed to marshal image", err, "image_id", image.ID, "layer", "service.impl")
		return "", err
	}

	if err := s.producer.Send(ctx, []byte(image.ID), payload); err != nil {
		s.logger.LogError("service — failed to enqueue image processing task", err, "image_id", image.ID, "layer", "service.impl")
		s.iStorageRollback(image)
		return "", err
	}

	s.logger.Debug("service — image uploaded, metadata saved and processing queued", "id", image.ID, "layer", "service.impl")

	return image.ID, nil

}

func initialize(image *models.Image) {
	image.ID = helpers.CreateUUID()
	image.Size = int64(len(image.File))
	image.Status = models.StatusPending
	image.ContentType = image.FileHeader.Header.Get("Content-Type")
	image.ObjectKey = prefix() + image.ID + filepath.Ext(image.FileHeader.Filename)
}

func prefix() string {
	now := time.Now().UTC()
	return fmt.Sprintf("unprocessed/%d/%02d/%02d/", now.Year(), int(now.Month()), now.Day())
}

func (s *Service) rollback(image *models.Image, err error) {

	if strings.Contains(err.Error(), "meta") {
		s.logger.LogError("service — failed to save image to meta storage", err, "image_id", image.ID, "layer", "service.impl")
		s.iStorageRollback(image)
		return
	}

	s.logger.LogError("service — failed to save image to image storage", err, "image_id", image.ID, "layer", "service.impl")

}

func (s *Service) iStorageRollback(image *models.Image) {

	ctx, cancel := context.WithTimeout(context.Background(), rbTimeout)
	defer cancel()

	if err := s.imageStorage.DeleteImage(ctx, image.ObjectKey); err != nil {
		s.logger.LogError("service — failed to delete orphan image from image storage", err, "image_id", image.ID, "layer", "service.impl")
	}
	s.logger.Debug("service — orphan image deleted from image storage", "image_id", image.ID, "layer", "service.impl")

}
