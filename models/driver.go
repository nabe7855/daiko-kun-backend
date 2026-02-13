package models

import (
	"time"
)

type Driver struct {
	ID            string    `json:"id"`
	Name          string    `json:"name" binding:"required"`
	PhoneNumber   string    `json:"phone_number" binding:"required"`
	LicenseNumber string    `json:"license_number"`
	CompanyID     *string   `json:"company_id"`
	Status        string    `json:"status"`
	CurrentLat    *float64  `json:"current_lat,omitempty"`
	CurrentLng    *float64  `json:"current_lng,omitempty"`
	AverageRating float64   `json:"average_rating"`
	RatingCount   int       `json:"rating_count"`
	FCMToken      *string   `json:"fcm_token"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
