package httpserver

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"context"
	"errors"
	"net/http"
	"time"
)

type HttpServer struct {
	shutdownTimeout time.Duration
	logger          logger.Logger
	cancel          context.CancelFunc
	instance        *http.Server
}

func NewServer(logger logger.Logger, config config.Server, handler http.Handler, cancel context.CancelFunc) *HttpServer {

	return &HttpServer{
		shutdownTimeout: config.ShutdownTimeout,
		logger:          logger,
		cancel:          cancel,
		instance: &http.Server{
			Addr:           ":" + config.Port,
			Handler:        handler,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes},
	}

}

func (s *HttpServer) Run() {
	s.logger.LogInfo("server — receiving requests", "layer", "server.httpserver")
	if err := s.instance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.LogError("server — fatal at ListenAndServe, initiating emergency shutdown", err, "layer", "server.httpserver")
		s.cancel()
	}
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.instance.Shutdown(ctx); err != nil {
		s.logger.LogError("server — failed to shutdown gracefully", err, "layer", "server.httpserver")
	} else {
		s.logger.LogInfo("server — shutdown complete", "layer", "server.httpserver")
	}
}
