package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DATABASE")
	user := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")

	connStr := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	migrationFile := flag.String("migration", "", "Path to the migration file")
	flag.Parse()

	if *migrationFile == "" {
		log.Fatalf("migration file path is required")
	}

	query, err := os.ReadFile(*migrationFile)
	if err != nil {
		log.Fatalf("failed to read SQL file: %v", err)
	}

	_, err = db.Exec(string(query))
	if err != nil {
		log.Fatalf("failed to execute SQL query: %v", err)
	}

	log.Println("Migration applied successfully")
}
