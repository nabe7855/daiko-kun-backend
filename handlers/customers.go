package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestOTP handles POST /customer/request-otp
func RequestOTP(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	// モックなので、実際には送らずコンソールに出力するだけ
	fmt.Printf("\n--- [SMS MOCK] ---\nTO: %s\nCODE: 1234\n------------------\n\n", req.PhoneNumber)

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

// VerifyOTP handles POST /customer/verify-otp
func VerifyOTP(c *gin.Context) {
	var req struct {
		PhoneNumber    string `json:"phone_number" binding:"required"`
		Code           string `json:"code" binding:"required"`
		SocialID       string `json:"social_id"`
		SocialProvider string `json:"social_provider"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number and code are required"})
		return
	}

	// モック認証: コードが 1234 なら通過
	if req.Code != "1234" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
		return
	}

	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	var customer models.Customer
	err := db.DB.QueryRow("SELECT id, name, phone_number, email, social_id, social_provider, status, created_at, updated_at FROM customers WHERE phone_number = $1", req.PhoneNumber).
		Scan(&customer.ID, &customer.Name, &customer.PhoneNumber, &customer.Email, &customer.SocialID, &customer.SocialProvider, &customer.Status, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		// 未登録の場合は新規登録
		customer.ID = uuid.New().String()
		customer.PhoneNumber = req.PhoneNumber
		customer.Status = "active"
		customer.CreatedAt = time.Now()
		customer.UpdatedAt = time.Now()

		sqlStatement := `
			INSERT INTO customers (id, phone_number, social_id, social_provider, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`
		err = db.DB.QueryRow(sqlStatement,
			customer.ID,
			customer.PhoneNumber,
			req.SocialID,
			req.SocialProvider,
			customer.Status,
			customer.CreatedAt,
			customer.UpdatedAt,
		).Scan(&customer.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, customer)
}

// UpdateCustomerProfile handles PATCH /customer/profile
func UpdateCustomerProfile(c *gin.Context) {
	var req struct {
		ID    string  `json:"id" binding:"required"`
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE customers SET "
	args := []interface{}{}
	updates := []string{}

	if req.Name != nil {
		args = append(args, *req.Name)
		updates = append(updates, fmt.Sprintf("name = $%d", len(args)))
	}
	if req.Email != nil {
		args = append(args, *req.Email)
		updates = append(updates, fmt.Sprintf("email = $%d", len(args)))
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	args = append(args, time.Now())
	updates = append(updates, fmt.Sprintf("updated_at = $%d", len(args)))

	args = append(args, req.ID)
	query += strings.Join(updates, ", ") + fmt.Sprintf(" WHERE id = $%d", len(args))

	_, err := db.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// ListSavedAddresses handles GET /customer/:id/addresses
func ListSavedAddresses(c *gin.Context) {
	customerID := c.Param("id")

	rows, err := db.DB.Query("SELECT id, customer_id, label, address, description, latitude, longitude, created_at, updated_at FROM customer_addresses WHERE customer_id = $1 ORDER BY created_at DESC", customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var addresses []models.SavedAddress
	for rows.Next() {
		var a models.SavedAddress
		err := rows.Scan(&a.ID, &a.CustomerID, &a.Label, &a.Address, &a.Description, &a.Latitude, &a.Longitude, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		addresses = append(addresses, a)
	}

	if addresses == nil {
		addresses = []models.SavedAddress{}
	}
	c.JSON(http.StatusOK, addresses)
}

// AddSavedAddress handles POST /customer/addresses
func AddSavedAddress(c *gin.Context) {
	var req models.SavedAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	sqlStatement := `
		INSERT INTO customer_addresses (id, customer_id, label, address, description, latitude, longitude, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := db.DB.Exec(sqlStatement, req.ID, req.CustomerID, req.Label, req.Address, req.Description, req.Latitude, req.Longitude, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// DeleteSavedAddress handles DELETE /customer/addresses/:id
func DeleteSavedAddress(c *gin.Context) {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM customer_addresses WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
}

// CustomerLogin - Deprecated (Unified into VerifyOTP)
func CustomerLogin(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{"error": "Please use /verify-otp instead"})
}

// UpdateCustomerFCMToken handles PATCH /customer/fcm-token
func UpdateCustomerFCMToken(c *gin.Context) {
	var req struct {
		ID       string `json:"id" binding:"required"`
		FCMToken string `json:"fcm_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE customers SET fcm_token = $1, updated_at = $2 WHERE id = $3", req.FCMToken, time.Now(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FCM token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
}

// DeleteCustomerAccount handles DELETE /customer/:id
func DeleteCustomerAccount(c *gin.Context) {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
