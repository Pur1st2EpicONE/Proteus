package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/wb-go/wbf/retry"
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
    WHERE uuid IN (%s)`,

		strings.Join(placeholders, ","))

	_, err := s.db.ExecWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	s.logger.Debug("postgres — meta batch deleted", "layer", "repository.meta_storage.postgres")

	return nil

}
