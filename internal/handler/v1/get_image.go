package v1

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"fmt"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) GetImage(c *ginext.Context) {

	id := c.Param("id")
	if id == "" {
		respondError(c, fmt.Errorf("id cannot be empty"))
		return
	}

	objectKey, status, err := h.service.GetImageMeta(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	if status == "deleted" {
		respondError(c, errs.ErrImageNotFound)
		return
	}

	if status == models.StatusPending {
		c.Header("Retry-After", "5")
		c.JSON(http.StatusAccepted, ginext.H{
			"status":  status,
			"message": "Image is not ready yet",
		})
		return
	}

	data, contentType, err := h.service.DownloadImage(c.Request.Context(), objectKey)
	if err != nil {
		respondError(c, err)
		return
	}

	c.Data(http.StatusOK, contentType, data)

}
