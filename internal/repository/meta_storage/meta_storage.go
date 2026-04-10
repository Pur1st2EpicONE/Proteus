// Package meta_storage provides a high-level abstraction for
// PostgreSQL-backed metadata operations (save, retrieve, status updates,
// soft-delete and cleanup). It defines the MetaStorage interface and
// wires the concrete PostgreSQL implementation.
package meta_storage

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/meta_storage/postgres"
	"context"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
)

// MetaStorage defines the contract for all metadata persistence
// operations used by the service layer (image lifecycle, status
// transitions and periodic cleanup).
type MetaStorage interface {
	SaveImageMeta(ctx context.Context, image *models.Image) error        // SaveImageMeta inserts initial metadata (UUID, object_key, pending status) for a newly uploaded image.
	GetImageMeta(ctx context.Context, id string) (string, string, error) // GetImageMeta returns the current object_key and status for the given image UUID.
	MarkAsReady(ctx context.Context, objectKey string, id string) error  // MarkAsReady updates the image to "ready" status and replaces the temporary object_key with the final processed one.
	MarkAsDeleted(ctx context.Context, id string) error                  // MarkAsDeleted performs a soft-delete by changing status to "deleted" (used by the DELETE /image/:id endpoint).
	GetDeleted(ctx context.Context) ([]models.Image, error)              // GetDeleted returns images that are either explicitly deleted or pending longer than PendingTimeout (used by the background cleaner).
	DeleteBatch(ctx context.Context, ids []string) error                 // DeleteBatch permanently removes a batch of images from the meta table (called by the cleaner after successful image storage deletion).
	Close()                                                              // Close closes the underlying database connection pool and logs the result.
}

// NewMetaStorage returns a concrete PostgreSQL-based implementation of
// the MetaStorage interface.
func NewMetaStorage(logger logger.Logger, config config.MetaStorage, db *dbpg.DB) MetaStorage {
	return postgres.NewMetaStorage(logger, config, db)
}

// ConnectDB creates a new *dbpg.DB connection pool using the provided
// configuration, validates the DSN and performs an initial ping.
func ConnectDB(config config.MetaStorage) (*dbpg.DB, error) {

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.DBName, config.SSLMode), nil, &dbpg.Options{
		MaxOpenConns: config.MaxOpenConns, MaxIdleConns: config.MaxIdleConns, ConnMaxLifetime: config.ConnMaxLifetime})
	if err != nil {
		return nil, fmt.Errorf("database driver not found or DSN invalid: %w", err)
	}

	if err := db.Master.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil

}
