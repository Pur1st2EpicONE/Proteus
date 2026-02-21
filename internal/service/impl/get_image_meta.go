package impl

import (
	"context"
)

func (s *Service) GetImageMeta(ctx context.Context, id string) (string, string, error) {
	return s.metaStorage.GetImageMeta(ctx, id)
}
