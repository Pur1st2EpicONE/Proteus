package impl

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) Cleaner(ctx context.Context) {

	if !s.config.Cleaner {
		return
	}

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	s.logger.LogInfo("service — cleaner started", "layer", "service.impl")

	for {
		select {

		case <-ctx.Done():
			s.logger.LogInfo("service — cleaner stopped", "layer", "service.impl")
			return
		case <-ticker.C:
			s.logger.Debug("service — cleanup cycle started", "layer", "service.impl")
			if err := s.clean(ctx); err != nil {
				s.logger.LogError("service — cleanup failed", err, "layer", "service.impl")
			} else {
				s.logger.Debug("service — cleanup cycle completed", "layer", "service.impl")
			}

		}
	}

}

func (s *Service) clean(ctx context.Context) error {

	deletedImages, err := s.metaStorage.GetDeleted(ctx)
	if err != nil {
		return fmt.Errorf("failed to get deleted images: %w", err)
	}

	if len(deletedImages) == 0 {
		return nil
	}

	objectKeys := make([]string, len(deletedImages))
	ids := make([]string, len(deletedImages))

	for i, img := range deletedImages {
		objectKeys[i] = img.ObjectKey
		ids[i] = img.ID
	}

	err = s.imageStorage.DeleteBatch(ctx, objectKeys)
	if err != nil {
		return fmt.Errorf("batch delete from image storage was unsuccessful: %w", err)
	}

	err = s.metaStorage.DeleteBatch(ctx, ids)
	if err != nil {
		return fmt.Errorf("batch delete from meta storage was unsuccessful: %w", err)
	}

	return nil

}
