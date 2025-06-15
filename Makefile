# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=saml-server
BINARY_UNIX=$(BINARY_NAME)_unix

# Main targets
.PHONY: all build clean test coverage deps run dev docker-up docker-down help

all: test build

## Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server

## Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/server

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Run the application
run: build
	./$(BINARY_NAME)

## Run in development mode (with auto-reload if you have air installed)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running normally..."; \
		$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server && ./$(BINARY_NAME); \
	fi

## Start database with Docker Compose
docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

## Stop database
docker-down:
	docker-compose -f deployments/docker-compose.yml down

## Format code
fmt:
	$(GOCMD) fmt ./...

## Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Generate SSL certificates for SAML
certs:
	@echo "Generating SSL certificates..."
	openssl req -x509 -newkey rsa:2048 -keyout sp.key -out sp.crt -days 365 -nodes \
		-subj "/C=US/ST=CA/L=San Francisco/O=SAML POC/CN=localhost"

## Setup project (install deps, generate certs, start database)
setup: deps certs docker-up
	@echo "Project setup complete!"
	@echo "Run 'make run' to start the server"

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) printf "  %-20s%s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST) 