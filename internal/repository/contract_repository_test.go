package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestCreateContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractRepository(db)

	now := time.Now()
	c := &models.Contract{
		ID:           "11111111-1111-1111-1111-111111111111",
		Parties:      []string{"A", "B"},
		Status:       "draft",
		ContractType: "service_agreement",
		Version:      "v1",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO contracts (
			id, parties, status, contract_type, version, parent_contract_id,
			expiration_date, renewal_terms, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10
		)`)).
		WithArgs(c.ID, pq.Array(c.Parties), c.Status, c.ContractType, c.Version, nil, c.ExpirationDate, c.RenewalTerms, c.CreatedAt, c.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CreateContract(context.Background(), c); err != nil {
		t.Fatalf("CreateContract error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetContractByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractRepository(db)

	rows := sqlmock.NewRows([]string{"id", "parties", "status", "contract_type", "version", "parent_contract_id", "expiration_date", "renewal_terms", "created_at", "updated_at"}).
		AddRow("11111111-1111-1111-1111-111111111111", pq.Array([]string{"A", "B"}), "draft", "service_agreement", "v1", nil, nil, "", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, parties, status, contract_type, version, parent_contract_id,
		       expiration_date, renewal_terms, created_at, updated_at
		FROM contracts
		WHERE id = $1`)).
		WithArgs("11111111-1111-1111-1111-111111111111").
		WillReturnRows(rows)

	c, err := repo.GetContractByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GetContractByID error: %v", err)
	}
	if c == nil || c.ID == "" || len(c.Parties) != 2 {
		t.Fatalf("unexpected contract result: %+v", c)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateAndDeleteContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractRepository(db)
	now := time.Now()
	c := &models.Contract{
		ID:           "11111111-1111-1111-1111-111111111111",
		Parties:      []string{"A", "B"},
		Status:       "active",
		ContractType: "service_agreement",
		Version:      "v2",
		UpdatedAt:    now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE contracts SET
			parties = $2, status = $3, contract_type = $4, version = $5,
			parent_contract_id = $6, expiration_date = $7, renewal_terms = $8, updated_at = $9
		WHERE id = $1`)).
		WithArgs(c.ID, pq.Array(c.Parties), c.Status, c.ContractType, c.Version, nil, c.ExpirationDate, c.RenewalTerms, c.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateContract(context.Background(), c); err != nil {
		t.Fatalf("UpdateContract error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM contracts WHERE id = $1`)).
		WithArgs(c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteContract(context.Background(), c.ID); err != nil {
		t.Fatalf("DeleteContract error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
