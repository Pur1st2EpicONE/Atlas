package handler

import (
	"Atlas/internal/config"
	"Atlas/internal/errs"
	"Atlas/internal/service"
	"context"
	"net/http"
	"strings"

	v1 "Atlas/internal/handler/v1"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

const header = "Authorization"

func NewHandler(config config.Server, service *service.Service) http.Handler {

	handler := ginext.New("")

	handler.Use(ginext.Recovery())
	handler.Static("/static", "./web/static")

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(config, *service)

	auth := apiV1.Group("/auth")
	auth.POST("/sign-up", handlerV1.SignUp)
	auth.POST("/sign-in", handlerV1.SignIn)

	protected := apiV1.Group("/")
	protected.Use(authJWT(service.AuthService))

	protected.POST("/events", handlerV1.CreateEvent)

	return handler

}

func authJWT(service service.AuthService) gin.HandlerFunc {

	return func(c *ginext.Context) {

		authHeader := c.GetHeader(header)
		if authHeader == "" {
			v1.RespondError(c, errs.ErrEmptyAuthHeader)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			v1.RespondError(c, errs.ErrInvalidAuthHeader)
			return
		}

		userID, err := service.ParseToken(parts[1])
		if err != nil {
			v1.RespondError(c, err)
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "userID", userID))
		c.Next()

	}

}
