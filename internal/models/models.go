package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatusDeleted = "deleted" // deleted
	StatusUpdated = "updated" // updated
)

const (
	Admin   = "admin"   // admin
	Manager = "manager" // manager
	Viewer  = "viewer"  // viewer
)

type User struct {
	ID       int64
	Login    string `json:"login"`
	Password string
	Role     string `json:"role"`
}

type Item struct {
	ID          int64           `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Quantity    int             `json:"quantity" db:"quantity"`
	Price       decimal.Decimal `json:"price" db:"price"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type Update struct {
	Name        *string          `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string          `json:"description,omitempty" binding:"omitempty,max=1000"`
	Quantity    *int             `json:"quantity,omitempty" binding:"omitempty,min=0"`
	Price       *decimal.Decimal `json:"price,omitempty" binding:"omitempty,min=0"`
}

type ItemHistory struct {
	ID        int64           `json:"id"`
	ItemID    int64           `json:"item_id"`
	UserID    int64           `json:"user_id"`
	Action    string          `json:"action"`
	ChangedAt time.Time       `json:"changed_at"`
	OldData   json.RawMessage `json:"old_data,omitempty"`
	NewData   json.RawMessage `json:"new_data,omitempty"`
}
