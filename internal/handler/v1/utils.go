package v1

import (
	"Proteus/internal/errs"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// validateHeader performs basic validation of the uploaded file header:
// non-empty file and supported image MIME types (jpeg, png, webp, gif).
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

// respondOK sends a 200 OK JSON response wrapped as {"result": ...}.
func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

// respondAccepted sends a 202 Accepted JSON response (used when the
// requested image is still being processed).
func respondAccepted(c *ginext.Context, response any) {
	c.JSON(http.StatusAccepted, ginext.H{"result": response})
}

// respondWithData sends raw image bytes with the correct Content-Type
// header (used for direct image downloads).
func respondWithData(c *ginext.Context, contentType string, data []byte) {
	c.Data(http.StatusOK, contentType, data)
}

// respondError maps any error to the appropriate HTTP status code and
// message, then aborts the request with a JSON error body.
func respondError(c *ginext.Context, err error) {
	if err != nil {
		status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(status, ginext.H{"error": msg})
	}
}

// mapErrorToStatus translates domain and standard library errors into
// the corresponding HTTP status code and user-facing message.
func mapErrorToStatus(err error) (int, string) {

	switch {
	case errors.Is(err, errs.ErrNoFile),
		errors.Is(err, errs.ErrReadFile),
		errors.Is(err, errs.ErrInvalidImageID),
		errors.Is(err, errs.ErrInvalidImageContent),
		errors.Is(err, errs.ErrUnsupportedImageFormat),
		errors.Is(err, errs.ErrInvalidImageDimensions),
		errors.Is(err, errs.ErrNoActionsProvided),
		errors.Is(err, errs.ErrUnsupportedAction),
		errors.Is(err, errs.ErrWatermarkTextRequired),
		errors.Is(err, errs.ErrResizeDimensionsRequired),
		errors.Is(err, errs.ErrNegativeResizeDimensions):
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

// rbTooLarge reports whether the error is caused by exceeding the
// configured maximum request body size.
func rbTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}
