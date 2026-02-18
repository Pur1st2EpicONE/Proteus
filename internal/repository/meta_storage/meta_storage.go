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

type MetaStorage interface {
	SaveImageMeta(ctx context.Context, image *models.Image) error
	Close()
}

func NewMetaStorage(logger logger.Logger, config config.MetaStorage, db *dbpg.DB) MetaStorage {
	return postgres.NewMetaStorage(logger, config, db)
}

func ConnectDB(config config.MetaStorage) (*dbpg.DB, error) {

	options := &dbpg.Options{
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.DBName, config.SSLMode), nil, options)
	if err != nil {
		return nil, fmt.Errorf("database driver not found or DSN invalid: %w", err)
	}

	if err := db.Master.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil

}
