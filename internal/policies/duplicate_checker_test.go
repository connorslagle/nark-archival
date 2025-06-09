package policies

import (
	"context"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

func TestContentHasher(t *testing.T) {
	hasher := &DefaultContentHasher{}

	tests := []struct {
		name      string
		event1    *nostr.Event
		event2    *nostr.Event
		sameHash  bool
	}{
		{
			name: "identical papers should have same hash",
			event1: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Distributed Systems Study"},
					{"abstract", "A comprehensive study of distributed systems"},
					{"author", "John Doe"},
				},
			},
			event2: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Distributed Systems Study"},
					{"abstract", "A comprehensive study of distributed systems"},
					{"author", "John Doe"},
				},
			},
			sameHash: true,
		},
		{
			name: "papers with different titles should have different hash",
			event1: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Distributed Systems Study"},
					{"abstract", "A comprehensive study"},
					{"author", "John Doe"},
				},
			},
			event2: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Cloud Computing Study"},
					{"abstract", "A comprehensive study"},
					{"author", "John Doe"},
				},
			},
			sameHash: false,
		},
		{
			name: "case differences should be normalized",
			event1: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "DISTRIBUTED SYSTEMS"},
					{"abstract", "A Study"},
					{"author", "JOHN DOE"},
				},
			},
			event2: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "distributed systems"},
					{"abstract", "a study"},
					{"author", "john doe"},
				},
			},
			sameHash: true,
		},
		{
			name: "different event types with same content have different hashes",
			event1: &nostr.Event{
				Kind:    AcademicPaperKind,
				Content: "Test content",
			},
			event2: &nostr.Event{
				Kind:    AcademicDataKind,
				Content: "Test content",
			},
			sameHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hasher.GenerateHash(tt.event1)
			hash2 := hasher.GenerateHash(tt.event2)

			if tt.sameHash && hash1 != hash2 {
				t.Errorf("Expected same hash, got different: %s != %s", hash1, hash2)
			}
			if !tt.sameHash && hash1 == hash2 {
				t.Errorf("Expected different hash, got same: %s", hash1)
			}
		})
	}
}

func TestInMemoryDuplicateChecker(t *testing.T) {
	ctx := context.Background()
	checker := NewInMemoryDuplicateChecker()

	event1 := &nostr.Event{
		ID:   "event1",
		Kind: AcademicPaperKind,
		Tags: nostr.Tags{
			{"title", "Test Paper"},
			{"abstract", "This is a test abstract for duplicate detection"},
			{"author", "Test Author"},
		},
	}

	event2 := &nostr.Event{
		ID:   "event2",
		Kind: AcademicPaperKind,
		Tags: nostr.Tags{
			{"title", "Test Paper"},
			{"abstract", "This is a test abstract for duplicate detection"},
			{"author", "Test Author"},
		},
	}

	event3 := &nostr.Event{
		ID:   "event3",
		Kind: AcademicPaperKind,
		Tags: nostr.Tags{
			{"title", "Different Paper"},
			{"abstract", "This is a different abstract"},
			{"author", "Another Author"},
		},
	}

	// Test initial state - no duplicates
	isDup, err := checker.IsDuplicate(ctx, event1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isDup {
		t.Error("Expected no duplicate for first event")
	}

	// Store first event
	if err := checker.StoreHash(ctx, event1, ""); err != nil {
		t.Fatalf("Failed to store hash: %v", err)
	}

	// Check duplicate with same content
	isDup, err = checker.IsDuplicate(ctx, event2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !isDup {
		t.Error("Expected duplicate for event with same content")
	}

	// Check non-duplicate
	isDup, err = checker.IsDuplicate(ctx, event3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if isDup {
		t.Error("Expected no duplicate for different content")
	}
}

func TestPreventDuplicatePapers(t *testing.T) {
	ctx := context.Background()
	checker := NewInMemoryDuplicateChecker()

	paper := &nostr.Event{
		Kind: AcademicPaperKind,
		Tags: nostr.Tags{
			{"title", "Test Paper"},
			{"abstract", "A sufficiently long abstract for testing duplicate detection in academic papers"},
			{"author", "Author Name"},
		},
	}

	// First submission should pass
	err := PreventDuplicatePapers(ctx, paper, checker)
	if err != nil {
		t.Errorf("First paper submission failed: %v", err)
	}

	// Store the hash
	checker.StoreHash(ctx, paper, "")

	// Second submission should fail
	err = PreventDuplicatePapers(ctx, paper, checker)
	if err == nil {
		t.Error("Expected error for duplicate paper")
	}
	if err != nil && !contains(err.Error(), "duplicate paper detected") {
		t.Errorf("Expected duplicate paper error, got: %v", err)
	}

	// Non-paper events should pass through
	review := &nostr.Event{
		Kind: AcademicReviewKind,
		Tags: nostr.Tags{
			{"e", "some-paper-id"},
		},
	}
	err = PreventDuplicatePapers(ctx, review, checker)
	if err != nil {
		t.Errorf("Non-paper event failed: %v", err)
	}
}