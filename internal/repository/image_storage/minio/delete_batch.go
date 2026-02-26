package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *ImageStorage) DeleteBatch(ctx context.Context, objectKeys []string) error {

	if len(objectKeys) == 0 {
		return nil
	}

	objectsCh := make(chan minio.ObjectInfo, len(objectKeys)*2)

	for _, key := range objectKeys {
		objectsCh <- minio.ObjectInfo{Key: key}
		objectsCh <- minio.ObjectInfo{Key: "un" + key}
	}
	close(objectsCh)

	errorCh := s.client.RemoveObjects(ctx, s.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	var errs []error
	for e := range errorCh {
		if e.Err != nil {
			errs = append(errs, e.Err)
			s.logger.Debug("minio — unable to delete object", e.Err, "object_key", e.ObjectName, "layer", "repository.image_storage.minio")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("batch delete had %d errors, last error: %w", len(errs), errs[len(errs)-1])
	}

	s.logger.Debug("minio — image batch deleted", "layer", "repository.image_storage.minio")

	return nil

}
