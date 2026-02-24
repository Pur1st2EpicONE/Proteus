package postgres

import (
	"context"
	"fmt"
)

func (s *MetaStorage) GetImageMeta(ctx context.Context, id string) (key string, status string, err error) {

	err = s.db.QueryRowContext(ctx, `
        
	SELECT COALESCE(object_key, ''), status
    FROM images
    WHERE uuid = $1`,

		id).Scan(&key, &status)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch image meta: %w", err)
	}

	return key, status, nil

}
