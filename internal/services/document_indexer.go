package services

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// DocumentIndexer defines minimal indexing operations for contract documents.
type DocumentIndexer interface {
	Index(ctx context.Context, contractID string, r io.Reader) error
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	Remove(ctx context.Context, contractID string) error
}

type SearchResult struct {
	ContractID string
	Score      int
}

// InMemoryDocumentIndexer is a simple inverted index stored in memory.
// Not suitable for production but great for tests/prototyping.
type InMemoryDocumentIndexer struct {
	mu    sync.RWMutex
	index map[string]map[string]int // term -> contractID -> freq
}

func NewInMemoryDocumentIndexer() *InMemoryDocumentIndexer {
	return &InMemoryDocumentIndexer{index: make(map[string]map[string]int)}
}

var tokenRegex = regexp.MustCompile(`[A-Za-z0-9]+`)

func tokenize(reader io.Reader) ([]string, error) {
	s := bufio.NewScanner(reader)
	s.Split(bufio.ScanLines)
	var tokens []string
	for s.Scan() {
		line := s.Text()
		for _, m := range tokenRegex.FindAllString(line, -1) {
			tokens = append(tokens, strings.ToLower(m))
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (i *InMemoryDocumentIndexer) Index(ctx context.Context, contractID string, r io.Reader) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	tokens, err := tokenize(r)
	if err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, t := range tokens {
		m, ok := i.index[t]
		if !ok {
			m = make(map[string]int)
			i.index[t] = m
		}
		m[contractID]++
	}
	return nil
}

func (i *InMemoryDocumentIndexer) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	qTokens := tokenRegex.FindAllString(strings.ToLower(query), -1)
	if len(qTokens) == 0 {
		return nil, nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	scores := make(map[string]int)
	for _, t := range qTokens {
		if postings, ok := i.index[t]; ok {
			for cid, freq := range postings {
				scores[cid] += freq
			}
		}
	}
	res := make([]SearchResult, 0, len(scores))
	for cid, sc := range scores {
		res = append(res, SearchResult{ContractID: cid, Score: sc})
	}
	sort.Slice(res, func(a, b int) bool {
		if res[a].Score == res[b].Score {
			return res[a].ContractID < res[b].ContractID
		}
		return res[a].Score > res[b].Score
	})
	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

func (i *InMemoryDocumentIndexer) Remove(ctx context.Context, contractID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for term, postings := range i.index {
		if _, ok := postings[contractID]; ok {
			delete(postings, contractID)
			if len(postings) == 0 {
				delete(i.index, term)
			}
		}
	}
	return nil
}
