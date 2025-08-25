package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get the database URL from environment or command line
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:nujMat-tykzos-xahqo5@db.hpoauyaopahsdxeryjlb.supabase.co:5432/postgres?sslmode=require"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check migration state
	log.Println("=== Migration State ===")
	rows, err := db.Query("SELECT version, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Printf("No schema_migrations table found: %v", err)
	} else {
		for rows.Next() {
			var version int
			var dirty bool
			rows.Scan(&version, &dirty)
			log.Printf("Migration version: %d, dirty: %v", version, dirty)
		}
		rows.Close()
	}

	// Execute missing table creation
	log.Println("\n=== Creating Missing Tables ===")
	missingTablesSQL := `
		-- Create missing tables from migration 3
		CREATE TABLE IF NOT EXISTS authorized_representatives (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
		    first_name VARCHAR(100) NOT NULL,
		    last_name VARCHAR(100) NOT NULL,
		    email VARCHAR(255) NOT NULL,
		    phone VARCHAR(50) NOT NULL,
		    position VARCHAR(100) NOT NULL,
		    is_active BOOLEAN DEFAULT true,
		    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS enterprise_documents (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
		    document_type VARCHAR(100) NOT NULL CHECK (document_type IN (
		        'certificate_of_incorporation',
		        'business_license',
		        'tax_certificate',
		        'bank_statement',
		        'director_id',
		        'proof_of_address',
		        'articles_of_association',
		        'memorandum_of_association'
		    )),
		    file_name VARCHAR(255) NOT NULL,
		    file_path VARCHAR(500) NOT NULL,
		    file_size BIGINT NOT NULL,
		    mime_type VARCHAR(100) NOT NULL,
		    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
		    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		    verified_at TIMESTAMP WITH TIME ZONE
		);

		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_authorized_representatives_enterprise_id ON authorized_representatives(enterprise_id);
		CREATE INDEX IF NOT EXISTS idx_authorized_representatives_email ON authorized_representatives(email);
		CREATE INDEX IF NOT EXISTS idx_enterprise_documents_enterprise_id ON enterprise_documents(enterprise_id);
		CREATE INDEX IF NOT EXISTS idx_enterprise_documents_document_type ON enterprise_documents(document_type);
		CREATE INDEX IF NOT EXISTS idx_enterprise_documents_status ON enterprise_documents(status);
	`

	_, err = db.Exec(missingTablesSQL)
	if err != nil {
		log.Printf("Warning: Could not create missing tables: %v", err)
	} else {
		log.Println("Successfully created missing tables")
	}

	// Check what tables exist
	log.Println("\n=== Updated Table List ===")
	rows, err = db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		log.Printf("Table: %s", tableName)
	}

	// Check enterprise table structure
	log.Println("\n=== Enterprise Table Columns ===")
	colRows, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'enterprises' 
		AND table_schema = 'public'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Error checking enterprise columns: %v", err)
	} else {
		defer colRows.Close()
		for colRows.Next() {
			var columnName, dataType string
			colRows.Scan(&columnName, &dataType)
			log.Printf("Column: %s (%s)", columnName, dataType)
		}
	}

	// Update enterprise table to match migration 3 structure
	log.Println("\n=== Updating Enterprise Table Structure ===")
	updateEnterpriseSQL := `
		-- Add missing columns to enterprises table
		ALTER TABLE enterprises 
		ADD COLUMN IF NOT EXISTS trade_name VARCHAR(255),
		ADD COLUMN IF NOT EXISTS registration_number VARCHAR(100),
		ADD COLUMN IF NOT EXISTS tax_id VARCHAR(100),
		ADD COLUMN IF NOT EXISTS business_type VARCHAR(100),
		ADD COLUMN IF NOT EXISTS industry VARCHAR(100),
		ADD COLUMN IF NOT EXISTS website VARCHAR(255),
		ADD COLUMN IF NOT EXISTS email VARCHAR(255),
		ADD COLUMN IF NOT EXISTS phone VARCHAR(50),
		ADD COLUMN IF NOT EXISTS street1 VARCHAR(255),
		ADD COLUMN IF NOT EXISTS street2 VARCHAR(255),
		ADD COLUMN IF NOT EXISTS city VARCHAR(100),
		ADD COLUMN IF NOT EXISTS state VARCHAR(100),
		ADD COLUMN IF NOT EXISTS postal_code VARCHAR(20),
		ADD COLUMN IF NOT EXISTS country VARCHAR(100),
		ADD COLUMN IF NOT EXISTS compliance_status VARCHAR(50) DEFAULT 'pending',
		ADD COLUMN IF NOT EXISTS aml_risk_score INTEGER DEFAULT 0,
		ADD COLUMN IF NOT EXISTS sanctions_check_date TIMESTAMP WITH TIME ZONE,
		ADD COLUMN IF NOT EXISTS pep_check_date TIMESTAMP WITH TIME ZONE,
		ADD COLUMN IF NOT EXISTS compliance_officer VARCHAR(255),
		ADD COLUMN IF NOT EXISTS last_review_date TIMESTAMP WITH TIME ZONE,
		ADD COLUMN IF NOT EXISTS next_review_date TIMESTAMP WITH TIME ZONE,
		ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE;

		-- Add constraints if they don't exist
		DO $$
		BEGIN
		    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'enterprises_registration_number_key') THEN
		        ALTER TABLE enterprises ADD CONSTRAINT enterprises_registration_number_key UNIQUE (registration_number);
		    END IF;
		END $$;

		-- Update check constraints
		ALTER TABLE enterprises 
		DROP CONSTRAINT IF EXISTS enterprises_kyb_status_check,
		ADD CONSTRAINT enterprises_kyb_status_check CHECK (kyb_status IN ('pending', 'in_review', 'verified', 'rejected', 'suspended'));

		DO $$
		BEGIN
		    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'enterprises_compliance_status_check') THEN
		        ALTER TABLE enterprises ADD CONSTRAINT enterprises_compliance_status_check CHECK (compliance_status IN ('pending', 'compliant', 'non_compliant', 'under_review'));
		    END IF;
		END $$;

		-- Create additional indexes
		CREATE INDEX IF NOT EXISTS idx_enterprises_registration_number ON enterprises(registration_number);
		CREATE INDEX IF NOT EXISTS idx_enterprises_compliance_status ON enterprises(compliance_status);
		CREATE INDEX IF NOT EXISTS idx_enterprises_created_at ON enterprises(created_at);
	`

	_, err = db.Exec(updateEnterpriseSQL)
	if err != nil {
		log.Printf("Warning: Could not update enterprise table: %v", err)
	} else {
		log.Println("Successfully updated enterprise table structure")
	}
}
