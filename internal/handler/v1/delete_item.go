package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) DeleteItem(c *ginext.Context) {

	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, errs.ErrInvalidItemID)
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	if err := h.service.DeleteItem(c.Request.Context(), userID, itemID); err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, models.StatusDeleted)

}
