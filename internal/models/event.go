package models

import "time"

type OrderCreatedEvent struct {
	OrderID     string         `json:"order_id"`
	UserID      string         `json:"user_id"`
	TotalAmount int            `json:"total_amount"`
	OrderDate   time.Time      `json:"order_date"`
	ProductIDs  []string       `json:"product_ids"`
	Quantities  map[string]int `json:"quantities"`
}
