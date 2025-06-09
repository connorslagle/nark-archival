package policies

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

const (
	// Academic event kinds
	AcademicPaperKind      = 31428
	AcademicCitationKind   = 31429
	AcademicReviewKind     = 31430
	AcademicDataKind       = 31431
	AcademicDiscussionKind = 31432
)

// ValidateAcademicEvent verifies required tags based on event kind
func ValidateAcademicEvent(event *nostr.Event) error {
	switch event.Kind {
	case AcademicPaperKind:
		return validatePaper(event)
	case AcademicCitationKind:
		return validateCitation(event)
	case AcademicReviewKind:
		return validateReview(event)
	case AcademicDataKind:
		return validateData(event)
	case AcademicDiscussionKind:
		return validateDiscussion(event)
	default:
		return fmt.Errorf("invalid academic event kind: %d. Only kinds 31428-31432 are accepted", event.Kind)
	}
}

// validatePaper ensures papers have required metadata
func validatePaper(event *nostr.Event) error {
	requiredTags := map[string]bool{
		"title":    false,
		"subject":  false,
		"abstract": false,
		"author":   false,
	}

	// Check for required tags
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		if _, ok := requiredTags[tag[0]]; ok {
			requiredTags[tag[0]] = true
		}
	}

	// Build error message for missing tags
	var missing []string
	for tag, found := range requiredTags {
		if !found {
			missing = append(missing, tag)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("academic paper missing required tags: %s. Papers must include title, subject, abstract, and at least one author tag", strings.Join(missing, ", "))
	}

	// Validate tag content
	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "title":
			if len(strings.TrimSpace(tag[1])) < 10 {
				return fmt.Errorf("paper title too short: must be at least 10 characters")
			}
		case "abstract":
			if len(strings.TrimSpace(tag[1])) < 50 {
				return fmt.Errorf("paper abstract too short: must be at least 50 characters to provide meaningful summary")
			}
		case "author":
			if len(strings.TrimSpace(tag[1])) < 3 {
				return fmt.Errorf("author name too short: must be at least 3 characters")
			}
		}
	}

	return nil
}

// validateCitation ensures citations reference valid papers
func validateCitation(event *nostr.Event) error {
	hasCitedPaper := false
	hasContext := false

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "e":
			// Reference to cited paper
			hasCitedPaper = true
		case "context":
			// Citation context
			hasContext = true
			if len(strings.TrimSpace(tag[1])) < 20 {
				return fmt.Errorf("citation context too short: must provide at least 20 characters of context")
			}
		}
	}

	if !hasCitedPaper {
		return fmt.Errorf("citation must reference a paper: missing 'e' tag pointing to the cited paper event")
	}

	if !hasContext {
		return fmt.Errorf("citation must provide context: missing 'context' tag explaining the citation")
	}

	return nil
}

// validateReview ensures reviews are properly linked to papers
func validateReview(event *nostr.Event) error {
	hasReferencedPaper := false
	hasRating := false
	hasContent := false

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "e":
			// Reference to reviewed paper
			hasReferencedPaper = true
		case "rating":
			hasRating = true
		case "content":
			hasContent = true
			if len(strings.TrimSpace(tag[1])) < 100 {
				return fmt.Errorf("review content too short: must provide at least 100 characters of substantive review")
			}
		}
	}

	if !hasReferencedPaper {
		return fmt.Errorf("review must reference a paper: missing 'e' tag pointing to the reviewed paper")
	}

	if !hasRating && !hasContent {
		return fmt.Errorf("review must include either a rating or content review (preferably both)")
	}

	return nil
}

// validateData ensures research data has proper metadata
func validateData(event *nostr.Event) error {
	hasDataType := false
	hasDescription := false
	hasRelatedPaper := false

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		switch tag[0] {
		case "data-type":
			hasDataType = true
		case "description":
			hasDescription = true
			if len(strings.TrimSpace(tag[1])) < 30 {
				return fmt.Errorf("data description too short: must provide at least 30 characters describing the dataset")
			}
		case "e":
			// Reference to related paper
			hasRelatedPaper = true
		}
	}

	if !hasDataType {
		return fmt.Errorf("research data must specify type: missing 'data-type' tag (e.g., 'dataset', 'code', 'supplementary')")
	}

	if !hasDescription {
		return fmt.Errorf("research data must have description: missing 'description' tag")
	}

	if !hasRelatedPaper {
		return fmt.Errorf("research data must reference related paper: missing 'e' tag pointing to associated paper")
	}

	return nil
}

// validateDiscussion ensures discussions are properly threaded
func validateDiscussion(event *nostr.Event) error {
	hasReference := false
	contentLength := len(strings.TrimSpace(event.Content))

	for _, tag := range event.Tags {
		if len(tag) < 2 {
			continue
		}
		if tag[0] == "e" {
			// Reference to paper or parent discussion
			hasReference = true
		}
	}

	if !hasReference {
		return fmt.Errorf("academic discussion must reference a paper or parent discussion: missing 'e' tag")
	}

	if contentLength < 50 {
		return fmt.Errorf("discussion content too short: must provide at least 50 characters for meaningful academic discourse")
	}

	return nil
}

// RequireMinimalMetadata ensures all academic events have basic required metadata
func RequireMinimalMetadata(event *nostr.Event) error {
	// First validate according to specific type
	if err := ValidateAcademicEvent(event); err != nil {
		return err
	}

	// Additional cross-type validations
	hasTimestamp := false
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "published_at" {
			hasTimestamp = true
			break
		}
	}

	if !hasTimestamp && event.CreatedAt == 0 {
		return fmt.Errorf("academic content must have a timestamp: either in created_at field or published_at tag")
	}

	return nil
}