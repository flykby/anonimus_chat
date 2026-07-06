#!/bin/sh
set -e

addr="${HTTP_ADDR:-:8080}"
host="127.0.0.1"
port="${addr##*:}"

if [ -n "${WEBHOOK_KEY_PATH:-}" ] && [ -r "${WEBHOOK_KEY_PATH}" ]; then
	exec wget -qO- --no-check-certificate "https://${host}:${port}/health"
fi

exec wget -qO- "http://${host}:${port}/health"
