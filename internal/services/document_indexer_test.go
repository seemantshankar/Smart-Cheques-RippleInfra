package services

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestInMemoryDocumentIndexer_IndexSearchRemove(t *testing.T) {
	idx := NewInMemoryDocumentIndexer()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c1 := "11111111-1111-1111-1111-111111111111"
	c2 := "22222222-2222-2222-2222-222222222222"

	text1 := "Master Service Agreement for Software Delivery and Support"
	text2 := "Purchase Order and Delivery Schedule for Hardware"

	if err := idx.Index(ctx, c1, bytes.NewReader([]byte(text1))); err != nil {
		t.Fatalf("index c1 error: %v", err)
	}
	if err := idx.Index(ctx, c2, bytes.NewReader([]byte(text2))); err != nil {
		t.Fatalf("index c2 error: %v", err)
	}

	// Query for 'delivery' should find both; c1 may have lower or equal score depending on token counts
	res, err := idx.Search(ctx, "delivery support", 10)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(res) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(res))
	}
	// Ensure contract IDs are among results
	seen := map[string]bool{}
	for _, r := range res {
		seen[r.ContractID] = true
	}
	if !seen[c1] || !seen[c2] {
		t.Fatalf("expected both contracts in results: %+v", res)
	}

	// Limit works
	res2, err := idx.Search(ctx, "delivery", 1)
	if err != nil {
		t.Fatalf("search2 error: %v", err)
	}
	if len(res2) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res2))
	}

	// Remove one contract and ensure it disappears from results
	if err := idx.Remove(ctx, c2); err != nil {
		t.Fatalf("remove error: %v", err)
	}
	res3, err := idx.Search(ctx, strings.ToLower("delivery"), 10)
	if err != nil {
		t.Fatalf("search3 error: %v", err)
	}
	for _, r := range res3 {
		if r.ContractID == c2 {
			t.Fatalf("expected c2 to be removed from index")
		}
	}
}
