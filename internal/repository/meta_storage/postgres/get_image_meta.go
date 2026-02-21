package postgres

import (
	"Proteus/internal/models"
	"context"
	"fmt"
)

func (s *MetaStorage) GetImageMeta(ctx context.Context, id string) (key string, status string, err error) {

	err = s.db.QueryRowContext(ctx, `
        
	SELECT COALESCE(processed_key, ''), status
    FROM images
    WHERE uuid = $1 AND status = $2 OR status = $3`,

		id, models.StatusPending, models.StatusReady).Scan(&key, &status)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch image meta: %w", err)
	}

	return key, status, nil

}
