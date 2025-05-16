#!/bin/bash

BASE_URL="http://localhost:8080"
TEST_FILE="test.txt" # Create a dummy file for testing uploads
DOWNLOADED_FILE_PREFIX="downloaded_test_file"

# Check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq could not be found. Please install jq to run this script."
    echo "For example, on Debian/Ubuntu: sudo apt install jq"
    echo "On macOS with Homebrew: brew install jq"
    exit 1
fi

# Create a dummy test file if it doesn't exist
if [ ! -f "$TEST_FILE" ]; then
    echo "This is a test file for upload." > "$TEST_FILE"
    echo "Created dummy file: $TEST_FILE"
fi

echo "### Testing API Endpoints ###"

# 1. Create Upload Job
echo -e "\n\n--- 1. Create Upload Job ---"
echo "POST $BASE_URL/upload-jobs"
CREATE_JOB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d '{"fileName": "mytestfile.txt"}' "$BASE_URL/upload-jobs")
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
curl -X GET "$BASE_URL/upload-jobs/$JOB_ID"
echo # Newline for better formatting


# 3. Upload File
echo -e "\n\n--- 3. Upload File ---"
echo "POST $BASE_URL/upload-jobs/$JOB_ID"
UPLOAD_RESPONSE=$(curl -s -X POST -F "file=@$TEST_FILE" "$BASE_URL/upload-jobs/$JOB_ID")
echo "Response: $UPLOAD_RESPONSE"

FILE_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.fileId')

if [ -z "$FILE_ID" ] || [ "$FILE_ID" == "null" ]; then
    echo "Error: Could not extract fileId from upload response. Job status might not be 'COMPLETED' yet or an error occurred."
    echo "Full upload response: $UPLOAD_RESPONSE"
    # Depending on the desired behavior, you might want to poll here or exit.
    # For now, we'll try to proceed, but download will likely fail.
    # exit 1 # Uncomment to make it a hard stop
fi

if [ -n "$FILE_ID" ] && [ "$FILE_ID" != "null" ]; then
    echo "Extracted FILE_ID: $FILE_ID"

    # 4. Download File (only if FILE_ID was extracted)
    echo -e "\n\n--- 4. Download File ---"
    DOWNLOADED_FILE_NAME="${DOWNLOADED_FILE_PREFIX}_${FILE_ID}.txt"
    echo "GET $BASE_URL/files/$FILE_ID -o $DOWNLOADED_FILE_NAME"
    curl -L -v --max-time 30 -X GET "$BASE_URL/files/$FILE_ID" -o "$DOWNLOADED_FILE_NAME"
    
    if [ -f "$DOWNLOADED_FILE_NAME" ]; then
        echo -e "\nFile downloaded as $DOWNLOADED_FILE_NAME"
        echo "Content of downloaded file:"
        cat "$DOWNLOADED_FILE_NAME"
        echo
    else
        echo -e "\nError: Downloaded file $DOWNLOADED_FILE_NAME not found."
    fi
else
    echo -e "\nSkipping file download because FILE_ID was not extracted."
fi


echo -e "\n\n### Testing Complete ###"
echo "Ensure your server is running on http://localhost:8080."
echo "If you saw errors, check the server logs and the script output."

echo "Remember to replace placeholder IDs (your_job_id, your_file_id) with actual values."
echo "You might need to make the script executable: chmod +x scripts/test-api.sh" 