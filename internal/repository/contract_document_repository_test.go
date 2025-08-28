package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sqlmock.Sqlmock, *ContractDocumentRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	repo := NewContractDocumentRepository(db)
	return &mock, repo
}

func TestContractDocumentRepository_UpdateDocumentFields(t *testing.T) {
	mock, repo := setupMockDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cid := "11111111-1111-1111-1111-111111111111"
	fn := "contract.pdf"
	sz := int64(1024)
	mt := "application/pdf"

	(*mock).ExpectExec(regexp.QuoteMeta(`UPDATE contracts SET
			original_filename = $2,
			file_size = $3,
			mime_type = $4
		WHERE id = $1`)).
		WithArgs(cid, &fn, &sz, &mt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateDocumentFields(ctx, cid, &fn, &sz, &mt); err != nil {
		t.Fatalf("UpdateDocumentFields error: %v", err)
	}
	if err := (*mock).ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestContractDocumentRepository_GetDocumentFields(t *testing.T) {
	mock, repo := setupMockDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cid := "22222222-2222-2222-2222-222222222222"
	rows := sqlmock.NewRows([]string{"original_filename", "file_size", "mime_type"}).
		AddRow("msa.docx", int64(2048), "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	(*mock).ExpectQuery(regexp.QuoteMeta(`SELECT original_filename, file_size, mime_type
		FROM contracts
		WHERE id = $1`)).
		WithArgs(cid).
		WillReturnRows(rows)

	meta, err := repo.GetDocumentFields(ctx, cid)
	if err != nil {
		t.Fatalf("GetDocumentFields error: %v", err)
	}
	if meta.OriginalFilename != "msa.docx" || meta.FileSize != 2048 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if err := (*mock).ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
