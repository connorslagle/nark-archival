package policies

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// ContentHasher generates a deterministic hash for academic content
type ContentHasher interface {
	GenerateHash(event *nostr.Event) string
}

// DuplicateChecker checks for duplicate academic content
type DuplicateChecker interface {
	IsDuplicate(ctx context.Context, event *nostr.Event) (bool, error)
	StoreHash(ctx context.Context, event *nostr.Event, hash string) error
}

// DefaultContentHasher implements ContentHasher
type DefaultContentHasher struct{}

// GenerateHash creates a deterministic hash for duplicate detection
func (h *DefaultContentHasher) GenerateHash(event *nostr.Event) string {
	// For papers, hash title + authors + abstract
	if event.Kind == PaperKind {
		return h.hashPaper(event)
	}
	
	// For other types, use content + key tags
	hasher := sha256.New()
	hasher.Write([]byte(event.Content))
	
	// Add relevant tags to hash
	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "title", "abstract", "author", "data-type", "description":
				hasher.Write([]byte(tag[0]))
				hasher.Write([]byte(tag[1]))
			}
		}
	}
	
	return hex.EncodeToString(hasher.Sum(nil))
}

// hashPaper creates a hash specifically for academic papers
func (h *DefaultContentHasher) hashPaper(event *nostr.Event) string {
	hasher := sha256.New()
	
	var title, abstract string
	var authors []string
	
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "title":
			title = strings.ToLower(strings.TrimSpace(tag[1]))
		case "abstract":
			abstract = strings.ToLower(strings.TrimSpace(tag[1]))
		case "author":
			authors = append(authors, strings.ToLower(strings.TrimSpace(tag[1])))
		}
	}
	
	// Hash normalized title
	hasher.Write([]byte(title))
	
	// Hash sorted authors to ensure consistency
	for _, author := range authors {
		hasher.Write([]byte(author))
	}
	
	// Hash first 500 chars of abstract (normalized)
	if len(abstract) > 500 {
		abstract = abstract[:500]
	}
	hasher.Write([]byte(abstract))
	
	return hex.EncodeToString(hasher.Sum(nil))
}

// PreventDuplicatePapers checks if a paper already exists based on content hash
func PreventDuplicatePapers(ctx context.Context, event *nostr.Event, checker DuplicateChecker) error {
	// Only check for papers and data
	if event.Kind != PaperKind && event.Kind != DataKind {
		return nil
	}
	
	isDuplicate, err := checker.IsDuplicate(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}
	
	if isDuplicate {
		if event.Kind == PaperKind {
			return fmt.Errorf("duplicate paper detected: a paper with the same title, authors, and abstract already exists in the archive")
		}
		return fmt.Errorf("duplicate research data detected: this dataset already exists in the archive")
	}
	
	return nil
}

// InMemoryDuplicateChecker is a simple in-memory implementation for testing
type InMemoryDuplicateChecker struct {
	hasher ContentHasher
	hashes map[string]bool
}

// NewInMemoryDuplicateChecker creates a new in-memory duplicate checker
func NewInMemoryDuplicateChecker() *InMemoryDuplicateChecker {
	return &InMemoryDuplicateChecker{
		hasher: &DefaultContentHasher{},
		hashes: make(map[string]bool),
	}
}

// IsDuplicate checks if content hash already exists
func (c *InMemoryDuplicateChecker) IsDuplicate(ctx context.Context, event *nostr.Event) (bool, error) {
	hash := c.hasher.GenerateHash(event)
	return c.hashes[hash], nil
}

// StoreHash stores a content hash
func (c *InMemoryDuplicateChecker) StoreHash(ctx context.Context, event *nostr.Event, hash string) error {
	if hash == "" {
		hash = c.hasher.GenerateHash(event)
	}
	c.hashes[hash] = true
	return nil
}