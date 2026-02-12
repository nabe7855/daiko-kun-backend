package main

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	db.InitDB()

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Simple Ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Admin Routes
	admin := r.Group("/admin")
	{
		admin.POST("/drivers", handlers.CreateDriver)
		admin.GET("/drivers", handlers.ListDrivers)
		admin.GET("/ride-requests", handlers.ListAllRideRequests)
		admin.GET("/drivers/:id/history", handlers.ListDriverHistory)
	}

	// Driver Routes
	driver := r.Group("/driver")
	{
		driver.POST("/login", handlers.DriverLogin)
		driver.PATCH("/drivers/:id/status", handlers.UpdateDriverStatus)
		driver.GET("/drivers/:id/history", handlers.ListDriverHistory)
		driver.GET("/available-requests", handlers.GetAvailableRideRequests)
		driver.POST("/ride-requests/:id/accept", handlers.AcceptRideRequest)
		driver.PATCH("/ride-requests/:id/status", handlers.UpdateRideStatus)
		driver.PATCH("/drivers/:id/location", handlers.UpdateDriverLocation)
	}

	// Customer Routes
	customer := r.Group("/customer")
	{
		customer.POST("/request-otp", handlers.RequestOTP)
		customer.POST("/verify-otp", handlers.VerifyOTP)
		customer.POST("/login", handlers.CustomerLogin) // Deprecated
		customer.POST("/ride-requests", handlers.CreateRideRequest)
		customer.POST("/ride-requests/:id/rate", handlers.SubmitRating)
	}

	// Message & Emergency Routes (Shared)
	r.GET("/ride-requests/:id/messages", handlers.ListMessages)
	r.POST("/ride-requests/:id/messages", handlers.SendMessage)
	r.POST("/ride-requests/:id/emergency", handlers.CreateEmergencyReport)
	r.POST("/ride-requests/:id/upload", handlers.UploadMessageImage)

	// Static files for image uploads
	r.Static("/uploads", "./uploads")

	r.Run() // listen and serve on 0.0.0.0:8080
}
