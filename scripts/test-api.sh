#!/bin/bash

BASE_URL="http://localhost:8080"
CLEAN_TEST_FILE="clean_test.txt"
VIRUS_TEST_FILE="virus_test.txt"
DOWNLOADED_FILE_PREFIX="downloaded_test_file"

KEYCLOAK_URL="http://localhost:8081"
REALM="file-storage"
CLIENT_ID="file-storage"
CLIENT_SECRET="test-secret" # This matches the secret in setup-keycloak.sh

get_keycloak_token() {
    local token_response=$(curl -s -X POST "${KEYCLOAK_URL}/realms/${REALM}/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials" \
        -d "client_id=${CLIENT_ID}" \
        -d "client_secret=${CLIENT_SECRET}")

    local token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ -z "$token" ] || [ "$token" == "null" ]; then
        echo "Error: Failed to obtain token from Keycloak"
        echo "Response: $token_response"
        exit 1
    fi
    
    echo "$token"
}

if ! command -v jq &> /dev/null
then
    echo "jq could not be found. Please install jq to run this script."
    echo "For example, on Debian/Ubuntu: sudo apt install jq"
    echo "On macOS with Homebrew: brew install jq"
    exit 1
fi

echo "Obtaining JWT token from Keycloak..."
JWT_TOKEN=$(get_keycloak_token)
if [ $? -ne 0 ]; then
    echo "Failed to obtain JWT token. Exiting."
    exit 1
fi
echo "Successfully obtained JWT token"

# Create test files
if [ ! -f "$CLEAN_TEST_FILE" ]; then
    echo "This is a clean test file." > "$CLEAN_TEST_FILE"
fi

if [ ! -f "$VIRUS_TEST_FILE" ]; then
    echo "virus" > "$VIRUS_TEST_FILE"
fi

echo "### Testing API Endpoints ###"

# Test 1: Clean File Upload
echo -e "\n\n--- Test 1: Clean File Upload ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d '{"fileName": "clean_file.txt"}' "$BASE_URL/upload-jobs")
echo "Response: $CREATE_JOB_RESPONSE"

JOB_ID=$(echo "$CREATE_JOB_RESPONSE" | jq -r '.jobId')

if [ -z "$JOB_ID" ] || [ "$JOB_ID" == "null" ]; then
    echo "Error: Could not extract jobId from create job response."
    exit 1
fi
echo "Extracted JOB_ID: $JOB_ID"

# Upload clean file
echo "POST $BASE_URL/upload-jobs/$JOB_ID"
UPLOAD_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -F "file=@$CLEAN_TEST_FILE" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Response: $UPLOAD_RESPONSE"

# Wait for virus check to complete
echo "Waiting for virus check to complete..."
sleep 0.5

# Check job status
echo "GET $BASE_URL/upload-jobs/$JOB_ID"
JOB_STATUS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Job Status: $JOB_STATUS"

# Verify clean file was accepted
if echo "$JOB_STATUS" | jq -e '.status == "COMPLETED"' > /dev/null; then
    echo "Clean file was accepted as expected!"
else
    echo "Error: Clean file was not accepted!"
    exit 1
fi

# Test 2: Virus File Upload
echo -e "\n\n--- Test 2: Virus File Upload ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d '{"fileName": "virus_file.txt"}' "$BASE_URL/upload-jobs")
echo "Response: $CREATE_JOB_RESPONSE"

JOB_ID=$(echo "$CREATE_JOB_RESPONSE" | jq -r '.jobId')

if [ -z "$JOB_ID" ] || [ "$JOB_ID" == "null" ]; then
    echo "Error: Could not extract jobId from create job response."
    exit 1
fi
echo "Extracted JOB_ID: $JOB_ID"

# Upload virus file
echo "POST $BASE_URL/upload-jobs/$JOB_ID"
UPLOAD_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -F "file=@$VIRUS_TEST_FILE" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Response: $UPLOAD_RESPONSE"

# Wait for virus check to complete
echo "Waiting for virus check to complete..."
sleep 0.5

# Check job status
echo "GET $BASE_URL/upload-jobs/$JOB_ID"
JOB_STATUS=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Job Status: $JOB_STATUS"

# Verify virus file was rejected
if echo "$JOB_STATUS" | jq -e '.status == "FAILED"' > /dev/null; then
    echo "Virus file was rejected as expected!"
else
    echo "Error: Virus file was not rejected!"
    exit 1
fi

# Cleanup
rm -f "$CLEAN_TEST_FILE" "$VIRUS_TEST_FILE"
echo -e "\n\nAll tests completed successfully!"

echo -e "\n\n### Testing Complete ###"
echo "Ensure your server is running on http://localhost:8080."
echo "If you saw errors, check the server logs and the script output."

echo "Remember to replace placeholder IDs (your_job_id, your_file_id) with actual values."
echo "You might need to make the script executable: chmod +x scripts/test-api.sh" 