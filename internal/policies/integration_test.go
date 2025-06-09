// +build integration

package policies

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// TestPolicyIntegration tests the full policy flow
func TestPolicyIntegration(t *testing.T) {
	ctx := context.Background()
	
	// Create a full policy engine
	engine := NewPolicyEngine(nil, nil, nil)
	
	// Simulate a paper submission workflow
	authorPubkey := "author_integration_test"
	
	// 1. Submit a valid paper
	paper := &nostr.Event{
		ID:        "paper_integration_1",
		PubKey:    authorPubkey,
		Kind:      AcademicPaperKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"title", "Integration Testing for Academic Relays"},
			{"abstract", "This paper demonstrates the importance of integration testing in academic relay systems. We present a comprehensive testing framework that validates all policy components working together."},
			{"subject", "Software Engineering"},
			{"author", "Integration Tester"},
			{"author-pubkey", "coauthor_pubkey"},
		},
	}
	
	err := engine.ValidateEvent(ctx, paper)
	if err != nil {
		t.Fatalf("Valid paper failed: %v", err)
	}
	
	// Store the paper for future reference
	err = engine.PostProcessEvent(ctx, paper)
	if err != nil {
		t.Fatalf("Failed to post-process paper: %v", err)
	}
	
	// 2. Try to submit a duplicate (should fail)
	duplicatePaper := &nostr.Event{
		ID:        "paper_integration_2",
		PubKey:    "different_author",
		Kind:      AcademicPaperKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"title", "Integration Testing for Academic Relays"}, // Same title
			{"abstract", "This paper demonstrates the importance of integration testing in academic relay systems. We present a comprehensive testing framework that validates all policy components working together."}, // Same abstract
			{"subject", "Computer Science"}, // Different subject
			{"author", "Integration Tester"}, // Same author name
		},
	}
	
	err = engine.ValidateEvent(ctx, duplicatePaper)
	if err == nil {
		t.Error("Duplicate paper should have been rejected")
	}
	if !contains(err.Error(), "duplicate") {
		t.Errorf("Expected duplicate error, got: %v", err)
	}
	
	// 3. Submit a citation
	citation := &nostr.Event{
		ID:        "citation_integration_1",
		PubKey:    "researcher123",
		Kind:      AcademicCitationKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"e", paper.ID},
			{"context", "This groundbreaking work on integration testing provides the foundation for our research"},
		},
	}
	
	err = engine.ValidateEvent(ctx, citation)
	if err != nil {
		t.Errorf("Valid citation failed: %v", err)
	}
	
	// 4. Submit a review from a non-author (should succeed)
	review := &nostr.Event{
		ID:        "review_integration_1",
		PubKey:    "reviewer_external",
		Kind:      AcademicReviewKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"e", paper.ID},
			{"content", "This paper provides a thorough examination of integration testing in academic relay systems. The methodology is sound and the results are convincing. The authors have made a significant contribution to the field of distributed academic infrastructure."},
			{"methodology-assessment", "The testing framework is well-designed and comprehensive"},
			{"strengths", "Clear presentation, thorough coverage of edge cases, practical implementation"},
			{"weaknesses", "Limited performance benchmarks, could benefit from more real-world examples"},
			{"recommendation", "Accept with minor revisions"},
		},
	}
	
	err = engine.ValidateEvent(ctx, review)
	if err != nil {
		t.Errorf("Valid review from non-author failed: %v", err)
	}
	
	// 5. Try author self-review (should fail)
	selfReview := &nostr.Event{
		ID:        "review_integration_2",
		PubKey:    authorPubkey, // Same as paper author
		Kind:      AcademicReviewKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"e", paper.ID},
			{"content", "This is an excellent paper that deserves immediate publication without any changes."},
			{"methodology-assessment", "Perfect"},
			{"strengths", "Everything"},
		},
	}
	
	err = engine.ValidateEvent(ctx, selfReview)
	if err == nil {
		t.Error("Self-review should have been rejected")
	}
	if !contains(err.Error(), "authors cannot review") {
		t.Errorf("Expected self-review error, got: %v", err)
	}
	
	// 6. Submit research data
	data := &nostr.Event{
		ID:        "data_integration_1",
		PubKey:    authorPubkey,
		Kind:      AcademicDataKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Tags: nostr.Tags{
			{"e", paper.ID},
			{"data-type", "dataset"},
			{"description", "Test results and benchmarks from the integration testing framework"},
			{"url", "https://example.com/data/integration-tests.csv"},
		},
	}
	
	err = engine.ValidateEvent(ctx, data)
	if err != nil {
		t.Errorf("Valid research data failed: %v", err)
	}
	
	// 7. Start a discussion thread
	discussion := &nostr.Event{
		ID:        "discussion_integration_1",
		PubKey:    "academic_user",
		Kind:      AcademicDiscussionKind,
		CreatedAt: nostr.Timestamp(time.Now().Unix()),
		Content:   "I'm curious about how this integration testing framework handles edge cases in distributed environments. Have you considered Byzantine failures?",
		Tags: nostr.Tags{
			{"e", paper.ID},
		},
	}
	
	err = engine.ValidateEvent(ctx, discussion)
	if err != nil {
		t.Errorf("Valid discussion failed: %v", err)
	}
	
	// 8. Test rate limiting by submitting many papers quickly
	for i := 0; i < 10; i++ {
		spamPaper := &nostr.Event{
			ID:        fmt.Sprintf("spam_paper_%d", i),
			PubKey:    "spammer",
			Kind:      AcademicPaperKind,
			CreatedAt: nostr.Timestamp(time.Now().Unix()),
			Tags: nostr.Tags{
				{"title", fmt.Sprintf("Spam Paper Number %d", i)},
				{"abstract", fmt.Sprintf("This is spam paper number %d with a sufficiently long abstract to pass validation but trigger rate limiting.", i)},
				{"subject", "Spam"},
				{"author", "Spammer"},
			},
		}
		
		err := engine.ValidateEvent(ctx, spamPaper)
		if i < 5 && err != nil {
			t.Errorf("Paper %d should have passed rate limit: %v", i, err)
		}
		if i >= 5 && err == nil {
			t.Errorf("Paper %d should have been rate limited", i)
		}
	}
}

// TestConcurrentPolicyValidation tests thread safety
func TestConcurrentPolicyValidation(t *testing.T) {
	ctx := context.Background()
	engine := NewPolicyEngine(nil, nil, nil)
	
	// Run multiple validations concurrently
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(index int) {
			paper := &nostr.Event{
				ID:        fmt.Sprintf("concurrent_paper_%d", index),
				PubKey:    fmt.Sprintf("author_%d", index),
				Kind:      AcademicPaperKind,
				CreatedAt: nostr.Timestamp(time.Now().Unix()),
				Tags: nostr.Tags{
					{"title", fmt.Sprintf("Concurrent Paper %d", index)},
					{"abstract", fmt.Sprintf("This is concurrent paper number %d with enough content to meet minimum requirements for testing thread safety.", index)},
					{"subject", "Concurrency"},
					{"author", fmt.Sprintf("Author %d", index)},
				},
			}
			
			err := engine.ValidateEvent(ctx, paper)
			if err != nil {
				t.Errorf("Concurrent validation %d failed: %v", index, err)
			}
			
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}