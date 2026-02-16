package app

import (
	"Proteus/internal/config"
	"Proteus/internal/handler"
	"Proteus/internal/logger"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"Proteus/internal/server"
	"Proteus/internal/service"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/pressly/goose/v3"
	"github.com/wb-go/wbf/dbpg"
)

type App struct {
	logger       logger.Logger
	logFile      *os.File
	server       server.Server
	ctx          context.Context
	cancel       context.CancelFunc
	metaStorage  meta_storage.MetaStorage
	imageStorage image_storage.ImageStorage
}

func Boot() *App {

	config, err := config.Load()
	if err != nil {
		log.Fatalf("app — failed to load configs: %v", err)
	}

	logger, logFile := logger.NewLogger(config.Logger)

	metaDb, imageDb, err := bootstrapRepository(logger, config.Repository)
	if err != nil {
		logger.LogFatal("app — failed to bootstrap repository layer", err, "layer", "app")
	}

	return wireApp(metaDb, imageDb, logger, logFile, config)

}

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

func wireApp(metaDb *dbpg.DB, imageDb *minio.Client, logger logger.Logger, logFile *os.File, config config.Config) *App {

	ctx, cancel := newContext(logger)
	metaStorge := meta_storage.NewMetaStorage(logger, config.Repository.MetaStorage, metaDb)
	imageStorage := image_storage.NewImageStorage(logger, config.Repository.ImageStorage, imageDb)
	service := service.NewService(logger, metaStorge, imageStorage)
	handler := handler.NewHandler(service)
	server := server.NewServer(logger, config.Server, handler)

	return &App{
		logger:       logger,
		logFile:      logFile,
		server:       server,
		ctx:          ctx,
		cancel:       cancel,
		metaStorage:  metaStorge,
		imageStorage: imageStorage,
	}

}

func newContext(logger logger.Logger) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		logger.LogInfo("app — received signal "+sig.String()+", initiating graceful shutdown", "layer", "app")
		cancel()
	}()

	return ctx, cancel

}

func (a *App) Run() {

	go func() {
		if err := a.server.Run(); err != nil {
			a.logger.LogFatal("server run failed", err, "layer", "app")
		}
	}()

	<-a.ctx.Done()

	a.Stop()

}

func (a *App) Stop() {

	a.server.Shutdown()
	a.metaStorage.Close()

	if a.logFile != nil && a.logFile != os.Stdout {
		_ = a.logFile.Close()
	}

}
