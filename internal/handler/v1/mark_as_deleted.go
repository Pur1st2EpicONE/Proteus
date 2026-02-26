package v1

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/helpers"
)

func (h *Handler) MarkAsDeleted(c *ginext.Context) {

	id := c.Param("id")
	if err := helpers.ParseUUID(id); err != nil {
		respondError(c, errs.ErrInvalidImageID)
		return
	}

	err := h.service.MarkAsDeleted(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, models.StatusDeleted)

}
