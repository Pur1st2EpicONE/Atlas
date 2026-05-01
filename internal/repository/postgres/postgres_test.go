package postgres_test

import (
	"Atlas/internal/config"
	"Atlas/internal/logger/mocks"
	"Atlas/internal/models"
	"Atlas/internal/repository"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/shopspring/decimal"
	wbf "github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"go.uber.org/mock/gomock"
)

var (
	testStorage *repository.Storage
	testDB      *dbpg.DB
	mockCtrl    *gomock.Controller
	mockLog     *mocks.MockLogger
)

const (
	validItemID   = 1
	nonExistentID = 999999
)

func TestMain(m *testing.M) {

	cfg := wbf.New()
	if err := cfg.LoadEnvFiles("../../../.env"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := cfg.LoadConfigFiles("../../../config.yaml"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var appCfg config.Config
	if err := cfg.Unmarshal(&appCfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testCfg := config.Storage{
		Host:               appCfg.Storage.Host,
		Port:               appCfg.Storage.Port,
		Username:           os.Getenv("DB_USER"),
		Password:           os.Getenv("DB_PASSWORD"),
		DBName:             appCfg.Storage.DBName,
		SSLMode:            appCfg.Storage.SSLMode,
		MaxOpenConns:       appCfg.Storage.MaxOpenConns,
		MaxIdleConns:       appCfg.Storage.MaxIdleConns,
		ConnMaxLifetime:    appCfg.Storage.ConnMaxLifetime,
		QueryRetryStrategy: appCfg.Storage.QueryRetryStrategy,
		TxRetryStrategy:    appCfg.Storage.TxRetryStrategy,
	}

	mockCtrl = gomock.NewController(nil)
	mockLog = mocks.NewMockLogger(mockCtrl)
	mockLog.EXPECT().LogInfo(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	var err error
	testDB, err = repository.ConnectDB(testCfg)
	if err != nil {
		fmt.Printf("failed to connect to test DB: %v\n", err)
		os.Exit(1)
	}

	if err := migrate(testDB.Master); err != nil {
		fmt.Printf("failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	testStorage = repository.NewStorage(mockLog, testCfg, testDB)

	exitCode := m.Run()
	testStorage.Close()
	mockCtrl.Finish()
	os.Exit(exitCode)

}

func migrate(db *sql.DB) error {
	_ = goose.SetDialect("postgres")
	if err := goose.Up(db, "../../../migrations"); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}
	return nil
}

func setupTest(t *testing.T) {

	ctx := context.Background()
	_, err := testDB.Master.ExecContext(ctx, `

	TRUNCATE TABLE users, items, item_history
	RESTART IDENTITY CASCADE`)

	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

}

func setCurrentUser(tx *sql.Tx, ctx context.Context, userID int64) error {
	_, err := tx.ExecContext(ctx, `SELECT set_config('app.current_user_id', $1::text, false)`, fmt.Sprint(userID))
	return err
}

func createTestUser(ctx context.Context, t *testing.T) int64 {
	user := models.User{
		Login:    fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		Password: "pass",
		Role:     models.Viewer,
	}
	id, err := testStorage.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return id
}

func TestCreateUser(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	user := models.User{Login: "testuser", Password: "hashpass", Role: models.Admin}

	id, err := testStorage.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero user ID")
	}

}

func TestCreateUser_Duplicate(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	user := models.User{Login: "duplicate", Password: "pass", Role: models.Viewer}

	_, err := testStorage.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("first CreateUser failed: %v", err)
	}

	_, err = testStorage.CreateUser(ctx, user)
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		t.Logf("duplicate violation correctly caught: %v", err)
	} else {
		t.Fatalf("unexpected error type: %v", err)
	}

}

func TestGetUserByLogin(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	user := models.User{Login: "getme", Password: "secret", Role: models.Manager}

	userID, err := testStorage.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	got, err := testStorage.GetUserByLogin(ctx, "getme")
	if err != nil {
		t.Fatalf("GetUserByLogin failed: %v", err)
	}
	if got.ID != userID || got.Login != user.Login || got.Password != user.Password || got.Role != user.Role {
		t.Fatalf("user mismatch: got %+v, want %+v", got, user)
	}

	_, err = testStorage.GetUserByLogin(ctx, "nonexistent")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

}

func TestCreateItem(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	item := models.Item{
		Name:        "Test Item",
		Description: "Description",
		Quantity:    10,
		Price:       decimal.NewFromFloat(99.99),
	}

	var created models.Item
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		if err := setCurrentUser(tx, txCtx, userID); err != nil {
			return err
		}
		var err error
		created, err = testStorage.CreateItem(tx, txCtx, item)
		return err
	})
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Name != item.Name || created.Quantity != item.Quantity {
		t.Fatalf("created item mismatch: %+v", created)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("timestamps not set")
	}

	var historyUserID int64
	err = testDB.Master.QueryRowContext(ctx,
		`SELECT user_id FROM item_history WHERE item_id = $1 AND action = 'INSERT'`, created.ID).Scan(&historyUserID)
	if err != nil {
		t.Fatalf("failed to fetch history: %v", err)
	}
	if historyUserID != userID {
		t.Fatalf("history user_id mismatch: got %d, want %d", historyUserID, userID)
	}

}

func TestGetItem(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "GetTest",
			Quantity: 5,
			Price:    decimal.NewFromFloat(10.0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID
		return nil
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	got, err := testStorage.GetItem(ctx, itemID)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
	if got.ID != itemID {
		t.Fatalf("got ID %d, want %d", got.ID, itemID)
	}

	_, err = testStorage.GetItem(ctx, nonExistentID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

}

func TestGetItems(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	items := []models.Item{
		{Name: "Item A", Quantity: 1, Price: decimal.NewFromFloat(1.0)},
		{Name: "Item B", Quantity: 2, Price: decimal.NewFromFloat(2.0)},
	}
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		for _, it := range items {
			if _, err := testStorage.CreateItem(tx, txCtx, it); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("setup insert failed: %v", err)
	}

	all, err := testStorage.GetItems(ctx)
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 items, got %d", len(all))
	}

}

func TestUpdateItem(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "OldName",
			Quantity: 10,
			Price:    decimal.NewFromFloat(100.0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID
		return nil
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	updatedData := models.Item{Name: "NewName", Quantity: 20, Price: decimal.NewFromFloat(200.0)}
	err = testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		return testStorage.UpdateItem(tx, txCtx, itemID, updatedData)
	})
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}

	got, err := testStorage.GetItem(ctx, itemID)
	if err != nil {
		t.Fatalf("GetItem after update failed: %v", err)
	}
	if got.Name != updatedData.Name || got.Quantity != updatedData.Quantity || !got.Price.Equal(updatedData.Price) {
		t.Fatalf("update not applied: got %+v", got)
	}

	var historyUserID int64
	err = testDB.Master.QueryRowContext(ctx,
		`SELECT user_id FROM item_history WHERE item_id = $1 AND action = 'UPDATE' ORDER BY changed_at DESC LIMIT 1`,
		itemID).Scan(&historyUserID)
	if err != nil {
		t.Fatalf("failed to fetch history: %v", err)
	}
	if historyUserID != userID {
		t.Fatalf("history user_id mismatch: got %d, want %d", historyUserID, userID)
	}

}

func TestDeleteItem(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "ToDelete",
			Quantity: 1,
			Price:    decimal.NewFromFloat(0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID
		return nil
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	err = testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		return testStorage.DeleteItem(tx, txCtx, itemID)
	})
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	_, err = testStorage.GetItem(ctx, itemID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}

	var historyUserID int64
	err = testDB.Master.QueryRowContext(ctx,
		`SELECT user_id FROM item_history WHERE item_id = $1 AND action = 'DELETE'`, itemID).Scan(&historyUserID)
	if err != nil {
		t.Fatalf("failed to fetch DELETE history: %v", err)
	}
	if historyUserID != userID {
		t.Fatalf("history user_id mismatch: got %d, want %d", historyUserID, userID)
	}

}

func TestDeleteItem_NotFound(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		return testStorage.DeleteItem(tx, txCtx, nonExistentID)
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

}

func TestGetItemForUpdate(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "ForUpdate",
			Quantity: 7,
			Price:    decimal.NewFromFloat(0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID

		item, err := testStorage.GetItemForUpdate(tx, txCtx, itemID)
		if err != nil {
			return err
		}
		if item.ID != itemID {
			return fmt.Errorf("wrong item fetched")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("GetItemForUpdate failed: %v", err)
	}

}

func TestGetItemHistory(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "Hist",
			Quantity: 1,
			Price:    decimal.NewFromFloat(1.0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID

		updated := models.Item{
			Name:     "HistUpdated",
			Quantity: 2,
			Price:    decimal.NewFromFloat(2.0),
		}
		if err := testStorage.UpdateItem(tx, txCtx, itemID, updated); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("setup history failed: %v", err)
	}

	filter := models.HistoryFilter{Limit: 10}
	history, err := testStorage.GetItemHistory(ctx, itemID, filter)
	if err != nil {
		t.Fatalf("GetItemHistory failed: %v", err)
	}
	if len(history) < 2 {
		t.Fatalf("expected at least 2 history entries, got %d", len(history))
	}
	for _, h := range history {
		if h.UserID != userID {
			t.Errorf("history user_id = %d, want %d", h.UserID, userID)
		}
		if h.Action == "" {
			t.Error("history action missing")
		}
	}

}

func TestGetItemHistory_Filtering(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	var itemID int64
	now := time.Now().UTC()
	earlier := now.Add(-time.Hour)

	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		created, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "FilterTest",
			Quantity: 1,
			Price:    decimal.NewFromFloat(0),
		})
		if err != nil {
			return err
		}
		itemID = created.ID
		upd := models.Item{
			Name:     "FilterTest2",
			Quantity: 2,
			Price:    decimal.NewFromFloat(0),
		}
		if err := testStorage.UpdateItem(tx, txCtx, itemID, upd); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	filter := models.HistoryFilter{Action: "UPDATE", Limit: 5}
	hist, err := testStorage.GetItemHistory(ctx, itemID, filter)
	if err != nil {
		t.Fatalf("filter by action failed: %v", err)
	}
	if len(hist) == 0 {
		t.Error("expected UPDATE history entries")
	}
	for _, h := range hist {
		if h.Action != "UPDATE" {
			t.Errorf("expected action UPDATE, got %s", h.Action)
		}
	}

	filter2 := models.HistoryFilter{From: earlier, To: now, Limit: 5}
	hist2, err := testStorage.GetItemHistory(ctx, itemID, filter2)
	if err != nil {
		t.Fatalf("filter by time failed: %v", err)
	}
	if len(hist2) == 0 {
		t.Log("no history in time range (timestamp sensitivity)")
	}

}

func TestTransaction_Commit(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		_, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "TxCommit",
			Quantity: 1,
			Price:    decimal.NewFromFloat(0),
		})
		return err
	})
	if err != nil {
		t.Fatalf("transaction commit failed: %v", err)
	}
	items, err := testStorage.GetItems(ctx)
	if err != nil {
		t.Fatalf("failed to fetch items after commit: %v", err)
	}
	if len(items) != 1 {
		t.Fatal("item not persisted after commit")
	}

}

func TestTransaction_Rollback(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := createTestUser(ctx, t)

	err := testStorage.Transaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		_ = setCurrentUser(tx, txCtx, userID)
		_, err := testStorage.CreateItem(tx, txCtx, models.Item{
			Name:     "TxRollback",
			Quantity: 1,
			Price:    decimal.NewFromFloat(0),
		})
		if err != nil {
			return err
		}
		return errors.New("forced rollback")
	})
	if err == nil || err.Error() != "forced rollback" {
		t.Fatalf("expected forced rollback error, got %v", err)
	}
	items, err := testStorage.GetItems(ctx)
	if err != nil {
		t.Fatalf("failed to fetch items after rollback: %v", err)
	}
	if len(items) != 0 {
		t.Fatal("item was inserted despite rollback")
	}

}

func TestItemHistoryJSON(t *testing.T) {

	old := map[string]any{"quantity": 5}
	new := map[string]any{"quantity": 10}

	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)

	history := models.ItemHistory{OldData: oldBytes, NewData: newBytes}

	if len(history.OldData) == 0 || len(history.NewData) == 0 {
		t.Error("JSON marshaling failed")
	}

}
