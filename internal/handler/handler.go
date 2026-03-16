package handler

import (
	"Atlas/internal/config"
	"Atlas/internal/errs"
	"Atlas/internal/service"
	"Atlas/internal/service/impl"
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	v1 "Atlas/internal/handler/v1"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
		tokenString := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 {
				tokenString = parts[1]
			}
		} else {
			if cookie, err := c.Cookie("token"); err == nil {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			v1.RespondError(c, errs.ErrEmptyAuthHeader)
			return
		}

		claims := &impl.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, service.KeyFunc)
		if err != nil || !token.Valid {
			v1.RespondError(c, errs.ErrInvalidToken)
			return
		}

		userID, _ := strconv.ParseInt(claims.Subject, 10, 64)

		ctx := context.WithValue(c.Request.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

	}

}

func RequireRole(allowed ...string) gin.HandlerFunc {

	return func(c *ginext.Context) {

		role := c.Request.Context().Value("role")
		if role == nil {
			v1.RespondError(c, errs.ErrInvalidToken)
			return
		}

		userRole := role.(string)
		if slices.Contains(allowed, userRole) {
			c.Next()
			return
		}

		v1.RespondError(c, errors.New("insufficient permissions"))

	}

}
