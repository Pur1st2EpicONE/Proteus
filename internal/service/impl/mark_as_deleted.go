package impl

import (
	"Proteus/internal/errs"
	"context"
	"database/sql"
	"errors"
)

// MarkAsDeleted delegates the soft-delete operation to the meta
// storage and maps sql.ErrNoRows to ErrImageNotFound.
func (s *Service) MarkAsDeleted(ctx context.Context, id string) error {
	if err := s.metaStorage.MarkAsDeleted(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrImageNotFound
		}
		s.logger.LogError("service — failed to mark image as deleted in meta storage", err, "image_id", id, "layer", "service.impl")
		return err
	}
	s.logger.Debug("service — image marked as deleted in meta storage", "image_id", id, "layer", "service.impl")
	return nil
}
