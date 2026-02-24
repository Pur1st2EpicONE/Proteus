package postgres

import (
	"Proteus/internal/models"
	"context"
)

func (s *MetaStorage) GetDeleted(ctx context.Context) ([]models.Image, error) {

	rows, err := s.db.QueryContext(ctx, `

        SELECT uuid, object_key 
        FROM images 
        WHERE status = $1`,

		models.StatusDeleted)
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
