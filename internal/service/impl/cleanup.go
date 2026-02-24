package impl

import (
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *Service) Cleanup(ctx context.Context) error {

	deletedImages, err := s.metaStorage.GetDeleted(ctx)
	if err != nil {
		return fmt.Errorf("failed to get deleted images: %w", err)
	}

	if len(deletedImages) == 0 {
		s.logger.Debug("no deleted images to clean up")
		return nil
	}

	objectKeys := make([]string, len(deletedImages))
	ids := make([]string, len(deletedImages))

	for i, img := range deletedImages {
		objectKeys[i] = img.ObjectKey
		ids[i] = img.ID
	}

	err = retry.Do(func() error {
		return s.imageStorage.DeleteBatch(ctx, objectKeys)
	}, retry.Strategy{Attempts: 3, Delay: 5, Backoff: 2})
	if err != nil {
		s.logger.LogError("batch delete from MinIO failed", err)
		return err
	}

	err = retry.Do(func() error {
		return s.metaStorage.DeleteBatch(ctx, ids)
	}, retry.Strategy{Attempts: 3, Delay: 5, Backoff: 2})
	if err != nil {
		s.logger.LogError("batch delete from PG failed", err)
		return err
	}

	s.logger.Debug("cleaned up %d images", len(deletedImages))

	return nil
}
