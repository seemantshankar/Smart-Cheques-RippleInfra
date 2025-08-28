package services

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/smart-payment-infrastructure/internal/models"
)

// ExtractFileMetadata reads up to maxRead bytes to detect MIME and computes SHA-256 hash and size by streaming.
// It returns DocumentMetadata and the hex-encoded SHA256 hash.
func ExtractFileMetadata(r io.Reader, originalFilename string) (*models.DocumentMetadata, string, error) {
	// Buffer to detect mime
	const sniff = 512
	peek := make([]byte, sniff)
	h := sha256.New()

	// TeeReader will feed hasher while we copy to a limited buffer for detection
	tee := io.TeeReader(r, h)
	n, err := io.ReadFull(tee, peek)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, "", err
	}
	mime := http.DetectContentType(peek[:n])

	// Continue hashing any remaining bytes
	m, err := io.Copy(h, r)
	if err != nil {
		return nil, "", err
	}
	total := int64(n) + m
	hashHex := hex.EncodeToString(h.Sum(nil))
	meta := &models.DocumentMetadata{
		OriginalFilename: originalFilename,
		FileSize:         total,
		MimeType:         mime,
	}
	return meta, hashHex, nil
}
