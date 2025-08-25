package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Seeder struct {
	postgres *PostgresDB
	mongo    *MongoDB
}

func NewSeeder(postgres *PostgresDB, mongo *MongoDB) *Seeder {
	return &Seeder{
		postgres: postgres,
		mongo:    mongo,
	}
}

func (s *Seeder) SeedDevelopmentData() error {
	log.Println("Seeding development data...")

	if err := s.seedEnterprises(); err != nil {
		return fmt.Errorf("failed to seed enterprises: %w", err)
	}

	if err := s.seedContracts(); err != nil {
		return fmt.Errorf("failed to seed contracts: %w", err)
	}

	log.Println("Development data seeded successfully")
	return nil
}

func (s *Seeder) seedEnterprises() error {
	enterprises := []struct {
		legalName    string
		jurisdiction string
		kybStatus    string
		xrplWallet   string
	}{
		{"Acme Corp", "US", "verified", "rAcmeCorpXRPLWallet123456789"},
		{"Global Trade Ltd", "UK", "verified", "rGlobalTradeXRPLWallet123456"},
		{"Tech Innovations Inc", "CA", "pending", "rTechInnovXRPLWallet123456"},
		{"Supply Chain Solutions", "DE", "verified", "rSupplyChainXRPLWallet123456"},
		{"Digital Commerce Co", "SG", "pending", "rDigitalCommerceXRPLWallet123"},
	}

	for _, enterprise := range enterprises {
		id := uuid.New()
		query := `
			INSERT INTO enterprises (id, legal_name, jurisdiction, kyb_status, xrpl_wallet, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT DO NOTHING
		`
		now := time.Now()
		_, err := s.postgres.DB.Exec(query, id, enterprise.legalName, enterprise.jurisdiction, 
			enterprise.kybStatus, enterprise.xrplWallet, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert enterprise %s: %w", enterprise.legalName, err)
		}
		log.Printf("Seeded enterprise: %s", enterprise.legalName)
	}

	return nil
}

func (s *Seeder) seedContracts() error {
	// Get some enterprise IDs for contract parties
	rows, err := s.postgres.DB.Query("SELECT id FROM enterprises LIMIT 3")
	if err != nil {
		return fmt.Errorf("failed to get enterprise IDs: %w", err)
	}
	defer rows.Close()

	var enterpriseIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan enterprise ID: %w", err)
		}
		enterpriseIDs = append(enterpriseIDs, id)
	}

	if len(enterpriseIDs) < 2 {
		log.Println("Not enough enterprises to create sample contracts")
		return nil
	}

	contracts := []struct {
		parties           []string
		contractHash      string
		aiAnalysisConfidence float64
	}{
		{
			parties:           []string{enterpriseIDs[0], enterpriseIDs[1]},
			contractHash:      "contract_hash_001_supply_agreement",
			aiAnalysisConfidence: 0.95,
		},
		{
			parties:           []string{enterpriseIDs[1], enterpriseIDs[2]},
			contractHash:      "contract_hash_002_service_agreement",
			aiAnalysisConfidence: 0.87,
		},
	}

	for _, contract := range contracts {
		id := uuid.New()
		query := `
			INSERT INTO contracts (id, parties, contract_hash, ai_analysis_confidence, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (contract_hash) DO NOTHING
		`
		now := time.Now()
		_, err := s.postgres.DB.Exec(query, id, pq.Array(contract.parties), contract.contractHash, 
			contract.aiAnalysisConfidence, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert contract %s: %w", contract.contractHash, err)
		}
		log.Printf("Seeded contract: %s", contract.contractHash)
	}

	return nil
}

func (s *Seeder) ClearDevelopmentData() error {
	log.Println("Clearing development data...")

	tables := []string{"audit_logs", "milestones", "smart_cheques", "contracts", "enterprises"}
	
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		_, err := s.postgres.DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
		log.Printf("Cleared table: %s", table)
	}

	log.Println("Development data cleared successfully")
	return nil
}