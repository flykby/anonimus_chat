#!/usr/bin/env bash
# One-time GHCR setup on the production VM (Option A: GitHub Actions + SSH deploy).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REGISTRY_URL="${REGISTRY_URL:-ghcr.io/flykby/anonimus}"
REGISTRY_HOST="${REGISTRY_URL%%/*}"

log() {
	echo "[setup-ghcr] $*"
}

die() {
	echo "[setup-ghcr] error: $*" >&2
	exit 1
}

if [[ ! -f .env ]]; then
	die ".env not found — copy .env.prod.example to .env first"
fi

if [[ -z "${GHCR_PAT:-}" ]]; then
	echo "Need a GitHub PAT with read:packages (and write:packages if you push from VM)."
	echo "Create: GitHub → Settings → Developer settings → Personal access tokens"
	read -rsp "Paste PAT (hidden): " GHCR_PAT
	echo
fi

[[ -n "$GHCR_PAT" ]] || die "GHCR_PAT is empty"

read -rp "GitHub username for docker login [flykby]: " GHCR_USER
GHCR_USER="${GHCR_USER:-flykby}"

log "logging in to $REGISTRY_HOST as $GHCR_USER"
echo "$GHCR_PAT" | docker login "$REGISTRY_HOST" -u "$GHCR_USER" --password-stdin

if grep -q '^REGISTRY_URL=' .env; then
	sed -i "s|^REGISTRY_URL=.*|REGISTRY_URL=${REGISTRY_URL}|" .env
else
	echo "REGISTRY_URL=${REGISTRY_URL}" >> .env
fi

if grep -q '^REGISTRY_USER=' .env; then
	sed -i "s|^REGISTRY_USER=.*|REGISTRY_USER=${GHCR_USER}|" .env
else
	echo "REGISTRY_USER=${GHCR_USER}" >> .env
fi

if grep -q '^REGISTRY_PASSWORD=' .env; then
	sed -i "s|^REGISTRY_PASSWORD=.*|REGISTRY_PASSWORD=${GHCR_PAT}|" .env
else
	echo "REGISTRY_PASSWORD=${GHCR_PAT}" >> .env
fi

log "updated .env: REGISTRY_URL=${REGISTRY_URL}"
log "testing pull (tag=latest)..."
docker pull "${REGISTRY_URL}/bot:latest" || log "pull failed — push an image from CI first (git push origin main)"

log "done. Next: configure GitHub Actions secrets (see docs/deploy.md)"
