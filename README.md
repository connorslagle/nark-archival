# NARK Archival Relay

A NOSTR archival relay for academic content built using the Khatru framework.

## Overview

NARK (NOSTR Academic Repository Kit) Archival is a specialized NOSTR relay designed for long-term preservation of academic content. It provides a permanent, decentralized storage solution for scholarly communications, research data, and academic discourse on the NOSTR protocol.

The NARK protocol enables researchers to publish and collaborate outside traditional academic institutions, removing gatekeepers and paywalls while maintaining rigorous peer review standards.

## 📚 Documentation

- 📊 **[Protocol Architecture](docs/architecture.md)** - Full system design with component details
- 🔄 **[Protocol Overview](docs/protocol-overview.md)** - Simplified flow diagram
- 💡 **[Use Cases](docs/use-cases.md)** - Real-world scenarios and benefits
- 👥 **[User Perspectives](docs/user-perspectives.md)** - Analysis from each user type's viewpoint
- 🗺️ **[Implementation Roadmap](docs/implementation-roadmap.md)** - Future development plans
- 📋 **[Quick Reference](docs/quick-reference.md)** - Event types and commands
- ✅ **[NIP Compliance](docs/nip-compliance-analysis.md)** - NOSTR protocol compatibility analysis

## Features

### Core Features
- Built on Khatru framework for high performance
- PostgreSQL backend for reliable data persistence
- Archival-focused: deletion requests are rejected to ensure content preservation
- Docker containerization for easy deployment
- Designed specifically for academic content preservation

### Academic Event Support
Supports specialized academic event kinds (31428-31432):
- **Academic Papers** (31428): Research papers with title, abstract, authors
- **Citations** (31429): References between academic works
- **Peer Reviews** (31430): Academic reviews with conflict-of-interest protection
- **Research Data** (31431): Datasets and supplementary materials
- **Academic Discussions** (31432): Scholarly discourse threads

### Content Policies

#### 1. **Event Validation**
- Papers require: title (10+ chars), abstract (50+ chars), subject, and authors
- Reviews require: paper reference, substantial content (100+ chars)
- Citations require: paper reference and context (20+ chars)
- Data requires: type, description (30+ chars), related paper
- Discussions require: reference and meaningful content (50+ chars)

#### 2. **Duplicate Prevention**
- Content-based hashing for papers and research data
- Prevents re-submission of identical content
- Case-insensitive matching

#### 3. **Review Integrity**
- Authors cannot review their own papers
- Co-authors blocked from reviewing
- Requires structured feedback (methodology, strengths, weaknesses)

#### 4. **Rate Limiting**
- Papers: 5 per day per pubkey
- Reviews: 10 per day per pubkey
- Data: 10 per day per pubkey
- Discussions: 50 per hour per pubkey
- General: 100 events per hour

### API Endpoints
- `ws://localhost:3334` - WebSocket relay endpoint
- `http://localhost:3334/health` - Health check endpoint
- `http://localhost:3334/policies` - Policy information endpoint

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/connorslagle/nark-archival
cd nark-archival
```

2. Run with Docker Compose:
```bash
docker-compose up -d
```

The relay will be available at `ws://localhost:3334`

## Testing

### Running Tests

Using Make (recommended):
```bash
# Run all tests
make test

# Generate coverage report
make test-coverage

# Test specific components
make test-policies
make test-relay

# Run integration tests
go test -tags=integration -v ./...
```

Using Go directly:
```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Test specific package
go test -v ./internal/policies
```

### Test Coverage

The test suite includes:
- **Unit tests** for all policy components
- **Integration tests** for full workflow validation
- **Concurrent access tests** for thread safety
- **Mock implementations** for database components

### Continuous Integration

The project uses GitHub Actions for automated testing:
- Runs on every push and pull request
- Tests against PostgreSQL service
- Includes linting and code coverage
- Builds Docker images

## Development

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (if running without Docker)
- Make (optional, for convenience commands)

### Quick Development Setup

```bash
# Clone and enter directory
git clone https://github.com/connorslagle/nark-archival
cd nark-archival

# Install dependencies
go mod download

# Run tests
make test

# Start PostgreSQL with Docker
docker run -d --name nark-postgres \
  -e POSTGRES_USER=nark \
  -e POSTGRES_PASSWORD=narkpass \
  -e POSTGRES_DB=nark_archival \
  -p 5432:5432 \
  postgres:16-alpine

# Run the relay
make run-dev
```

### Project Structure

```
nark-archival/
├── cmd/relay/              # Main application entry point
│   ├── main.go            # Relay server implementation
│   └── main_test.go       # Main package tests
├── internal/
│   └── policies/          # Academic content policies
│       ├── academic_validator.go     # Event validation
│       ├── duplicate_checker.go      # Duplicate prevention
│       ├── review_integrity.go       # Review validation
│       ├── rate_limiter.go          # Rate limiting
│       ├── policies.go              # Policy engine
│       └── *_test.go               # Policy tests
├── .github/workflows/     # CI/CD configuration
├── docker-compose.yml     # Docker composition
├── Dockerfile            # Container definition
├── Makefile             # Development commands
└── go.mod               # Go module definition
```

### Available Make Commands

```bash
make build          # Build the binary
make test           # Run tests
make test-coverage  # Generate coverage report
make run           # Build and run locally
make run-dev       # Run with development settings
make docker-build  # Build Docker image
make docker-up     # Start with Docker Compose
make docker-down   # Stop Docker Compose
make lint          # Run linter
make fmt           # Format code
```

### Configuration

Environment variables:
- `PORT`: Relay listening port (default: 3334)
- `DATABASE_URL`: PostgreSQL connection string

Example:
```bash
export PORT=3334
export DATABASE_URL="postgres://nark:narkpass@localhost:5432/nark_archival?sslmode=disable"
```

## API Examples

### Submit an Academic Paper
```json
{
  "id": "...",
  "pubkey": "...",
  "created_at": 1234567890,
  "kind": 31428,
  "tags": [
    ["title", "Distributed Consensus in Academic Networks"],
    ["abstract", "This paper presents a novel approach to achieving consensus in distributed academic networks..."],
    ["subject", "Computer Science"],
    ["author", "Jane Doe"],
    ["author", "John Smith"],
    ["published_at", "2024-01-15"]
  ],
  "content": "",
  "sig": "..."
}
```

### Submit a Peer Review
```json
{
  "kind": 31430,
  "tags": [
    ["e", "paper-event-id"],
    ["content", "This paper provides a thorough examination of consensus mechanisms..."],
    ["methodology-assessment", "The experimental design is sound..."],
    ["strengths", "Clear presentation, novel approach"],
    ["weaknesses", "Limited evaluation scenarios"],
    ["recommendation", "Accept with minor revisions"]
  ]
}
```

### Check Relay Policies
```bash
curl http://localhost:3334/policies
```

Response:
```json
{
  "rate_limits": {
    "general": "100 events per 1h",
    "papers": "5 per 24h",
    "reviews": "10 per 24h"
  },
  "content_requirements": {
    "papers": ["title (min 10 chars)", "abstract (min 50 chars)", "subject tag", "at least one author"],
    "reviews": ["reference to paper", "substantial content (100+ chars)", "no self-reviews", "structured feedback"]
  },
  "duplicate_prevention": "Active for papers and research data",
  "retention_policy": "Permanent - no deletions allowed"
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines

- Write tests for all new functionality
- Follow Go best practices and idioms
- Use meaningful commit messages
- Update documentation as needed
- Ensure CI passes before merging

## Troubleshooting

### Common Issues

**"go: command not found"**
- Go is not installed or not in your PATH. Reinstall Go and restart your terminal.

**"dial tcp: connect: connection refused"**
- PostgreSQL is not running. Start PostgreSQL or use Docker.

**"duplicate key value violates unique constraint"**
- The event_hashes table might have stale data. Clear it or use a fresh database.

**Rate limit errors during testing**
- Tests might be running too fast. Add delays or use separate pubkeys for each test.

### Debug Mode

For verbose logging:
```bash
DEBUG=true ./relay
```

## License

[License information to be added]

## Acknowledgments

- Built with [Khatru](https://github.com/fiatjaf/khatru) relay framework
- Uses [go-nostr](https://github.com/nbd-wtf/go-nostr) for NOSTR protocol
- PostgreSQL storage via [eventstore](https://github.com/fiatjaf/eventstore)