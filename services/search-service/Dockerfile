# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /workspace
COPY proto /workspace/proto
COPY services/search-service/go.mod services/search-service/go.sum /workspace/services/search-service/

WORKDIR /workspace/services/search-service
COPY certs/ /tmp/certs/
RUN find /tmp/certs -name '*.crt' -exec cat {} >> /etc/ssl/certs/ca-certificates.crt \;
ENV GODEBUG=http2client=0
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOFLAGS=-mod=mod
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY services/search-service /workspace/services/search-service
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /workspace/services/search-service/main .
EXPOSE 50054
CMD ["./main"]
