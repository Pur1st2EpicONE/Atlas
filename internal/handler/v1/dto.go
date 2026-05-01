// Package v1 provides data transfer objects (DTOs) for API version 1.
// These structs define the request/response shapes for authentication and item management.
package v1

import (
	"github.com/shopspring/decimal"
)

// RegisterDTO represents the request body for user registration.
type RegisterDTO struct {
	Login    string `json:"login"`    // Login is the desired username
	Password string `json:"password"` // Password is the user's plain-text password (will be hashed)
	Role     string `json:"role"`     // Role is the permission level (admin, manager, viewer)
}

// LoginDTO represents the request body for user authentication.
type LoginDTO struct {
	Login    string `json:"login"`    // Login is the username
	Password string `json:"password"` // Password is the plain-text password for verification
}

// CreateItemDTO represents the request body for creating a new item.
type CreateItemDTO struct {
	Name        string          `json:"name"`        // Name of the item
	Description string          `json:"description"` // Description of the item
	Quantity    int             `json:"quantity"`    // Quantity in stock
	Price       decimal.Decimal `json:"price"`       // Price as a decimal
}

// UpdateItemDTO represents the request body for partially updating an item.
// All fields are pointers to allow omitting fields (nil = no update).
type UpdateItemDTO struct {
	Name        *string          `json:"name"`        // New name (if provided)
	Description *string          `json:"description"` // New description (if provided)
	Quantity    *int             `json:"quantity"`    // New quantity (if provided)
	Price       *decimal.Decimal `json:"price"`       // New price (if provided)
}
