.PHONY: help start docker-compose test-api local setup-azure setup-azurite setup setup-vault vault-init setup-db migrate

help:
	@echo "Available commands:"
	@echo "  make start          - Starts the Go application"
	@echo "  make docker-compose - Starts the docker-compose services in detached mode"
	@echo "  make test-api       - Executes the API tests"
	@echo "  make local          - Starts the Go application with Azurite for local development"
	@echo "  make setup-azure    - Ensures Azure CLI container is running"
	@echo "  make setup-azurite  - Ensures Azurite container is running and creates the 'files' container"
	@echo "  make setup-vault    - Ensures Vault container is running"
	@echo "  make vault-init     - Initializes Vault with storage credentials"
	@echo "  make setup          - Runs all setup targets (azure, azurite, vault, db)"
	@echo "  make setup-db       - Ensures PostgreSQL is running"
	@echo "  make migrate        - Runs database migrations"

start:
	@echo "Starting Go application..."
	go run main.go

docker-compose:
	@echo "Starting docker-compose services..."
	docker-compose up -d

test-api:
	@echo "Executing API tests..."
	./scripts/test-api.sh

setup-azure:
	@echo "Ensuring Azure CLI container is running..."
	docker-compose up -d azure-cli

setup-azurite:
	@echo "Ensuring Azurite is running..."
	docker-compose up -d azurite
	@echo "Creating 'files' container in Azurite..."
	sleep 2
	docker exec azure-cli az storage container create --name files --connection-string 'DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://azurite:10000/devstoreaccount1;'

setup-vault:
	@echo "Ensuring Vault is running..."
	docker-compose up -d vault
	@echo "Waiting for Vault to be ready..."
	@until curl -fs "http://localhost:8200/v1/sys/health" > /dev/null 2>&1; do \
		echo "Waiting for Vault to become available..."; \
		sleep 1; \
	done

vault-init:
	@echo "Initializing Vault with storage credentials..."
	./scripts/init-vault.sh

setup-db:
	@echo "Ensuring PostgreSQL is running..."
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; do \
		echo "Waiting for PostgreSQL to become available..."; \
		sleep 1; \
	done

migrate:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go

setup: setup-azure setup-azurite setup-vault vault-init setup-db migrate
	@echo "Setup complete!" 