package v1

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"io"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// UploadImage handles POST /api/v1/upload requests.
// It enforces the configured request and file size limits, binds form
// data, validates the uploaded file header, reads the file into memory
// and passes the image to the service for asynchronous processing.
func (h *Handler) UploadImage(c *ginext.Context) {

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.config.MaxRequestSize)

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
	defer func() { _ = multipartFile.Close() }()

	if err := h.validateHeader(header); err != nil {
		respondError(c, err)
		return
	}

	file, err := io.ReadAll(io.LimitReader(multipartFile, h.config.MaxFileSize+1))
	if err != nil || int64(len(file)) > h.config.MaxFileSize {
		respondError(c, errs.ErrFileTooLarge)
		return
	}

	imageID, err := h.service.UploadImage(c.Request.Context(), &models.Image{
		File:       file,
		FileHeader: header,
		Request: models.Request{
			Action:    request.Action,
			Watermark: request.Watermark,
			Height:    request.Height,
			Width:     request.Width}})
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, imageID)

}
