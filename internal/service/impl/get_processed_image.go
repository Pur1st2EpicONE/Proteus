package impl

import (
	"context"
	"path/filepath"
	"strings"
)

func (s *Service) DownloadImage(ctx context.Context, key string) ([]byte, string, error) {
	data, err := s.imageStorage.DownloadImage(ctx, key)
	if err != nil {
		return nil, "", err
	}

	contentType := "image/jpeg"
	ext := filepath.Ext(key)
	switch strings.ToLower(ext) {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	return data, contentType, nil
}
