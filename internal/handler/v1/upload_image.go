package v1

import (
	"Proteus/internal/models"
	"io"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

const maxFileSize = 10 << 20 // 10 MB
const maxRequestSize = maxFileSize + (2 << 20)

func (h *Handler) UploadImage(c *ginext.Context) {

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestSize)

	var request UploadImageDTO
	if err := c.ShouldBind(&request); err != nil {
		respondError(c, err)
		return
	}

	multipartFile, header, err := c.Request.FormFile("image")
	if err != nil {
		respondError(c, err)
		return
	}
	defer multipartFile.Close()

	if err := validateHeader(header); err != nil {
		respondError(c, err)
		return
	}

	file, err := io.ReadAll(io.LimitReader(multipartFile, maxFileSize+1))
	if err != nil {
		respondError(c, err)
		return
	}

	imageID, err := h.service.UploadImage(c.Request.Context(), &models.Image{
		File:       file,
		FileHeader: header,
		Request: models.Request{
			Action:    request.Action,
			Watermark: request.Watermark,
			Height:    request.Height,
			Width:     request.Width,
			Quality:   request.Quality}})
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, ginext.H{"image_id": imageID})

}
