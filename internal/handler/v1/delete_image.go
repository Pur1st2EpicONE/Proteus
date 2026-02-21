package v1

import (
	"fmt"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) DeleteImage(c *ginext.Context) {

	id := c.Param("id")
	if id == "" {
		respondError(c, fmt.Errorf("id cannot be empty"))
		return
	}

	err := h.service.DeleteImage(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "Image deleted successfully"})

}
