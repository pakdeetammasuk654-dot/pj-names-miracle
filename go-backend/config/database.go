package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	// Load local development variables when .env is present. Existing process
	// environment variables take precedence, so deployment configuration is not
	// overridden by a file.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	host, err := requiredEnv("DB_HOST")
	if err != nil {
		return nil, err
	}
	port := getEnv("DB_PORT", "5432")
	user, err := requiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}
	pass, err := requiredEnv("DB_PASS")
	if err != nil {
		return nil, err
	}
	dbname, err := requiredEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	sslMode := strings.ToLower(getEnv("DB_SSLMODE", "disable"))
	switch sslMode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		return nil, fmt.Errorf("DB_SSLMODE must be disable, require, verify-ca, or verify-full")
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbname, sslMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if _, err := db.Exec("SELECT set_limit(0.15)"); err != nil {
		log.Printf("[Database] Note: set_limit(0.15) warning: %v", err)
	}

	log.Println("[Database] Successfully connected to PostgreSQL!")
	return db, nil
}

func requiredEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %s is not set", key)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
