package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateDriver handles POST /admin/drivers
func CreateDriver(c *gin.Context) {
	var driver models.Driver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim whitespace
	driver.Name = strings.TrimSpace(driver.Name)
	driver.PhoneNumber = strings.TrimSpace(driver.PhoneNumber)
	driver.LicenseNumber = strings.TrimSpace(driver.LicenseNumber)

	// Generate ID and timestamps
	driver.ID = uuid.New().String()
	driver.CreatedAt = time.Now()
	driver.UpdatedAt = time.Now()
	
	// Default status
	if driver.Status == "" {
		driver.Status = "inactive"
	}

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	sqlStatement := `
		INSERT INTO drivers (id, name, phone_number, license_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	err := db.DB.QueryRow(sqlStatement, 
		driver.ID, 
		driver.Name, 
		driver.PhoneNumber, 
		driver.LicenseNumber, 
		driver.Status, 
		driver.CreatedAt, 
		driver.UpdatedAt,
	).Scan(&driver.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert driver: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// ListDrivers handles GET /admin/drivers
func ListDrivers(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	rows, err := db.DB.Query("SELECT id, name, phone_number, license_number, status, created_at, updated_at FROM drivers ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers: " + err.Error()})
		return
	}
	defer rows.Close()

	var drivers []models.Driver
	for rows.Next() {
		var d models.Driver
		err := rows.Scan(&d.ID, &d.Name, &d.PhoneNumber, &d.LicenseNumber, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan driver: " + err.Error()})
			return
		}
		drivers = append(drivers, d)
	}

	if drivers == nil {
		drivers = []models.Driver{}
	}

	c.JSON(http.StatusOK, drivers)
}

// DriverLogin handles POST /driver/login
func DriverLogin(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	var driver models.Driver
	err := db.DB.QueryRow("SELECT id, name, phone_number, license_number, status, created_at, updated_at FROM drivers WHERE phone_number = $1", req.PhoneNumber).
		Scan(&driver.ID, &driver.Name, &driver.PhoneNumber, &driver.LicenseNumber, &driver.Status, &driver.CreatedAt, &driver.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// UpdateDriverStatus handles PATCH /driver/drivers/:id/status
func UpdateDriverStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	_, err := db.DB.Exec("UPDATE drivers SET status = $1, updated_at = $2 WHERE id = $3", req.Status, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// UpdateDriverLocation handles PATCH /driver/drivers/:id/location
func UpdateDriverLocation(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Lat float64 `json:"lat" binding:"required"`
		Lng float64 `json:"lng" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	_, err := db.DB.Exec("UPDATE drivers SET current_lat = $1, current_lng = $2, updated_at = $3 WHERE id = $4", 
		body.Lat, body.Lng, time.Now(), id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver location: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location updated successfully"})
}
