package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) SignIn(c *ginext.Context) {

	var request LoginDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, errs.ErrInvalidJSON)
		return
	}

	user, err := h.service.GetUser(c.Request.Context(),
		models.User{Login: request.Login, Password: request.Password})
	if err != nil {
		RespondError(c, err)
		return
	}

	token, err := h.service.CreateToken(user)
	if err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, token)

}
