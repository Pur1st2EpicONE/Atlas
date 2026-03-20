package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) UpdateItem(c *ginext.Context) {

	var request UpdateItemDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, errs.ErrInvalidJSON)
		return
	}

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

	if err := h.service.UpdateItem(c.Request.Context(), userID, itemID, models.Update{
		Name:        request.Name,
		Description: request.Description,
		Quantity:    request.Quantity,
		Price:       request.Price}); err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, models.StatusUpdated)

}
