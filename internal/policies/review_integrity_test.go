package policies

import (
	"context"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

func TestValidateReviewIntegrity(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryPaperStore()

	// Create a test paper
	paperAuthorPubkey := "author123"
	coAuthorPubkey := "coauthor456"
	paper := &nostr.Event{
		ID:     "paper123",
		Kind:   AcademicPaperKind,
		PubKey: paperAuthorPubkey,
		Tags: nostr.Tags{
			{"title", "Test Paper"},
			{"p", coAuthorPubkey}, // co-author
		},
	}
	store.StoreEvent(paper)

	tests := []struct {
		name    string
		review  *nostr.Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid review from non-author",
			review: &nostr.Event{
				Kind:   AcademicReviewKind,
				PubKey: "reviewer789",
				Tags: nostr.Tags{
					{"e", "paper123"},
					{"content", "This paper presents a thorough analysis of the problem space. The methodology is sound and the results are well-presented. However, the conclusions could be strengthened with additional data."},
					{"methodology-assessment", "Sound experimental design"},
					{"strengths", "Clear presentation, thorough analysis"},
					{"weaknesses", "Limited dataset size"},
					{"recommendation", "Accept with minor revisions"},
				},
			},
			wantErr: false,
		},
		{
			name: "review from paper author",
			review: &nostr.Event{
				Kind:   AcademicReviewKind,
				PubKey: paperAuthorPubkey,
				Tags: nostr.Tags{
					{"e", "paper123"},
					{"content", "Great paper!"},
				},
			},
			wantErr: true,
			errMsg:  "authors cannot review their own papers",
		},
		{
			name: "review from co-author",
			review: &nostr.Event{
				Kind:   AcademicReviewKind,
				PubKey: coAuthorPubkey,
				Tags: nostr.Tags{
					{"e", "paper123"},
					{"content", "Excellent work!"},
				},
			},
			wantErr: true,
			errMsg:  "co-authors cannot review their own papers",
		},
		{
			name: "review without paper reference",
			review: &nostr.Event{
				Kind:   AcademicReviewKind,
				PubKey: "reviewer789",
				Tags: nostr.Tags{
					{"content", "Good paper"},
				},
			},
			wantErr: true,
			errMsg:  "no paper reference found",
		},
		{
			name: "review with insufficient quality",
			review: &nostr.Event{
				Kind:   AcademicReviewKind,
				PubKey: "reviewer789",
				Tags: nostr.Tags{
					{"e", "paper123"},
					{"content", "This paper is okay but needs work."},
					{"rating", "3/5"},
				},
			},
			wantErr: true,
			errMsg:  "review quality insufficient",
		},
		{
			name: "non-review event passes through",
			review: &nostr.Event{
				Kind:   AcademicPaperKind,
				PubKey: "anyone",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReviewIntegrity(ctx, tt.review, store)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReviewIntegrity() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestExtractAuthorsFromEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    *nostr.Event
		expected []string
	}{
		{
			name: "single author",
			event: &nostr.Event{
				PubKey: "author1",
				Tags:   nostr.Tags{},
			},
			expected: []string{"author1"},
		},
		{
			name: "multiple authors with p tags",
			event: &nostr.Event{
				PubKey: "author1",
				Tags: nostr.Tags{
					{"p", "author2"},
					{"p", "author3"},
				},
			},
			expected: []string{"author1", "author2", "author3"},
		},
		{
			name: "authors with author-pubkey tags",
			event: &nostr.Event{
				PubKey: "author1",
				Tags: nostr.Tags{
					{"author-pubkey", "author2"},
					{"author-pubkey", "author3"},
				},
			},
			expected: []string{"author1", "author2", "author3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authors := extractAuthorsFromEvent(tt.event)
			if len(authors) != len(tt.expected) {
				t.Errorf("Expected %d authors, got %d", len(tt.expected), len(authors))
			}
			for i, author := range tt.expected {
				if i >= len(authors) || authors[i] != author {
					t.Errorf("Expected author[%d] = %s, got %s", i, author, authors[i])
				}
			}
		})
	}
}

func TestInMemoryPaperStore(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryPaperStore()

	// Test storing and retrieving event
	event := &nostr.Event{
		ID:     "test123",
		Kind:   AcademicPaperKind,
		PubKey: "author123",
		Tags: nostr.Tags{
			{"p", "coauthor456"},
		},
	}

	store.StoreEvent(event)

	// Test GetEvent
	retrieved, err := store.GetEvent(ctx, "test123")
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}
	if retrieved.ID != event.ID {
		t.Errorf("Expected event ID %s, got %s", event.ID, retrieved.ID)
	}

	// Test GetPaperAuthors
	authors, err := store.GetPaperAuthors(ctx, "test123")
	if err != nil {
		t.Fatalf("Failed to get authors: %v", err)
	}
	if len(authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(authors))
	}

	// Test non-existent event
	_, err = store.GetEvent(ctx, "nonexistent")
	if err != nil {
		t.Error("Expected nil error for non-existent event")
	}
}