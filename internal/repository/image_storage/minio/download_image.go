package minio

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *ImageStorage) DownloadImage(ctx context.Context, objectKey string) ([]byte, error) {

	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get object: %w", err)
	}
	defer object.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(object); err != nil {
		return nil, fmt.Errorf("unable to read object data: %w", err)
	}

	s.logger.Debug("minio — object downloaded", "object_key", objectKey, "layer", "repository.image_storage.minio")

	return buf.Bytes(), nil

}
