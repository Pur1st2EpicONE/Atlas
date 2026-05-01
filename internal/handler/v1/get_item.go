package v1

import (
	"Atlas/internal/errs"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

// GetItem handles GET /api/v1/items/:id requests.
// It extracts the item ID from the URL path and returns the corresponding item.
// Returns an error if the item does not exist or the ID is invalid.
func (h *Handler) GetItem(c *ginext.Context) {

	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, errs.ErrInvalidItemID)
		return
	}

	item, err := h.service.GetItem(c.Request.Context(), itemID)
	if err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, item)

}
