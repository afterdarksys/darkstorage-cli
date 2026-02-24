# Dark Storage CLI - Docker Image
# Multi-stage build for minimal image size

FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o darkstorage main.go

# Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Create non-root user
RUN addgroup -g 1000 darkstorage && \
    adduser -D -u 1000 -G darkstorage darkstorage

# Copy binary
COPY --from=builder /build/darkstorage /usr/local/bin/darkstorage

# Switch to non-root user
USER darkstorage

# Create config directory
RUN mkdir -p /home/darkstorage/.darkstorage

# Set working directory
WORKDIR /data

# Entry point
ENTRYPOINT ["darkstorage"]
CMD ["--help"]
