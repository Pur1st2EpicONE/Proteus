package postgres

import (
	"Proteus/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) MarkAsReady(ctx context.Context, objectKey string, uuid string) error {

	result, err := s.db.ExecWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    UPDATE images
    SET object_key = $1, status = $2, updated_at = $3 
    WHERE uuid = $4`,

		objectKey, models.StatusReady, time.Now().UTC(), uuid)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with uuid %s not found", uuid)
	}

	s.logger.Debug("postgres — image marked as ready", "image_id", uuid, "layer", "repository.meta_storage.postgres")

	return nil

}
