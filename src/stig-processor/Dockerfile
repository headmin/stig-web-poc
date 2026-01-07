# Multi-stage build for STIG processor with Go 1.25
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code with proper structure
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Build the binary with version information and Go 1.25 optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -trimpath \
    -o stig-processor \
    ./cmd

# Verify the binary was built correctly
RUN ./stig-processor -version

# Final stage - minimal runtime image
FROM alpine:latest

# Security: Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Install only essential runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Set up application directory structure
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/stig-processor .

# Create directories for input/output with proper permissions
RUN mkdir -p /app/input /app/output && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set default environment variables
ENV INPUT_FILE=/app/input/microsoft-windows-11-security-technical-implementation-guide.json \
    OUTPUT_DIR=/app/output \
    FORMAT=yaml \
    VERBOSE=false \
    DRY_RUN=false \
    TIMEOUT=5m \
    GOMAXPROCS=0

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./stig-processor -version || exit 1

# Expose volumes for input/output
VOLUME ["/app/input", "/app/output"]

# Set metadata labels
LABEL maintainer="STIG Processor Team" \
      description="DISA STIG processor for Fleet/osquery policy generation with Go 1.25" \
      version="${VERSION}" \
      go.version="1.25" \
      org.opencontainers.image.title="STIG Processor" \
      org.opencontainers.image.description="Convert DISA STIG rules to Fleet/osquery policies using Go 1.25" \
      org.opencontainers.image.vendor="STIG Processor Team" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.source="https://github.com/your-org/stig-processor" \
      org.opencontainers.image.documentation="https://github.com/your-org/stig-processor/blob/main/README.md"

# Default command with flexible argument support
ENTRYPOINT ["./stig-processor"]
CMD ["-input", "/app/input/microsoft-windows-11-security-technical-implementation-guide.json", \
     "-output", "/app/output", \
     "-verbose"]
