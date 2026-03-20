package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) CreateItem(c *ginext.Context) {

	var request CreateItemDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, errs.ErrInvalidJSON)
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	item, err := h.service.CreateItem(c.Request.Context(), userID, models.Item{
		Name:        request.Name,
		Description: request.Description,
		Quantity:    request.Quantity,
		Price:       request.Price})
	if err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, item)

}
