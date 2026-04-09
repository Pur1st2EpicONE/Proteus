// Package app wires application components for the Proteus image processing service,
// provides lifecycle management for the HTTP server, Kafka consumer/producer,
// dual storages (meta DB and MinIO) and exposes the entry point for booting
// and running the service with graceful shutdown.
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
	logger       logger.Logger              // logger is the structured logger used across all application layers.
	logFile      *os.File                   // logFile is the file handle where logs are written (nil or os.Stdout if logging to console).
	server       server.Server              // server is the HTTP server instance handling incoming requests.
	consumer     broker.Consumer            // consumer is the Kafka consumer responsible for image processing tasks.
	producer     broker.Producer            // producer is the Kafka producer used to publish new image tasks.
	ctx          context.Context            // ctx is the root context used to coordinate graceful shutdown across components.
	service      service.Service            // service is the business-logic layer that orchestrates image processing.
	metaStorage  meta_storage.MetaStorage   // metaStorage is the PostgreSQL-backed repository for image metadata.
	imageStorage image_storage.ImageStorage // imageStorage is the MinIO-backed repository for raw and processed images.
}

// Boot loads configuration, initializes logger, bootstraps both meta and image
// repositories (PostgreSQL + migrations and MinIO + bucket) and wires all
// components. It returns a fully constructed *App ready to run.
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

// wireApp constructs application components (meta storage, image storage,
// service, server, Kafka producer and consumer), creates a cancellable context
// and returns the assembled *App.
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

// newContext creates a context that is cancelled when the process
// receives SIGINT or SIGTERM. It also logs receipt of the signal
// and initiates graceful shutdown by calling the cancel function.
func newContext(logger logger.Logger) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		sigString := sig.String()
		if sig == syscall.SIGTERM {
			sigString = "terminate" // sig.String() returns the SIGTERM string in past tense for some reason
		}
		logger.LogInfo("app — received signal "+sigString+", initiating graceful shutdown", "layer", "app")
		cancel()
	}()

	return ctx, cancel

}

// processFunc returns a closure that adapts the service.ProcessImage method
// to the signature expected by the Kafka consumer.
func processFunc(service service.Service) func(ctx context.Context, image models.ImageProcessTask) error {
	return func(ctx context.Context, image models.ImageProcessTask) error {
		return service.ProcessImage(ctx, image)
	}
}

// Run starts the server, consumer and service cleaner in background goroutines
// and blocks until the application's context is cancelled. After cancellation
// it invokes Stop.
func (a *App) Run() {

	go a.server.Run()
	go a.consumer.Run()
	go a.service.Cleaner(a.ctx)

	<-a.ctx.Done()

	a.Stop()

}

// Stop performs an orderly shutdown of application components: it shuts down
// the server, closes the Kafka consumer and producer, closes both storages
// and the log file if it is not os.Stdout. It also logs the stop event.
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
