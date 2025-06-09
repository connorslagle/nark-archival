package policies

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRateLimiter(t *testing.T) {
	ctx := context.Background()
	
	// Create rate limiter with test config
	config := &RateLimitConfig{
		EventsPerWindow: 3,
		WindowDuration:  100 * time.Millisecond,
		KindLimits: map[int]KindLimit{
			AcademicPaperKind: {
				EventsPerWindow: 2,
				WindowDuration:  200 * time.Millisecond,
			},
		},
	}
	
	limiter := NewMemoryRateLimiter(config)
	pubkey := "testpubkey123"

	// Test general rate limit
	t.Run("general rate limit", func(t *testing.T) {
		// First 3 requests should pass
		for i := 0; i < 3; i++ {
			err := limiter.AllowRequest(ctx, pubkey, AcademicDiscussionKind)
			if err != nil {
				t.Errorf("Request %d failed: %v", i+1, err)
			}
		}

		// 4th request should fail
		err := limiter.AllowRequest(ctx, pubkey, AcademicDiscussionKind)
		if err == nil {
			t.Error("Expected rate limit error, got nil")
		}
		if err != nil && !contains(err.Error(), "rate limit exceeded") {
			t.Errorf("Expected rate limit error, got: %v", err)
		}

		// Wait for window to expire
		time.Sleep(150 * time.Millisecond)

		// Should be able to make requests again
		err = limiter.AllowRequest(ctx, pubkey, AcademicDiscussionKind)
		if err != nil {
			t.Errorf("Request after window expired failed: %v", err)
		}
	})

	// Reset for next test
	limiter.Reset(pubkey)

	// Test kind-specific rate limit
	t.Run("kind-specific rate limit", func(t *testing.T) {
		// First 2 paper requests should pass
		for i := 0; i < 2; i++ {
			err := limiter.AllowRequest(ctx, pubkey, AcademicPaperKind)
			if err != nil {
				t.Errorf("Paper request %d failed: %v", i+1, err)
			}
		}

		// 3rd paper request should fail
		err := limiter.AllowRequest(ctx, pubkey, AcademicPaperKind)
		if err == nil {
			t.Error("Expected rate limit error for papers, got nil")
		}
		if err != nil && !contains(err.Error(), "academic papers") {
			t.Errorf("Expected paper-specific error, got: %v", err)
		}

		// Other kinds should still work
		err = limiter.AllowRequest(ctx, pubkey, AcademicReviewKind)
		if err != nil {
			t.Errorf("Review request failed: %v", err)
		}
	})

	// Test multiple pubkeys
	t.Run("multiple pubkeys", func(t *testing.T) {
		limiter.Reset(pubkey)
		pubkey2 := "anotherpubkey456"

		// Both pubkeys should have independent limits
		err := limiter.AllowRequest(ctx, pubkey, AcademicPaperKind)
		if err != nil {
			t.Errorf("First pubkey request failed: %v", err)
		}

		err = limiter.AllowRequest(ctx, pubkey2, AcademicPaperKind)
		if err != nil {
			t.Errorf("Second pubkey request failed: %v", err)
		}
	})
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	// Check general config
	if config.EventsPerWindow != 100 {
		t.Errorf("Expected 100 events per window, got %d", config.EventsPerWindow)
	}
	if config.WindowDuration != time.Hour {
		t.Errorf("Expected 1 hour window, got %v", config.WindowDuration)
	}

	// Check paper limit
	paperLimit, ok := config.KindLimits[AcademicPaperKind]
	if !ok {
		t.Fatal("No limit configured for papers")
	}
	if paperLimit.EventsPerWindow != 5 {
		t.Errorf("Expected 5 papers per window, got %d", paperLimit.EventsPerWindow)
	}
	if paperLimit.WindowDuration != 24*time.Hour {
		t.Errorf("Expected 24 hour window for papers, got %v", paperLimit.WindowDuration)
	}
}

func TestFilterTimestamps(t *testing.T) {
	now := time.Now()
	window := time.Minute

	timestamps := []time.Time{
		now.Add(-2 * time.Minute), // outside window
		now.Add(-30 * time.Second), // inside window
		now.Add(-10 * time.Second), // inside window
		now.Add(-90 * time.Second), // outside window
	}

	filtered := filterTimestamps(timestamps, now, window)
	
	if len(filtered) != 2 {
		t.Errorf("Expected 2 timestamps within window, got %d", len(filtered))
	}
}

func TestGetEventTypeName(t *testing.T) {
	tests := []struct {
		kind     int
		expected string
	}{
		{AcademicPaperKind, "academic papers"},
		{AcademicCitationKind, "citations"},
		{AcademicReviewKind, "peer reviews"},
		{AcademicDataKind, "research data"},
		{AcademicDiscussionKind, "discussions"},
		{99999, "events"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := getEventTypeName(tt.kind)
			if name != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, name)
			}
		})
	}
}

func TestCheckRateLimit(t *testing.T) {
	ctx := context.Background()

	// Test with nil limiter (should pass)
	event := &nostr.Event{
		PubKey: "test",
		Kind:   AcademicPaperKind,
	}
	
	err := CheckRateLimit(ctx, event, nil)
	if err != nil {
		t.Errorf("Expected nil error with nil limiter, got: %v", err)
	}

	// Test with actual limiter
	limiter := NewMemoryRateLimiter(&RateLimitConfig{
		EventsPerWindow: 1,
		WindowDuration:  100 * time.Millisecond,
	})

	err = CheckRateLimit(ctx, event, limiter)
	if err != nil {
		t.Errorf("First request failed: %v", err)
	}

	err = CheckRateLimit(ctx, event, limiter)
	if err == nil {
		t.Error("Expected rate limit error")
	}
}