#!/bin/bash

# Default host for Keycloak
KEYCLOAK_HOST=${KEYCLOAK_HOST:-keycloak}
KEYCLOAK_PORT=8080
MAX_RETRIES=30
RETRY_INTERVAL=5


# Function to check if realm exists
check_realm_exists() {
    local realm=$1
    local response=$(curl -s -w "\n%{http_code}" http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/admin/realms/${realm} \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    local http_code=$(echo "$response" | tail -n1)
    [ "$http_code" = "200" ]
}

# Function to delete realm if exists
delete_realm() {
    local realm=$1
    if check_realm_exists "$realm"; then
        echo "Deleting existing realm: $realm"
        curl -s -X DELETE http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/admin/realms/${realm} \
            -H "Authorization: Bearer $ADMIN_TOKEN"
    fi
}


# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
for i in $(seq 1 $MAX_RETRIES); do
    RESPONSE=$(curl -s http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/realms/master/.well-known/openid-configuration)
    if [ $? -eq 0 ]; then
        echo "Keycloak is ready!"
        break
    fi
    if [ $i -eq $MAX_RETRIES ]; then
        echo "Keycloak failed to become ready after $MAX_RETRIES attempts"
        exit 1
    fi
    echo "Attempt $i/$MAX_RETRIES: Keycloak not ready yet, waiting ${RETRY_INTERVAL}s..."
    sleep $RETRY_INTERVAL
done


# Get admin token
echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/realms/master/protocol/openid-connect/token \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=admin" \
    -d "password=admin" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')

if [ -z "$ADMIN_TOKEN" ]; then
    echo "Failed to get admin token"
    exit 1
fi

# Delete existing realm if it exists
delete_realm "file-storage"

# Create realm
echo "Creating realm..."
REALM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/admin/realms \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "realm": "file-storage",
        "enabled": true
    }')
HTTP_CODE=$(echo "$REALM_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" != "201" ]; then
    echo "Failed to create realm: $REALM_RESPONSE"
    exit 1
fi

# Create client
echo "Creating client..."
CLIENT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/admin/realms/file-storage/clients \
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
    }')
HTTP_CODE=$(echo "$CLIENT_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" != "201" ]; then
    echo "Failed to create client: $CLIENT_RESPONSE"
    exit 1
fi

# Create test user
echo "Creating test user..."
USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://${KEYCLOAK_HOST}:${KEYCLOAK_PORT}/admin/realms/file-storage/users \
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
    }')
HTTP_CODE=$(echo "$USER_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" != "201" ]; then
    echo "Failed to create test user: $USER_RESPONSE"
    exit 1
fi

echo "Keycloak setup complete!"
