package v1

import (
	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) GetItems(c *ginext.Context) {
	items, err := h.service.GetItems(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	respondOK(c, items)
}
