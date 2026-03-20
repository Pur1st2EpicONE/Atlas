package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) SignUp(c *ginext.Context) {

	var request RegisterDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, errs.ErrInvalidJSON)
		return
	}

	userID, err := h.service.CreateUser(c.Request.Context(), models.User{
		Login:    request.Login,
		Password: request.Password,
		Role:     request.Role})
	if err != nil {
		RespondError(c, err)
		return
	}

	token, err := h.service.CreateToken(models.User{
		ID:   userID,
		Role: request.Role,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, token)

}
