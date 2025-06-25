#!/bin/bash

BASE_URL="http://app:8080"
CLEAN_TEST_FILE="clean_test.txt"
VIRUS_TEST_FILE="virus_test.txt"
DOWNLOADED_FILE_PREFIX="downloaded_test_file"

KEYCLOAK_URL="http://keycloak:8080"
REALM="file-storage"
CLIENT_ID="file-storage"
CLIENT_SECRET="test-secret" # This matches the secret in setup-keycloak.sh
JWT_TOKEN="mock-token"


if ! command -v jq &> /dev/null
then
    echo "jq could not be found. Please install jq to run this script."
    echo "For example, on Debian/Ubuntu: sudo apt install jq"
    echo "On macOS with Homebrew: brew install jq"
    exit 1
fi

# Create test files
if [ ! -f "$CLEAN_TEST_FILE" ]; then
    echo "This is a clean test file." > "$CLEAN_TEST_FILE"
fi

if [ ! -f "$VIRUS_TEST_FILE" ]; then
    echo "virus" > "$VIRUS_TEST_FILE"
fi

echo "### Testing API Endpoints ###"

# 0. Health Check
echo -e "\n\n--- 0. Health Check ---"
echo "GET $BASE_URL/health"
echo "Testing connection to app..."
curl -v "$BASE_URL/health"
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ "$HEALTH_STATUS" -eq 200 ]; then
    echo "Health check successful - app is running!"
else
    echo "Error: Health check failed. Status code: $HEALTH_STATUS"
    exit 1
fi

# Test 1: Clean File Upload
echo -e "\n\n--- Test 1: Clean File Upload ---"

# 1. Create Upload Job
echo -e "\n\n--- 1. Create Upload Job ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -v -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -d '{"filename": "clean_file.txt", "fileType": "some_filetype", "linkedResourceType": "company", "linkedResourceID": "3"}' \
    "$BASE_URL/upload-jobs")
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

# Extract file ID for further testing
FILE_ID=$(echo "$JOB_STATUS" | jq -r '.fileId')
if [ -z "$FILE_ID" ] || [ "$FILE_ID" == "null" ]; then
    echo "Error: Could not extract fileId from job status."
    exit 1
fi
echo "Extracted FILE_ID: $FILE_ID"

# Test 1.5: Get File Info
echo -e "\n\n--- Test 1.5: Get File Info ---"
echo "GET $BASE_URL/files/$FILE_ID"
FILE_INFO_RESPONSE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/files/$FILE_ID")
echo "File Info Response: $FILE_INFO_RESPONSE"

# Verify file info was retrieved successfully
if echo "$FILE_INFO_RESPONSE" | jq -e '.id' > /dev/null; then
    echo "File info retrieved successfully!"
else
    echo "Error: Failed to retrieve file info!"
    exit 1
fi

# Test 1.6: Delete File
echo -e "\n\n--- Test 1.6: Delete File ---"
echo "DELETE $BASE_URL/files/$FILE_ID"
DELETE_RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/files/$FILE_ID")
echo "Delete Response Status: $DELETE_RESPONSE"

# Verify file was deleted successfully (should return 204 No Content)
if [ -z "$DELETE_RESPONSE" ]; then
    echo "File deleted successfully!"
else
    echo "Error: Failed to delete file!"
    exit 1
fi

# Test 1.7: Verify File Info is Gone
echo -e "\n\n--- Test 1.7: Verify File Info is Gone ---"
echo "GET $BASE_URL/files/$FILE_ID"
FILE_INFO_AFTER_DELETE=$(curl -s -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/files/$FILE_ID")
echo "File Info After Delete: $FILE_INFO_AFTER_DELETE"

# Verify file info is no longer accessible
if echo "$FILE_INFO_AFTER_DELETE" | jq -e '.error' > /dev/null; then
    echo "File info correctly no longer accessible!"
else
    echo "Error: File info still accessible after deletion!"
    exit 1
fi

# Test 2: Virus File Upload
echo -e "\n\n--- Test 2: Virus File Upload ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWT_TOKEN" -d '{"filename": "virus_file.txt", "fileType": "some_filetype", "linkedResourceType": "company", "linkedResourceID": "3"}' "$BASE_URL/upload-jobs")
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
