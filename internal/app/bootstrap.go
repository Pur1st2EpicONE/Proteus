package app

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/pressly/goose/v3"
	"github.com/wb-go/wbf/dbpg"
)

func bootstrapRepository(logger logger.Logger, config config.Repository) (*dbpg.DB, *minio.Client, error) {

	metaDb, err := bootstrapMetaDB(logger, config.MetaStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("fatal at metaDb: %w", err)
	}

	imageDb, err := bootstrapImageDB(logger, config.ImageStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("fatal at imageDb: %w", err)

	}
	return metaDb, imageDb, nil

}

func bootstrapMetaDB(logger logger.Logger, config config.MetaStorage) (*dbpg.DB, error) {

	metaDb, err := meta_storage.ConnectDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meta storage: %w", err)
	}

	logger.LogInfo("app — connected to meta database", "layer", "app")

	if err := goose.SetDialect(config.Dialect); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(metaDb.Master, config.MigrationsDir); err != nil {
		return nil, fmt.Errorf("failed to apply goose migrations: %w", err)
	}

	logger.Debug("app — migrations applied", "layer", "app")

	return metaDb, nil

}

func bootstrapImageDB(logger logger.Logger, config config.ImageStorage) (*minio.Client, error) {

	imageDb, err := image_storage.ConnectDB(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to image storage: %w", err)
	}

	logger.LogInfo("app — connected to image database", "layer", "app")

	err = imageDb.MakeBucket(context.Background(), config.MinIOBucket, minio.MakeBucketOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot create bucket %q: %w", config.MinIOBucket, err)
	}

	logger.LogInfo("MinIO bucket created", "bucket", config.MinIOBucket)

	return imageDb, nil

}
