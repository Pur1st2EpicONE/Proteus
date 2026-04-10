// Package postgres contains the concrete implementation of the
// meta_storage.MetaStorage interface using the wb-go/wbf/dbpg
// PostgreSQL driver with retry support.
package postgres

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

type MetaStorage struct {
	db     *dbpg.DB           // db is the underlying connection pool.
	logger logger.Logger      // logger is used for structured logging of all repository operations.
	config config.MetaStorage // config holds retry strategy, pending timeout and other meta-specific settings.
}

// NewMetaStorage constructs a new PostgreSQL-backed MetaStorage
// with the provided logger, configuration and DB connection.
func NewMetaStorage(logger logger.Logger, config config.MetaStorage, db *dbpg.DB) *MetaStorage {
	return &MetaStorage{logger: logger, config: config, db: db}
}

// Close gracefully shuts down the master connection pool and logs
// success or failure.
func (m *MetaStorage) Close() {
	if err := m.db.Master.Close(); err != nil {
		m.logger.LogError("postgres — failed to close properly", err, "layer", "repository.meta_storage.postgres")
	} else {
		m.logger.LogInfo("postgres — database closed", "layer", "repository.meta_storage.postgres")
	}
}

// DB returns the underlying *dbpg.DB instance (exposed for tests only).
func (m *MetaStorage) DB() *dbpg.DB {
	return m.db
}
