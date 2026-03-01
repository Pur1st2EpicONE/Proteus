package handler

import (
	"Proteus/internal/config"
	v1 "Proteus/internal/handler/v1"
	"Proteus/internal/service"
	"html/template"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

const templatePath = "web/templates/index.html"

func NewHandler(config config.Server, service service.Service) http.Handler {

	handler := ginext.New("")

	handler.Use(ginext.Recovery())
	handler.Static("/static", "./web/static")

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(config, service)

	apiV1.POST("/upload", handlerV1.UploadImage)
	apiV1.GET("/image/:id", handlerV1.GetImage)
	apiV1.DELETE("/image/:id", handlerV1.MarkAsDeleted)

	handler.GET("/", homePage(template.Must(template.ParseFiles(templatePath))))

	return handler

}

func homePage(t *template.Template) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		if err := t.Execute(c.Writer, nil); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ginext.H{"error": "Failed to render page"})
		}
	}
}
