package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	db       *sql.DB
	migrator *migrate.Migrate
}

func NewMigrator(db *sql.DB, migrationsPath string) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{
		db:       db,
		migrator: m,
	}, nil
}

func (m *Migrator) Up() error {
	log.Println("Running database migrations...")
	err := m.migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	if err == migrate.ErrNoChange {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Database migrations completed successfully")
	}
	return nil
}

func (m *Migrator) Down() error {
	log.Println("Rolling back database migrations...")
	err := m.migrator.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	log.Println("Database migrations rolled back successfully")
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.migrator.Version()
}

func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrator.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}