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
)

type App struct {
	logger       logger.Logger
	logFile      *os.File
	server       server.Server
	consumer     broker.Consumer
	producer     broker.Producer
	ctx          context.Context
	service      service.Service
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
	producer := broker.NewProducer(logger, wbfProducer)

	service := service.NewService(logger, config.Service, producer, mStorage, iStorage)
	server := server.NewServer(logger, config.Server, handler.NewHandler(config.Server, service), cancel)

	wbfConsumer := wbf.NewConsumer(config.Consumer.Brokers, config.Consumer.Topic, config.Consumer.GroupID)
	consumer := broker.NewConsumer(ctx, logger, config.Consumer, wbfConsumer, processFunc(service), iStorage)

	return &App{
		logger:       logger,
		logFile:      logFile,
		server:       server,
		consumer:     consumer,
		producer:     producer,
		ctx:          ctx,
		service:      service,
		metaStorage:  mStorage,
		imageStorage: iStorage,
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

func processFunc(service service.Service) func(ctx context.Context, image models.ImageProcessTask) error {
	return func(ctx context.Context, image models.ImageProcessTask) error {
		return service.ProcessImage(ctx, image)
	}
}

func (a *App) Run() {

	go a.server.Run()
	go a.consumer.Run()
	go a.service.Cleaner(a.ctx)

	<-a.ctx.Done()

	a.Stop()

}

func (a *App) Stop() {

	a.server.Shutdown()

	a.consumer.Close()
	a.producer.Close()

	a.metaStorage.Close()
	a.imageStorage.Close()

	a.logger.LogInfo("app — stopped", "layer", "app")

	if a.logFile != nil && a.logFile != os.Stdout {
		_ = a.logFile.Close()
	}

}
