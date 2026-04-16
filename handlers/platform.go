package handlers

import (
	"daiko-kun-backend/db"
	"daiko-kun-backend/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListCompanies handles GET /admin/platform/companies
func ListCompanies(c *gin.Context) {
	log.Println("Handling GET /admin/platform/companies")
	query := "SELECT id, name, COALESCE(address, ''), COALESCE(phone_number, ''), COALESCE(email, ''), COALESCE(status, 'active'), COALESCE(commission_rate, 0), created_at, updated_at FROM companies ORDER BY created_at DESC"
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Printf("Error querying companies: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var comp models.Company
		err := rows.Scan(&comp.ID, &comp.Name, &comp.Address, &comp.PhoneNumber, &comp.Email, &comp.Status, &comp.CommissionRate, &comp.CreatedAt, &comp.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning company row %s: %v\n", comp.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		companies = append(companies, comp)
	}

	if companies == nil {
		companies = []models.Company{}
	}
	c.JSON(http.StatusOK, companies)
}

// CreateCompany handles POST /admin/platform/companies
func CreateCompany(c *gin.Context) {
	var req struct {
		models.Company
		AdminUsername string `json:"admin_username"`
		AdminPassword string `json:"admin_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	companyID := uuid.New().String()
	now := time.Now()

	// 1. Create Company
	sqlComp := `INSERT INTO companies (id, name, address, phone_number, email, status, commission_rate, created_at, updated_at) 
	            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.Exec(sqlComp, companyID, req.Name, req.Address, req.PhoneNumber, req.Email, "active", req.CommissionRate, now, now)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company: " + err.Error()})
		return
	}

	// 2. Create Admin User for this company
	if req.AdminUsername != "" && req.AdminPassword != "" {
		adminID := uuid.New().String()
		sqlAdmin := `INSERT INTO admin_users (id, company_id, username, password_hash, role, name, created_at, updated_at)
		             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = tx.Exec(sqlAdmin, adminID, companyID, req.AdminUsername, req.AdminPassword, "company_admin", req.Name+" 管理者", now, now)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": companyID, "message": "Company and admin user created successfully"})
}

// UpdateCompanyStatus handles PATCH /admin/platform/companies/:id/status
func UpdateCompanyStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec("UPDATE companies SET status = $1, updated_at = $2 WHERE id = $3", req.Status, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company status updated"})
}

// GetPlatformStats handles GET /admin/platform/stats
func GetPlatformStats(c *gin.Context) {
	var stats struct {
		TotalCompanies  int     `json:"total_companies"`
		TotalDrivers    int     `json:"total_drivers"`
		TotalRequests   int     `json:"total_requests"`
		TotalSales      float64 `json:"total_sales"`
		PlatformRevenue float64 `json:"platform_revenue"`
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM companies").Scan(&stats.TotalCompanies)
	db.DB.QueryRow("SELECT COUNT(*) FROM drivers").Scan(&stats.TotalDrivers)
	db.DB.QueryRow("SELECT COUNT(*) FROM ride_requests WHERE status = 'completed'").Scan(&stats.TotalRequests)
	db.DB.QueryRow("SELECT COALESCE(SUM(actual_fare), 0) FROM ride_requests WHERE status = 'completed'").Scan(&stats.TotalSales)

	// Calculate Platform Revenue (based on each company's commission rate)
	revenueQuery := `
		SELECT COALESCE(SUM(r.actual_fare * c.commission_rate / 100.0), 0)
		FROM ride_requests r
		JOIN drivers d ON r.driver_id = d.id
		JOIN companies c ON d.company_id = c.id
		WHERE r.status = 'completed'
	`
	db.DB.QueryRow(revenueQuery).Scan(&stats.PlatformRevenue)

	c.JSON(http.StatusOK, stats)
}

type CompanySettlement struct {
	CompanyID      string  `json:"company_id"`
	CompanyName    string  `json:"company_name"`
	TotalSales     float64 `json:"total_sales"`
	PlatformFee    float64 `json:"platform_fee"`
	NetProfit      float64 `json:"net_profit"`
	CompletedRides int     `json:"completed_rides"`
}

// GetSettlements handles GET /admin/platform/settlements
func GetSettlements(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")

	whereClause := "r.status = 'completed'"
	var args []interface{}

	if year != "" && month != "" {
		// Assuming 'updated_at' is used as the completion timestamp
		// PostgreSQL: date_part('year', updated_at) = $1 AND date_part('month', updated_at) = $2
		whereClause += " AND date_part('year', r.updated_at) = $1 AND date_part('month', r.updated_at) = $2"
		args = append(args, year, month)
	}

	query := fmt.Sprintf(`
		SELECT 
			c.id, c.name,
			COUNT(r.id) as completed_rides,
			COALESCE(SUM(r.actual_fare), 0) as total_sales,
			COALESCE(SUM(r.actual_fare * c.commission_rate / 100.0), 0) as platform_fee
		FROM companies c
		LEFT JOIN drivers d ON c.id = d.company_id
		LEFT JOIN ride_requests r ON d.id = r.driver_id AND %s
		GROUP BY c.id, c.name
		ORDER BY total_sales DESC
	`, whereClause)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var settlements []CompanySettlement
	for rows.Next() {
		var s CompanySettlement
		err := rows.Scan(&s.CompanyID, &s.CompanyName, &s.CompletedRides, &s.TotalSales, &s.PlatformFee)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		s.NetProfit = s.TotalSales - s.PlatformFee
		settlements = append(settlements, s)
	}

	if settlements == nil {
		settlements = []CompanySettlement{}
	}
	c.JSON(http.StatusOK, settlements)
}

// ExportSettlementsCSV handles GET /admin/platform/settlements/export
func ExportSettlementsCSV(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")

	whereClause := "r.status = 'completed'"
	var args []interface{}

	if year != "" && month != "" {
		whereClause += " AND date_part('year', r.updated_at) = $1 AND date_part('month', r.updated_at) = $2"
		args = append(args, year, month)
	}

	query := fmt.Sprintf(`
		SELECT 
			c.name,
			COUNT(r.id) as completed_rides,
			COALESCE(SUM(r.actual_fare), 0) as total_sales,
			COALESCE(SUM(r.actual_fare * c.commission_rate / 100.0), 0) as platform_fee
		FROM companies c
		LEFT JOIN drivers d ON c.id = d.company_id
		LEFT JOIN ride_requests r ON d.id = r.driver_id AND %s
		GROUP BY c.id, c.name
		ORDER BY total_sales DESC
	`, whereClause)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	csvData := "会社名,完了配車数,総売上,プラットフォーム手数料,純利益\n"
	for rows.Next() {
		var name string
		var count int
		var sales, fee float64
		err := rows.Scan(&name, &count, &sales, &fee)
		if err != nil {
			continue
		}
		profit := sales - fee
		csvData += fmt.Sprintf("%s,%d,%.0f,%.0f,%.0f\n", name, count, sales, fee, profit)
	}

	filename := fmt.Sprintf("settlement_%s_%s.csv", year, month)
	if year == "" || month == "" {
		filename = "settlement_all.csv"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, csvData)
}

// GetCompanyStats handles GET /admin/company/stats
func GetCompanyStats(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	// Security: Prefer company_id from context (JWT) over query param
	var companyID string
	companyIDInToken, exists := c.Get("company_id")
	if exists && companyIDInToken != nil {
		// Enforce tenant isolation
		companyID = companyIDInToken.(string)
	} else {
		// Fallback for Super Admin
		companyID = c.Query("company_id")
	}

	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}

	var stats struct {
		TotalDrivers  int     `json:"total_drivers"`
		TotalRequests int     `json:"total_requests"`
		TotalSales    float64 `json:"total_sales"`
	}

	db.DB.QueryRow("SELECT COUNT(*) FROM drivers WHERE company_id = $1", companyID).Scan(&stats.TotalDrivers)
	
	// Join with drivers to filter ride_requests by company_id
	queryRequests := `
		SELECT COUNT(*), COALESCE(SUM(r.actual_fare), 0) 
		FROM ride_requests r
		JOIN drivers d ON r.driver_id = d.id
		WHERE d.company_id = $1 AND r.status = 'completed'
	`
	db.DB.QueryRow(queryRequests, companyID).Scan(&stats.TotalRequests, &stats.TotalSales)

	c.JSON(http.StatusOK, stats)
}
