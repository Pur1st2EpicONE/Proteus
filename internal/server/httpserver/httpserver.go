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
	srv             *http.Server
	shutdownTimeout time.Duration
	logger          logger.Logger
}

func NewServer(logger logger.Logger, config config.Server, handler http.Handler) *HttpServer {

	server := &HttpServer{
		shutdownTimeout: config.ShutdownTimeout,
		logger:          logger,
	}

	server.srv = &http.Server{
		Addr:           ":" + config.Port,
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return server

}

func (s *HttpServer) Run() error {
	s.logger.LogInfo("server — receiving requests", "layer", "server")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.LogError("server — failed to shutdown gracefully", err, "layer", "server")
	} else {
		s.logger.LogInfo("server — shutdown complete", "layer", "server")
	}
}
