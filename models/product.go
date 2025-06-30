package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          string `gorm:"type:uuid" json:"id"`
	SellerID    string `gorm:"type:uuid" json:"seller_id"`
	Name        string `gorm:"type:varchar(100)" json:"name"`
	Price       int    `json:"price" `
	Stock       int    `json:"stock"`
	Discount    int    `json:"discount"`
	Type        string `json:"type"`
	Description string `gorm:"type:text" json:"description"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

type ProductRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Price       int    `json:"price" validate:"required,gt=0"` // gt=0 berarti harus lebih besar dari 0
	Stock       int    `json:"stock" validate:"required,gte=0"`
	Discount    int    `json:"discount" validate:"gte=0,lte=100"` // gt=0,lte=100 (diskon 0-100%)
	Type        string `json:"type" validate:"required,alpha"`    // alpha berarti hanya huruf
	Description string `json:"description"`
}
type ProductResponse struct {
	ID          string `json:"id"`
	SellerID    string `json:"seller_id"`
	SellerName  string `json:"seller_name"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
	Discount    int    `json:"discount"`
	Type        string `json:"type"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ProductWithSeller struct {
	ID          string    `json:"id"`
	SellerID    string    `json:"seller_id"`
	SellerName  string    `json:"seller_name"` // hasil JOIN ke users.name
	Name        string    `json:"name"`
	Price       int       `json:"price"`
	Stock       int       `json:"stock"`
	Discount    int       `json:"discount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
