package v1

import (
	"Proteus/internal/errs"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) validateHeader(header *multipart.FileHeader) error {

	if header.Size == 0 {
		return errs.ErrNoFile
	}

	if header.Size > h.config.MaxFileSize {
		return errs.ErrFileTooLarge
	}

	switch header.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return errs.ErrUnsupportedImageFormat
	}

}

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

func respondAccepted(c *ginext.Context, response any) {
	c.JSON(http.StatusAccepted, ginext.H{"result": response})
}

func respondWithData(c *ginext.Context, contentType string, data []byte) {
	c.Data(http.StatusOK, contentType, data)
}

func respondError(c *ginext.Context, err error) {
	if err != nil {
		status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(status, ginext.H{"error": msg})
	}
}

func mapErrorToStatus(err error) (int, string) {

	switch {
	case errors.Is(err, errs.ErrNoFile),
		errors.Is(err, errs.ErrReadFile),
		errors.Is(err, errs.ErrInvalidImageContent),
		errors.Is(err, errs.ErrUnsupportedImageFormat),
		errors.Is(err, errs.ErrInvalidImageDimensions),
		errors.Is(err, errs.ErrNoActionsProvided),
		errors.Is(err, errs.ErrUnsupportedAction),
		errors.Is(err, errs.ErrWatermarkTextRequired),
		errors.Is(err, errs.ErrResizeDimensionsRequired),
		errors.Is(err, errs.ErrNegativeResizeDimensions),
		errors.Is(err, errs.ErrInvalidQualityRange):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, http.ErrMissingFile):
		return http.StatusBadRequest, errs.ErrNoFile.Error()

	case rbTooLarge(err),
		errors.Is(err, errs.ErrFileTooLarge),
		errors.Is(err, errs.ErrImageTooLargeDimensions):
		return http.StatusRequestEntityTooLarge, err.Error()

	case errors.Is(err, errs.ErrImageNotFound):
		return http.StatusNotFound, err.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}

func rbTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
