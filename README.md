# NARK Archival Relay

A NOSTR archival relay for academic content built using the Khatru framework.

## Overview

NARK (NOSTR Academic Repository Kit) Archival is a specialized NOSTR relay designed for long-term preservation of academic content. It provides a permanent, decentralized storage solution for scholarly communications, research data, and academic discourse on the NOSTR protocol.

## Features

- Built on Khatru framework for high performance
- PostgreSQL backend for reliable data persistence
- Archival-focused: deletion requests are rejected to ensure content preservation
- Docker containerization for easy deployment
- Designed specifically for academic content preservation

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/yourusername/nark-archival
cd nark-archival
```

2. Run with Docker Compose:
```bash
docker-compose up -d
```

The relay will be available at `ws://localhost:3334`

## Development

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (if running without Docker)

### Project Structure

```
nark-archival/
├── cmd/relay/          # Main application entry point
├── internal/
│   ├── archive/        # Archival logic and policies
│   ├── policies/       # Content filtering and validation
│   └── storage/        # Database abstraction layer
├── docker-compose.yml  # Docker composition with PostgreSQL
└── Dockerfile         # Container definition
```

### Building from Source

```bash
go mod download
go build -o relay cmd/relay/main.go
./relay
```

## Configuration

Environment variables:
- `PORT`: Relay listening port (default: 3334)
- `DATABASE_URL`: PostgreSQL connection string

## License

[License information to be added]