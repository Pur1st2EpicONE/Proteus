package postgres

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type MetaStorage struct {
	db     *dbpg.DB
	logger logger.Logger
	config config.MetaStorage
}

func NewMetaStorage(logger logger.Logger, config config.MetaStorage, db *dbpg.DB) *MetaStorage {
	return &MetaStorage{db: db, logger: logger, config: config}
}

func (s *MetaStorage) SaveImageMeta(ctx context.Context, image *models.Image) error {

	_, err := s.db.QueryRowWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, `
		
		INSERT INTO images (uuid, object_key, original_filename, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`,

		image.ID, image.ObjectKey, image.FileHeader.Filename, models.StatusPending, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to save image metadata: %w", err)
	}

	return nil

}

func (s *MetaStorage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database closed", "layer", "repository.postgres")
	}
}

func (s *MetaStorage) DB() *dbpg.DB {
	return s.db
}

func (s *MetaStorage) Config() *config.MetaStorage {
	return &s.config
}
