package v1

import (
	"Atlas/internal/config"
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"Atlas/internal/service"
	"Atlas/internal/service/mocks"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/wb-go/wbf/ginext"
	"go.uber.org/mock/gomock"
)

const (
	testUserID   = int64(1)
	testItemID   = int64(100)
	testLogin    = "abobus"
	testPassword = "qwerty"
	testRole     = models.Admin
)

func setupRouter(mockAuth *mocks.MockAuthService, mockCore *mocks.MockCoreService) http.Handler {

	gin.SetMode(gin.TestMode)
	router := ginext.New("")

	router.Use(func(c *ginext.Context) {
		userIDHeader := c.GetHeader("X-User-ID")
		if userIDHeader != "" {
			userID, err := strconv.ParseInt(userIDHeader, 10, 64)
			if err == nil {
				c.Set("userID", userID)
			}
		}
		roleHeader := c.GetHeader("X-Role")
		if roleHeader != "" {
			c.Set("role", roleHeader)
		}
		c.Next()
	})

	handler := NewHandler(config.Server{}, service.Service{
		AuthService: mockAuth,
		CoreService: mockCore,
	})

	auth := router.Group("/api/v1/auth")
	auth.POST("/sign-up", handler.SignUp)
	auth.POST("/sign-in", handler.SignIn)

	protected := router.Group("/api/v1")
	protected.Use(func(c *ginext.Context) {
		if _, exists := c.Get("userID"); !exists {
			RespondError(c, errs.ErrInvalidToken)
			c.Abort()
			return
		}
		if _, exists := c.Get("role"); !exists {
			RespondError(c, errs.ErrInvalidToken)
			c.Abort()
			return
		}
		c.Next()
	})

	items := protected.Group("/items")
	items.GET("", handler.GetItems)
	items.GET("/:id", handler.GetItem)
	items.POST("", handler.CreateItem)
	items.PUT("/:id", handler.UpdateItem)
	items.GET("/:id/history", handler.GetItemHistory)
	items.DELETE("/:id", handler.DeleteItem)

	return router

}

func TestHandler_SignUp(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewBufferString(`{"login":"test"`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidJSON.Error())
	})

	t.Run("service CreateUser error", func(t *testing.T) {
		mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("internal error"))
		body := `{"login":"test","password":"pass","role":"viewer"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(int64(0), errs.ErrUserAlreadyExists)
		body := `{"login":"test","password":"pass","role":"viewer"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusConflict, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrUserAlreadyExists.Error())
	})

	t.Run("CreateToken error", func(t *testing.T) {
		mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(testUserID, nil)
		mockAuth.EXPECT().CreateToken(gomock.Any()).Return("", errors.New("token error"))
		body := `{"login":"test","password":"pass","role":"viewer"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(testUserID, nil)
		mockAuth.EXPECT().CreateToken(gomock.Any()).Return("token123", nil)
		body := `{"login":"test","password":"pass","role":"viewer"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-up", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "token123")
	})

}

func TestHandler_SignIn(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewBufferString(`{"login":"test"`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidJSON.Error())
	})

	t.Run("GetUser error", func(t *testing.T) {
		mockAuth.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(models.User{}, errs.ErrInvalidCredentials)
		body := `{"login":"test","password":"wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidCredentials.Error())
	})

	t.Run("CreateToken error", func(t *testing.T) {
		mockAuth.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(models.User{ID: testUserID, Role: models.Viewer}, nil)
		mockAuth.EXPECT().CreateToken(gomock.Any()).Return("", errors.New("token error"))
		body := `{"login":"test","password":"pass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockAuth.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(models.User{ID: testUserID, Role: models.Admin}, nil)
		mockAuth.EXPECT().CreateToken(gomock.Any()).Return("token123", nil)
		body := `{"login":"test","password":"pass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-in", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "token123")
	})

}

func TestHandler_GetItems(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("no userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().GetItems(gomock.Any()).Return(nil, errors.New("db error"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		items := []models.Item{
			{ID: 1, Name: "Item1", Quantity: 5, Price: decimal.NewFromInt(100)},
			{ID: 2, Name: "Item2", Quantity: 3, Price: decimal.NewFromInt(200)},
		}
		mockCore.EXPECT().GetItems(gomock.Any()).Return(items, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "Item1")
	})

}

func TestHandler_GetItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid item ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/abc", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidItemID.Error())
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().GetItem(gomock.Any(), testItemID).Return(models.Item{}, errors.New("db error"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("item not found", func(t *testing.T) {
		mockCore.EXPECT().GetItem(gomock.Any(), testItemID).Return(models.Item{}, errs.ErrItemNotFound)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusNotFound, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrItemNotFound.Error())
	})

	t.Run("success", func(t *testing.T) {
		item := models.Item{ID: testItemID, Name: "Test Item", Quantity: 10, Price: decimal.NewFromInt(99)}
		mockCore.EXPECT().GetItem(gomock.Any(), testItemID).Return(item, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Viewer)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "Test Item")
	})

}

func TestHandler_CreateItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("no userID", func(t *testing.T) {
		body := `{"name":"item","quantity":1,"price":10.5}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(`{"name":123}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidJSON.Error())
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().CreateItem(gomock.Any(), testUserID, gomock.Any()).Return(models.Item{}, errors.New("db error"))
		body := `{"name":"new item","description":"desc","quantity":5,"price":19.99}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		created := models.Item{ID: testItemID, Name: "new item", Quantity: 5, Price: decimal.NewFromFloat(19.99)}
		mockCore.EXPECT().CreateItem(gomock.Any(), testUserID, gomock.Any()).Return(created, nil)
		body := `{"name":"new item","description":"desc","quantity":5,"price":19.99}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), strconv.FormatInt(testItemID, 10))
	})

}

func TestHandler_UpdateItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid item ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/abc", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidItemID.Error())
	})

	t.Run("no userID", func(t *testing.T) {
		body := `{"name":"updated"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), bytes.NewBufferString(`{"name":123}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidJSON.Error())
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().UpdateItem(gomock.Any(), testUserID, testItemID, gomock.Any()).Return(errors.New("db error"))
		body := `{"name":"updated"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockCore.EXPECT().UpdateItem(gomock.Any(), testUserID, testItemID, gomock.Any()).Return(nil)
		body := `{"name":"updated","quantity":10}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Manager)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), models.StatusUpdated)
	})

}

func TestHandler_DeleteItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid item ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/abc", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidItemID.Error())
	})

	t.Run("no userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().DeleteItem(gomock.Any(), testUserID, testItemID).Return(errors.New("db error"))
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockCore.EXPECT().DeleteItem(gomock.Any(), testUserID, testItemID).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+strconv.FormatInt(testItemID, 10), nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), models.StatusDeleted)
	})

}

func TestHandler_GetItemHistory(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockAuth := mocks.NewMockAuthService(controller)
	mockCore := mocks.NewMockCoreService(controller)

	router := setupRouter(mockAuth, mockCore)

	t.Run("invalid item ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/abc/history", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidItemID.Error())
	})

	t.Run("no userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10)+"/history", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("invalid query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10)+"/history?limit=abc", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Contains(t, resp.Body.String(), errs.ErrInvalidLimit.Error())
	})

	t.Run("service error", func(t *testing.T) {
		mockCore.EXPECT().GetItemHistory(gomock.Any(), testItemID, gomock.Any()).Return(nil, errors.New("db error"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10)+"/history", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success with JSON", func(t *testing.T) {
		history := []models.ItemHistory{
			{ID: 1, ItemID: testItemID, UserID: testUserID, Action: "INSERT", ChangedAt: time.Now()},
		}
		mockCore.EXPECT().GetItemHistory(gomock.Any(), testItemID, gomock.Any()).Return(history, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10)+"/history", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "INSERT")
	})

	t.Run("success with CSV export", func(t *testing.T) {
		history := []models.ItemHistory{
			{ID: 1, ItemID: testItemID, UserID: testUserID, Action: "UPDATE", ChangedAt: time.Now(), OldData: []byte(`{"name":"old"}`), NewData: []byte(`{"name":"new"}`)},
		}
		mockCore.EXPECT().GetItemHistory(gomock.Any(), testItemID, gomock.Any()).Return(history, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+strconv.FormatInt(testItemID, 10)+"/history?export=csv", nil)
		req.Header.Set("X-User-ID", strconv.FormatInt(testUserID, 10))
		req.Header.Set("X-Role", models.Admin)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Header().Get("Content-Type"), "text/csv")
		require.Contains(t, resp.Body.String(), "ItemID,UserID,Action")
	})

}

func TestGetRole(t *testing.T) {

	t.Run("success - role found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("role", models.Admin)

		role, err := getRole(c)
		require.NoError(t, err)
		require.Equal(t, models.Admin, role)
	})

	t.Run("error - role not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		role, err := getRole(c)
		require.Error(t, err)
		require.Equal(t, "", role)
		require.Equal(t, errs.ErrInvalidToken, err)
	})

}

func TestParseTime(t *testing.T) {

	t.Run("empty string", func(t *testing.T) {
		_, err := parseTime("")
		require.Error(t, err)
		require.Equal(t, errs.ErrMissingDate, err)
	})

	t.Run("RFC3339 format", func(t *testing.T) {
		timeStr := "2024-01-15T10:30:00Z"
		result, err := parseTime(timeStr)
		require.NoError(t, err)
		require.Equal(t, "2024-01-15 10:30:00 +0000 UTC", result.String())
	})

	t.Run("YYYY-MM-DDTHH:MM:SS format", func(t *testing.T) {
		timeStr := "2024-01-15T10:30:00"
		result, err := parseTime(timeStr)
		require.NoError(t, err)
		require.Equal(t, "2024-01-15 10:30:00 +0000 UTC", result.String())
	})

	t.Run("YYYY-MM-DD format", func(t *testing.T) {
		timeStr := "2024-01-15"
		result, err := parseTime(timeStr)
		require.NoError(t, err)
		require.Equal(t, "2024-01-15 00:00:00 +0000 UTC", result.String())
	})

	t.Run("with timezone offset", func(t *testing.T) {
		timeStr := "2024-01-15T10:30:00+03:00"
		result, err := parseTime(timeStr)
		require.NoError(t, err)
		require.Equal(t, "2024-01-15 07:30:00 +0000 UTC", result.String())
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := parseTime("invalid")
		require.Error(t, err)
		require.Equal(t, errs.ErrInvalidDate, err)
	})

}

func TestFmtRespond(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	t.Run("JSON response for non-CSV", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		data := map[string]string{"key": "value"}
		fmtRespond(c, data)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "value")
	})

	t.Run("CSV response with empty history", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test?export=csv", nil)

		fmtRespond(c, []models.ItemHistory{})

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		require.Contains(t, w.Body.String(), "ID,ItemID,UserID")
	})

	t.Run("CSV response with non-history data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test?export=csv", nil)

		fmtRespond(c, "not history")

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("CSV response with history data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test?export=csv", nil)

		history := []models.ItemHistory{
			{
				ID:        1,
				ItemID:    100,
				UserID:    200,
				Action:    "INSERT",
				ChangedAt: time.Now(),
			},
		}

		fmtRespond(c, history)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Disposition"), "item_100_history")
	})

}
