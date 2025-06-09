package policies

import (
	"context"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func TestPolicyEngine(t *testing.T) {
	ctx := context.Background()
	
	// Create policy engine with test components
	rateLimiter := NewMemoryRateLimiter(&RateLimitConfig{
		EventsPerWindow: 10,
		WindowDuration:  100 * time.Millisecond,
		KindLimits: map[int]KindLimit{
			AcademicPaperKind: {
				EventsPerWindow: 2,
				WindowDuration:  200 * time.Millisecond,
			},
		},
	})
	
	duplicateChecker := NewInMemoryDuplicateChecker()
	paperStore := NewInMemoryPaperStore()
	
	engine := NewPolicyEngine(rateLimiter, duplicateChecker, paperStore)

	t.Run("valid paper passes all policies", func(t *testing.T) {
		paper := &nostr.Event{
			ID:     "paper1",
			PubKey: "author1",
			Kind:   AcademicPaperKind,
			Tags: nostr.Tags{
				{"title", "A Comprehensive Study of Policy Engines"},
				{"abstract", "This paper presents a detailed analysis of policy engines in distributed systems, focusing on their implementation and performance characteristics."},
				{"subject", "Computer Science"},
				{"author", "John Doe"},
			},
			CreatedAt: nostr.Timestamp(1234567890),
		}

		err := engine.ValidateEvent(ctx, paper)
		if err != nil {
			t.Errorf("Valid paper failed validation: %v", err)
		}

		// Post process should also succeed
		err = engine.PostProcessEvent(ctx, paper)
		if err != nil {
			t.Errorf("Post process failed: %v", err)
		}
	})

	t.Run("rate limit enforcement", func(t *testing.T) {
		// Submit multiple papers quickly
		for i := 0; i < 3; i++ {
			paper := &nostr.Event{
				ID:     "paper" + string(rune(i)),
				PubKey: "ratelimited",
				Kind:   AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Paper Number " + string(rune(i))},
					{"abstract", "This is a sufficiently long abstract for paper number " + string(rune(i)) + " to meet the minimum requirements."},
					{"subject", "Testing"},
					{"author", "Test Author"},
				},
				CreatedAt: nostr.Timestamp(1234567890),
			}

			err := engine.ValidateEvent(ctx, paper)
			if i < 2 && err != nil {
				t.Errorf("Paper %d should have passed: %v", i, err)
			}
			if i >= 2 && err == nil {
				t.Errorf("Paper %d should have been rate limited", i)
			}
			if i >= 2 && err != nil && !contains(err.Error(), "rate limit") {
				t.Errorf("Expected rate limit error, got: %v", err)
			}
		}
	})

	t.Run("duplicate prevention", func(t *testing.T) {
		paper1 := &nostr.Event{
			ID:     "dup1",
			PubKey: "author2",
			Kind:   AcademicPaperKind,
			Tags: nostr.Tags{
				{"title", "Unique Paper Title"},
				{"abstract", "This is a unique abstract that has never been submitted before and meets all the requirements."},
				{"subject", "Testing"},
				{"author", "Unique Author"},
			},
			CreatedAt: nostr.Timestamp(1234567890),
		}

		// First submission should pass
		err := engine.ValidateEvent(ctx, paper1)
		if err != nil {
			t.Errorf("First submission failed: %v", err)
		}

		// Store the hash
		engine.PostProcessEvent(ctx, paper1)

		// Duplicate submission should fail
		paper2 := &nostr.Event{
			ID:     "dup2",
			PubKey: "author3",
			Kind:   AcademicPaperKind,
			Tags: nostr.Tags{
				{"title", "Unique Paper Title"}, // Same title
				{"abstract", "This is a unique abstract that has never been submitted before and meets all the requirements."}, // Same abstract
				{"subject", "Different Subject"}, // Different subject doesn't matter
				{"author", "Unique Author"}, // Same author
			},
			CreatedAt: nostr.Timestamp(1234567890),
		}

		err = engine.ValidateEvent(ctx, paper2)
		if err == nil {
			t.Error("Duplicate paper should have failed")
		}
		if err != nil && !contains(err.Error(), "duplicate") {
			t.Errorf("Expected duplicate error, got: %v", err)
		}
	})

	t.Run("review integrity check", func(t *testing.T) {
		// Create a paper
		paper := &nostr.Event{
			ID:     "paper_for_review",
			PubKey: "paper_author",
			Kind:   AcademicPaperKind,
			Tags: nostr.Tags{
				{"title", "Paper for Review Testing"},
				{"abstract", "This paper will be used to test review integrity checks and ensure authors cannot review their own work."},
				{"subject", "Testing"},
				{"author", "Paper Author"},
			},
		}
		paperStore.StoreEvent(paper)

		// Author trying to review own paper
		review := &nostr.Event{
			PubKey: "paper_author", // Same as paper author
			Kind:   AcademicReviewKind,
			Tags: nostr.Tags{
				{"e", "paper_for_review"},
				{"content", "This is an excellent paper with groundbreaking results. The methodology is perfect and there are no flaws whatsoever."},
				{"methodology-assessment", "Perfect"},
				{"strengths", "Everything"},
			},
		}

		err := engine.ValidateEvent(ctx, review)
		if err == nil {
			t.Error("Self-review should have failed")
		}
		if err != nil && !contains(err.Error(), "review policy") {
			t.Errorf("Expected review policy error, got: %v", err)
		}
	})

	t.Run("metadata validation", func(t *testing.T) {
		invalidPaper := &nostr.Event{
			PubKey: "author4",
			Kind:   AcademicPaperKind,
			Tags: nostr.Tags{
				{"title", "Short"}, // Too short
				{"abstract", "Also too short"}, // Too short
				{"subject", "Testing"},
				{"author", "A"}, // Too short
			},
			CreatedAt: 0, // No timestamp
		}

		err := engine.ValidateEvent(ctx, invalidPaper)
		if err == nil {
			t.Error("Invalid metadata should have failed")
		}
		if err != nil && !contains(err.Error(), "metadata policy") {
			t.Errorf("Expected metadata policy error, got: %v", err)
		}
	})
}

func TestPolicyEngineDefaults(t *testing.T) {
	// Test that nil components get replaced with defaults
	engine := NewPolicyEngine(nil, nil, nil)
	
	if engine.rateLimiter == nil {
		t.Error("Rate limiter should have been initialized")
	}
	if engine.duplicateChecker == nil {
		t.Error("Duplicate checker should have been initialized")
	}
	if engine.paperStore == nil {
		t.Error("Paper store should have been initialized")
	}
}

func TestGetPolicyInfo(t *testing.T) {
	engine := NewPolicyEngine(nil, nil, nil)
	info := engine.GetPolicyInfo()

	// Check structure
	rateLimits, ok := info["rate_limits"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limits not found or wrong type")
	}

	general, ok := rateLimits["general"].(string)
	if !ok || general == "" {
		t.Error("general rate limit info missing")
	}

	contentReqs, ok := info["content_requirements"].(map[string]interface{})
	if !ok {
		t.Fatal("content_requirements not found or wrong type")
	}

	papers, ok := contentReqs["papers"].([]string)
	if !ok || len(papers) == 0 {
		t.Error("paper requirements missing")
	}

	dupPrevention, ok := info["duplicate_prevention"].(string)
	if !ok || dupPrevention == "" {
		t.Error("duplicate_prevention info missing")
	}

	retention, ok := info["retention_policy"].(string)
	if !ok || retention == "" {
		t.Error("retention_policy info missing")
	}
}