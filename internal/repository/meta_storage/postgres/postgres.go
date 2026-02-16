package postgres

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

type MetaStorage struct {
	db     *dbpg.DB
	logger logger.Logger
	config config.MetaStorage
}

func NewMetaStorage(logger logger.Logger, config config.MetaStorage, db *dbpg.DB) *MetaStorage {
	return &MetaStorage{db: db, logger: logger, config: config}
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
