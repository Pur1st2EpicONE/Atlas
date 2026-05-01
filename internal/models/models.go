// Package models defines data structures and constants used across the application.
// It includes User, Item, Update, ItemHistory, and role/status constants.
package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Status constants represent operation results.
const (
	StatusDeleted = "deleted" // StatusDeleted indicates an item was successfully deleted
	StatusUpdated = "updated" // StatusUpdated indicates an item was successfully updated
)

// Role constants define user permission levels.
const (
	Admin   = "admin"   // Admin role has full access (including deletion and history)
	Manager = "manager" // Manager role can create and update items
	Viewer  = "viewer"  // Viewer role can only read items
)

// User represents an application user.
type User struct {
	ID       int64  // ID is the unique identifier of the user
	Login    string `json:"login"` // Login is the username
	Password string // Password is the hashed password (never serialized to JSON)
	Role     string `json:"role"` // Role is the user's permission level
}

// Item represents a product or inventory item.
type Item struct {
	ID          int64           `json:"id" db:"id"`                   // ID is the unique identifier
	Name        string          `json:"name" db:"name"`               // Name of the item
	Description string          `json:"description" db:"description"` // Description of the item
	Quantity    int             `json:"quantity" db:"quantity"`       // Quantity in stock
	Price       decimal.Decimal `json:"price" db:"price"`             // Price as a decimal (supports currency)
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`   // CreatedAt is the creation timestamp
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`   // UpdatedAt is the last modification timestamp
}

// Update represents a partial update to an item.
// All fields are pointers to distinguish between omission and zero values.
type Update struct {
	Name        *string          `json:"name,omitempty" binding:"omitempty,min=1,max=255"`   // Name (if provided) must be 1-255 chars
	Description *string          `json:"description,omitempty" binding:"omitempty,max=1000"` // Description max 1000 chars
	Quantity    *int             `json:"quantity,omitempty" binding:"omitempty,min=0"`       // Quantity non-negative
	Price       *decimal.Decimal `json:"price,omitempty" binding:"omitempty,min=0"`          // Price non-negative
}

// ItemHistory records a change to an item.
type ItemHistory struct {
	ID        int64           `json:"id"`                 // ID is the unique history record ID
	ItemID    int64           `json:"item_id"`            // ItemID references the changed item
	UserID    int64           `json:"user_id"`            // UserID is the user who performed the action
	Action    string          `json:"action"`             // Action describes the change (e.g., "CREATE", "UPDATE", "DELETE")
	ChangedAt time.Time       `json:"changed_at"`         // ChangedAt is the timestamp of the change
	OldData   json.RawMessage `json:"old_data,omitempty"` // OldData is the previous state (JSON)
	NewData   json.RawMessage `json:"new_data,omitempty"` // NewData is the new state (JSON)
}

// HistoryFilter defines query parameters for fetching item history.
type HistoryFilter struct {
	From   time.Time // From is the start time
	To     time.Time // To is the end time
	UserID int64     // UserID filters by the user who performed the action
	Action string    // Action filters by the type of action
	Limit  int       // Limit restricts the number of returned records
}
