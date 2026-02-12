package models

import (
	"time"
)

type Customer struct {
	ID             string    `json:"id"`
	Name           *string   `json:"name"`
	PhoneNumber    string    `json:"phone_number" binding:"required"`
	Email          *string   `json:"email"`
	SocialID       *string   `json:"social_id"`
	SocialProvider *string   `json:"social_provider"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
