# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /api /usr/local/bin/api
EXPOSE 8000
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8000/health || exit 1
USER nobody
ENTRYPOINT ["/usr/local/bin/api"]
