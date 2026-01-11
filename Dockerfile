# ==================================
# Build Stage
# ==================================
FROM golang:1.25-alpine AS builder

# Install ca-certificates for HTTPS and git for private deps
RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /app

# Cache dependencies layer
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations and security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -installsuffix cgo \
    -o main \
    ./cmd/api/main.go

# ==================================
# Run Stage (Distroless for security)
# ==================================
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=nonroot:nonroot /app/main .

# Use distroless nonroot user (UID 65532)
USER nonroot:nonroot

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/main", "--health-check"] || exit 1

# Run application
ENTRYPOINT ["/app/main"]
