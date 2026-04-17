package main

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/handlers"
	"daiko-kun-backend/middleware"
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
	admin.POST("/login", handlers.AdminLogin)

	// Protected routes
	adminProtected := admin.Group("/")
	adminProtected.Use(middleware.AuthMiddleware())
	{
		// Platform Management (Super Admin)
		platform := adminProtected.Group("/platform")
		{
			platform.GET("/stats", handlers.GetPlatformStats)
			platform.GET("/companies", handlers.ListCompanies)
			platform.POST("/companies", handlers.CreateCompany)
			platform.PUT("/companies/:id", handlers.UpdateCompany)
			platform.PATCH("/companies/:id/status", handlers.UpdateCompanyStatus)
			platform.GET("/settlements", handlers.GetSettlements)
			platform.GET("/settlements/export", handlers.ExportSettlementsCSV)
			platform.GET("/reports", handlers.ListReports)
		}

		// Company Management (Individual Companies)
		company := adminProtected.Group("/company")
		{
			company.GET("/stats", handlers.GetCompanyStats)
		}

		adminProtected.POST("/drivers", handlers.CreateDriver)
		adminProtected.GET("/drivers", handlers.ListDrivers)
		adminProtected.GET("/ride-requests", handlers.ListAllRideRequests)
		adminProtected.GET("/drivers/:id/history", handlers.ListDriverHistory)
	}

	// Driver Routes
	driver := r.Group("/driver")
	{
		driver.POST("/login", handlers.DriverLogin)
		driver.GET("/drivers/:id", handlers.GetDriver)
		driver.GET("/drivers/:id/reservations", handlers.ListDriverReservations)
		driver.PATCH("/drivers/:id/status", handlers.UpdateDriverStatus)
		driver.GET("/drivers/:id/history", handlers.ListDriverHistory)
		driver.GET("/available-requests", handlers.GetAvailableRideRequests)
		driver.GET("/reserved-requests", handlers.GetReservedRideRequests)
		driver.POST("/ride-requests/:id/accept", handlers.AcceptRideRequest)
		driver.PATCH("/ride-requests/:id/status", handlers.UpdateRideStatus)
		driver.POST("/ride-requests/:id/rate-customer", handlers.SubmitCustomerRating)
		driver.PATCH("/drivers/:id/location", handlers.UpdateDriverLocation)
		driver.PATCH("/fcm-token", handlers.UpdateDriverFCMToken)
		driver.DELETE("/drivers/:id", handlers.DeleteDriverAccount)
	}

	// Customer Routes
	customer := r.Group("/customer")
	{
		customer.POST("/request-otp", handlers.RequestOTP)
		customer.POST("/verify-otp", handlers.VerifyOTP)
		customer.POST("/login", handlers.CustomerLogin) // Deprecated
		customer.GET("/ride-requests", handlers.ListAllRideRequests)
		customer.GET("/ride-requests/:id", handlers.GetRideRequest)
		customer.GET("/ride-requests/reserved", handlers.ListCustomerReservations)
		customer.POST("/ride-requests", handlers.CreateRideRequest)
		customer.PATCH("/ride-requests/:id/status", handlers.UpdateRideStatus)
		customer.POST("/ride-requests/:id/rate", handlers.SubmitRating)
		customer.PATCH("/profile", handlers.UpdateCustomerProfile)
		customer.GET("/:id/addresses", handlers.ListSavedAddresses)
		customer.POST("/addresses", handlers.AddSavedAddress)
		customer.DELETE("/addresses/:id", handlers.DeleteSavedAddress)
		customer.PATCH("/fcm-token", handlers.UpdateCustomerFCMToken)
		customer.DELETE("/:id", handlers.DeleteCustomerAccount)
	}

	// Message & Emergency Routes (Shared)
	r.GET("/ride-requests/:id/messages", handlers.ListMessages)
	r.POST("/ride-requests/:id/messages", handlers.SendMessage)
	r.POST("/ride-requests/:id/emergency", handlers.CreateEmergencyReport)
	r.POST("/ride-requests/:id/report", handlers.SubmitUserReport)
	r.POST("/ride-requests/:id/upload", handlers.UploadMessageImage)

	// Static files for image uploads
	r.Static("/uploads", "./uploads")

	r.Run() // listen and serve on 0.0.0.0:8080
}
