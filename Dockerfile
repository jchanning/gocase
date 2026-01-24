# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for ARM64
# CGO_ENABLED=0 ensures a static binary
RUN CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/

# Final stage - using alpine for minimal image with shell access
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy templates and assets
COPY --from=builder /app/views ./views
COPY --from=builder /app/assets ./assets

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./server"]
