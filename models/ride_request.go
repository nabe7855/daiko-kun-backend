package models

import "time"

type RideRequest struct {
	ID                 string    `json:"id"`
	CustomerID         string    `json:"customer_id"`
	DriverID           *string   `json:"driver_id"` // Nullable
	PickupLat          float64   `json:"pickup_lat"`
	PickupLng          float64   `json:"pickup_lng"`
	DestinationLat     float64   `json:"destination_lat"`
	DestinationLng     float64   `json:"destination_lng"`
	PickupAddress      string    `json:"pickup_address"`
	DestinationAddress string    `json:"destination_address"`
	EstimatedFare      float64   `json:"estimated_fare"`
	Status             string    `json:"status"`
	ActualFare         *float64  `json:"actual_fare"`
	PaymentMethod      *string   `json:"payment_method"`
	RatingToDriver     *int      `json:"rating_to_driver"`
	RatingToCustomer   *int      `json:"rating_to_customer"`
	ReviewComment      *string   `json:"review_comment"`
	ScheduledAt        *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	
	// Driver info (joined)
	DriverName       *string  `json:"driver_name,omitempty"`
	DriverPhone      *string  `json:"driver_phone,omitempty"`
	LicenseNumber    *string  `json:"license_number,omitempty"`
	DriverCurrentLat *float64 `json:"driver_current_lat,omitempty"`
	DriverCurrentLng *float64 `json:"driver_current_lng,omitempty"`
	DriverAverageRating *float64 `json:"driver_average_rating,omitempty"`
	DriverRatingCount   *int     `json:"driver_rating_count,omitempty"`
	CustomerName        *string  `json:"customer_name,omitempty"`
}
