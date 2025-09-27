# Build Go binaries
build-api:
	@echo "Building..."
	@go build -o main cmd/api/main.go

build-worker:
	@echo "Building..."
	@go build -o worker cmd/worker/main.go

# Run locally
run-api:
	@go run cmd/api/main.go

run-worker:
	@go run cmd/worker/main.go

# Docker commands
docker-build:
    @docker build -t goodfood-be:latest

docker-run:
	@docker compose up --build

docker-down:
	@docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if ! command -v air &> /dev/null; then \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi; \
	air

.PHONY: build run test clean watch docker-run docker-down itest