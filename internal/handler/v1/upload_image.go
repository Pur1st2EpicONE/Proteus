package v1

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

const maxFileSize = 10 << 20 // 10 MB
const maxRequestSize = maxFileSize + (2 << 20)

func (h *Handler) UploadImage(c *ginext.Context) {

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestSize)

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

	imageID, err := h.service.UploadImage(c.Request.Context(), &models.Image{File: file, FileHeader: header})
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, ginext.H{"image_id": imageID})

}

func validateHeader(header *multipart.FileHeader) error {

	if header.Size == 0 {
		return errs.ErrEmptyFile
	}

	if header.Size > maxFileSize {
		return errs.ErrFileTooLarge
	}

	switch header.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return errs.ErrUnsupportedImageFormat
	}

}
