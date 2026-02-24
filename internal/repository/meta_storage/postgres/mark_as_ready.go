package postgres

import (
	"context"
	"fmt"
)

func (s *MetaStorage) MarkAsReady(ctx context.Context, objectKey string, uuid string) error {

	result, err := s.db.ExecContext(ctx, `

    UPDATE images
    SET object_key = $1, status = 'ready', updated_at = NOW() 
    WHERE uuid = $2`,

		objectKey, uuid)
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
