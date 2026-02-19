package app

import (
	"Proteus/internal/broker"
	"Proteus/internal/config"
	"Proteus/internal/handler"
	"Proteus/internal/logger"
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
	km "github.com/segmentio/kafka-go"
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

func wireApp(metaDb *dbpg.DB, imageDb *minio.Client, logger logger.Logger, logFile *os.File, cfg config.Config) *App {

	ctx, cancel := newContext(logger)
	consumer := broker.NewConsumer(logger, cfg.Consumer, wbf.NewConsumer(cfg.Consumer.Brokers, cfg.Consumer.Topic, cfg.Consumer.GroupID))
	producer := broker.NewProducer(logger, cfg.Producer, wbf.NewProducer(cfg.Consumer.Brokers, cfg.Consumer.Topic))
	metaStorge := meta_storage.NewMetaStorage(logger, cfg.Repository.MetaStorage, metaDb)
	imageStorage := image_storage.NewImageStorage(logger, cfg.Repository.ImageStorage, imageDb)
	service := service.NewService(logger, producer, metaStorge, imageStorage)
	handler := handler.NewHandler(service)
	server := server.NewServer(logger, cfg.Server, handler)

	return &App{
		logger:       logger,
		logFile:      logFile,
		server:       server,
		consumer:     consumer,
		producer:     producer,
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

	c := make(chan km.Message)

	go a.consumer.Run(a.ctx, c, retry.Strategy{Attempts: 3, Delay: 5, Backoff: 2})

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
