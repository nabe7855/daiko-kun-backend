package models

import "time"

type Company struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	PhoneNumber    string    `json:"phone_number"`
	Email          string    `json:"email"`
	Status         string    `json:"status"` // pending, active, suspended
	CommissionRate float64   `json:"commission_rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AdminUser struct {
	ID           string    `json:"id"`
	CompanyID    *string   `json:"company_id"` // NULL for Super Admin
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never export password
	Role         string    `json:"role"` // super_admin, company_admin
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
