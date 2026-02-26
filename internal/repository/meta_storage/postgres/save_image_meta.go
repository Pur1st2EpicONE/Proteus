package postgres

import (
	"Proteus/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) SaveImageMeta(ctx context.Context, image *models.Image) error {

	_, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `
		
	INSERT INTO images (uuid, object_key, status, updated_at)
	VALUES ($1, $2, $3, $4)`,

		image.ID, image.ObjectKey, image.Status, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	s.logger.Debug("postgres — meta saved", "image_id", image.ID, "layer", "repository.meta_storage.postgres")

	return nil

}
