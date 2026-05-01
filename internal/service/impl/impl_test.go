package impl

import (
	"Atlas/internal/config"
	"Atlas/internal/errs"
	mockLogger "Atlas/internal/logger/mocks"
	"Atlas/internal/models"
	"Atlas/internal/repository/mocks"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func newTestAuthConfig() config.Auth {
	return config.Auth{
		TokenSignedString: "test-secret-aboba-string",
		TokenTTL:          time.Hour,
		MinLoginLength:    3,
		MaxLoginLength:    20,
		MinPasswordLength: 6,
	}
}

func newTestCoreConfig() config.Core {
	return config.Core{
		MinItemNameLength:        1,
		MaxItemNameLength:        100,
		MaxItemDescriptionLength: 500,
		MinItemQuantity:          0,
		MaxItemQuantity:          1000,
		MaxItemPrice:             1000000,
	}
}

func TestAuthService_CreateUser(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockAuthStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, mockStorage)

	t.Run("validation error", func(t *testing.T) {
		user := models.User{Login: "ab", Password: "qweqwe", Role: models.Viewer}
		_, err := service.CreateUser(context.Background(), user)
		require.ErrorIs(t, err, errs.ErrLoginTooShort)
	})

	t.Run("hash password error", func(t *testing.T) {
		user := models.User{Login: "validlogin", Password: string(make([]byte, 73)), Role: models.Viewer}
		_, err := service.CreateUser(context.Background(), user)
		require.ErrorIs(t, err, errs.ErrPasswordTooLong)
	})

	t.Run("storage duplicate error", func(t *testing.T) {
		user := models.User{Login: "existing", Password: "qweqwe", Role: models.Viewer}
		mockStorage.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(int64(0), &pq.Error{Code: "23505"})
		_, err := service.CreateUser(context.Background(), user)
		require.ErrorIs(t, err, errs.ErrUserAlreadyExists)
	})

	t.Run("storage other error", func(t *testing.T) {
		user := models.User{Login: "newuser", Password: "qweqwe", Role: models.Viewer}
		mockStorage.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("db down"))
		mockLogger.EXPECT().LogError("service — failed to create new user", errors.New("db down"), "layer", "service.impl")
		_, err := service.CreateUser(context.Background(), user)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		user := models.User{Login: "validuser", Password: "qweqwe", Role: models.Admin}
		mockStorage.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u models.User) (int64, error) {
			require.Len(t, u.Password, 60, "password should be hashed")
			return 42, nil
		})
		id, err := service.CreateUser(context.Background(), user)
		require.NoError(t, err)
		require.Equal(t, int64(42), id)
	})

}

func TestAuthService_CreateToken(t *testing.T) {

	config := newTestAuthConfig()
	service := &AuthService{config: config}
	user := models.User{ID: 123, Role: models.Manager}

	token, err := service.CreateToken(user)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, service.KeyFunc)
	require.NoError(t, err)

	claims, ok := parsed.Claims.(*Claims)
	require.True(t, ok)
	require.Equal(t, "123", claims.Subject)
	require.Equal(t, models.Manager, claims.Role)

}

func TestAuthService_GetUser(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockAuthStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, mockStorage)

	t.Run("validation error", func(t *testing.T) {
		_, err := service.GetUser(context.Background(), models.User{Login: "", Password: "pass"})
		require.ErrorIs(t, err, errs.ErrEmptyLogin)
	})

	t.Run("user not found", func(t *testing.T) {
		mockStorage.EXPECT().GetUserByLogin(gomock.Any(), "unknown").Return(models.User{}, sql.ErrNoRows)
		_, err := service.GetUser(context.Background(), models.User{Login: "unknown", Password: "pass"})
		require.ErrorIs(t, err, errs.ErrInvalidCredentials)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage.EXPECT().GetUserByLogin(gomock.Any(), "error").Return(models.User{}, errors.New("db error"))
		mockLogger.EXPECT().LogError("service — failed to get userID by login", errors.New("db error"), "layer", "service.impl")
		_, err := service.GetUser(context.Background(), models.User{Login: "error", Password: "pass"})
		require.Error(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		mockStorage.EXPECT().GetUserByLogin(gomock.Any(), "john").Return(models.User{Login: "john", Password: string(hash)}, nil)
		_, err := service.GetUser(context.Background(), models.User{Login: "john", Password: "wrong"})
		require.ErrorIs(t, err, errs.ErrInvalidCredentials)
	})

	t.Run("success", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		expectedUser := models.User{ID: 10, Login: "john", Password: string(hash), Role: models.Viewer}
		mockStorage.EXPECT().GetUserByLogin(gomock.Any(), "john").Return(expectedUser, nil)
		user, err := service.GetUser(context.Background(), models.User{Login: "john", Password: "secret"})
		require.NoError(t, err)
		require.Equal(t, expectedUser.ID, user.ID)
		require.Equal(t, expectedUser.Role, user.Role)
	})

}

func TestAuthService_ParseToken(t *testing.T) {

	config := newTestAuthConfig()
	service := &AuthService{config: config}

	t.Run("invalid token", func(t *testing.T) {
		_, err := service.ParseToken("invalid")
		require.ErrorIs(t, err, errs.ErrInvalidToken)
	})

	t.Run("valid token", func(t *testing.T) {
		tokenStr, _ := service.CreateToken(models.User{ID: 99, Role: models.Admin})
		userID, err := service.ParseToken(tokenStr)
		require.NoError(t, err)
		require.Equal(t, int64(99), userID)
	})

}

func TestCoreService_CreateItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	config := newTestCoreConfig()
	service := NewCoreService(mockLogger, config, mockStorage)
	validItem := models.Item{Name: "item", Description: "desc", Quantity: 5, Price: decimal.NewFromInt(100)}

	t.Run("validation error", func(t *testing.T) {
		item := models.Item{Name: "", Quantity: 5, Price: decimal.NewFromInt(100)}
		_, err := service.CreateItem(context.Background(), 1, item)
		require.ErrorIs(t, err, errs.ErrMissingItemName)
	})

	t.Run("transaction error", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).Return(errors.New("tx failed"))
		mockLogger.EXPECT().LogError("service — failed to save item in storage", errors.New("tx failed"), "layer", "service.impl")
		_, err := service.CreateItem(context.Background(), 1, validItem)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(*sql.Tx, context.Context) error) error {
			return nil
		})
		_, err := service.CreateItem(context.Background(), 1, validItem)
		require.NoError(t, err)
	})

}

func TestCoreService_UpdateItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	config := newTestCoreConfig()
	service := NewCoreService(mockLogger, config, mockStorage)

	update := models.Update{
		Name:        strPtr("new name"),
		Description: strPtr("new desc"),
		Quantity:    intPtr(10),
		Price:       decimalPtr(decimal.NewFromInt(200)),
	}

	t.Run("item not found during update", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(*sql.Tx, context.Context) error) error {
			return errs.ErrItemNotFound
		})
		err := service.UpdateItem(context.Background(), 1, 100, update)
		require.ErrorIs(t, err, errs.ErrItemNotFound)
	})

	t.Run("unexpected error", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).Return(errors.New("db crash"))
		mockLogger.EXPECT().LogError("service — failed to update item in storage", errors.New("db crash"), "layer", "service.impl")
		err := service.UpdateItem(context.Background(), 1, 100, update)
		require.Error(t, err)
	})

}

func TestCoreService_DeleteItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	service := &CoreService{logger: mockLogger, storage: mockStorage}

	t.Run("item not found", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
		err := service.DeleteItem(context.Background(), 1, 100)
		require.ErrorIs(t, err, errs.ErrItemNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		mockStorage.EXPECT().Transaction(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
		mockLogger.EXPECT().LogError("service — failed to delete item from storage", errors.New("some error"), "itemID", int64(100), "layer", "service.impl")
		err := service.DeleteItem(context.Background(), 1, 100)
		require.Error(t, err)
	})

}

func TestCoreService_GetItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	service := &CoreService{logger: mockLogger, storage: mockStorage}

	t.Run("not found", func(t *testing.T) {
		mockStorage.EXPECT().GetItem(gomock.Any(), int64(1)).Return(models.Item{}, sql.ErrNoRows)
		_, err := service.GetItem(context.Background(), 1)
		require.ErrorIs(t, err, errs.ErrItemNotFound)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage.EXPECT().GetItem(gomock.Any(), int64(2)).Return(models.Item{}, errors.New("db error"))
		mockLogger.EXPECT().LogError("service — failed to get item from storage", errors.New("db error"), "itemID", int64(2), "layer", "service.impl")
		_, err := service.GetItem(context.Background(), 2)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		expected := models.Item{ID: 3, Name: "test"}
		mockStorage.EXPECT().GetItem(gomock.Any(), int64(3)).Return(expected, nil)
		item, err := service.GetItem(context.Background(), 3)
		require.NoError(t, err)
		require.Equal(t, expected, item)
	})

}

func TestCoreService_GetItems(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	service := &CoreService{logger: mockLogger, storage: mockStorage}

	t.Run("storage error", func(t *testing.T) {
		mockStorage.EXPECT().GetItems(gomock.Any()).Return(nil, errors.New("db error"))
		mockLogger.EXPECT().LogError("service — failed to get items from storage", errors.New("db error"), "layer", "service.impl")
		_, err := service.GetItems(context.Background())
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		items := []models.Item{{ID: 1}, {ID: 2}}
		mockStorage.EXPECT().GetItems(gomock.Any()).Return(items, nil)
		result, err := service.GetItems(context.Background())
		require.NoError(t, err)
		require.Equal(t, items, result)
	})

}

func TestCoreService_GetItemHistory(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStorage := mocks.NewMockCoreStorage(controller)
	mockLogger := mockLogger.NewMockLogger(controller)

	service := &CoreService{logger: mockLogger, storage: mockStorage}

	t.Run("invalid filter", func(t *testing.T) {
		from := time.Now()
		to := from.Add(-time.Hour)
		filter := models.HistoryFilter{From: from, To: to}
		_, err := service.GetItemHistory(context.Background(), 1, filter)
		require.ErrorIs(t, err, errs.ErrInvalidDateRange)
	})

	t.Run("storage error", func(t *testing.T) {
		filter := models.HistoryFilter{Limit: 100}
		mockStorage.EXPECT().GetItemHistory(gomock.Any(), int64(1), filter).Return(nil, errors.New("db error"))
		mockLogger.EXPECT().LogError("service — failed to get item history from storage", errors.New("db error"), "layer", "service.impl")
		_, err := service.GetItemHistory(context.Background(), 1, filter)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		history := []models.ItemHistory{{ID: 1}}
		filter := models.HistoryFilter{Limit: 100}
		mockStorage.EXPECT().GetItemHistory(gomock.Any(), int64(2), filter).Return(history, nil)
		result, err := service.GetItemHistory(context.Background(), 2, filter)
		require.NoError(t, err)
		require.Equal(t, history, result)
	})

}

func TestValidateName(t *testing.T) {

	config := newTestCoreConfig()
	service := &CoreService{config: config}

	tests := []struct {
		name string
		err  error
	}{
		{"", errs.ErrMissingItemName},
		{"a", nil},
		{strings.Repeat("a", 100), nil},
		{strings.Repeat("a", 101), errs.ErrItemNameTooLong},
	}
	for _, tt := range tests {
		err := service.validateName(tt.name)
		require.Equal(t, tt.err, err)
	}

}

func TestValidateDescription(t *testing.T) {
	config := newTestCoreConfig()
	service := &CoreService{config: config}
	err := service.validateDescription(strings.Repeat("a", 500))
	require.NoError(t, err)
	err = service.validateDescription(strings.Repeat("a", 501))
	require.ErrorIs(t, err, errs.ErrItemDescriptionTooLong)
}

func TestValidateQuantity(t *testing.T) {

	config := newTestCoreConfig()
	service := &CoreService{config: config}

	tests := []struct {
		q   int
		err error
	}{
		{-1, errs.ErrItemQuantityTooLow},
		{0, nil},
		{1000, nil},
		{1001, errs.ErrItemQuantityTooHigh},
	}
	for _, tt := range tests {
		err := service.validateQuantity(tt.q)
		require.Equal(t, tt.err, err)
	}

}

func TestValidatePrice(t *testing.T) {

	config := newTestCoreConfig()
	service := &CoreService{config: config}

	tests := []struct {
		price decimal.Decimal
		err   error
	}{
		{decimal.NewFromInt(-1), errs.ErrNegativeItemPrice},
		{decimal.NewFromInt(0), errs.ErrItemZeroPrice},
		{decimal.NewFromInt(1000000), nil},
		{decimal.NewFromInt(1000001), errs.ErrItemPriceTooLarge},
		{decimal.NewFromFloat(1.234), errs.ErrItemPriceInvalidPrecision},
		{decimal.NewFromFloat(1.23), nil},
	}
	for _, tt := range tests {
		err := service.validatePrice(tt.price)
		require.Equal(t, tt.err, err)
	}

}

func TestValidateFilter(t *testing.T) {

	service := &CoreService{}
	from := time.Now()
	to := from.Add(time.Hour)
	filter := models.HistoryFilter{From: from, To: to}

	err := service.validateFilter(filter)
	require.NoError(t, err)

	to = from.Add(-time.Hour)
	filter.To = to

	err = service.validateFilter(filter)
	require.ErrorIs(t, err, errs.ErrInvalidDateRange)

}

func TestUpdateFields(t *testing.T) {
	item := models.Item{Name: "old", Description: "old", Quantity: 1, Price: decimal.NewFromInt(10)}
	update := models.Update{
		Name:        strPtr("new"),
		Description: strPtr("new desc"),
		Quantity:    intPtr(5),
		Price:       decimalPtr(decimal.NewFromInt(20)),
	}
	updateFields(&item, update)
	require.Equal(t, "new", item.Name)
	require.Equal(t, "new desc", item.Description)
	require.Equal(t, 5, item.Quantity)
	require.Equal(t, decimal.NewFromInt(20), item.Price)
	item = models.Item{Name: "old", Quantity: 1}
	update = models.Update{Name: strPtr("only name")}
	updateFields(&item, update)
	require.Equal(t, "only name", item.Name)
	require.Equal(t, 1, item.Quantity)
}

func TestUnexpectedError(t *testing.T) {
	require.False(t, unexpectedError(nil))
	require.False(t, unexpectedError(errs.ErrMissingItemName))
	require.False(t, unexpectedError(errs.ErrItemNotFound))
	require.True(t, unexpectedError(errors.New("some other error")))
}

func strPtr(s string) *string                       { return &s }
func intPtr(i int) *int                             { return &i }
func decimalPtr(d decimal.Decimal) *decimal.Decimal { return &d }

func TestAuthService_validateRole(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, nil)

	t.Run("empty role", func(t *testing.T) {
		err := service.validateRole("")
		require.ErrorIs(t, err, errs.ErrEmptyRole)
	})

	t.Run("invalid role", func(t *testing.T) {
		err := service.validateRole("superuser")
		require.ErrorIs(t, err, errs.ErrInvalidRole)
	})

	t.Run("valid roles", func(t *testing.T) {
		require.NoError(t, service.validateRole(models.Admin))
		require.NoError(t, service.validateRole(models.Manager))
		require.NoError(t, service.validateRole(models.Viewer))
	})

}

func TestAuthService_validateLogin(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, nil)

	t.Run("empty login", func(t *testing.T) {
		err := service.validateLogin("")
		require.ErrorIs(t, err, errs.ErrEmptyLogin)
	})

	t.Run("login too short", func(t *testing.T) {
		err := service.validateLogin("ab")
		require.ErrorIs(t, err, errs.ErrLoginTooShort)
	})

	t.Run("login too long", func(t *testing.T) {
		longLogin := strings.Repeat("a", config.MaxLoginLength+1)
		err := service.validateLogin(longLogin)
		require.ErrorIs(t, err, errs.ErrLoginTooLong)
	})

	t.Run("valid login", func(t *testing.T) {
		err := service.validateLogin("valid")
		require.NoError(t, err)
	})

}

func TestAuthService_validatePassword(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, nil)

	t.Run("empty password", func(t *testing.T) {
		err := service.validatePassword("")
		require.ErrorIs(t, err, errs.ErrEmptyPassword)
	})

	t.Run("password too short", func(t *testing.T) {
		err := service.validatePassword("pass")
		require.ErrorIs(t, err, errs.ErrPasswordTooShort)
	})

	t.Run("password too long", func(t *testing.T) {
		longPass := strings.Repeat("a", 73)
		err := service.validatePassword(longPass)
		require.ErrorIs(t, err, errs.ErrPasswordTooLong)
	})

	t.Run("valid password", func(t *testing.T) {
		err := service.validatePassword("qweqwe")
		require.NoError(t, err)
	})

}

func TestAuthService_validateUser(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	config := newTestAuthConfig()
	service := NewAuthService(mockLogger, config, nil)

	t.Run("empty login", func(t *testing.T) {
		err := service.validateUser(models.User{Login: "", Password: "pass"})
		require.ErrorIs(t, err, errs.ErrEmptyLogin)
	})

	t.Run("empty password", func(t *testing.T) {
		err := service.validateUser(models.User{Login: "user", Password: ""})
		require.ErrorIs(t, err, errs.ErrEmptyPassword)
	})

	t.Run("valid user", func(t *testing.T) {
		err := service.validateUser(models.User{Login: "user", Password: "pass"})
		require.NoError(t, err)
	})

}

func TestCoreService_validateName_TooShort(t *testing.T) {
	config := config.Core{MinItemNameLength: 3, MaxItemNameLength: 10}
	service := &CoreService{config: config}
	err := service.validateName("ab")
	require.ErrorIs(t, err, errs.ErrItemNameTooShort)
}

func TestCoreService_validateItem(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	config := newTestCoreConfig()
	service := NewCoreService(mockLogger, config, nil)

	t.Run("description too long", func(t *testing.T) {
		item := models.Item{
			Name:        "name",
			Description: strings.Repeat("a", config.MaxItemDescriptionLength+1),
			Quantity:    5,
			Price:       decimal.NewFromInt(100),
		}
		err := service.validateItem(item)
		require.ErrorIs(t, err, errs.ErrItemDescriptionTooLong)
	})

	t.Run("quantity too low", func(t *testing.T) {
		item := models.Item{
			Name:        "name",
			Description: "desc",
			Quantity:    -1,
			Price:       decimal.NewFromInt(100),
		}
		err := service.validateItem(item)
		require.ErrorIs(t, err, errs.ErrItemQuantityTooLow)
	})

	t.Run("quantity too high", func(t *testing.T) {
		item := models.Item{
			Name:        "name",
			Description: "desc",
			Quantity:    config.MaxItemQuantity + 1,
			Price:       decimal.NewFromInt(100),
		}
		err := service.validateItem(item)
		require.ErrorIs(t, err, errs.ErrItemQuantityTooHigh)
	})

	t.Run("price invalid (negative)", func(t *testing.T) {
		item := models.Item{
			Name:        "name",
			Description: "desc",
			Quantity:    5,
			Price:       decimal.NewFromInt(-1),
		}
		err := service.validateItem(item)
		require.ErrorIs(t, err, errs.ErrNegativeItemPrice)
	})

}
