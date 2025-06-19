# NARK Protocol NIP Compliance Summary

## Changes Made for NIP Compliance

### 1. Event Kind Adjustments

We've restructured our event kinds to properly align with NIP-01 specifications:

#### Regular Events (Immutable)
- **Citation**: 31429 → 8429
- **Review**: 31430 → 8430
- **Discussion**: 31432 → 8432
- **Question**: NEW → 8434

#### Addressable Events (Replaceable by d-tag)
- **Paper**: 31428 (kept)
- **Data**: 31431 (kept)
- **Paper Update**: 31433 (kept)
- **Mentorship**: 31435 (kept)
- **Proposal**: 31436 (kept)
- **Progress**: 31437 (kept)
- **Citizen Project**: 31438 (kept)
- **Media Summary**: 31439 (kept)

### 2. Required Tag Implementation

All addressable events now require a `d` tag:
```json
{
  "kind": 31428,
  "tags": [
    ["d", "unique-paper-identifier"],
    ["title", "Paper Title"],
    ...
  ]
}
```

### 3. Deletion Handling (NIP-09)

Proper rejection of deletion requests:
- Returns: `["OK", event_id, false, "blocked: deletion not allowed on archival relay"]`
- Also rejects deletion request events (kind 5)

### 4. Relay Information (NIP-11)

Added complete relay information:
```json
{
  "name": "NARK Academic Archive",
  "description": "A permanent archival relay for academic content on NOSTR",
  "supported_nips": [1, 9, 11, 25, 40, 57, 58],
  "limitation": {
    "max_message_length": 65536,
    "max_event_tags": 100,
    "max_limit": 500,
    "auth_required": false,
    "payment_required": false
  }
}
```

### 5. Expiration Handling (NIP-40)

Events with expiration tags are rejected:
- Returns: `"blocked: expiring events not allowed on archival relay"`

### 6. Integration Opportunities

The protocol is now compatible with:
- **NIP-25**: Reactions for paper feedback
- **NIP-57**: Lightning zaps for funding
- **NIP-58**: Badges for academic reputation
- **NIP-23**: Long-form content (alternative for papers)

### 7. Validation Functions

Added comprehensive validation for all new event types:
- `validateQuestion()`
- `validatePaperUpdate()`
- `validateMentorship()`
- `validateProposal()`
- `validateProgress()`
- `validateCitizenProject()`
- `validateMediaSummary()`

### 8. Constants File

Created `nip_compliant_kinds.go` with:
- All event kind constants
- Helper functions (`IsAcademicKind`, `RequiresDTag`, `GetKindName`)
- Clear documentation of which kinds are regular vs addressable

## Benefits Achieved

1. **Wider Compatibility**: Any NIP-compliant client can now interact with NARK relays
2. **Future Proof**: Following standards makes adopting new NIPs easier
3. **Clear Semantics**: Using proper event ranges clarifies behavior
4. **Ecosystem Integration**: Can leverage existing tools for reactions, payments, badges

## Migration Path

For existing implementations:
1. Map old event kinds to new ones
2. Add d-tags to addressable events
3. Update validation logic
4. Test with standard NOSTR clients

## Next Steps

1. Consider using NIP-23 for paper content
2. Implement NIP-42 for authenticated publishing
3. Add NIP-65 relay list support
4. Integrate NIP-51 for paper collections

The NARK protocol is now fully compliant with core NOSTR standards while maintaining its specialized academic features.