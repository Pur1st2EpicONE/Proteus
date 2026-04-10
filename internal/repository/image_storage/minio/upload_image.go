package minio

import (
	"Proteus/internal/models"
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

// UploadImage stores the image bytes in MinIO using the object's key
// and the original content type.
func (s *ImageStorage) UploadImage(ctx context.Context, image *models.Image) error {

	_, err := s.client.PutObject(ctx, s.bucketName, image.ObjectKey, bytes.NewReader(image.File), image.Size,
		minio.PutObjectOptions{ContentType: image.ContentType})
	if err != nil {
		return fmt.Errorf("unable to put object: %w", err)
	}

	s.logger.Debug("minio — object uploaded", "object_key", image.ObjectKey, "layer", "repository.image_storage.minio")

	return nil

}
