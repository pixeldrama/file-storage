.PHONY: help start docker-compose test-api local

help:
	@echo "Available commands:"
	@echo "  make start          - Starts the Go application"
	@echo "  make docker-compose - Starts the docker-compose services in detached mode"
	@echo "  make test-api       - Executes the API tests"
	@echo "  make local          - Starts the Go application with Azurite for local development"

start:
	@echo "Starting Go application..."
	go run main.go

docker-compose:
	@echo "Starting docker-compose services..."
	docker-compose up -d

test-api:
	@echo "Executing API tests..."
	./scripts/test-api.sh

local:
	@echo "Starting Go application with Azurite..."
	USE_AZURITE=true go run main.go 