package v1

import (
	"github.com/shopspring/decimal"
)

type RegisterDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CreateItemDTO struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Quantity    int             `json:"quantity"`
	Price       decimal.Decimal `json:"price"`
}

type UpdateItemDTO struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Quantity    *int             `json:"quantity"`
	Price       *decimal.Decimal `json:"price"`
}
