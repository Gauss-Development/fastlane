# syntax=docker/dockerfile:1
# go.mod requires go >= 1.25.0; base must match (rest of the fleet is 1.24).
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set working directory
WORKDIR /workspace

# Copy shared proto module and service dependencies
COPY proto /workspace/proto
COPY services/post-service/go.mod services/post-service/go.sum /workspace/services/post-service/

# Download dependencies
WORKDIR /workspace/services/post-service
COPY certs/ /tmp/certs/
RUN find /tmp/certs -name '*.crt' -exec cat {} >> /etc/ssl/certs/ca-certificates.crt \;
ENV GODEBUG=http2client=0
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOFLAGS=-mod=mod
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source code
COPY services/post-service /workspace/services/post-service

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /workspace/services/post-service/main .

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Run the application
CMD ["./main"]
