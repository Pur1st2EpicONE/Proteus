package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *MetaStorage) GetImageMeta(ctx context.Context, id string) (processedKey string, status string, err error) {

	query := `
        
	SELECT COALESCE(processed_key, ''), status
    FROM images
    WHERE uuid = $1`

	err = s.db.QueryRowContext(ctx, query, id).Scan(&processedKey, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", err
		}
		return "", "", fmt.Errorf("get image meta failed: %w", err)
	}

	return processedKey, status, nil
}
