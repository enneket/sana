package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func initDB() {
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable must be set")
	}

	var err error
	db, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Test connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	// Create schema
	if err := createSchema(); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database connected and schema initialized")
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sanas (
		id SERIAL PRIMARY KEY,
		uid TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL REFERENCES users(id),
		content TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_sanas_user_id ON sanas(user_id);
	CREATE INDEX IF NOT EXISTS idx_sanas_updated_at ON sanas(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sanas_content_gin ON sanas USING gin(to_tsvector('simple', content));
	`
	_, err := db.Exec(context.Background(), schema)
	return err
}

func closeDB() {
	db.Close()
}

// Memo 模型
type Sana struct {
	ID        int       `json:"id"`
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SanaResponse 对外API响应格式
type SanaResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedTs int64  `json:"created_ts"`
	UpdatedTs int64  `json:"updated_ts"`
}

// ToResponse converts Sana to SanaResponse
func (s *Sana) ToResponse() SanaResponse {
	return SanaResponse{
		ID:        s.UID,
		Content:   s.Content,
		CreatedTs: s.CreatedAt.Unix(),
		UpdatedTs: s.UpdatedAt.Unix(),
	}
}
