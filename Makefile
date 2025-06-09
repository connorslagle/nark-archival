.PHONY: build test clean run docker-build docker-up docker-down lint

# Build the binary
build:
	go build -o relay cmd/relay/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f relay
	rm -f coverage.out coverage.html

# Run the relay locally
run: build
	./relay

# Run with specific environment variables
run-dev:
	PORT=3334 DATABASE_URL="postgres://nark:narkpass@localhost:5432/nark_archival?sslmode=disable" go run cmd/relay/main.go

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development setup
setup:
	go mod download
	go mod tidy

# Lint the code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run specific tests
test-policies:
	go test -v ./internal/policies/...

test-relay:
	go test -v ./cmd/relay/...

# Benchmark tests
bench:
	go test -bench=. ./...

# Check for vulnerabilities
vuln-check:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...