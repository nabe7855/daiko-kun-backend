package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"net/http"
	"strings"
	"time"

	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListMessages handles GET /ride-requests/:id/messages
func ListMessages(c *gin.Context) {
	rideID := c.Param("id")

	rows, err := db.DB.Query(`
		SELECT id, ride_id, sender_id, sender_type, content, image_url, created_at 
		FROM messages 
		WHERE ride_id = $1 
		ORDER BY created_at ASC
	`, rideID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.RideID, &m.SenderID, &m.SenderType, &m.Content, &m.ImageURL, &m.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		messages = append(messages, m)
	}

	if messages == nil {
		messages = []models.Message{}
	}
	c.JSON(http.StatusOK, messages)
}

// SendMessage handles POST /ride-requests/:id/messages
func SendMessage(c *gin.Context) {
	rideID := c.Param("id")
	var body struct {
		SenderID   string `json:"sender_id" binding:"required"`
		SenderType string `json:"sender_type" binding:"required"`
		Content    string `json:"content"` // Now optional
		ImageURL   string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := models.Message{
		ID:         uuid.New().String(),
		RideID:     rideID,
		SenderID:   body.SenderID,
		SenderType: body.SenderType,
		Content:    body.Content,
		ImageURL:   body.ImageURL,
		CreatedAt:  time.Now(),
	}

	_, err := db.DB.Exec(`
		INSERT INTO messages (id, ride_id, sender_id, sender_type, content, image_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, m.ID, m.RideID, m.SenderID, m.SenderType, m.Content, m.ImageURL, m.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

// CreateEmergencyReport handles POST /ride-requests/:id/emergency
func CreateEmergencyReport(c *gin.Context) {
	rideID := c.Param("id")
	var body struct {
		ReporterID   string `json:"reporter_id" binding:"required"`
		ReporterType string `json:"reporter_type" binding:"required"`
		Reason       string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e := models.EmergencyReport{
		ID:           uuid.New().String(),
		RideID:       rideID,
		ReporterID:   body.ReporterID,
		ReporterType: body.ReporterType,
		Reason:       body.Reason,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	// トランザクションで緊急停止と報告を記録
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO ride_emergencies (id, ride_id, reporter_id, reporter_type, reason, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ID, e.RideID, e.ReporterID, e.ReporterType, e.Reason, e.Status, e.CreatedAt)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record emergency: " + err.Error()})
		return
	}

	// 依頼自体のステータスを緊急停止的なものに変える（オプションだが、今回はステータスをキャンセルに近い扱いに）
	_, err = tx.Exec("UPDATE ride_requests SET status = 'cancelled', updated_at = $1 WHERE id = $2", time.Now(), rideID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel ride: " + err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, e)
}

// UploadMessageImage handles POST /ride-requests/:id/upload
func UploadMessageImage(c *gin.Context) {
	// ファイルサイズ制限 (最大10MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded or file too large: " + err.Error()})
		return
	}
	defer file.Close()

	// 拡張子の簡易バリデーション
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only jpg, png, gif allowed."})
		return
	}

	// Ensure uploads directory exists
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory: " + err.Error()})
			return
		}
	}

	// Create unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dst := filepath.Join(uploadDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file: " + err.Error()})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file content: " + err.Error()})
		return
	}

	// URL構築 (本来はドメインを変数化すべきだが、一旦相対パスまたは固定ホストで返す)
	// フロントエンドは http://localhost:8080 をつけてアクセスする想定
	url := fmt.Sprintf("/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
