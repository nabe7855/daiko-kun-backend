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
		INSERT INTO ride_requests (id, customer_id, pickup_lat, pickup_lng, destination_lat, destination_lng, pickup_address, destination_address, estimated_fare, status, scheduled_at, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := db.DB.Exec(sqlStatement,
		req.ID, req.CustomerID, req.PickupLat, req.PickupLng, req.DestinationLat, req.DestinationLng,
		req.PickupAddress, req.DestinationAddress, req.EstimatedFare, req.Status, req.ScheduledAt, req.PaymentMethod, req.CreatedAt, req.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ride request: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// ListAllRideRequests handles GET /admin/ride-requests
func ListAllRideRequests(c *gin.Context) {
	companyID := c.Query("company_id")

	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.scheduled_at, r.created_at, r.updated_at,
			r.actual_fare, r.payment_method, r.rating_to_driver, r.rating_to_customer, r.review_comment,
			d.name, d.phone_number, d.license_number,
			d.current_lat, d.current_lng,
			d.average_rating, d.rating_count
		FROM ride_requests r
		LEFT JOIN drivers d ON r.driver_id = d.id
	`
	args := []interface{}{}
	if companyID != "" {
		query += " WHERE d.company_id = $1"
		args = append(args, companyID)
	}
	query += " ORDER BY r.created_at DESC"

	rows, err := db.DB.Query(query, args...)
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
			&r.ScheduledAt, &r.CreatedAt, &r.UpdatedAt,
			&r.ActualFare, &r.PaymentMethod, &r.RatingToDriver, &r.RatingToCustomer, &r.ReviewComment,
			&r.DriverName, &r.DriverPhone, &r.LicenseNumber,
			&r.DriverCurrentLat, &r.DriverCurrentLng,
			&r.DriverAverageRating, &r.DriverRatingCount,
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

// GetRideRequest handles GET /customer/ride-requests/:id
func GetRideRequest(c *gin.Context) {
	id := c.Param("id")

	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.scheduled_at, r.created_at, r.updated_at,
			r.actual_fare, r.payment_method, r.rating_to_driver, r.rating_to_customer, r.review_comment,
			d.name, d.phone_number, d.license_number,
			d.current_lat, d.current_lng,
			d.average_rating, d.rating_count
		FROM ride_requests r
		LEFT JOIN drivers d ON r.driver_id = d.id
		WHERE r.id = $1
	`

	var r models.RideRequest
	err := db.DB.QueryRow(query, id).Scan(
		&r.ID, &r.CustomerID, &r.DriverID, &r.PickupLat, &r.PickupLng, 
		&r.DestinationLat, &r.DestinationLng, &r.PickupAddress, 
		&r.DestinationAddress, &r.EstimatedFare, &r.Status, 
		&r.ScheduledAt, &r.CreatedAt, &r.UpdatedAt,
		&r.ActualFare, &r.PaymentMethod, &r.RatingToDriver, &r.RatingToCustomer, &r.ReviewComment,
		&r.DriverName, &r.DriverPhone, &r.LicenseNumber,
		&r.DriverCurrentLat, &r.DriverCurrentLng,
		&r.DriverAverageRating, &r.DriverRatingCount,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride request not found"})
		return
	}

	c.JSON(http.StatusOK, r)
}

// GetAvailableRideRequests handles GET /driver/available-requests
func GetAvailableRideRequests(c *gin.Context) {
	// Statusがpendingかつ予約ではない（scheduled_atがNULL）の依頼を取得
	rows, err := db.DB.Query("SELECT id, customer_id, pickup_lat, pickup_lng, destination_lat, destination_lng, pickup_address, destination_address, estimated_fare, status, created_at, updated_at FROM ride_requests WHERE status = 'pending' AND (scheduled_at IS NULL OR scheduled_at <= NOW() + INTERVAL '1 hour') ORDER BY created_at DESC")
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

// GetReservedRideRequests handles GET /driver/reserved-requests
func GetReservedRideRequests(c *gin.Context) {
	// scheduled_at が設定されている未受諾の依頼を取得
	query := `
		SELECT id, customer_id, pickup_lat, pickup_lng, destination_lat, destination_lng, pickup_address, destination_address, estimated_fare, status, scheduled_at, created_at, updated_at 
		FROM ride_requests 
		WHERE status = 'pending' AND scheduled_at IS NOT NULL 
		ORDER BY scheduled_at ASC
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
		err := rows.Scan(&r.ID, &r.CustomerID, &r.PickupLat, &r.PickupLng, &r.DestinationLat, &r.DestinationLng, &r.PickupAddress, &r.DestinationAddress, &r.EstimatedFare, &r.Status, &r.ScheduledAt, &r.CreatedAt, &r.UpdatedAt)
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
		"pending":   true, // 辞退用
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

	if body.Status == "pending" {
		updateFields += ", driver_id = NULL"
	}
	
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

	// Use a transaction to ensure both ride_requests and drivers tables are updated
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// 1. Get driver_id
	var driverID *string
	err = tx.QueryRow("SELECT driver_id FROM ride_requests WHERE id = $1", id).Scan(&driverID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride request not found"})
		return
	}

	if driverID == nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "No driver assigned to this ride request"})
		return
	}

	// 2. Update ride_request rating
	_, err = tx.Exec("UPDATE ride_requests SET rating_to_driver = $1, review_comment = $2, updated_at = $3 WHERE id = $4",
		body.Rating, body.Comment, time.Now(), id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rating"})
		return
	}

	// 3. Recalculate driver average rating and count
	var avgRating float64
	var count int
	err = tx.QueryRow(`
		SELECT COALESCE(AVG(rating_to_driver), 0), COUNT(rating_to_driver)
		FROM ride_requests
		WHERE driver_id = $1 AND rating_to_driver IS NOT NULL
	`, *driverID).Scan(&avgRating, &count)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate driver rating"})
		return
	}

	// 4. Update drivers table
	_, err = tx.Exec("UPDATE drivers SET average_rating = $1, rating_count = $2, updated_at = $3 WHERE id = $4",
		avgRating, count, time.Now(), *driverID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver statistics"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
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
// ListCustomerReservations handles GET /customer/ride-requests/reserved
func ListCustomerReservations(c *gin.Context) {
	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
		return
	}

	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.scheduled_at, r.created_at, r.updated_at,
			d.name, d.phone_number
		FROM ride_requests r
		LEFT JOIN drivers d ON r.driver_id = d.id
		WHERE r.customer_id = $1 AND r.scheduled_at IS NOT NULL AND r.status NOT IN ('completed', 'cancelled')
		ORDER BY r.scheduled_at ASC
	`
	rows, err := db.DB.Query(query, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var requests []models.RideRequest
	for rows.Next() {
		var r models.RideRequest
		var driverName, driverPhone *string
		err := rows.Scan(
			&r.ID, &r.CustomerID, &r.DriverID, &r.PickupLat, &r.PickupLng, 
			&r.DestinationLat, &r.DestinationLng, &r.PickupAddress, 
			&r.DestinationAddress, &r.EstimatedFare, &r.Status, 
			&r.ScheduledAt, &r.CreatedAt, &r.UpdatedAt,
			&driverName, &driverPhone,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row: " + err.Error()})
			return
		}
		r.DriverName = driverName
		r.DriverPhone = driverPhone
		requests = append(requests, r)
	}

	if requests == nil {
		requests = []models.RideRequest{}
	}
	c.JSON(http.StatusOK, requests)
}

// ListDriverReservations handles GET /driver/drivers/:id/reservations
func ListDriverReservations(c *gin.Context) {
	driverID := c.Param("id")

	query := `
		SELECT 
			r.id, r.customer_id, r.driver_id, r.pickup_lat, r.pickup_lng, 
			r.destination_lat, r.destination_lng, r.pickup_address, 
			r.destination_address, r.estimated_fare, r.status, 
			r.scheduled_at, r.created_at, r.updated_at
		FROM ride_requests r
		WHERE r.driver_id = $1 AND r.status = 'accepted' AND r.scheduled_at IS NOT NULL
		ORDER BY r.scheduled_at ASC
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
			&r.ScheduledAt, &r.CreatedAt, &r.UpdatedAt,
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
