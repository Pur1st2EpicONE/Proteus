package v1

import (
	"Proteus/internal/errs"
	"errors"
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

func parseParam(c *ginext.Context) (int64, error) {

	idStr := c.Param("id")
	if idStr == "" {
		return 0, errs.ErrEmptyCommentID
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errs.ErrInvalidCommentID
	}

	return id, nil

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
	case errors.Is(err, errs.ErrInvalidJSON),
		errors.Is(err, errs.ErrEmptyContent),
		errors.Is(err, errs.ErrEmptyAuthor),
		errors.Is(err, errs.ErrInvalidParentID),
		errors.Is(err, errs.ErrInvalidPage),
		errors.Is(err, errs.ErrInvalidLimit),
		errors.Is(err, errs.ErrEmptyCommentID),
		errors.Is(err, errs.ErrInvalidCommentID),
		errors.Is(err, errs.ErrInvalidSort):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, errs.ErrParentNotFound),
		errors.Is(err, errs.ErrCommentNotFound):
		return http.StatusNotFound, err.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}
