package handler

import (
	"Atlas/internal/config"
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"Atlas/internal/service"
	"Atlas/internal/service/impl"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"

	v1 "Atlas/internal/handler/v1"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/wb-go/wbf/ginext"
)

const (
	indexPath  = "web/templates/index.html"
	loginPath  = "web/templates/login.html"
	signupPath = "web/templates/signup.html"
	header     = "Authorization"
)

func NewHandler(config config.Server, service *service.Service) http.Handler {

	handler := ginext.New("")
	handler.Use(ginext.Recovery())
	handler.Static("/static", "./web/static")

	handler.GET("/", renderPage(template.Must(template.ParseFiles(indexPath))))
	handler.GET("/login", renderPage(template.Must(template.ParseFiles(loginPath))))
	handler.GET("/signup", renderPage(template.Must(template.ParseFiles(signupPath))))

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(config, *service)

	auth := apiV1.Group("/auth")
	auth.POST("/sign-up", handlerV1.SignUp)
	auth.POST("/sign-in", handlerV1.SignIn)

	protected := apiV1.Group("/")
	protected.Use(authJWT(service.AuthService))

	items := protected.Group("/items")

	viewGroup := items.Group("").Use(requireRole(models.Viewer, models.Manager, models.Admin))
	viewGroup.GET("", handlerV1.GetItems)
	viewGroup.GET("/:id", handlerV1.GetItem)

	editGroup := items.Group("").Use(requireRole(models.Manager, models.Admin))
	editGroup.POST("", handlerV1.CreateItem)
	editGroup.PUT("/:id", handlerV1.UpdateItem)

	sudoGroup := items.Group("").Use(requireRole(models.Admin))
	sudoGroup.GET("/:id/history", handlerV1.GetItemHistory)
	sudoGroup.DELETE("/:id", handlerV1.DeleteItem)

	return handler

}

func renderPage(tmpl *template.Template) gin.HandlerFunc {
	return func(c *ginext.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(c.Writer, nil); err != nil {
			c.String(http.StatusInternalServerError, errs.ErrInternal.Error())
		}
	}
}

func authJWT(service service.AuthService) gin.HandlerFunc {

	return func(c *ginext.Context) {

		authHeader := c.GetHeader(header)
		tokenString := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = parts[1]
			}
		} else {
			if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
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

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			v1.RespondError(c, errs.ErrInvalidUserID)
			return
		}

		c.Set("userID", userID)
		c.Set("role", claims.Role)

		c.Next()

	}

}

func requireRole(allowed ...string) gin.HandlerFunc {

	return func(c *ginext.Context) {

		role, exists := c.Get("role")
		if !exists {
			v1.RespondError(c, errs.ErrInvalidToken)
			return
		}

		userRole, ok := role.(string)
		if !ok || !slices.Contains(allowed, userRole) {
			v1.RespondError(c, errs.ErrInsufficientPermissions)
			return
		}

		c.Next()

	}

}
