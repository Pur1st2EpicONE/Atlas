// Package handler provides HTTP handlers for the Atlas application.
// It sets up the Gin router, serves static files, renders HTML templates,
// and defines API routes with authentication and role-based authorization.
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
	indexPath  = "web/templates/index.html"  // indexPath is the template for the home page
	loginPath  = "web/templates/login.html"  // loginPath is the template for the login page
	signupPath = "web/templates/signup.html" // signupPath is the template for the signup page
	header     = "Authorization"             // header is the HTTP header name for the auth token
)

// NewHandler creates and configures the HTTP handler.
// It sets up static file serving, HTML page routes, and API version 1 routes
// with JWT authentication and role-based access control.
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

// renderPage executes an HTML template and writes the result to the response.
// It sets the Content-Type header and returns an internal server error on failure.
func renderPage(tmpl *template.Template) gin.HandlerFunc {
	return func(c *ginext.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(c.Writer, nil); err != nil {
			c.String(http.StatusInternalServerError, errs.ErrInternal.Error())
		}
	}
}

// authJWT validates the JWT token from the Authorization header or cookie.
// It extracts userID and role, stores them in the request context,
// and rejects requests with missing or invalid tokens.
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

// requireRole ensures that the authenticated user has one of the allowed roles.
// It reads the "role" value from the context and denies access if the role is missing or not permitted.
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
