#!/bin/bash

BASE_URL="http://localhost:8080"
TEST_FILE="test.txt" 
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

if [ ! -f "$TEST_FILE" ]; then
    echo "This is a test file for upload." > "$TEST_FILE"
    echo "Created dummy file: $TEST_FILE"
fi

echo "### Testing API Endpoints ###"

# 1. Create Upload Job
echo -e "\n\n--- 1. Create Upload Job ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d '{"fileName": "mytestfile.txt"}' "$BASE_URL/upload-jobs")
echo "Response: $CREATE_JOB_RESPONSE"

JOB_ID=$(echo "$CREATE_JOB_RESPONSE" | jq -r '.jobId')

if [ -z "$JOB_ID" ] || [ "$JOB_ID" == "null" ]; then
    echo "Error: Could not extract jobId from create job response."
    exit 1
fi
echo "Extracted JOB_ID: $JOB_ID"


# 2. Get Upload Job Status
echo -e "\n\n--- 2. Get Upload Job Status ---"
echo "GET $BASE_URL/upload-jobs/$JOB_ID"
curl -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/upload-jobs/$JOB_ID" > /dev/null
echo # Newline for better formatting


# 3. Upload File
echo -e "\n\n--- 3. Upload File ---"
echo "POST $BASE_URL/upload-jobs/$JOB_ID"
UPLOAD_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -F "file=@$TEST_FILE" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Response: $UPLOAD_RESPONSE"

FILE_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.fileId')

if [ -z "$FILE_ID" ] || [ "$FILE_ID" == "null" ]; then
    echo "Error: Could not extract fileId from upload response."
    exit 1
fi
echo "Extracted FILE_ID: $FILE_ID"

# 4. Download File
echo -e "\n\n--- 4. Download File ---"
echo "GET $BASE_URL/files/$FILE_ID"
DOWNLOADED_FILE="${DOWNLOADED_FILE_PREFIX}_${FILE_ID}"
curl -s -H "Authorization: Bearer $JWT_TOKEN" -o "$DOWNLOADED_FILE" "$BASE_URL/files/$FILE_ID"
echo "File downloaded to: $DOWNLOADED_FILE"

# Compare original and downloaded files
if cmp -s "$TEST_FILE" "$DOWNLOADED_FILE"; then
    echo "Files are identical - download successful!"
else
    echo "Error: Downloaded file differs from original!"
    exit 1
fi

# 5. Delete File
echo -e "\n\n--- 5. Delete File ---"
echo "DELETE $BASE_URL/files/$FILE_ID"
DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" -X DELETE "$BASE_URL/files/$FILE_ID")
if [ "$DELETE_STATUS" -eq 204 ]; then
    echo "File deleted successfully!"
else
    echo "Error: Failed to delete file. Status code: $DELETE_STATUS"
    exit 1
fi

# 6. Verify Delete
echo -e "\n\n--- 6. Verify Delete ---"
echo "GET $BASE_URL/files/$FILE_ID (should fail)"
VERIFY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/files/$FILE_ID")
if [ "$VERIFY_STATUS" -eq 404 ]; then
    echo "Verification successful - file no longer exists!"
else
    echo "Error: File still accessible after deletion. Status code: $VERIFY_STATUS"
    exit 1
fi

# Cleanup
rm -f "$TEST_FILE" "$DOWNLOADED_FILE"
echo -e "\n\nAll tests completed successfully!"

echo -e "\n\n### Testing Complete ###"
echo "Ensure your server is running on http://localhost:8080."
echo "If you saw errors, check the server logs and the script output."

echo "Remember to replace placeholder IDs (your_job_id, your_file_id) with actual values."
echo "You might need to make the script executable: chmod +x scripts/test-api.sh" 