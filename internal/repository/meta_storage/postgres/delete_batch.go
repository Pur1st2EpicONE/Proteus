package postgres

import (
	"Proteus/internal/models"
	"context"
	"fmt"
	"strings"
)

func (s *MetaStorage) DeleteBatch(ctx context.Context, ids []string) error {

	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))

	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = ids[i]
	}

	query := fmt.Sprintf(`

    DELETE FROM images 
    WHERE uuid IN (%s) AND status = $%d`,

		strings.Join(placeholders, ","), len(ids)+1)
	args = append(args, models.StatusDeleted)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch delete images from PG: %w", err)
	}

	return nil
}
