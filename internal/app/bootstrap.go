package app

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pressly/goose/v3"
	"github.com/wb-go/wbf/dbpg"
)

const startupTimeout = 10 * time.Second

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

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	if err := initBucket(ctx, imageDb, config.MinIOBucket, logger); err != nil {
		return nil, fmt.Errorf("unable to init MinIO bucket: %w", err)
	}

	return imageDb, nil

}

func initBucket(ctx context.Context, client *minio.Client, bucketName string, logger logger.Logger) error {

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed check bucket existence: %w", err)
	}

	if exists {
		return nil
	}

	if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("failed to create bucket %q: %w", bucketName, err)
	}

	logger.LogInfo("app — MinIO bucket created", "bucket", bucketName, "layer", "app")

	return nil

}
