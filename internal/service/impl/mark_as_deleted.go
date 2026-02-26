package impl

import (
	"Proteus/internal/errs"
	"context"
	"database/sql"
	"errors"
)

func (s *Service) MarkAsDeleted(ctx context.Context, id string) error {
	if err := s.metaStorage.MarkAsDeleted(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrImageNotFound
		}
		s.logger.LogError("service — failed to mark image as deleted in meta storage", err, "image_id", id, "layer", "service")
		return err
	}
	return nil
}
