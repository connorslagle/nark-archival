# NARK Protocol NIP Compliance Analysis

## Overview

This document analyzes how the NARK protocol aligns with existing NOSTR Implementation Possibilities (NIPs) and proposes adjustments to ensure maximum compatibility with the NOSTR ecosystem.

## Event Kind Analysis

### Current NARK Event Kinds (31428-31439)

According to NIP-01, the kind ranges are:
- `30000 <= n < 40000`: **Addressable events** - for each combination of `kind`, `pubkey` and `d` tag, only the latest event MUST be stored

Our chosen range (31428-31439) falls within the addressable events range, which is **partially appropriate** for our use case.

### Recommended Adjustments

1. **Academic Papers (31428)** ✅ **Keep as addressable**
   - Papers can be updated/versioned
   - Use `d` tag for paper identifier
   - Updates replace previous versions

2. **Citations (31429)** ❌ **Should be regular event**
   - Citations should not be replaceable
   - Recommend: Move to kind 1429 (regular range)

3. **Reviews (31430)** ❌ **Should be regular event**
   - Reviews should be immutable once published
   - Recommend: Move to kind 1430 (regular range)

4. **Research Data (31431)** ✅ **Keep as addressable**
   - Datasets can be updated
   - Use `d` tag for dataset identifier

5. **Discussions (31432)** ❌ **Should be regular event**
   - Discussion posts should not be editable
   - Recommend: Move to kind 1432 (regular range)

6. **New Event Types (31433-31439)** - Mixed recommendations

## NIP Compliance Requirements

### NIP-01: Basic Protocol ✅

**Current Compliance:**
- Following event structure
- Using proper tags (`e`, `p`, `a`)
- Implementing relay communication

**Required Adjustments:**
- None - fully compliant

### NIP-09: Event Deletion ⚠️

**Current Implementation:**
- We reject all deletion requests

**Required Adjustments:**
- Return proper error message: `["OK", <event_id>, false, "blocked: deletion not allowed on archival relay"]`
- Document this policy clearly in relay information

### NIP-11: Relay Information ✅

**Current Compliance:**
- Implementing relay information endpoint
- Providing proper metadata

**Required Adjustments:**
- Add `limitation` field:
```json
{
  "limitation": {
    "max_message_length": 65536,
    "max_event_tags": 100,
    "min_prefix": 0,
    "auth_required": false,
    "payment_required": false,
    "created_at_lower_limit": 0,
    "created_at_upper_limit": 0
  }
}
```

### NIP-23: Long-form Content 🔄

**Recommendation:** Consider using NIP-23 for papers instead of custom kinds

**Advantages:**
- Already supported by many clients
- Designed for articles/blog posts
- Supports markdown content
- Has title, summary, image tags

**Proposed Hybrid Approach:**
- Use kind 30023 for paper content
- Add academic-specific tags:
  - `subject`: Academic field
  - `doi`: Digital Object Identifier
  - `keywords`: Research keywords
  - `peer-reviewed`: "true"/"false"

### NIP-25: Reactions ✅

**Integration Opportunity:**
- Use kind 7 reactions for quick paper feedback
- `+` for "interesting/useful"
- `-` for "needs improvement"
- Custom emojis for specific reactions (🔬, 📊, ✅)

### NIP-33: Addressable Events ✅

**Current Compliance:**
- Using addressable events correctly
- Need to add `d` tags properly

**Required Implementation:**
```json
{
  "kind": 31428,
  "tags": [
    ["d", "unique-paper-identifier"],
    ["title", "Paper Title"],
    ["subject", "Computer Science"]
  ]
}
```

### NIP-40: Expiration ⚠️

**Conflict with Archival Nature:**
- We should explicitly reject events with expiration tags
- Return error: `["OK", <event_id>, false, "blocked: expiring events not allowed on archival relay"]`

### NIP-42: Authentication 🔄

**Future Implementation:**
- Could require authentication for publishing
- Useful for verified researcher accounts

### NIP-51: Lists 🎯

**Integration Opportunity:**
- Researchers can create paper collections
- Reading lists for students
- Curated bibliographies

### NIP-57: Lightning Zaps ✅

**Current Support:**
- Already planned for research funding
- Fully compatible with our design

### NIP-58: Badges ✅

**Perfect for Academic Reputation:**
- "Peer Reviewer" badge
- "Published Researcher" badge
- "Top Contributor" badge
- Field-specific expertise badges

### NIP-65: Relay List Metadata 🎯

**Implementation Need:**
- Researchers should specify preferred NARK relays
- Helps with relay discovery

## Proposed Protocol Adjustments

### 1. Revised Event Kinds

```javascript
// Regular Events (immutable)
const CITATION_KIND = 8429;        // Was 31429
const REVIEW_KIND = 8430;          // Was 31430
const DISCUSSION_KIND = 8432;      // Was 31432
const QUESTION_KIND = 8434;        // Was 31434

// Addressable Events (updatable)
const PAPER_KIND = 31428;          // Keep as-is
const DATA_KIND = 31431;           // Keep as-is
const PAPER_UPDATE_KIND = 31433;   // Keep as-is
const MENTORSHIP_KIND = 31435;     // Keep as-is
const PROPOSAL_KIND = 31436;       // Keep as-is
const PROGRESS_KIND = 31437;       // Keep as-is
const CITIZEN_PROJECT_KIND = 31438; // Keep as-is
const MEDIA_SUMMARY_KIND = 31439;  // Keep as-is

// Consider using existing kinds
const LONG_FORM_PAPER = 30023;     // NIP-23 long-form content
const BADGE_DEFINITION = 30009;    // NIP-58 badges
const BADGE_AWARD = 8;             // NIP-58 badge awards
```

### 2. Enhanced Tag Structure

```json
{
  "kind": 31428,
  "tags": [
    // Required addressable event tag
    ["d", "distributed-consensus-2024"],
    
    // NIP-23 compatible tags
    ["title", "Distributed Consensus in Academic Networks"],
    ["summary", "Abstract text here..."],
    ["published_at", "1704119029"],
    
    // Academic-specific tags
    ["subject", "Computer Science", "Distributed Systems"],
    ["author", "Jane Doe", "jane@university.edu", "0000-0001-2345-6789"], // name, email, ORCID
    ["author", "John Smith", "john@institution.org", "0000-0002-3456-7890"],
    ["keywords", "consensus", "byzantine-fault-tolerance", "blockchain"],
    ["license", "CC-BY-4.0"],
    ["doi", "10.1234/example.doi"],
    ["arxiv", "2401.12345"],
    
    // Peer review status
    ["peer-reviewed", "true"],
    ["review-type", "double-blind"],
    
    // Funding acknowledgments
    ["funding", "NSF Grant 12345"],
    ["funding", "Lightning Grant", "50000", "sats"],
    
    // Related content
    ["a", "31431:pubkey:dataset-id"], // Related dataset
    ["e", "previous-version-event-id"], // Previous version
    
    // Media
    ["image", "https://example.com/figure1.png", "1024x768"],
    ["video", "https://example.com/presentation.mp4"]
  ],
  "content": "Full paper content in markdown...",
  "created_at": 1704119029
}
```

### 3. Relay Response Standards

```javascript
// Successful storage
["OK", "event-id", true, ""]

// Validation failures
["OK", "event-id", false, "invalid: missing required tag 'title'"]
["OK", "event-id", false, "invalid: abstract too short (minimum 50 characters)"]
["OK", "event-id", false, "rate-limited: maximum 5 papers per day"]
["OK", "event-id", false, "duplicate: paper with same content hash already exists"]
["OK", "event-id", false, "blocked: deletion not allowed on archival relay"]
["OK", "event-id", false, "blocked: authors cannot review their own papers"]
```

## Implementation Priority

### Phase 1: Core Compliance
1. ✅ Fix event kind assignments
2. ✅ Implement proper NIP-01 tags
3. ✅ Add NIP-11 relay information
4. ✅ Handle NIP-09 deletion requests properly

### Phase 2: Enhanced Features
1. 🔄 Integrate NIP-23 for papers
2. 🔄 Implement NIP-57 zaps
3. 🔄 Add NIP-58 badges
4. 🔄 Support NIP-25 reactions

### Phase 3: Advanced Integration
1. 📋 NIP-51 lists for collections
2. 📋 NIP-42 authentication
3. 📋 NIP-65 relay discovery
4. 📋 NIP-88 for notifications

## Benefits of NIP Compliance

1. **Wider Client Support**: Existing NOSTR clients can display papers
2. **Ecosystem Integration**: Zaps, badges, reactions work out-of-box
3. **Future Proof**: New NIPs will be easier to adopt
4. **Network Effects**: Leverage existing NOSTR infrastructure

## Conclusion

By adjusting our event kinds and ensuring full NIP compliance, NARK can become a seamless part of the NOSTR ecosystem while maintaining its specialized academic features. The key changes are:

1. Move immutable events (citations, reviews, discussions) to regular event kinds
2. Keep updatable content (papers, data, proposals) as addressable events
3. Consider using NIP-23 for paper content with academic extensions
4. Fully implement standard NOSTR tags and responses
5. Leverage existing NIPs for reactions, badges, and payments

This approach gives us the best of both worlds: specialized academic functionality with broad NOSTR compatibility.