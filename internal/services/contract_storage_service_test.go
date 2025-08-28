package services

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalContractStorage_StoreRetrieveDelete(t *testing.T) {
	// Setup temp base dir
	baseDir := t.TempDir()
	store, err := NewLocalContractStorage(baseDir)
	if err != nil {
		t.Fatalf("NewLocalContractStorage error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	contractID := "c-123"
	content := []byte("Hello Contract Document! This is a PDF-like payload but plain text.")
	r := bytes.NewReader(content)
	meta, uri, err := store.Store(ctx, contractID, "testdoc.pdf", r)
	if err != nil {
		t.Fatalf("Store error: %v", err)
	}
	if meta == nil || meta.FileSize != int64(len(content)) || meta.OriginalFilename != "testdoc.pdf" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if meta.MimeType == "" {
		t.Fatalf("expected mime type to be detected")
	}
	if !strings.HasPrefix(uri, "file://") {
		t.Fatalf("unexpected uri: %s", uri)
	}

	// Retrieve and compare content
	rc, gotMeta, err := store.Retrieve(ctx, contractID)
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	defer rc.Close()
	gotContent, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(gotContent, content) {
		t.Fatalf("content mismatch: got %q want %q", string(gotContent), string(content))
	}
	if gotMeta == nil || gotMeta.FileSize != int64(len(content)) {
		// FileSize may be from Stat; ensure non-zero
		t.Fatalf("unexpected gotMeta: %+v", gotMeta)
	}

	// Ensure path exists on disk
	path := filepath.Join(baseDir, "contracts", contractID, "document")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file on disk: %v", err)
	}

	// Delete and ensure gone
	if err := store.Delete(ctx, contractID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat err=%v", err)
	}
}
