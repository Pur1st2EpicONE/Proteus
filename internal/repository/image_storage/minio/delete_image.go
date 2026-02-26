package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *ImageStorage) DeleteImage(ctx context.Context, objectKey string) error {

	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete object: %w", err)
	}

	s.logger.Debug("minio — object deleted", "object_key", objectKey, "layer", "repository.image_storage.minio")

	return nil

}
