# File Storage Service

A Go service for asynchronous file uploads and downloads using Azure Blob Storage.

## Prerequisites

- Go (for running the application directly with `make start`)
- Docker and Docker Compose (for running all other commands)

All other tools (Azure CLI, Azurite, etc.) are containerized and don't need to be installed locally.

## Setting Docker Registry

Configure your Docker registry by setting the `DOCKER_PROXY` environment variable:

```bash
# Set your registry URL
export DOCKER_PROXY=your-registry.com

# Configure individual images
export AZURITE_IMAGE=${DOCKER_PROXY}/azure-storage/azurite
export AZURE_CLI_IMAGE=${DOCKER_PROXY}/azure-cli
export VAULT_IMAGE=${DOCKER_PROXY}/hashicorp/vault:1.16.0
export POSTGRES_IMAGE=${DOCKER_PROXY}/postgres:16
export MIGRATE_IMAGE=${DOCKER_PROXY}/migrate/migrate:latest
export KEYCLOAK_IMAGE=${DOCKER_PROXY}/keycloak/keycloak:26.2.4
export ALPINE_IMAGE=${DOCKER_PROXY}/alpine:latest
```

This will automatically configure all Docker images to be pulled from your registry. For example:
- `docker.io/hashicorp/vault:1.16.0` becomes `your-registry.com/hashicorp/vault:1.16.0`
- `quay.io/keycloak/keycloak:26.2.4` becomes `your-registry.com/keycloak/keycloak:26.2.4`

## Quick Start

The easiest way to get started is using the provided Makefile commands:

```bash
# Start all services (database, vault, keycloak, app)
make start-app

# Run the complete setup (includes all prerequisites)
make setup

# Start the application locally with Azurite
make local

# Run API tests
make test-api
```

## Available Commands

```bash
make start          # Starts the Go application directly (requires Go installed)
make docker-compose # Starts all docker-compose services
make test-api      # Executes the API tests
make local         # Starts the Go application with Azurite
make setup-azure   # Ensures Azure CLI container is running
make setup-azurite # Ensures Azurite container is running and creates the 'files' container
make setup-vault   # Ensures Vault container is running
make vault-init    # Initializes Vault with storage credentials
make setup         # Runs all setup targets (azure, azurite, vault, db)
make migrate       # Runs database migrations
make setup-keycloak # Sets up Keycloak realm and client
make start-app     # Starts the main application
```

## Features

- Asynchronous file uploads with job tracking
- Azure Blob Storage integration
- Prometheus metrics endpoint
- Environment-based configuration
- Mocked vault integration for development
- Automatic antivirus scanning of uploaded files

## Configuration

The service can be configured using environment variables:

```bash
export SERVER_PORT=8080
export BLOB_STORAGE_URL="https://your-storage-account.blob.core.windows.net"
export CONTAINER_NAME="files"
export STORAGE_KEY="your-storage-key"  # In production, this will be fetched from vault
```

For local development with Azurite, set:
```bash
export USE_AZURITE=true
```

## Vault Integration

The service now supports HashiCorp Vault for secure storage of credentials. To use Vault:

1. Start the Vault server using Docker Compose:
```bash
docker-compose up -d vault
```

2. Initialize Vault with your storage credentials:
```bash
# Set your storage credentials as environment variables
export AZURE_STORAGE_ACCOUNT="your_account_name"
export AZURE_STORAGE_KEY="your_storage_key"
export BLOB_STORAGE_URL="your_storage_url"
export CONTAINER_NAME="your_container_name"

# Run the initialization script
./scripts/init-vault.sh
```

3. Configure the application to use Vault:
```bash
export VAULT_ADDRESS="http://localhost:8200"
export VAULT_TOKEN="dev-token"  # Use a secure token in production
```

The service will now fetch storage credentials from Vault instead of environment variables.

## API Endpoints

### Create Upload Job
```