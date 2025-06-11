# Multi-stage Dockerfile for E2E testing
# Optimized for CI execution speed and environment isolation

# Build stage - compile dependencies and prepare Go environment
FROM golang:1.23-alpine AS builder

# Install git for Go modules and ca-certificates for HTTPS
RUN apk --no-cache add git ca-certificates

# Set working directory
WORKDIR /app

# Copy Go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with module cache
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the main binary for testing
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o thinktank .

# Runtime stage - minimal image for test execution
FROM golang:1.23-alpine AS runtime

# Install essential packages for E2E testing
RUN apk --no-cache add \
    git \
    ca-certificates \
    tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S thinktank && \
    adduser -u 1001 -S thinktank -G thinktank

# Set working directory
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Copy source code (needed for test execution)
COPY . .

# Copy built binary from builder stage
COPY --from=builder /app/thinktank ./thinktank

# Set proper ownership
RUN chown -R thinktank:thinktank /app

# Switch to non-root user
USER thinktank

# Set Go environment variables for optimal testing
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GO111MODULE=on
ENV GOCACHE=/tmp/go-cache
ENV GOTMPDIR=/tmp

# Ensure Go modules are downloaded and verified
RUN go mod download && go mod verify

# Verify thinktank binary works
RUN ./thinktank --help > /dev/null

# Default command runs E2E tests
# This can be overridden by docker run command
CMD ["go", "test", "-v", "-tags=manual_api_test", "./internal/e2e/..."]
