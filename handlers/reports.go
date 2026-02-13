package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitUserReport handles POST /ride-requests/:id/report
func SubmitUserReport(c *gin.Context) {
	rideID := c.Param("id")
	var req models.UserReport
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.RideID = rideID
	req.Status = "pending"
	req.CreatedAt = time.Now()

	sql := `INSERT INTO user_reports (id, ride_id, reporter_id, reported_user_id, reporter_role, reason, status, created_at)
	        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := db.DB.Exec(sql, req.ID, req.RideID, req.ReporterID, req.ReportedUserID, req.ReporterRole, req.Reason, req.Status, req.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit report: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// ListReports handles GET /admin/platform/reports
func ListReports(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, ride_id, reporter_id, reported_user_id, reporter_role, reason, status, created_at FROM user_reports ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reports []models.UserReport
	for rows.Next() {
		var r models.UserReport
		err := rows.Scan(&r.ID, &r.RideID, &r.ReporterID, &r.ReportedUserID, &r.ReporterRole, &r.Reason, &r.Status, &r.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reports = append(reports, r)
	}

	if reports == nil {
		reports = []models.UserReport{}
	}
	c.JSON(http.StatusOK, reports)
}
