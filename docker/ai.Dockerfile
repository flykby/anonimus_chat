# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ai ./cmd/ai

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /ai /usr/local/bin/ai
EXPOSE 8001
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8001/health || exit 1
USER nobody
ENTRYPOINT ["/usr/local/bin/ai"]
