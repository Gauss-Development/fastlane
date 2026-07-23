# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for HTTPS requests)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set working directory
WORKDIR /workspace

# Copy shared proto module and service dependencies
COPY proto /workspace/proto
COPY services/auth-service/go.mod services/auth-service/go.sum /workspace/services/auth-service/

# Download dependencies
WORKDIR /workspace/services/auth-service
COPY certs/ /tmp/certs/
RUN find /tmp/certs -name '*.crt' -exec cat {} >> /etc/ssl/certs/ca-certificates.crt \;
ENV GODEBUG=http2client=0
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOFLAGS=-mod=mod
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source code
COPY services/auth-service /workspace/services/auth-service

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /workspace/services/auth-service/main .

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]
