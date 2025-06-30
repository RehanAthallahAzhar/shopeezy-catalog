package models

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID   string `gorm:"type:uuid;not null" json:"product_id"`
	UserID      string `gorm:"type:uuid;not null" json:"user_id"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

type RedisCartItem struct {
	Quantity        int       `json:"quantity"`
	CartDescription string    `json:"cart_description "`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CartItem struct {
	ID              string    `json:"id"`
	SellerID        string    `json:"seller_id"`
	SellerName      string    `json:"seller_name"`
	Quantity        int       `json:"quantity"`
	ProductID       string    `json:"product_id"`
	ProductName     string    `json:"product_name"`
	ProductPrice    int       `json:"product_price"`
	CartDescription string    `json:"cart_description"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CartItemResponse struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	UserID      string `json:"user_id"`
	ProductName string `json:"product_name"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty"`
}

type CartRequest struct {
	ProductID   string `json:"product_id" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required,min=1"`
	Description string `json:"description"`
}

type UpdateCartRequest struct {
	Quantity    int    `json:"quantity" validate:"required"`
	Description string `json:"description"`
}
