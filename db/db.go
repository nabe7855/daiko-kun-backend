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
		// Default for local development
		connStr = "postgres://daiko_user:daiko_password@localhost:5432/daiko_db?sslmode=disable"
		log.Println("DATABASE_URL not set, using default for local development")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	fmt.Println("Connected to database")
}
