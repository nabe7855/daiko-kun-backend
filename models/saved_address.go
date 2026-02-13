package models

import "time"

type SavedAddress struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Label      string    `json:"label"` // e.g. "自宅", "職場"
	Address     string    `json:"address"`
	Description string    `json:"description"`
	Latitude    *float64  `json:"latitude"`
	Longitude  *float64  `json:"longitude"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
