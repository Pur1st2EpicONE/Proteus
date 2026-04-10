// Package httpserver contains the concrete implementation of the
// server.Server interface using Go's standard library http.Server.
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
	shutdownTimeout time.Duration      // shutdownTimeout is the maximum time allowed for graceful shutdown.
	logger          logger.Logger      // logger is used for all server lifecycle and error events.
	cancel          context.CancelFunc // cancel is the root application cancel function; called on fatal ListenAndServe errors.
	instance        *http.Server       // instance is the underlying http.Server configured with timeouts and the request handler.
}

// NewServer constructs a new HttpServer with the given configuration.
// It configures the standard http.Server with all timeouts and limits
// from the config and stores the cancel function for emergency shutdown.
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

// Run starts the HTTP server on the configured address and blocks
// until the server is closed (either gracefully or due to a fatal error).
// If ListenAndServe returns any error other than http.ErrServerClosed,
// it logs the error and initiates an emergency application shutdown.
func (s *HttpServer) Run() {
	s.logger.LogInfo("server — receiving requests", "layer", "server.httpserver")
	if err := s.instance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.LogError("server — fatal at ListenAndServe, initiating emergency shutdown", err, "layer", "server.httpserver")
		s.cancel()
	}
}

// Shutdown performs a graceful shutdown of the HTTP server using the
// configured shutdown timeout. It logs success or failure.
func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.instance.Shutdown(ctx); err != nil {
		s.logger.LogError("server — failed to shutdown gracefully", err, "layer", "server.httpserver")
	} else {
		s.logger.LogInfo("server — shutdown complete", "layer", "server.httpserver")
	}
}
