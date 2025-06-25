ARG DOCKER_PROXY=docker.io
FROM  ${DOCKER_PROXY}/golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download -x

# Copy the source code
COPY . .

# Run go vet and go test
RUN go vet -v ./...
RUN go test -v ./...

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Create a minimal production image
FROM ${DOCKER_PROXY}/alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"] 