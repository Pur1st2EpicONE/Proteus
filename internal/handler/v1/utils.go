package v1

import (
	"Proteus/internal/errs"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

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

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
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
		errors.Is(err, errs.ErrFileTooLarge),
		errors.Is(err, errs.ErrReadFile),
		errors.Is(err, errs.ErrEmptyFile),
		errors.Is(err, errs.ErrInvalidImageContent),
		errors.Is(err, errs.ErrUnsupportedImageFormat),
		errors.Is(err, errs.ErrInvalidImageDimensions),
		errors.Is(err, errs.ErrImageTooLargeDimensions),
		errors.Is(err, errs.ErrNoActionsProvided),
		errors.Is(err, errs.ErrUnsupportedAction),
		errors.Is(err, errs.ErrWatermarkTextRequired),
		errors.Is(err, errs.ErrResizeDimensionsRequired),
		errors.Is(err, errs.ErrNegativeResizeDimensions),
		errors.Is(err, errs.ErrInvalidQualityRange):
		return http.StatusBadRequest, err.Error()

	case rbTooLarge(err):
		return http.StatusRequestEntityTooLarge, errs.ErrRequestBodyTooLarge.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}

func rbTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
