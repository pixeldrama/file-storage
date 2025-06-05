#!/bin/bash

echo "Waiting for Keycloak to be ready..."
until curl -s http://keycloak:8080/health/ready; do
    sleep 5
done

echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST http://keycloak:8080/realms/master/protocol/openid-connect/token \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=admin" \
    -d "password=admin" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | jq -r '.access_token')

echo "Creating realm..."
curl -s -X POST http://keycloak:8080/admin/realms \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "realm": "file-storage",
        "enabled": true
    }'

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
        "standardFlowEnabled": false,
        "directAccessGrantsEnabled": false,
        "serviceAccountsEnabled": true
    }'

echo "Keycloak initialization completed" 