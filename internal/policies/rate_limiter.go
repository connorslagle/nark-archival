package policies

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// RateLimiter manages rate limiting per pubkey
type RateLimiter interface {
	AllowRequest(ctx context.Context, pubkey string, eventKind int) error
	Reset(pubkey string)
}

// RateLimitConfig defines rate limit parameters
type RateLimitConfig struct {
	// Events per time window
	EventsPerWindow int
	// Time window duration
	WindowDuration time.Duration
	// Different limits per event kind
	KindLimits map[int]KindLimit
}

// KindLimit defines limits for specific event kinds
type KindLimit struct {
	EventsPerWindow int
	WindowDuration  time.Duration
}

// DefaultRateLimitConfig returns sensible defaults for academic content
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		EventsPerWindow: 100, // General limit
		WindowDuration:  time.Hour,
		KindLimits: map[int]KindLimit{
			AcademicPaperKind: {
				EventsPerWindow: 5, // Max 5 papers per day
				WindowDuration:  24 * time.Hour,
			},
			AcademicReviewKind: {
				EventsPerWindow: 10, // Max 10 reviews per day
				WindowDuration:  24 * time.Hour,
			},
			AcademicDataKind: {
				EventsPerWindow: 10, // Max 10 datasets per day
				WindowDuration:  24 * time.Hour,
			},
			AcademicDiscussionKind: {
				EventsPerWindow: 50, // More lenient for discussions
				WindowDuration:  time.Hour,
			},
		},
	}
}

// MemoryRateLimiter implements in-memory rate limiting
type MemoryRateLimiter struct {
	config   *RateLimitConfig
	mu       sync.RWMutex
	requests map[string]*userRequests
}

type userRequests struct {
	mu         sync.Mutex
	timestamps []time.Time
	kindCounts map[int]*kindRequests
}

type kindRequests struct {
	timestamps []time.Time
}

// NewMemoryRateLimiter creates a new in-memory rate limiter
func NewMemoryRateLimiter(config *RateLimitConfig) *MemoryRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	
	limiter := &MemoryRateLimiter{
		config:   config,
		requests: make(map[string]*userRequests),
	}
	
	// Start cleanup goroutine
	go limiter.cleanup()
	
	return limiter
}

// AllowRequest checks if a request from a pubkey is allowed
func (rl *MemoryRateLimiter) AllowRequest(ctx context.Context, pubkey string, eventKind int) error {
	rl.mu.Lock()
	user, exists := rl.requests[pubkey]
	if !exists {
		user = &userRequests{
			timestamps: make([]time.Time, 0),
			kindCounts: make(map[int]*kindRequests),
		}
		rl.requests[pubkey] = user
	}
	rl.mu.Unlock()
	
	user.mu.Lock()
	defer user.mu.Unlock()
	
	now := time.Now()
	
	// Check general rate limit
	user.timestamps = filterTimestamps(user.timestamps, now, rl.config.WindowDuration)
	if len(user.timestamps) >= rl.config.EventsPerWindow {
		return fmt.Errorf("rate limit exceeded: maximum %d events per %v. Please wait before submitting more content",
			rl.config.EventsPerWindow, rl.config.WindowDuration)
	}
	
	// Check kind-specific rate limit
	if kindLimit, hasLimit := rl.config.KindLimits[eventKind]; hasLimit {
		if user.kindCounts[eventKind] == nil {
			user.kindCounts[eventKind] = &kindRequests{
				timestamps: make([]time.Time, 0),
			}
		}
		
		kindReqs := user.kindCounts[eventKind]
		kindReqs.timestamps = filterTimestamps(kindReqs.timestamps, now, kindLimit.WindowDuration)
		
		if len(kindReqs.timestamps) >= kindLimit.EventsPerWindow {
			return fmt.Errorf("rate limit exceeded for %s: maximum %d per %v. Academic content requires careful review - please pace your submissions",
				getEventTypeName(eventKind), kindLimit.EventsPerWindow, kindLimit.WindowDuration)
		}
		
		// Add timestamp for kind-specific tracking
		kindReqs.timestamps = append(kindReqs.timestamps, now)
	}
	
	// Add timestamp for general tracking
	user.timestamps = append(user.timestamps, now)
	
	return nil
}

// Reset clears rate limit data for a pubkey
func (rl *MemoryRateLimiter) Reset(pubkey string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.requests, pubkey)
}

// cleanup periodically removes old entries
func (rl *MemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		
		// Remove users with no recent activity
		for pubkey, user := range rl.requests {
			user.mu.Lock()
			hasRecent := false
			
			// Check if any timestamps are recent
			for _, ts := range user.timestamps {
				if now.Sub(ts) < 24*time.Hour {
					hasRecent = true
					break
				}
			}
			
			user.mu.Unlock()
			
			if !hasRecent {
				delete(rl.requests, pubkey)
			}
		}
		
		rl.mu.Unlock()
	}
}

// filterTimestamps returns only timestamps within the window
func filterTimestamps(timestamps []time.Time, now time.Time, window time.Duration) []time.Time {
	filtered := make([]time.Time, 0, len(timestamps))
	cutoff := now.Add(-window)
	
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	
	return filtered
}

// getEventTypeName returns human-readable event type names
func getEventTypeName(kind int) string {
	switch kind {
	case AcademicPaperKind:
		return "academic papers"
	case AcademicCitationKind:
		return "citations"
	case AcademicReviewKind:
		return "peer reviews"
	case AcademicDataKind:
		return "research data"
	case AcademicDiscussionKind:
		return "discussions"
	default:
		return "events"
	}
}

// CheckRateLimit is a convenience function for rate limiting
func CheckRateLimit(ctx context.Context, event *nostr.Event, limiter RateLimiter) error {
	if limiter == nil {
		return nil
	}
	
	return limiter.AllowRequest(ctx, event.PubKey, event.Kind)
}