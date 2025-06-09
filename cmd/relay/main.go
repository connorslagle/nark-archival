package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fiatjaf/eventstore/postgresql"
	"github.com/fiatjaf/khatru"
	"github.com/jmoiron/sqlx"
	"github.com/nbd-wtf/go-nostr"
	_ "github.com/lib/pq"
	
	"github.com/yourusername/nark-archival/internal/policies"
)

const (
	// Academic event kinds (NIP-78 range)
	AcademicPaperKind       = 31428
	AcademicCitationKind    = 31429
	AcademicReviewKind      = 31430
	AcademicDataKind        = 31431
	AcademicDiscussionKind  = 31432
)

var academicKinds = []int{
	AcademicPaperKind,
	AcademicCitationKind,
	AcademicReviewKind,
	AcademicDataKind,
	AcademicDiscussionKind,
}

// PostgreSQLPaperStore adapts PostgreSQL backend for policy checks
type PostgreSQLPaperStore struct {
	store *postgresql.PostgresBackend
}

func (ps *PostgreSQLPaperStore) GetEvent(ctx context.Context, id string) (*nostr.Event, error) {
	filter := nostr.Filter{
		IDs: []string{id},
	}
	
	events, err := ps.store.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Get first event from channel
	select {
	case event := <-events:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, nil
	}
}

func (ps *PostgreSQLPaperStore) GetPaperAuthors(ctx context.Context, paperID string) ([]string, error) {
	event, err := ps.GetEvent(ctx, paperID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("paper not found")
	}
	
	var authors []string
	authors = append(authors, event.PubKey)
	
	for _, tag := range event.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "p", "author-pubkey":
				authors = append(authors, tag[1])
			}
		}
	}
	
	return authors, nil
}

// PostgreSQLDuplicateChecker adapts PostgreSQL for duplicate detection
type PostgreSQLDuplicateChecker struct {
	db     *sqlx.DB
	hasher policies.ContentHasher
}

func NewPostgreSQLDuplicateChecker(db *sqlx.DB) *PostgreSQLDuplicateChecker {
	return &PostgreSQLDuplicateChecker{
		db:     db,
		hasher: &policies.DefaultContentHasher{},
	}
}

func (dc *PostgreSQLDuplicateChecker) IsDuplicate(ctx context.Context, event *nostr.Event) (bool, error) {
	hash := dc.hasher.GenerateHash(event)
	
	// Check if hash exists in database
	var count int
	err := dc.db.GetContext(ctx, &count, 
		"SELECT COUNT(*) FROM event_hashes WHERE content_hash = $1", hash)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (dc *PostgreSQLDuplicateChecker) StoreHash(ctx context.Context, event *nostr.Event, hash string) error {
	if hash == "" {
		hash = dc.hasher.GenerateHash(event)
	}
	
	_, err := dc.db.ExecContext(ctx,
		"INSERT INTO event_hashes (event_id, content_hash, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		event.ID, hash, time.Now())
	
	return err
}

func (dc *PostgreSQLDuplicateChecker) Init(ctx context.Context) error {
	// Create table for content hashes
	_, err := dc.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS event_hashes (
			event_id TEXT PRIMARY KEY,
			content_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	
	// Create index separately
	_, err = dc.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_content_hash ON event_hashes(content_hash)
	`)
	return err
}

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nark:narkpass@localhost:5432/nark_archival?sslmode=disable"
	}

	// Connect to PostgreSQL using sqlx
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize PostgreSQL event store
	store := &postgresql.PostgresBackend{
		DB:                db,
		DatabaseURL:       dbURL,
		QueryLimit:        500,
		QueryIDsLimit:     500,
		QueryAuthorsLimit: 500,
		QueryKindsLimit:   10,
		QueryTagsLimit:    10,
	}
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize event store: %v", err)
	}

	// Initialize duplicate checker
	duplicateChecker := NewPostgreSQLDuplicateChecker(db)
	if err := duplicateChecker.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize duplicate checker: %v", err)
	}

	// Create policy engine
	policyEngine := policies.NewPolicyEngine(
		policies.NewMemoryRateLimiter(policies.DefaultRateLimitConfig()),
		duplicateChecker,
		&PostgreSQLPaperStore{store: store},
	)

	// Create Khatru relay
	relay := khatru.NewRelay()

	// Set relay information (NIP-11)
	relay.Info.Name = "NARK Academic Archive"
	relay.Info.Description = "A permanent archival relay for academic content on NOSTR"
	relay.Info.PubKey = ""
	relay.Info.Contact = "admin@nark-archive.org"
	relay.Info.SupportedNIPs = []int{1, 11, 78}
	relay.Info.Software = "https://github.com/yourusername/nark-archival"
	relay.Info.Version = "0.1.0"

	// Configure storage backend with policy enforcement
	relay.StoreEvent = append(relay.StoreEvent, func(ctx context.Context, event *nostr.Event) error {
		// Only accept academic event kinds
		if !isAcademicEvent(event) {
			return fmt.Errorf("only academic events (kinds %v) are accepted", academicKinds)
		}

		// Validate event signature
		if ok, err := event.CheckSignature(); !ok || err != nil {
			return fmt.Errorf("invalid event signature")
		}

		// Run policy validation
		if err := policyEngine.ValidateEvent(ctx, event); err != nil {
			return err
		}

		// Store in PostgreSQL
		if err := store.SaveEvent(ctx, event); err != nil {
			return fmt.Errorf("storage error: %w", err)
		}

		// Post-process (store hashes, update indexes)
		if err := policyEngine.PostProcessEvent(ctx, event); err != nil {
			log.Printf("Post-process error for event %s: %v", event.ID, err)
		}

		return nil
	})

	// Configure event queries
	relay.QueryEvents = append(relay.QueryEvents, func(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
		// Filter to only return academic events
		if len(filter.Kinds) == 0 {
			filter.Kinds = academicKinds
		} else {
			// Ensure only academic kinds are queried
			filteredKinds := []int{}
			for _, kind := range filter.Kinds {
				if isAcademicKind(kind) {
					filteredKinds = append(filteredKinds, kind)
				}
			}
			filter.Kinds = filteredKinds
		}

		return store.QueryEvents(ctx, filter)
	})

	// Implement retention policy - never delete academic events
	relay.DeleteEvent = append(relay.DeleteEvent, func(ctx context.Context, event *nostr.Event) error {
		return fmt.Errorf("deletion not allowed: this is a permanent archival relay for academic content")
	})

	// Count events handler
	relay.CountEvents = append(relay.CountEvents, func(ctx context.Context, filter nostr.Filter) (int64, error) {
		// Filter to only count academic events
		if len(filter.Kinds) == 0 {
			filter.Kinds = academicKinds
		} else {
			filteredKinds := []int{}
			for _, kind := range filter.Kinds {
				if isAcademicKind(kind) {
					filteredKinds = append(filteredKinds, kind)
				}
			}
			filter.Kinds = filteredKinds
		}

		return store.CountEvents(ctx, filter)
	})

	// Add health check endpoint
	relay.Router().HandleFunc("/health", healthCheckHandler)
	
	// Add policy info endpoint
	relay.Router().HandleFunc("/policies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policyEngine.GetPolicyInfo())
	})

	// Get port from environment
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "3334"
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port number: %v", err)
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting NARK Academic Archive relay on :%d", port)
		if err := relay.Start("0.0.0.0", port); err != nil {
			log.Printf("Relay error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, shutting down gracefully...")
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Close database connection
	store.Close()

	log.Println("Shutdown complete")
}

// isAcademicEvent checks if an event is an academic event kind
func isAcademicEvent(event *nostr.Event) bool {
	return isAcademicKind(event.Kind)
}

// isAcademicKind checks if a kind is an academic kind
func isAcademicKind(kind int) bool {
	for _, k := range academicKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// healthCheckHandler returns the health status of the relay
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"nark-archival-relay","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}