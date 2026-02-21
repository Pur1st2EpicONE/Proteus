package impl

import (
	"Proteus/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/wb-go/wbf/helpers"
	"golang.org/x/sync/errgroup"
)

func (s *Service) UploadImage(ctx context.Context, image *models.Image) (string, error) {

	if err := validate(image); err != nil {
		return "", err
	}

	initialize(image)
	var g errgroup.Group

	g.Go(func() error { return s.metaStorage.SaveImageMeta(ctx, image) })
	g.Go(func() error { return s.imageStorage.UploadImage(ctx, image) })

	if err := g.Wait(); err != nil {
		return "", err
	}

	payload, err := json.Marshal(models.ImageProcessTask{
		ID:           image.ID,
		ObjectKey:    image.ObjectKey,
		OriginalName: image.FileHeader.Filename,
		MimeType:     image.FileHeader.Header.Get("Content-Type"),
		FileSize:     image.Size,
		Action:       image.Request.Action,
		Watermark:    image.Request.Watermark,
		Height:       image.Request.Height,
		Width:        image.Request.Width,
		Quality:      image.Request.Quality,
	})
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
	image.ObjectKey = prefix() + image.ID + filepath.Ext(image.FileHeader.Filename)
}

func prefix() string {
	now := time.Now().UTC()
	return fmt.Sprintf("unprocessed/%d/%02d/%02d/", now.Year(), int(now.Month()), now.Day())
}
