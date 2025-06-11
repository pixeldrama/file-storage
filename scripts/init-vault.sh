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
        -d '{"type": "kv", "options": {"version": "2"}}' || true
fi

# Check if AppRole auth method is already enabled
if ! curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/sys/auth/approle" > /dev/null 2>&1; then
    echo "Enabling AppRole auth method..."
    curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/sys/auth/approle" \
        -d '{"type": "approle"}' || true
fi

echo "Creating app role..."
if ! curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/auth/approle/role/file-storage" > /dev/null 2>&1; then
    curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/auth/approle/role/file-storage" \
        -d '{
            "policies": ["file-storage-policy"],
            "token_ttl": "1h",
            "token_max_ttl": "4h"
        }' || true
fi

echo "Creating policy..."
if ! curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR}/v1/sys/policies/acl/file-storage-policy" > /dev/null 2>&1; then
    curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/sys/policies/acl/file-storage-policy" \
        -d '{
            "policy": "path \"secret/data/storage\" { capabilities = [\"read\"] }"
        }' || true
fi

STATIC_ROLE_ID="test-role-id"
STATIC_SECRET_ID="test-secret-id"

curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/auth/approle/role/file-storage/role-id" \
    -d "{\"role_id\": \"${STATIC_ROLE_ID}\"}" || true

curl -fs -H "X-Vault-Token: ${VAULT_TOKEN}" -X POST "${VAULT_ADDR}/v1/auth/approle/role/file-storage/custom-secret-id" \
    -d "{\"secret_id\": \"${STATIC_SECRET_ID}\"}" || true

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
    }" || true

echo "Vault initialization complete!"
echo "Role ID: ${STATIC_ROLE_ID}"
echo "Secret ID: ${STATIC_SECRET_ID}"

exit 0
