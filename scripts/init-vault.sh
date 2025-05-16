#!/bin/bash

# Exit on any error
set -e

# Default values
VAULT_ADDR=${VAULT_ADDR:-"http://localhost:8200"}
VAULT_TOKEN=${VAULT_TOKEN:-"dev-token"}
AZURE_STORAGE_ACCOUNT=${AZURE_STORAGE_ACCOUNT:-"devstoreaccount1"}
AZURE_STORAGE_KEY=${AZURE_STORAGE_KEY:-"Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="}
BLOB_STORAGE_URL=${BLOB_STORAGE_URL:-"http://127.0.0.1:10000/devstoreaccount1"}
CONTAINER_NAME=${CONTAINER_NAME:-"files"}

# Wait for Vault to be ready
echo "Waiting for Vault to start..."
until curl -fs "${VAULT_ADDR}/v1/sys/health" > /dev/null 2>&1; do
    echo "Waiting for Vault to become available..."
    sleep 1
done

# Enable the KV secrets engine version 2
echo "Enabling KV secrets engine..."
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-token vault vault secrets enable -path=secret kv-v2 || true

# Store the storage credentials
echo "Storing storage credentials in Vault..."
docker exec -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_TOKEN=dev-token vault vault kv put secret/storage \
    credentials="$(cat << EOF
{
    "account_name": "${AZURE_STORAGE_ACCOUNT}",
    "storage_key": "${AZURE_STORAGE_KEY}",
    "storage_url": "${BLOB_STORAGE_URL}",
    "container_name": "${CONTAINER_NAME}"
}
EOF
)"

echo "Vault initialization complete!" 