package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Default to Neon database provided by user
		connStr = "postgresql://neondb_owner:npg_7X9NEisWRGIU@ep-odd-sun-a11xn5a1-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require"
		log.Println("DATABASE_URL not set, using Neon default")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	fmt.Println("Connected to Neon database")

	// Create tables if they don't exist
	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			company_id UUID,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT DEFAULT 'admin',
			name TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			contact_phone TEXT,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			phone_number TEXT UNIQUE NOT NULL,
			name TEXT,
			email TEXT,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS drivers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			company_id UUID REFERENCES companies(id),
			phone_number TEXT UNIQUE NOT NULL,
			name TEXT,
			car_model TEXT,
			current_lat FLOAT,
			current_lng FLOAT,
			status TEXT DEFAULT 'offline',
			fcm_token TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ride_requests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			customer_id UUID REFERENCES users(id),
			driver_id UUID REFERENCES drivers(id),
			pickup_address TEXT,
			destination_address TEXT,
			pickup_lat FLOAT,
			pickup_lng FLOAT,
			dest_lat FLOAT,
			dest_lng FLOAT,
			scheduled_at TIMESTAMP,
			status TEXT DEFAULT 'pending',
			price INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ride_request_id UUID REFERENCES ride_requests(id),
			sender_id UUID,
			sender_type TEXT,
			content TEXT,
			image_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS otp_codes (
			phone_number TEXT PRIMARY KEY,
			code TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL
		)`,
	}

	for _, q := range queries {
		_, err := DB.Exec(q)
		if err != nil {
			log.Printf("Error creating table: %v", err)
		}
	}

	// Insert default admin if not exists
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = 'admin'").Scan(&count)
	if count == 0 {
		DB.Exec("INSERT INTO admin_users (username, password_hash, role, name) VALUES ('admin', 'admin123', 'super_admin', 'System Admin')")
		fmt.Println("Default admin user (admin/admin123) created in admin_users")
	}

	// Insert test driver if not exists
	DB.QueryRow("SELECT COUNT(*) FROM drivers WHERE phone_number = '07000000000'").Scan(&count)
	if count == 0 {
		// First create a default company
		var companyId string
		DB.QueryRow("INSERT INTO companies (name) VALUES ('テスト代行社') RETURNING id").Scan(&companyId)
		DB.Exec("INSERT INTO drivers (company_id, phone_number, name, status) VALUES ($1, '07000000000', 'テスト', 'offline')", companyId)
		fmt.Println("Test driver (07000000000) created")
	}

	fmt.Println("Database tables initialized successfully")
}
