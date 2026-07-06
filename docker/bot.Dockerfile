# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget \
    && mkdir -p /app/certs \
    && chown nobody:nobody /app/certs
COPY --from=builder /bot /usr/local/bin/bot
COPY docker/bot-entrypoint.sh /usr/local/bin/bot-entrypoint.sh
RUN chmod +x /usr/local/bin/bot-entrypoint.sh
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- --no-check-certificate https://127.0.0.1:8080/health 2>/dev/null || wget -qO- http://127.0.0.1:8080/health
USER nobody
ENTRYPOINT ["/usr/local/bin/bot-entrypoint.sh"]
