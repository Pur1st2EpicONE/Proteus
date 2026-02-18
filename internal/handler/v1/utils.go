package v1

import (
	"Proteus/internal/errs"
	"errors"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

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
		errors.Is(err, errs.ErrImageTooLargeDimensions):
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
