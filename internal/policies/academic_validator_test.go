package policies

import (
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func TestValidateAcademicEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   *nostr.Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid paper",
			event: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "A Study on Distributed Systems Performance"},
					{"abstract", "This paper presents a comprehensive analysis of distributed systems performance under various load conditions and network topologies."},
					{"subject", "Computer Science"},
					{"author", "John Doe"},
				},
			},
			wantErr: false,
		},
		{
			name: "paper missing title",
			event: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"abstract", "This paper presents a comprehensive analysis of distributed systems performance under various load conditions."},
					{"subject", "Computer Science"},
					{"author", "John Doe"},
				},
			},
			wantErr: true,
			errMsg:  "missing required tags: title",
		},
		{
			name: "paper with short title",
			event: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "Study"},
					{"abstract", "This paper presents a comprehensive analysis of distributed systems performance under various load conditions."},
					{"subject", "Computer Science"},
					{"author", "John Doe"},
				},
			},
			wantErr: true,
			errMsg:  "title too short",
		},
		{
			name: "paper with short abstract",
			event: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "A Study on Distributed Systems"},
					{"abstract", "Short abstract"},
					{"subject", "Computer Science"},
					{"author", "John Doe"},
				},
			},
			wantErr: true,
			errMsg:  "abstract too short",
		},
		{
			name: "valid citation",
			event: &nostr.Event{
				Kind: AcademicCitationKind,
				Tags: nostr.Tags{
					{"e", "referenced-paper-id"},
					{"context", "This work builds upon the foundational research presented in the referenced paper"},
				},
			},
			wantErr: false,
		},
		{
			name: "citation missing paper reference",
			event: &nostr.Event{
				Kind: AcademicCitationKind,
				Tags: nostr.Tags{
					{"context", "This work builds upon previous research"},
				},
			},
			wantErr: true,
			errMsg:  "must reference a paper",
		},
		{
			name: "valid review",
			event: &nostr.Event{
				Kind: AcademicReviewKind,
				Tags: nostr.Tags{
					{"e", "reviewed-paper-id"},
					{"content", "This paper presents an innovative approach to distributed consensus. The methodology is sound and the experimental results are convincing. The authors have made a significant contribution to the field."},
					{"rating", "4/5"},
				},
			},
			wantErr: false,
		},
		{
			name: "review without paper reference",
			event: &nostr.Event{
				Kind: AcademicReviewKind,
				Tags: nostr.Tags{
					{"content", "Great paper!"},
				},
			},
			wantErr: true,
			errMsg:  "must reference a paper",
		},
		{
			name: "valid research data",
			event: &nostr.Event{
				Kind: AcademicDataKind,
				Tags: nostr.Tags{
					{"e", "related-paper-id"},
					{"data-type", "dataset"},
					{"description", "Experimental results from distributed systems performance testing"},
				},
			},
			wantErr: false,
		},
		{
			name: "data missing type",
			event: &nostr.Event{
				Kind: AcademicDataKind,
				Tags: nostr.Tags{
					{"e", "related-paper-id"},
					{"description", "Some experimental data"},
				},
			},
			wantErr: true,
			errMsg:  "must specify type",
		},
		{
			name: "valid discussion",
			event: &nostr.Event{
				Kind:    AcademicDiscussionKind,
				Content: "I found the methodology section particularly interesting. Have you considered applying this approach to edge computing scenarios?",
				Tags: nostr.Tags{
					{"e", "paper-or-discussion-id"},
				},
			},
			wantErr: false,
		},
		{
			name: "discussion too short",
			event: &nostr.Event{
				Kind:    AcademicDiscussionKind,
				Content: "Nice work!",
				Tags: nostr.Tags{
					{"e", "paper-id"},
				},
			},
			wantErr: true,
			errMsg:  "content too short",
		},
		{
			name: "invalid event kind",
			event: &nostr.Event{
				Kind: 12345,
			},
			wantErr: true,
			errMsg:  "invalid academic event kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAcademicEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAcademicEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestRequireMinimalMetadata(t *testing.T) {
	tests := []struct {
		name    string
		event   *nostr.Event
		wantErr bool
	}{
		{
			name: "event with timestamp",
			event: &nostr.Event{
				Kind:      AcademicPaperKind,
				CreatedAt: nostr.Timestamp(time.Now().Unix()),
				Tags: nostr.Tags{
					{"title", "A Valid Title for Testing"},
					{"abstract", "This is a sufficiently long abstract that meets the minimum character requirement for academic papers."},
					{"subject", "Testing"},
					{"author", "Test Author"},
				},
			},
			wantErr: false,
		},
		{
			name: "event with published_at tag",
			event: &nostr.Event{
				Kind: AcademicPaperKind,
				Tags: nostr.Tags{
					{"title", "A Valid Title for Testing"},
					{"abstract", "This is a sufficiently long abstract that meets the minimum character requirement for academic papers."},
					{"subject", "Testing"},
					{"author", "Test Author"},
					{"published_at", "2024-01-01"},
				},
			},
			wantErr: false,
		},
		{
			name: "event without any timestamp",
			event: &nostr.Event{
				Kind:      AcademicPaperKind,
				CreatedAt: 0,
				Tags: nostr.Tags{
					{"title", "A Valid Title for Testing"},
					{"abstract", "This is a sufficiently long abstract that meets the minimum character requirement for academic papers."},
					{"subject", "Testing"},
					{"author", "Test Author"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireMinimalMetadata(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireMinimalMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}