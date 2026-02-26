package impl

import (
	"context"
	"path/filepath"
	"strings"
)

func (s *Service) DownloadImage(ctx context.Context, key string) ([]byte, string, error) {

	data, err := s.imageStorage.DownloadImage(ctx, key)
	if err != nil {
		s.logger.LogError("service — failed to download processed image from image storage", err, "object_key", key, "layer", "service.impl")
		return nil, "", err
	}

	contentType := "image/jpeg"

	switch strings.ToLower(filepath.Ext(key)) {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	return data, contentType, nil

}
