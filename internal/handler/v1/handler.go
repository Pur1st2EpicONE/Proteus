// Package v1 contains version 1 of the Proteus REST API handlers,
// request DTOs and response helper functions.
package v1

import (
	"Proteus/internal/config"
	"Proteus/internal/service"
)

// Handler encapsulates dependencies required by all v1 API endpoint
// implementations (server configuration limits and the business-logic
// service layer).
type Handler struct {
	config  config.Server   // config provides request size limits and timeouts used during upload handling.
	service service.Service // service is the application service that performs image processing logic.
}

// NewHandler creates a new v1 Handler instance wired with the given
// server configuration and service layer.
func NewHandler(config config.Server, service service.Service) *Handler {
	return &Handler{config: config, service: service}
}
