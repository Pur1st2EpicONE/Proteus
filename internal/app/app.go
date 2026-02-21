package app

import (
	"Proteus/internal/broker"
	"Proteus/internal/config"
	"Proteus/internal/handler"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"Proteus/internal/repository/meta_storage"
	"Proteus/internal/server"
	"Proteus/internal/service"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/wb-go/wbf/dbpg"
	wbf "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type App struct {
	logger       logger.Logger
	logFile      *os.File
	server       server.Server
	consumer     broker.Consumer
	producer     broker.Producer
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
		logger.LogFatal("app — failed to bootstrap repository", err, "layer", "app")
	}

	return wireApp(metaDb, imageDb, logger, logFile, config)

}

func wireApp(metaDb *dbpg.DB, imageDb *minio.Client, logger logger.Logger, logFile *os.File, config config.Config) *App {

	ctx, cancel := newContext(logger)

	mStorage := meta_storage.NewMetaStorage(logger, config.Repository.MetaStorage, metaDb)
	iStorage := image_storage.NewImageStorage(logger, config.Repository.ImageStorage, imageDb)

	wbfProducer := wbf.NewProducer(config.Consumer.Brokers, config.Consumer.Topic)
	producer := broker.NewProducer(logger, config.Producer, wbfProducer)

	service := service.NewService(logger, producer, mStorage, iStorage)

	wbfConsumer := wbf.NewConsumer(config.Consumer.Brokers, config.Consumer.Topic, config.Consumer.GroupID)
	consumer := broker.NewConsumer(logger, config.Consumer, wbfConsumer, processFunc(service), iStorage)

	handler := handler.NewHandler(service)
	server := server.NewServer(logger, config.Server, handler)

	return &App{
		logger:       logger,
		logFile:      logFile,
		server:       server,
		consumer:     consumer,
		producer:     producer,
		ctx:          ctx,
		cancel:       cancel,
		metaStorage:  mStorage,
		imageStorage: iStorage,
	}

}

func processFunc(service service.Service) func(ctx context.Context, image models.ImageProcessTask) error {
	return func(ctx context.Context, image models.ImageProcessTask) error {
		return service.ProcessImage(ctx, image)
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

	go a.consumer.Run(a.ctx, retry.Strategy{Attempts: 3, Delay: 5, Backoff: 2})

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
