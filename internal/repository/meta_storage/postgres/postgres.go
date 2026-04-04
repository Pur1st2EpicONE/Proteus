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
	return &MetaStorage{logger: logger, config: config, db: db}
}

func (m *MetaStorage) Close() {
	if err := m.db.Master.Close(); err != nil {
		m.logger.LogError("postgres — failed to close properly", err, "layer", "repository.meta_storage.postgres")
	} else {
		m.logger.LogInfo("postgres — database closed", "layer", "repository.meta_storage.postgres")
	}
}

func (m *MetaStorage) DB() *dbpg.DB {
	return m.db
}
