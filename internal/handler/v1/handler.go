package v1

import (
	"Proteus/internal/config"
	"Proteus/internal/service"
)

type Handler struct {
	config  config.Server
	service service.Service
}

func NewHandler(config config.Server, service service.Service) *Handler {
	return &Handler{config: config, service: service}
}
