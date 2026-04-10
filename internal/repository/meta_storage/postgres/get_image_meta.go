package postgres

import (
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// GetImageMeta retrieves the current object_key and status for a given
// image UUID. Used by the GET /image/:id endpoint to decide whether
// to serve the image or return 202 Accepted (pending).
func (m *MetaStorage) GetImageMeta(ctx context.Context, id string) (key string, status string, err error) {

	row, err := m.db.QueryRowWithRetry(ctx, retry.Strategy(m.config.QueryRetryStrategy), `

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
