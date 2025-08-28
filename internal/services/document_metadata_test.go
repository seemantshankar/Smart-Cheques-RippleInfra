package services

import (
	"bytes"
	"io"
	"testing"
)

func TestExtractFileMetadata(t *testing.T) {
	content := []byte("%PDF-1.4 fake pdf header followed by content")
	meta, hashHex, err := ExtractFileMetadata(bytes.NewReader(content), "contract.pdf")
	if err != nil {
		t.Fatalf("ExtractFileMetadata error: %v", err)
	}
	if meta == nil || meta.OriginalFilename != "contract.pdf" {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if meta.FileSize != int64(len(content)) {
		t.Fatalf("wrong size: got %d want %d", meta.FileSize, len(content))
	}
	if meta.MimeType == "" {
		t.Fatalf("mime not detected")
	}
	if len(hashHex) == 0 {
		t.Fatalf("hash not computed")
	}

	// Ensure idempotent over same content
	meta2, hashHex2, err := ExtractFileMetadata(bytes.NewReader(content), "contract.pdf")
	if err != nil {
		t.Fatalf("ExtractFileMetadata(2) error: %v", err)
	}
	if meta2.FileSize != meta.FileSize || hashHex2 != hashHex {
		t.Fatalf("unexpected second run: meta2=%+v hash2=%s", meta2, hashHex2)
	}

	// Empty content
	meta3, hash3, err := ExtractFileMetadata(bytes.NewReader(nil), "empty.txt")
	if err != nil && err != io.EOF { // function should not error on empty; it returns nil error in our impl
		t.Fatalf("unexpected error for empty: %v", err)
	}
	if meta3 == nil || hash3 == "" {
		t.Fatalf("expected metadata and hash for empty input")
	}
}
