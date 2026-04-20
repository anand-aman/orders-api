package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderId     uint64     `json:"orderid"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	LineItems   []LineItem `json:"line_items"`
	CreatedAt   *time.Time `json:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type LineItem struct {
	ItemId   uuid.UUID `json:"itemid"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}
