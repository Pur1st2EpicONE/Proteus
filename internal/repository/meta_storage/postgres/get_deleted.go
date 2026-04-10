package postgres

import (
	"Proteus/internal/models"
	"context"
	"time"

	"github.com/wb-go/wbf/retry"
)

// GetDeleted returns a list of images that should be cleaned up:
// either explicitly marked as deleted or pending longer than the
// configured PendingTimeout. It is used by the service.Cleaner goroutine.
func (m *MetaStorage) GetDeleted(ctx context.Context) ([]models.Image, error) {

	rows, err := m.db.QueryWithRetry(ctx, retry.Strategy(m.config.QueryRetryStrategy), `

    SELECT uuid, object_key 
    FROM images 
    WHERE status = $1
	OR (status = $2 AND updated_at < $3)`,

		models.StatusDeleted, models.StatusPending, time.Now().UTC().Add(-m.config.PendingTimeout))
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	var images []models.Image

	for rows.Next() {
		var img models.Image
		if err := rows.Scan(&img.ID, &img.ObjectKey); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, rows.Err()

}
