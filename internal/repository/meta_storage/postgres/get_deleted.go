package postgres

import (
	"Proteus/internal/models"
	"context"
	"time"

	"github.com/wb-go/wbf/retry"
)

func (s *MetaStorage) GetDeleted(ctx context.Context) ([]models.Image, error) {

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    SELECT uuid, object_key 
    FROM images 
    WHERE status = $1
	OR (status = $2 AND updated_at < $3)`,

		models.StatusDeleted, models.StatusPending, time.Now().UTC().Add(-s.config.PendingTimeout))
	if err != nil {
		return nil, err
	}

	defer rows.Close()
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
