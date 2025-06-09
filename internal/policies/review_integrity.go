package policies

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// PaperAuthorStore retrieves paper authors for validation
type PaperAuthorStore interface {
	GetPaperAuthors(ctx context.Context, paperID string) ([]string, error)
	GetEvent(ctx context.Context, id string) (*nostr.Event, error)
}

// ValidateReviewIntegrity ensures reviews are not from paper authors (conflict of interest)
func ValidateReviewIntegrity(ctx context.Context, event *nostr.Event, store PaperAuthorStore) error {
	if event.Kind != AcademicReviewKind {
		return nil
	}

	// Find the paper being reviewed
	var paperID string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "e" {
			paperID = tag[1]
			break
		}
	}

	if paperID == "" {
		return fmt.Errorf("review integrity check failed: no paper reference found")
	}

	// Get the paper event
	paperEvent, err := store.GetEvent(ctx, paperID)
	if err != nil {
		return fmt.Errorf("review integrity check failed: cannot verify paper: %w", err)
	}

	if paperEvent == nil {
		return fmt.Errorf("review integrity check failed: referenced paper not found")
	}

	// Check if reviewer is paper author (conflict of interest)
	if paperEvent.PubKey == event.PubKey {
		return fmt.Errorf("review integrity violation: authors cannot review their own papers (conflict of interest)")
	}

	// Check co-authors
	authors, err := store.GetPaperAuthors(ctx, paperID)
	if err != nil {
		// If we can't get authors, extract from paper event
		authors = extractAuthorsFromEvent(paperEvent)
	}

	for _, authorPubkey := range authors {
		if authorPubkey == event.PubKey {
			return fmt.Errorf("review integrity violation: co-authors cannot review their own papers (conflict of interest)")
		}
	}

	// Validate review has substantial content
	if err := validateReviewQuality(event); err != nil {
		return err
	}

	return nil
}

// extractAuthorsFromEvent gets author pubkeys from paper tags
func extractAuthorsFromEvent(event *nostr.Event) []string {
	var authors []string
	
	// Primary author is the event creator
	authors = append(authors, event.PubKey)
	
	// Look for co-author tags
	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "p": // Co-author pubkey reference
				authors = append(authors, tag[1])
			case "author-pubkey": // Explicit author pubkey
				authors = append(authors, tag[1])
			}
		}
	}
	
	return authors
}

// validateReviewQuality ensures reviews meet quality standards
func validateReviewQuality(event *nostr.Event) error {
	hasMethodology := false
	hasStrengths := false
	hasWeaknesses := false
	hasRecommendation := false
	
	// Check for structured review elements
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "methodology-assessment":
			hasMethodology = true
		case "strengths":
			hasStrengths = true
		case "weaknesses":
			hasWeaknesses = true
		case "recommendation":
			hasRecommendation = true
		}
	}
	
	// Require at least some structured feedback
	structuredElements := 0
	if hasMethodology {
		structuredElements++
	}
	if hasStrengths {
		structuredElements++
	}
	if hasWeaknesses {
		structuredElements++
	}
	if hasRecommendation {
		structuredElements++
	}
	
	if structuredElements < 2 {
		return fmt.Errorf("review quality insufficient: please include at least 2 of the following: methodology-assessment, strengths, weaknesses, or recommendation tags")
	}
	
	return nil
}

// InMemoryPaperStore is a simple implementation for testing
type InMemoryPaperStore struct {
	events  map[string]*nostr.Event
	authors map[string][]string
}

// NewInMemoryPaperStore creates a new in-memory store
func NewInMemoryPaperStore() *InMemoryPaperStore {
	return &InMemoryPaperStore{
		events:  make(map[string]*nostr.Event),
		authors: make(map[string][]string),
	}
}

// GetEvent retrieves an event by ID
func (s *InMemoryPaperStore) GetEvent(ctx context.Context, id string) (*nostr.Event, error) {
	event, ok := s.events[id]
	if !ok {
		return nil, nil
	}
	return event, nil
}

// GetPaperAuthors retrieves paper authors
func (s *InMemoryPaperStore) GetPaperAuthors(ctx context.Context, paperID string) ([]string, error) {
	authors, ok := s.authors[paperID]
	if !ok {
		// Try to get from event
		event, err := s.GetEvent(ctx, paperID)
		if err != nil || event == nil {
			return nil, fmt.Errorf("paper not found")
		}
		return extractAuthorsFromEvent(event), nil
	}
	return authors, nil
}

// StoreEvent stores an event (for testing)
func (s *InMemoryPaperStore) StoreEvent(event *nostr.Event) {
	s.events[event.ID] = event
	if event.Kind == AcademicPaperKind {
		s.authors[event.ID] = extractAuthorsFromEvent(event)
	}
}