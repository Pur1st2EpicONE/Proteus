package impl

import (
	"Proteus/internal/errs"
	"context"
	"database/sql"
	"errors"
)

func (s *Service) GetImageMeta(ctx context.Context, id string) (string, string, error) {
	key, status, err := s.metaStorage.GetImageMeta(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errs.ErrImageNotFound
		}
		s.logger.LogError("service — failed to get image info from meta storage", err, "image_id", id, "layer", "service")
		return "", "", err
	}
	return key, status, nil
}
