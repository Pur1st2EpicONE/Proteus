package v1

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/helpers"
)

// GetImage handles GET /api/v1/image/:id requests.
// It validates the UUID, retrieves image metadata via the service,
// and either returns the processed image binary or an appropriate
// status code (202 for pending, 404 for not found, etc.).
func (h *Handler) GetImage(c *ginext.Context) {

	id := c.Param("id")
	if err := helpers.ParseUUID(id); err != nil {
		respondError(c, errs.ErrInvalidImageID)
		return
	}

	objectKey, status, err := h.service.GetImageMeta(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	if status == models.StatusPending {
		respondAccepted(c, status)
		return
	}

	data, contentType, err := h.service.DownloadImage(c.Request.Context(), objectKey)
	if err != nil {
		respondError(c, err)
		return
	}

	respondWithData(c, contentType, data)

}
