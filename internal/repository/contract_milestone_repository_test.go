package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestCreateMilestone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractMilestoneRepository(db)
	now := time.Now()
	est := time.Hour * 24
	m := &models.ContractMilestone{
		ID:                   "22222222-2222-2222-2222-222222222222",
		ContractID:           "11111111-1111-1111-1111-111111111111",
		MilestoneID:          "m1",
		SequenceOrder:        1,
		TriggerConditions:    "on_acceptance",
		VerificationCriteria: "evidence_uploaded",
		EstimatedDuration:    est,
		ActualDuration:       nil,
		RiskLevel:            "low",
		CriticalityScore:     10,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO contract_milestones (
			id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
			estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)`)).
		WithArgs(m.ID, m.ContractID, m.MilestoneID, m.SequenceOrder, m.TriggerConditions, m.VerificationCriteria, int64(m.EstimatedDuration), nil, m.RiskLevel, m.CriticalityScore, m.CreatedAt, m.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CreateMilestone(context.Background(), m); err != nil {
		t.Fatalf("CreateMilestone error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetMilestoneByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractMilestoneRepository(db)

	now := time.Now()
	est := int64((time.Hour * 48))
	rows := sqlmock.NewRows([]string{"id", "contract_id", "milestone_id", "sequence_order", "trigger_conditions", "verification_criteria", "estimated_duration", "actual_duration", "risk_level", "criticality_score", "created_at", "updated_at"}).
		AddRow("22222222-2222-2222-2222-222222222222", "11111111-1111-1111-1111-111111111111", "m1", 1, "on_acceptance", "evidence_uploaded", est, nil, "low", 10, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
		       estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE id = $1`)).
		WithArgs("22222222-2222-2222-2222-222222222222").
		WillReturnRows(rows)

	m, err := repo.GetMilestoneByID(context.Background(), "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("GetMilestoneByID error: %v", err)
	}
	if m == nil || m.ID == "" || m.EstimatedDuration != time.Duration(est) || m.ActualDuration != nil {
		t.Fatalf("unexpected milestone result: %+v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateAndDeleteMilestone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractMilestoneRepository(db)
	now := time.Now()
	est := time.Hour * 12
	act := time.Hour * 10
	m := &models.ContractMilestone{
		ID:                   "22222222-2222-2222-2222-222222222222",
		SequenceOrder:        2,
		TriggerConditions:    "on_delivery",
		VerificationCriteria: "accepted",
		EstimatedDuration:    est,
		ActualDuration:       &act,
		RiskLevel:            "medium",
		CriticalityScore:     50,
		UpdatedAt:            now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE contract_milestones SET
			sequence_order = $2, trigger_conditions = $3, verification_criteria = $4,
			estimated_duration = $5, actual_duration = $6, risk_level = $7, criticality_score = $8,
			updated_at = $9
		WHERE id = $1`)).
		WithArgs(m.ID, m.SequenceOrder, m.TriggerConditions, m.VerificationCriteria, int64(m.EstimatedDuration), int64(*m.ActualDuration), m.RiskLevel, m.CriticalityScore, m.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateMilestone(context.Background(), m); err != nil {
		t.Fatalf("UpdateMilestone error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM contract_milestones WHERE id = $1`)).
		WithArgs(m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteMilestone(context.Background(), m.ID); err != nil {
		t.Fatalf("DeleteMilestone error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetMilestonesByContractID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgresContractMilestoneRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "contract_id", "milestone_id", "sequence_order", "trigger_conditions", "verification_criteria", "estimated_duration", "actual_duration", "risk_level", "criticality_score", "created_at", "updated_at"}).
		AddRow("m-1", "c-1", "m1", 1, "t1", "v1", int64(time.Hour), nil, "low", 5, now, now).
		AddRow("m-2", "c-1", "m2", 2, "t2", "v2", int64(time.Hour*2), int64(time.Hour), "medium", 50, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
		       estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE contract_id = $1
		ORDER BY sequence_order`)).
		WithArgs("c-1").
		WillReturnRows(rows)

	list, err := repo.GetMilestonesByContractID(context.Background(), "c-1")
	if err != nil {
		t.Fatalf("GetMilestonesByContractID error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 milestones, got %d", len(list))
	}
	if list[0].EstimatedDuration != time.Hour || list[0].ActualDuration != nil {
		t.Fatalf("unexpected first milestone: %+v", list[0])
	}
	if list[1].EstimatedDuration != time.Hour*2 || list[1].ActualDuration == nil || *list[1].ActualDuration != time.Hour {
		t.Fatalf("unexpected second milestone: %+v", list[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
