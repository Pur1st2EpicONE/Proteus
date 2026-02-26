package postgres

import (
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) GetImageMeta(ctx context.Context, id string) (key string, status string, err error) {

	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    SELECT COALESCE(object_key, ''), status
    FROM images
    WHERE uuid = $1`,

		id)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute query: %w", err)
	}

	if err := row.Scan(&key, &status); err != nil {
		return "", "", fmt.Errorf("failed to scan row: %w", err)
	}

	return key, status, nil

}
