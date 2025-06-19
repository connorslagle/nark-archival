package policies

// NIP-compliant event kinds for NARK protocol

const (
	// Regular Events (1000-9999) - Immutable, should be stored forever
	CitationKind   = 8429 // Academic citations (was 31429)
	ReviewKind     = 8430 // Peer reviews (was 31430)
	DiscussionKind = 8432 // Academic discussions (was 31432)
	QuestionKind   = 8434 // Academic questions (was 31434)

	// Addressable Events (30000-39999) - Replaceable by kind+pubkey+d-tag
	PaperKind         = 31428 // Academic papers (keep as-is)
	DataKind          = 31431 // Research data (keep as-is)
	PaperUpdateKind   = 31433 // Paper version updates
	MentorshipKind    = 31435 // Mentorship offers/requests
	ProposalKind      = 31436 // Funding proposals
	ProgressKind      = 31437 // Progress reports
	CitizenProjKind   = 31438 // Citizen science projects
	MediaSummaryKind  = 31439 // Media summaries

	// Consider using these standard kinds
	LongFormKind       = 30023 // NIP-23 long-form content (for papers)
	BadgeDefinitionKind = 30009 // NIP-58 badge definitions
	BadgeAwardKind     = 8     // NIP-58 badge awards
	ReactionKind       = 7     // NIP-25 reactions
	ZapRequestKind     = 9734  // NIP-57 zap requests
	ZapReceiptKind     = 9735  // NIP-57 zap receipts
)

// IsAcademicKind checks if a kind is an academic event kind
func IsAcademicKind(kind int) bool {
	switch kind {
	case CitationKind, ReviewKind, DiscussionKind, QuestionKind,
		PaperKind, DataKind, PaperUpdateKind, MentorshipKind,
		ProposalKind, ProgressKind, CitizenProjKind, MediaSummaryKind:
		return true
	default:
		return false
	}
}

// IsAddressableKind checks if a kind requires a d-tag
func IsAddressableKind(kind int) bool {
	return kind >= 30000 && kind < 40000
}

// RequiresDTag returns true if the event kind requires a d-tag
func RequiresDTag(kind int) bool {
	switch kind {
	case PaperKind, DataKind, PaperUpdateKind, MentorshipKind,
		ProposalKind, ProgressKind, CitizenProjKind, MediaSummaryKind:
		return true
	default:
		return false
	}
}

// GetKindName returns a human-readable name for the event kind
func GetKindName(kind int) string {
	switch kind {
	case CitationKind:
		return "Citation"
	case ReviewKind:
		return "Peer Review"
	case DiscussionKind:
		return "Discussion"
	case QuestionKind:
		return "Question"
	case PaperKind:
		return "Academic Paper"
	case DataKind:
		return "Research Data"
	case PaperUpdateKind:
		return "Paper Update"
	case MentorshipKind:
		return "Mentorship"
	case ProposalKind:
		return "Funding Proposal"
	case ProgressKind:
		return "Progress Report"
	case CitizenProjKind:
		return "Citizen Science Project"
	case MediaSummaryKind:
		return "Media Summary"
	case LongFormKind:
		return "Long-form Content"
	case BadgeDefinitionKind:
		return "Badge Definition"
	case BadgeAwardKind:
		return "Badge Award"
	case ReactionKind:
		return "Reaction"
	case ZapRequestKind:
		return "Zap Request"
	case ZapReceiptKind:
		return "Zap Receipt"
	default:
		return "Unknown"
	}
}