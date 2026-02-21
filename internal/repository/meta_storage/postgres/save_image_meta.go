package postgres

import (
	"Proteus/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) SaveImageMeta(ctx context.Context, image *models.Image) error {

	_, err := s.db.QueryRowWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, `
		
		INSERT INTO images (uuid, object_key, status, created_at)
		VALUES ($1, $2, $3, $4)`,

		image.ID, image.ObjectKey, models.StatusPending, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to save image metadata: %w", err)
	}

	return nil

}
