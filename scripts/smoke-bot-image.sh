#!/usr/bin/env bash
# Smoke-test bot Docker image: health endpoint without Telegram API.
set -euo pipefail

IMAGE="${BOT_IMAGE:-anonimus/bot:local}"
PORT="${SMOKE_PORT:-18080}"
HEALTH_URL="http://127.0.0.1:${PORT}/health"

CID=""
cleanup() {
  if [[ -n "$CID" ]]; then
    docker rm -f "$CID" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

CID="$(docker run -d \
  --rm \
  -e BOT_HEALTH_ONLY=1 \
  -p "${PORT}:8080" \
  "$IMAGE")"

for _ in $(seq 1 15); do
  if curl -fsS "$HEALTH_URL" >/dev/null 2>&1; then
    echo "smoke ok: $HEALTH_URL"
    curl -fsS "$HEALTH_URL"
    echo
    exit 0
  fi
  sleep 1
done

echo "smoke failed: $HEALTH_URL not ready" >&2
docker logs "$CID" >&2 || true
exit 1
