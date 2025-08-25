package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresConnection(connectionString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
	log.Println("Closing PostgreSQL database connection")
	return p.DB.Close()
}

func (p *PostgresDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return p.DB.PingContext(ctx)
}

func (p *PostgresDB) GetStats() sql.DBStats {
	return p.DB.Stats()
}