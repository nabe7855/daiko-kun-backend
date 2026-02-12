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

// CustomerLogin - Deprecated (Unified into VerifyOTP)
func CustomerLogin(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{"error": "Please use /verify-otp instead"})
}
