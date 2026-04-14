package models

import "time"

type UserReport struct {
	ID             string    `json:"id"`
	RideID         string    `json:"ride_id"`
	ReporterID     string    `json:"reporter_id"`
	ReportedUserID string    `json:"reported_user_id"`
	ReporterRole   string    `json:"reporter_role"` // 'customer' or 'driver'
	Reason         string    `json:"reason"`
	IsBlocking     bool      `json:"is_blocking"`
	Status         string    `json:"status"` // 'pending', 'reviewed', 'action_taken'
	CreatedAt      time.Time `json:"created_at"`
}
