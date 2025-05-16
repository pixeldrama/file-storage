# File Storage Service

A Go service for asynchronous file uploads and downloads using Azure Blob Storage.

## Features

- Asynchronous file uploads with job tracking
- Azure Blob Storage integration
- Prometheus metrics endpoint
- Environment-based configuration
- Mocked vault integration for development
- Automatic antivirus scanning of uploaded files

## Prerequisites

### Installing Required Tools

#### Mac OS

1. Install Azure CLI:
```bash
brew install azure-cli
```

2. Install Azurite (Azure Storage Emulator):
```bash
npm install -g azurite
```

3. Start Azurite in the background:
```bash
azurite &
```

4. Create the 'files' container in Azurite:
```bash
az storage container create --name files --connection-string 'DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;'
```

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