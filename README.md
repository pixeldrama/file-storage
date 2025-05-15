# File Storage Service

A Go service for asynchronous file uploads and downloads using Azure Blob Storage.

## Features

- Asynchronous file uploads with job tracking
- Azure Blob Storage integration
- Prometheus metrics endpoint
- Environment-based configuration
- Mocked vault integration for development
- Automatic antivirus scanning of uploaded files

## Prerequisites

### Installing Required Tools

#### Mac OS

1. Install Azure CLI:
```bash
brew install azure-cli
```

2. Install Azurite (Azure Storage Emulator):
```bash
npm install -g azurite
```

3. Start Azurite in the background:
```bash
azurite &
```

4. Create the 'files' container in Azurite:
```bash
az storage container create --name files --connection-string 'DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;'
```

## Configuration

The service can be configured using environment variables:

```bash
export SERVER_PORT=8080
export BLOB_STORAGE_URL="https://your-storage-account.blob.core.windows.net"
export CONTAINER_NAME="files"
export STORAGE_KEY="your-storage-key"  # In production, this will be fetched from vault
```

For local development with Azurite, set:
```bash
export USE_AZURITE=true
```

## API Endpoints

### Create Upload Job
```
POST /api/upload-jobs
```
Creates a new upload job and returns a UUID.

### Get Upload Job Status
```
GET /api/upload-jobs/{jobId}
```
Returns the current status of an upload job.

### Upload File
```
POST /api/upload-jobs/{jobId}
Content-Type: multipart/form-data
```
Uploads a file for the specified job ID. Files are automatically scanned for viruses after upload.

### Download File
```
GET /api/files/{fileId}
```
Downloads a file by its ID. Files are only available for download after passing antivirus scanning. If a file is still being scanned, the request will return a 404 status code.

### Metrics
```
GET /metrics
```
Prometheus metrics endpoint.

## Building and Running

1. Install dependencies:
```bash
go mod download
```

2. Build the service:
```bash
go build
```

3. Run the service:
```bash
./file-storage-go
```

## Development

For development purposes, the vault integration is mocked. The storage key can be provided through the `STORAGE_KEY` environment variable or will default to a mock value.

Use `make local` to run the service with Azurite for local development.

## Testing

Use `make test-api` to run API tests against a running instance of the service. Ensure Azurite is running and the 'files' container exists before testing.

## Metrics

The service exposes the following Prometheus metrics:

- `file_upload_duration_seconds`: Histogram of file upload durations
- `file_upload_size_bytes`: Histogram of uploaded file sizes

## License

MIT 