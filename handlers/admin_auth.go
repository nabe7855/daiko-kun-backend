package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID        string  `json:"id"`
	CompanyID *string `json:"company_id"`
	Role      string  `json:"role"`
	Name      string  `json:"name"`
}

// AdminLogin handles POST /admin/login
func AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.AdminUser
	// Note: In a real production app, use bcrypt to compare hashes.
	// For this phase, we compare plain text as a placeholder if stored as such.
	err := db.DB.QueryRow("SELECT id, company_id, username, role, name FROM admin_users WHERE username = $1 AND password_hash = $2", req.Username, req.Password).
		Scan(&user.ID, &user.CompanyID, &user.Username, &user.Role, &user.Name)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		ID:        user.ID,
		CompanyID: user.CompanyID,
		Role:      user.Role,
		Name:      user.Name,
	})
}
