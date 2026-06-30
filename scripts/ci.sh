#!/usr/bin/env bash
# Full CI pipeline: lint → test → build → smoke → push (optional).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

STAGE="${1:-all}"

export CI=true
export GIT_SHA="${GIT_SHA:-$(git rev-parse HEAD)}"
export GIT_SHA_SHORT="${GIT_SHA_SHORT:-$(git rev-parse --short HEAD)}"
export GIT_BRANCH="${GIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"

log() {
  echo "[ci] $*"
}

run_lint() {
  log "stage: lint"
  make tidy
  make lint
}

run_test() {
  log "stage: test"
  make test
}

run_build() {
  log "stage: build"
  make build-docker
}

run_smoke() {
  log "stage: smoke"
  if ! command -v docker >/dev/null 2>&1; then
    log "docker not available, skipping smoke test"
    return 0
  fi
  bash "$ROOT/scripts/smoke-bot-image.sh"
}

run_push() {
  log "stage: push"
  if [[ -z "${REGISTRY_URL:-}" ]]; then
    log "REGISTRY_URL not set, skipping push"
    return 0
  fi
  make push
}

case "$STAGE" in
  lint) run_lint ;;
  test) run_test ;;
  build) run_build ;;
  smoke) run_smoke ;;
  push) run_push ;;
  all)
    run_lint
    run_test
    run_build
    run_smoke
    run_push
    ;;
  *)
    echo "usage: $0 [lint|test|build|smoke|push|all]" >&2
    exit 1
    ;;
esac

log "done"
