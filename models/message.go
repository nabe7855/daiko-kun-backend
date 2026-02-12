package models

import "time"

type Message struct {
	ID         string    `json:"id"`
	RideID     string    `json:"ride_id"`
	SenderID   string    `json:"sender_id"`
	SenderType string    `json:"sender_type"` // 'customer' or 'driver'
	Content    string    `json:"content"`
	ImageURL   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type EmergencyReport struct {
	ID           string    `json:"id"`
	RideID       string    `json:"ride_id"`
	ReporterID   string    `json:"reporter_id"`
	ReporterType string    `json:"reporter_type"` // 'customer' or 'driver'
	Reason       string    `json:"reason"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
