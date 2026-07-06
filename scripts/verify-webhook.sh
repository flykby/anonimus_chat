#!/usr/bin/env bash
# Quick webhook diagnostics on the production VM.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env}"
[[ -f "$ENV_FILE" ]] || { echo "missing $ENV_FILE"; exit 1; }
# shellcheck disable=SC1090
source "$ENV_FILE"

echo "=== Public IP (use in WEBHOOK_URL) ==="
curl -4 -s ifconfig.me || true
echo ""

echo "=== .env webhook settings ==="
grep -E '^WEBHOOK_' "$ENV_FILE" || true
echo ""

echo "=== Cert files ==="
ls -la certs/webhook.pem certs/webhook.key 2>/dev/null || echo "certs/ missing — run ./scripts/gen-webhook-cert.sh"
echo ""

echo "=== Local health (host :8443) ==="
curl -k -sS "https://127.0.0.1:8443/health" || echo "health check failed"
echo ""

echo "=== Bot container cert readability ==="
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec -T bot \
  sh -c 'id; ls -la /app/certs/ 2>/dev/null; test -r /app/certs/webhook.key && echo "key: readable" || echo "key: NOT readable"' \
  2>/dev/null || echo "bot container not running"
echo ""

echo "=== Recent bot logs ==="
docker logs --tail 20 anonimus-bot 2>&1 || true
