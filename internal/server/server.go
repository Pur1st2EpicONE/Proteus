package server

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/server/httpserver"
	"context"
	"net/http"
)

type Server interface {
	Run()
	Shutdown()
}

func NewServer(logger logger.Logger, config config.Server, handler http.Handler, cancel context.CancelFunc) Server {
	return httpserver.NewServer(logger, config, handler, cancel)
}
