package policies

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

// Event kinds are now defined in nip_compliant_kinds.go
// Using the NIP-compliant constants from that file

// ValidateAcademicEvent verifies required tags based on event kind
func ValidateAcademicEvent(event *nostr.Event) error {
	// First check if it's an academic event
	if !IsAcademicKind(event.Kind) {
		return fmt.Errorf("invalid academic event kind: %d (%s)", event.Kind, GetKindName(event.Kind))
	}

	// Check for d-tag on addressable events
	if RequiresDTag(event.Kind) {
		hasDTag := false
		for _, tag := range event.Tags {
			if len(tag) >= 2 && tag[0] == "d" && tag[1] != "" {
				hasDTag = true
				break
			}
		}
		if !hasDTag {
			return fmt.Errorf("%s requires a 'd' tag for addressable event identification", GetKindName(event.Kind))
		}
	}

	// Validate based on specific kind
	switch event.Kind {
	case PaperKind:
		return validatePaper(event)
	case CitationKind:
		return validateCitation(event)
	case ReviewKind:
		return validateReview(event)
	case DataKind:
		return validateData(event)
	case DiscussionKind:
		return validateDiscussion(event)
	case QuestionKind:
		return validateQuestion(event)
	case PaperUpdateKind:
		return validatePaperUpdate(event)
	case MentorshipKind:
		return validateMentorship(event)
	case ProposalKind:
		return validateProposal(event)
	case ProgressKind:
		return validateProgress(event)
	case CitizenProjKind:
		return validateCitizenProject(event)
	case MediaSummaryKind:
		return validateMediaSummary(event)
	default:
		return fmt.Errorf("validation not implemented for kind %d", event.Kind)
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

// validateQuestion ensures questions are properly formatted
func validateQuestion(event *nostr.Event) error {
	hasReference := false
	hasQuestionType := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e", "a":
				hasReference = true
			case "question-type":
				hasQuestionType = true
			}
		}
	}

	if !hasReference {
		return fmt.Errorf("question must reference a paper or discussion: missing 'e' or 'a' tag")
	}

	if !hasQuestionType {
		return fmt.Errorf("question must specify type: missing 'question-type' tag (methodology/clarification/data/theory)")
	}

	if len(strings.TrimSpace(event.Content)) < 20 {
		return fmt.Errorf("question too short: must be at least 20 characters")
	}

	return nil
}

// validatePaperUpdate ensures paper updates reference the original
func validatePaperUpdate(event *nostr.Event) error {
	hasOriginal := false
	hasVersion := false
	hasChanges := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e":
				hasOriginal = true
			case "version":
				hasVersion = true
			case "changes":
				hasChanges = true
			}
		}
	}

	if !hasOriginal {
		return fmt.Errorf("paper update must reference original: missing 'e' tag to original paper")
	}

	if !hasVersion {
		return fmt.Errorf("paper update must specify version: missing 'version' tag")
	}

	if !hasChanges {
		return fmt.Errorf("paper update must describe changes: missing 'changes' tag")
	}

	return nil
}

// validateMentorship ensures mentorship offers/requests are complete
func validateMentorship(event *nostr.Event) error {
	hasMentorType := false
	hasFields := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "mentor-type":
				hasMentorType = true
				if tag[1] != "offer" && tag[1] != "request" {
					return fmt.Errorf("mentor-type must be 'offer' or 'request', got '%s'", tag[1])
				}
			case "fields":
				hasFields = true
			}
		}
	}

	if !hasMentorType {
		return fmt.Errorf("mentorship must specify type: missing 'mentor-type' tag (offer/request)")
	}

	if !hasFields {
		return fmt.Errorf("mentorship must specify fields: missing 'fields' tag")
	}

	return nil
}

// validateProposal ensures funding proposals have required information
func validateProposal(event *nostr.Event) error {
	hasAmount := false
	hasDuration := false
	hasAbstract := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "funding-amount":
				hasAmount = true
			case "duration":
				hasDuration = true
			case "abstract":
				hasAbstract = true
				if len(tag[1]) < 50 {
					return fmt.Errorf("proposal abstract too short: must be at least 50 characters")
				}
			}
		}
	}

	if !hasAmount {
		return fmt.Errorf("proposal must specify amount: missing 'funding-amount' tag")
	}

	if !hasDuration {
		return fmt.Errorf("proposal must specify duration: missing 'duration' tag")
	}

	if !hasAbstract {
		return fmt.Errorf("proposal must have abstract: missing 'abstract' tag")
	}

	return nil
}

// validateProgress ensures progress reports reference proposals
func validateProgress(event *nostr.Event) error {
	hasProposal := false
	hasMilestone := false
	hasCompletion := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e", "a":
				hasProposal = true
			case "milestone":
				hasMilestone = true
			case "completion":
				hasCompletion = true
			}
		}
	}

	if !hasProposal {
		return fmt.Errorf("progress report must reference proposal: missing 'e' or 'a' tag")
	}

	if !hasMilestone {
		return fmt.Errorf("progress report must specify milestone: missing 'milestone' tag")
	}

	if !hasCompletion {
		return fmt.Errorf("progress report must specify completion: missing 'completion' tag")
	}

	return nil
}

// validateCitizenProject ensures citizen science projects have proper structure
func validateCitizenProject(event *nostr.Event) error {
	hasProjectType := false
	hasRequirements := false
	hasDataFormat := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "project-type":
				hasProjectType = true
			case "requirements":
				hasRequirements = true
			case "data-format":
				hasDataFormat = true
			}
		}
	}

	if !hasProjectType {
		return fmt.Errorf("citizen project must specify type: missing 'project-type' tag")
	}

	if !hasRequirements {
		return fmt.Errorf("citizen project must specify requirements: missing 'requirements' tag")
	}

	if !hasDataFormat {
		return fmt.Errorf("citizen project must specify data format: missing 'data-format' tag")
	}

	return nil
}

// validateMediaSummary ensures media summaries reference papers
func validateMediaSummary(event *nostr.Event) error {
	hasPaperRef := false
	hasSummaryType := false
	hasLanguage := false

	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "e", "a":
				hasPaperRef = true
			case "summary-type":
				hasSummaryType = true
			case "language":
				hasLanguage = true
			}
		}
	}

	if !hasPaperRef {
		return fmt.Errorf("media summary must reference paper: missing 'e' or 'a' tag")
	}

	if !hasSummaryType {
		return fmt.Errorf("media summary must specify type: missing 'summary-type' tag")
	}

	if !hasLanguage {
		return fmt.Errorf("media summary must specify language: missing 'language' tag")
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