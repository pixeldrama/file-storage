.PHONY: help start docker-compose test-api local setup-azure setup-azurite setup setup-vault vault-init migrate setup-keycloak start-app

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
	@echo "  make migrate        - Runs database migrations"
	@echo "  make setup-keycloak - Sets up Keycloak realm and client"
	@echo "  make start-app      - Starts the main application"

start:
	@echo "Starting Go application..."
	go run main.go

docker-compose:
	@echo "Starting docker-compose services..."
	docker-compose up -d

test-api: start-app
	@echo "Executing API tests..."
	docker-compose run --rm test-api

setup-azure:
	@echo "Ensuring Azure CLI container is running..."
	docker-compose up -d azurite-init

setup-azurite:
	@echo "Ensuring Azurite is running..."
	docker-compose up -d azurite
	@echo "Creating 'files' container in Azurite..."
	sleep 2

setup-vault:
	@echo "Ensuring Vault is running..."
	docker-compose up -d vault
	@echo "Waiting for Vault to be ready..."
	@until docker-compose run --rm curl "curl -v http://vault:8200/v1/sys/health"; do \
		echo "Waiting for Vault to become available..."; \
		sleep 1; \
	done

vault-init:
	@echo "Initializing Vault with storage credentials..."
	./scripts/init-vault.sh

migrate:
	@echo "Running database migrations..."
	docker-compose up db-init

setup-keycloak:
	@echo "Ensuring Keycloak is running..."
	docker-compose up -d keycloak
	@echo "Waiting for Keycloak to be ready..."
	@until docker-compose run --rm curl "curl -s -f http://keycloak:8080/realms/master/.well-known/openid-configuration"; do \
		echo "Waiting for Keycloak to become available..."; \
		sleep 5; \
	done
	@echo "Setting up Keycloak realm and client..."
	docker-compose run --rm -e KEYCLOAK_HOST=keycloak -e KEYCLOAK_PORT=8080 curl "sh /scripts/setup-keycloak.sh"

start-app:
	@echo "Starting main application..."
	docker-compose up -d app
	@echo "Waiting for application to be ready..."
	@until docker-compose run --rm curl "curl -v http://app:8080/health"; do \
		echo "Waiting for application to become available..."; \
		sleep 1; \
	done

setup: setup-azure setup-azurite setup-vault migrate setup-keycloak start-app
	@echo "Setup complete!" 