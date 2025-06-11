.PHONY: build run test clean docker-build docker-run setup

# Initial setup
setup:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
	fi
	go mod download
	go mod tidy

# Build the application
build: setup
	go build -o ethereum-validator-api .

# Run the application
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f ethereum-validator-api
	go clean

# Build Docker image
docker-build:
	docker build -t ethereum-validator-api:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 --env ETH_RPC_URL ethereum-validator-api:latest

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run the application with hot reload (requires air)
dev:
	air

# Run Docker container with everything built and running
up:
	make docker-build
	make docker-run

# Stop and remove Docker containers
down:
	docker stop ethereum-validator-api
	docker rm ethereum-validator-api