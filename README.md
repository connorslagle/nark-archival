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

### Go Installation Guide (For Beginners)

If you're new to Go, here's how to get started:

#### 1. Install Go

**macOS (using Homebrew):**
```bash
brew install go
```

**macOS/Linux (manual installation):**
```bash
# Download and install Go from https://golang.org/dl/
# Or use your system's package manager
```

**Verify Go installation:**
```bash
go version
# Should output something like: go version go1.21.4 darwin/amd64
```

#### 2. Set up your Go environment

Go should automatically set up your workspace, but you can verify:
```bash
go env GOPATH
go env GOROOT
```

#### 3. Clone and compile the project

```bash
# Clone the repository
git clone https://github.com/yourusername/nark-archival
cd nark-archival

# Download dependencies
go mod tidy

# Compile the program
go build -o relay cmd/relay/main.go
```

This creates an executable file called `relay` in your current directory.

#### 4. Set up PostgreSQL (if running locally)

You'll need a PostgreSQL database running. Either:

**Option A: Use Docker for just PostgreSQL:**
```bash
docker run -d \
  --name nark-postgres \
  -e POSTGRES_USER=nark \
  -e POSTGRES_PASSWORD=narkpass \
  -e POSTGRES_DB=nark_archival \
  -p 5432:5432 \
  postgres:16-alpine
```

**Option B: Install PostgreSQL locally:**
```bash
# macOS
brew install postgresql
brew services start postgresql

# Create database and user
psql postgres
CREATE USER nark WITH PASSWORD 'narkpass';
CREATE DATABASE nark_archival OWNER nark;
\q
```

#### 5. Run the compiled program

```bash
# Set environment variables (optional)
export PORT=3334
export DATABASE_URL="postgres://nark:narkpass@localhost:5432/nark_archival?sslmode=disable"

# Run the relay
./relay
```

You should see:
```
Starting NARK Academic Archive relay on :3334
```

#### 6. Test your relay

Open another terminal and test the health endpoint:
```bash
curl http://localhost:3334/health
```

You should get a response like:
```json
{"status":"healthy","service":"nark-archival-relay","timestamp":"2025-06-09T05:01:24Z"}
```

#### Troubleshooting

**"go: command not found"**
- Go is not installed or not in your PATH. Reinstall Go and restart your terminal.

**"dial tcp: connect: connection refused"**
- PostgreSQL is not running. Start PostgreSQL or use the Docker command above.

**"missing go.sum entry"**
- Run `go mod tidy` to download and verify dependencies.

**Permission denied on `./relay`**
- Make the binary executable: `chmod +x relay`

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

### Building from Source (Quick Reference)

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