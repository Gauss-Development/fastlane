# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /workspace

COPY proto /workspace/proto
COPY services/design-service/go.mod services/design-service/go.sum /workspace/services/design-service/

WORKDIR /workspace/services/design-service
COPY certs/ /tmp/certs/
RUN find /tmp/certs -name '*.crt' -exec cat {} >> /etc/ssl/certs/ca-certificates.crt \;
ENV GODEBUG=http2client=0
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOFLAGS=-mod=mod
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY services/design-service /workspace/services/design-service

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /workspace/services/design-service/main .

EXPOSE 8086

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8086/health || exit 1

CMD ["./main"]
