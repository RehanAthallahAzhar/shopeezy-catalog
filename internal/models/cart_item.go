package models

import (
	"time"

	"github.com/google/uuid"
)

type RedisCartItem struct {
	Quantity    int       `json:"quantity"`
	Description string    `json:"description,omitempty"`
	Checked     bool      `json:"checked"`
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type CartItem struct {
	ProductID       uuid.UUID `json:"product_id"`
	SellerID        uuid.UUID `json:"seller_id"`
	SellerName      string    `json:"seller_name"`
	Quantity        int       `json:"quantity"`
	ProductName     string    `json:"product_name"`
	ProductPrice    int       `json:"product_price"`
	CartDescription string    `json:"cart_description"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CartItemResponse struct {
	SellerName   string  `json:"seller_name"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	Description  string  `json:"description"`
	Checked      bool    `json:"checked"`
}

type CartResponse struct {
	UserID     string             `json:"user_id"`
	TotalItems int                `json:"total_items"`
	Items      []CartItemResponse `json:"items"`
}

type CartRequest struct {
	Quantity    int    `json:"quantity" validate:"required,min=1"`
	Description string `json:"description"`
}

type UpdateCartRequest struct {
	Quantity    int    `json:"quantity" validate:"required"`
	Description string `json:"description"`
}
