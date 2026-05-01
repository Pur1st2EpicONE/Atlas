package v1

import (
	"github.com/wb-go/wbf/ginext"
)

// GetItems handles GET /api/v1/items requests.
// It returns the full list of items.
func (h *Handler) GetItems(c *ginext.Context) {
	items, err := h.service.GetItems(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	respondOK(c, items)
}
