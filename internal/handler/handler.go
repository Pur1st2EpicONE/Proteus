package handler

import (
	v1 "Proteus/internal/handler/v1"
	"Proteus/internal/service"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

const templatePath = "web/templates/index.html"

func NewHandler(service service.Service) http.Handler {

	handler := ginext.New("")

	handler.Use(ginext.Recovery())
	handler.Static("/static", "./web/static")

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(service)

	_ = apiV1
	_ = handlerV1

	return handler

}
