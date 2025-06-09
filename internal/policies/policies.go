package policies

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// PolicyEngine combines all policy checks for academic events
type PolicyEngine struct {
	rateLimiter      RateLimiter
	duplicateChecker DuplicateChecker
	paperStore       PaperAuthorStore
}

// NewPolicyEngine creates a new policy engine with all validators
func NewPolicyEngine(
	rateLimiter RateLimiter,
	duplicateChecker DuplicateChecker,
	paperStore PaperAuthorStore,
) *PolicyEngine {
	// Use defaults if not provided
	if rateLimiter == nil {
		rateLimiter = NewMemoryRateLimiter(DefaultRateLimitConfig())
	}
	if duplicateChecker == nil {
		duplicateChecker = NewInMemoryDuplicateChecker()
	}
	if paperStore == nil {
		paperStore = NewInMemoryPaperStore()
	}
	
	return &PolicyEngine{
		rateLimiter:      rateLimiter,
		duplicateChecker: duplicateChecker,
		paperStore:       paperStore,
	}
}

// ValidateEvent runs all policy checks on an academic event
func (pe *PolicyEngine) ValidateEvent(ctx context.Context, event *nostr.Event) error {
	// 1. Check rate limits first (least expensive)
	if err := CheckRateLimit(ctx, event, pe.rateLimiter); err != nil {
		return fmt.Errorf("rate limit policy: %w", err)
	}
	
	// 2. Validate event structure and metadata
	if err := RequireMinimalMetadata(event); err != nil {
		return fmt.Errorf("metadata policy: %w", err)
	}
	
	// 3. Check for duplicates (papers and data only)
	if event.Kind == AcademicPaperKind || event.Kind == AcademicDataKind {
		if err := PreventDuplicatePapers(ctx, event, pe.duplicateChecker); err != nil {
			return fmt.Errorf("duplicate prevention: %w", err)
		}
	}
	
	// 4. Validate review integrity (reviews only)
	if event.Kind == AcademicReviewKind {
		if err := ValidateReviewIntegrity(ctx, event, pe.paperStore); err != nil {
			return fmt.Errorf("review policy: %w", err)
		}
	}
	
	return nil
}

// PostProcessEvent handles post-storage operations
func (pe *PolicyEngine) PostProcessEvent(ctx context.Context, event *nostr.Event) error {
	// Store event for future reference (papers for review validation)
	if store, ok := pe.paperStore.(*InMemoryPaperStore); ok {
		store.StoreEvent(event)
	}
	
	// Store content hash for duplicate detection
	if event.Kind == AcademicPaperKind || event.Kind == AcademicDataKind {
		hasher := &DefaultContentHasher{}
		hash := hasher.GenerateHash(event)
		if err := pe.duplicateChecker.StoreHash(ctx, event, hash); err != nil {
			return fmt.Errorf("failed to store content hash: %w", err)
		}
	}
	
	return nil
}

// GetPolicyInfo returns human-readable policy information
func (pe *PolicyEngine) GetPolicyInfo() map[string]interface{} {
	config := DefaultRateLimitConfig()
	
	policies := map[string]interface{}{
		"rate_limits": map[string]interface{}{
			"general": fmt.Sprintf("%d events per %v", config.EventsPerWindow, config.WindowDuration),
			"papers": fmt.Sprintf("%d per %v", 
				config.KindLimits[AcademicPaperKind].EventsPerWindow,
				config.KindLimits[AcademicPaperKind].WindowDuration),
			"reviews": fmt.Sprintf("%d per %v",
				config.KindLimits[AcademicReviewKind].EventsPerWindow,
				config.KindLimits[AcademicReviewKind].WindowDuration),
			"data": fmt.Sprintf("%d per %v",
				config.KindLimits[AcademicDataKind].EventsPerWindow,
				config.KindLimits[AcademicDataKind].WindowDuration),
			"discussions": fmt.Sprintf("%d per %v",
				config.KindLimits[AcademicDiscussionKind].EventsPerWindow,
				config.KindLimits[AcademicDiscussionKind].WindowDuration),
		},
		"content_requirements": map[string]interface{}{
			"papers": []string{
				"title (min 10 chars)",
				"abstract (min 50 chars)", 
				"subject tag",
				"at least one author",
			},
			"reviews": []string{
				"reference to paper",
				"substantial content (100+ chars)",
				"no self-reviews",
				"structured feedback",
			},
			"citations": []string{
				"reference to cited paper",
				"context (min 20 chars)",
			},
			"data": []string{
				"data-type tag",
				"description (min 30 chars)",
				"reference to related paper",
			},
			"discussions": []string{
				"reference to paper or parent",
				"content (min 50 chars)",
			},
		},
		"duplicate_prevention": "Active for papers and research data",
		"retention_policy": "Permanent - no deletions allowed",
	}
	
	return policies
}