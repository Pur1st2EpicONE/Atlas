package v1

import (
	"Atlas/internal/errs"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) GetItemHistory(c *ginext.Context) {

	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, errs.ErrInvalidItemID)
		return
	}

	filter, err := parseQuery(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	history, err := h.service.GetItemHistory(c.Request.Context(), itemID, filter)
	if err != nil {
		RespondError(c, err)
		return
	}

	fmtRespond(c, history)

}
