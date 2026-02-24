package postgres

import (
	"Proteus/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) MarkAsDeleted(ctx context.Context, id string) error {

	tx, err := s.db.BeginTxWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, nil)

	if err != nil {
		return fmt.Errorf("failed to begin transaction for image deletion: %w", err)
	}
	defer tx.Rollback()

	var existingImage models.Image
	err = tx.QueryRowContext(ctx, `
        
	SELECT uuid, status 
    FROM images 
    WHERE uuid = $1 AND (status = $2 OR status = $3)
    FOR UPDATE`,

		id, models.StatusPending, models.StatusReady).Scan(&existingImage.ID, &existingImage.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return fmt.Errorf("failed to fetch image for update: %w", err)
	}

	_, err = tx.ExecContext(ctx, `

    UPDATE images 
    SET status = $1, updated_at = NOW() 
    WHERE uuid = $2`,

		models.StatusDeleted, id)
	if err != nil {
		return fmt.Errorf("failed to update image status to deleted: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit image deletion transaction: %w", err)
	}

	return nil

}
