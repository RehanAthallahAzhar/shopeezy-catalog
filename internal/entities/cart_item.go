package entities

import (
	"github.com/google/uuid"
)

type CartItem struct {
	ProductID       uuid.UUID
	ProductName     string
	ProductImageURL string // soon
	Price           float64
	Stock           int
	SellerID        uuid.UUID
	SellerName      string
	Quantity        int
	Description     string
	Checked         bool
}

type Cart struct {
	UserID     uuid.UUID
	Items      []CartItem
	TotalItems int
}
