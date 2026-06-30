#!/usr/bin/env bash
# Run CI inside docker/ci.Dockerfile with host Docker socket.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CI_IMAGE="${CI_IMAGE:-anonimus/ci:local}"
STAGE="${1:-all}"

if [[ ! -S /var/run/docker.sock ]]; then
  echo "error: /var/run/docker.sock not found; run on a VM with Docker" >&2
  exit 1
fi

docker build -f docker/ci.Dockerfile -t "$CI_IMAGE" .

ENV_FILE_ARGS=()
if [[ -f "$ROOT/.env.ci" ]]; then
  ENV_FILE_ARGS=(--env-file "$ROOT/.env.ci")
fi

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$ROOT:/workspace" \
  -w /workspace \
  "${ENV_FILE_ARGS[@]}" \
  -e REGISTRY_URL \
  -e REGISTRY_USER \
  -e REGISTRY_PASSWORD \
  -e BOT_TOKEN \
  -e GIT_SHA \
  -e GIT_SHA_SHORT \
  -e GIT_BRANCH \
  "$CI_IMAGE" \
  ./scripts/ci.sh "$STAGE"
