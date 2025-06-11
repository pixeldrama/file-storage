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

# Check if KV secrets engine is already enabled
if ! curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/sys/mounts/secret" > /dev/null 2>&1; then
    echo "Enabling KV secrets engine..."
    curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/sys/mounts/secret" \
        -d '{"type": "kv", "options": {"version": "2"}}'
fi

# Check if AppRole auth method is already enabled
if ! curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/sys/auth/approle" > /dev/null 2>&1; then
    echo "Enabling AppRole auth method..."
    curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/sys/auth/approle" \
        -d '{"type": "approle"}'
fi

echo "Creating app role..."
curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/auth/approle/role/file-storage" \
    -d '{
        "policies": ["file-storage-policy"],
        "token_ttl": "1h",
        "token_max_ttl": "4h"
    }'

echo "Creating policy..."
curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/sys/policies/acl/file-storage-policy" \
    -d '{
        "policy": "path \"secret/data/storage\" { capabilities = [\"read\"] }"
    }'

echo "Getting role ID..."
ROLE_ID=$(curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/auth/approle/role/file-storage/role-id" | jq -r '.data.role_id')

echo "Getting secret ID..."
SECRET_ID=$(curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/auth/approle/role/file-storage/secret-id" | jq -r '.data.secret_id')

# Store the storage credentials
echo "Storing storage credentials in Vault..."
curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/secret/data/storage" \
    -d "{
        \"data\": {
            \"account_name\": \"${AZURE_STORAGE_ACCOUNT}\",
            \"storage_key\": \"${AZURE_STORAGE_KEY}\",
            \"storage_url\": \"${BLOB_STORAGE_URL}\",
            \"container_name\": \"${CONTAINER_NAME}\"
        }
    }"

echo "Vault initialization complete!"
echo "Role ID: ${ROLE_ID}"
echo "Secret ID: ${SECRET_ID}"