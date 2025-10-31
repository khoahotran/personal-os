package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/khoahotran/personal-os/pkg/auth"
)

func main() {
	fmt.Println("adding owner into database...")

	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found, use system environment variables.")
	}

	DSN := os.Getenv("DB_DSN")
	OWNER_EMAIl := os.Getenv("OWNER_EMAIL")
	OWNER_PASSWORD := os.Getenv("OWNER_PASSWORD")

	hash, err := auth.HashPassword(OWNER_PASSWORD)
	if err != nil {
		log.Fatalf("cannot has password: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), DSN)
	if err != nil {
		log.Fatalf("cannot connect DB: %v", err)
	}
	defer pool.Close()

	query := `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET password_hash = $3
	`
	_, err = pool.Exec(context.Background(), query, uuid.New(), OWNER_EMAIl, hash)
	if err != nil {
		log.Fatalf("cannot add user: %v", err)
	}

	fmt.Printf("added or updated owner '%s' successfully!\n", OWNER_EMAIl)
}
