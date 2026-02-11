package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRideRequest handles POST /customer/ride-requests
func CreateRideRequest(c *gin.Context) {
	var req models.RideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.Status = "pending"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	sqlStatement := `
		INSERT INTO ride_requests (id, customer_id, pickup_lat, pickup_lng, destination_lat, destination_lng, pickup_address, destination_address, estimated_fare, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := db.DB.Exec(sqlStatement,
		req.ID, req.CustomerID, req.PickupLat, req.PickupLng, req.DestinationLat, req.DestinationLng,
		req.PickupAddress, req.DestinationAddress, req.EstimatedFare, req.Status, req.CreatedAt, req.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ride request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// ListAllRideRequests handles GET /admin/ride-requests
func ListAllRideRequests(c *gin.Context) {
	// Admin用だが、ユーザーアプリもこれを使ってポーリングしているため、ドライバー情報もJOINして返す
	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.created_at, r.updated_at,
			r.actual_fare, r.payment_method, r.rating_to_driver, r.rating_to_customer, r.review_comment,
			d.name, d.phone_number, d.license_number,
			d.current_lat, d.current_lng
		FROM ride_requests r
		LEFT JOIN drivers d ON r.driver_id = d.id
		ORDER BY r.created_at DESC
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var requests []models.RideRequest
	for rows.Next() {
		var r models.RideRequest
		err := rows.Scan(
			&r.ID, &r.CustomerID, &r.DriverID, &r.PickupLat, &r.PickupLng, 
			&r.DestinationLat, &r.DestinationLng, &r.PickupAddress, 
			&r.DestinationAddress, &r.EstimatedFare, &r.Status, 
			&r.CreatedAt, &r.UpdatedAt,
			&r.ActualFare, &r.PaymentMethod, &r.RatingToDriver, &r.RatingToCustomer, &r.ReviewComment,
			&r.DriverName, &r.DriverPhone, &r.LicenseNumber,
			&r.DriverCurrentLat, &r.DriverCurrentLng,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row: " + err.Error()})
			return
		}
		requests = append(requests, r)
	}

	if requests == nil {
		requests = []models.RideRequest{}
	}
	c.JSON(http.StatusOK, requests)
}

// GetAvailableRideRequests handles GET /driver/available-requests
func GetAvailableRideRequests(c *gin.Context) {
	// Statusがpending（まだドライバーが決まっていない）の依頼を取得
	rows, err := db.DB.Query("SELECT id, customer_id, pickup_lat, pickup_lng, destination_lat, destination_lng, pickup_address, destination_address, estimated_fare, status, created_at, updated_at FROM ride_requests WHERE status = 'pending' ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var requests []models.RideRequest
	for rows.Next() {
		var r models.RideRequest
		err := rows.Scan(&r.ID, &r.CustomerID, &r.PickupLat, &r.PickupLng, &r.DestinationLat, &r.DestinationLng, &r.PickupAddress, &r.DestinationAddress, &r.EstimatedFare, &r.Status, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		requests = append(requests, r)
	}

	if requests == nil {
		requests = []models.RideRequest{}
	}
	c.JSON(http.StatusOK, requests)
}

// AcceptRideRequest handles POST /driver/ride-requests/:id/accept
func AcceptRideRequest(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		DriverID string `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// トランザクションでステータスがpendingであることを確認してから、statusをacceptedに変更し、driver_idをセットする
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 依頼の状態を確認
	var currentStatus string
	err = tx.QueryRow("SELECT status FROM ride_requests WHERE id = $1 FOR UPDATE", id).Scan(&currentStatus)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride request not found"})
		return
	}

	if currentStatus != "pending" {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Ride request is no longer available"})
		return
	}

	// 更新
	_, err = tx.Exec("UPDATE ride_requests SET driver_id = $1, status = 'accepted', updated_at = $2 WHERE id = $3", body.DriverID, time.Now(), id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept ride request"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Ride request accepted successfully"})
}

// UpdateRideStatus handles PATCH /driver/ride-requests/:id/status
func UpdateRideStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status        string   `json:"status" binding:"required"`
		ActualFare    *float64 `json:"actual_fare"`
		PaymentMethod *string  `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ステータスのバリデーション
	allowedStatuses := map[string]bool{
		"accepted":  true,
		"arrived":   true,
		"started":   true,
		"completed": true,
		"cancelled": true,
	}
	if !allowedStatuses[body.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	updateFields := "status = $1, updated_at = $2"
	args := []interface{}{body.Status, time.Now()}
	
	if body.ActualFare != nil {
		args = append(args, *body.ActualFare)
		updateFields += fmt.Sprintf(", actual_fare = $%d", len(args))
	}
	if body.PaymentMethod != nil {
		args = append(args, *body.PaymentMethod)
		updateFields += fmt.Sprintf(", payment_method = $%d", len(args))
	}
	
	args = append(args, id)
	query := fmt.Sprintf("UPDATE ride_requests SET %s WHERE id = $%d", updateFields, len(args))

	_, err := db.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "status": body.Status})
}

// SubmitRating handles POST /customer/ride-requests/:id/rate
func SubmitRating(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Rating  int    `json:"rating" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE ride_requests SET rating_to_driver = $1, review_comment = $2, updated_at = $3 WHERE id = $4", 
		body.Rating, body.Comment, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating submitted successfully"})
}
// ListDriverHistory handles GET /drivers/:id/history
func ListDriverHistory(c *gin.Context) {
	driverID := c.Param("id")

	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.created_at, r.updated_at,
			r.actual_fare, r.payment_method, r.rating_to_driver, r.rating_to_customer, r.review_comment
		FROM ride_requests r
		WHERE r.driver_id = $1 AND r.status = 'completed'
		ORDER BY r.created_at DESC
	`
	rows, err := db.DB.Query(query, driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var requests []models.RideRequest
	for rows.Next() {
		var r models.RideRequest
		err := rows.Scan(
			&r.ID, &r.CustomerID, &r.DriverID, &r.PickupLat, &r.PickupLng, 
			&r.DestinationLat, &r.DestinationLng, &r.PickupAddress, 
			&r.DestinationAddress, &r.EstimatedFare, &r.Status, 
			&r.CreatedAt, &r.UpdatedAt,
			&r.ActualFare, &r.PaymentMethod, &r.RatingToDriver, &r.RatingToCustomer, &r.ReviewComment,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row: " + err.Error()})
			return
		}
		requests = append(requests, r)
	}

	if requests == nil {
		requests = []models.RideRequest{}
	}
	c.JSON(http.StatusOK, requests)
}
