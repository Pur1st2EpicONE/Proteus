package postgres

import (
	"context"
	"fmt"
)

func (s *MetaStorage) MarkAsReady(ctx context.Context, uuid string, processedKey string) error {

	query := `

    UPDATE images
    SET processed_key = $1, status = 'ready'
    WHERE uuid = $2`

	result, err := s.db.ExecContext(ctx, query, processedKey, uuid)
	if err != nil {
		return fmt.Errorf("failed to update image meta to ready: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with uuid %s not found", uuid)
	}

	return nil
}
