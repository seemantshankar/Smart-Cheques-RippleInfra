package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/smart-payment-infrastructure/internal/models"
)

// ContractStorageService defines storage operations for contract documents.
// Implementations can back onto local filesystem, S3, GCS, etc.
type ContractStorageService interface {
	// Store saves the provided content for a contract and returns metadata plus a storage URI.
	Store(ctx context.Context, contractID, originalFilename string, r io.Reader) (*models.DocumentMetadata, string, error)
	// Retrieve returns a ReadCloser to the stored content as well as metadata if available.
	Retrieve(ctx context.Context, contractID string) (io.ReadCloser, *models.DocumentMetadata, error)
	// Delete removes the stored content for the contract, if present.
	Delete(ctx context.Context, contractID string) error
}

// LocalContractStorage implements ContractStorageService on the local filesystem.
// Files are stored at: <baseDir>/contracts/<contractID>/document
// A sidecar metadata file is written as: <baseDir>/contracts/<contractID>/metadata.json (not required for now)
type LocalContractStorage struct {
	baseDir string
}

func NewLocalContractStorage(baseDir string) (*LocalContractStorage, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("baseDir cannot be empty")
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "contracts"), 0o755); err != nil {
		return nil, fmt.Errorf("ensure baseDir: %w", err)
	}
	return &LocalContractStorage{baseDir: baseDir}, nil
}

func (s *LocalContractStorage) docPath(contractID string) string {
	return filepath.Join(s.baseDir, "contracts", contractID, "document")
}

func (s *LocalContractStorage) Store(ctx context.Context, contractID, originalFilename string, r io.Reader) (*models.DocumentMetadata, string, error) {
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	default:
	}

	dir := filepath.Join(s.baseDir, "contracts", contractID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", fmt.Errorf("create contract dir: %w", err)
	}

	// Write to temp then atomically move
	tmpFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		return nil, "", fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	hasher := sha256.New()
	multi := io.MultiWriter(tmpFile, hasher)
	n, err := io.Copy(multi, r)
	if err != nil {
		return nil, "", fmt.Errorf("write content: %w", err)
	}

	// Detect MIME using the first 512 bytes
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, "", fmt.Errorf("seek temp: %w", err)
	}
	buf := make([]byte, 512)
	m, _ := tmpFile.Read(buf)
	mime := http.DetectContentType(buf[:m])

	hashHex := hex.EncodeToString(hasher.Sum(nil))
	meta := &models.DocumentMetadata{
		OriginalFilename: originalFilename,
		FileSize:         n,
		MimeType:         mime,
	}

	finalPath := s.docPath(contractID)
	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		return nil, "", fmt.Errorf("finalize file: %w", err)
	}

	_ = hashHex // reserved for future when we extend metadata to include hash
	uri := fmt.Sprintf("file://%s", finalPath)
	return meta, uri, nil
}

func (s *LocalContractStorage) Retrieve(ctx context.Context, contractID string) (io.ReadCloser, *models.DocumentMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}
	p := s.docPath(contractID)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("document not found for contract %s", contractID)
		}
		return nil, nil, err
	}
	// Best-effort metadata from stat
	st, _ := f.Stat()
	mime := "application/octet-stream"
	buf := make([]byte, 512)
	if n, _ := f.ReadAt(buf, 0); n > 0 {
		mime = http.DetectContentType(buf[:n])
	}
	meta := &models.DocumentMetadata{
		OriginalFilename: st.Name(),
		FileSize:         st.Size(),
		MimeType:         mime,
	}
	return f, meta, nil
}

func (s *LocalContractStorage) Delete(ctx context.Context, contractID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	p := s.docPath(contractID)
	if err := os.Remove(p); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// try to remove directory if empty
	_ = os.Remove(filepath.Dir(p))
	return nil
}
