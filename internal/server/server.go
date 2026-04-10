// Package server defines the high-level Server interface for the
// Proteus HTTP server and provides a factory that wires the concrete
// implementation (based on the standard library http.Server).
package server

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/server/httpserver"
	"context"
	"net/http"
)

// Server abstracts the HTTP server lifecycle. It is used by the
// application root to start the server in a background goroutine
// and to perform a graceful shutdown when a termination signal is received.
type Server interface {
	Run()      // Run starts the HTTP server and blocks until it is shut down. On fatal errors it triggers application shutdown via the cancel function.
	Shutdown() // Shutdown performs a graceful shutdown of the server with the configured timeout.
}

// NewServer returns a concrete HTTP server implementation wired with
// the provided logger, configuration, request handler and root context
// cancel function (used for emergency shutdown on fatal server errors).
func NewServer(logger logger.Logger, config config.Server, handler http.Handler, cancel context.CancelFunc) Server {
	return httpserver.NewServer(logger, config, handler, cancel)
}
