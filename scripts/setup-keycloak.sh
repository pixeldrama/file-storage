#!/bin/bash

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
until curl -s http://keycloak:8080/health/ready; do
    sleep 5
done

# Get admin token
echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST http://keycloak:8080/realms/master/protocol/openid-connect/token \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=admin" \
    -d "password=admin" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')

# Create realm
echo "Creating realm..."
curl -s -X POST http://keycloak:8080/admin/realms \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "realm": "file-storage",
        "enabled": true
    }'

# Create client
echo "Creating client..."
curl -s -X POST http://keycloak:8080/admin/realms/file-storage/clients \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "clientId": "file-storage",
        "secret": "test-secret",
        "enabled": true,
        "protocol": "openid-connect",
        "publicClient": false,
        "standardFlowEnabled": true,
        "directAccessGrantsEnabled": true,
        "serviceAccountsEnabled": true
    }'

# Create test user
echo "Creating test user..."
curl -s -X POST http://keycloak:8080/admin/realms/file-storage/users \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "test-user",
        "enabled": true,
        "credentials": [
            {
                "type": "password",
                "value": "test-password",
                "temporary": false
            }
        ]
    }'

echo "Keycloak setup complete!" 