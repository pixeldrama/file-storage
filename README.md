# File Storage Service

A Go service for asynchronous file uploads and downloads using Azure Blob Storage.

## Features

- Asynchronous file uploads with job tracking
- Azure Blob Storage integration
- Prometheus metrics endpoint
- Environment-based configuration
- Mocked vault integration for development
- Automatic antivirus scanning of uploaded files

## Configuration

The service can be configured using environment variables:

```bash
export SERVER_PORT=8080
export BLOB_STORAGE_URL="https://your-storage-account.blob.core.windows.net"
export CONTAINER_NAME="files"
export STORAGE_KEY="your-storage-key"  # In production, this will be fetched from vault
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

## Metrics

The service exposes the following Prometheus metrics:

- `file_upload_duration_seconds`: Histogram of file upload durations
- `file_upload_size_bytes`: Histogram of uploaded file sizes

## License

MIT 