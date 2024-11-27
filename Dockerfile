# Base stage for shared build dependencies
FROM golang:1.21-alpine AS base
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app

# Development stage for debugging and testing
FROM base AS development
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY . .
RUN go mod download
CMD ["go", "run", "./cmd/server"]

# Test stage for running tests
FROM base AS test
COPY . .
RUN go mod download
RUN go test -v ./...

# Builder stage for compiling the application
FROM base AS builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server

# Security scanner stage
FROM golang:1.21-alpine AS security-check
RUN apk add --no-cache git
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
COPY . .
RUN go mod download
RUN govulncheck ./...

# Final production stage
FROM alpine:3.18 AS production

# Add non root user
RUN adduser -D -g '' appuser

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Use non root user
USER appuser

# Expose the server port
EXPOSE 8080

# Set default environment variables
ENV SERVER_ADDRESS=":8080" \
    POW_COMPLEXITY="100000" \
    CHALLENGE_EXPIRATION_SECONDS="300"

# Health check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD nc -zv localhost 8080 || exit 1

# Run the application
CMD ["./server"]
