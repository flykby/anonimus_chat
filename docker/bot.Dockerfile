# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY bot/go.mod bot/go.sum* ./
RUN go mod download
COPY bot/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /bot /usr/local/bin/bot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1
USER nobody
ENTRYPOINT ["/usr/local/bin/bot"]
