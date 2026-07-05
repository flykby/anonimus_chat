#!/usr/bin/env bash
# Apply goose migrations on production VM (postgres is internal-only in prod compose).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env}"
GOOSE_IMAGE="${GOOSE_IMAGE:-ghcr.io/kukymbr/goose-docker:3.27.1}"
COMPOSE_NETWORK="${COMPOSE_NETWORK:-anonimus-prod_default}"
GOOSE_COMMAND="${1:-up}"

log() {
	echo "[migrate-prod] $*"
}

die() {
	echo "[migrate-prod] error: $*" >&2
	exit 1
}

[[ -f "$ENV_FILE" ]] || die "env file not found: $ENV_FILE"
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

POSTGRES_USER="${POSTGRES_USER:-anonimus}"
POSTGRES_DB="${POSTGRES_DB:-anonimus}"
[[ -n "${POSTGRES_PASSWORD:-}" ]] || die "POSTGRES_PASSWORD is not set in $ENV_FILE"

GOOSE_DBSTRING="host=postgres port=5432 user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable"

log "pulling $GOOSE_IMAGE"
docker pull "$GOOSE_IMAGE"

log "running goose $GOOSE_COMMAND on network $COMPOSE_NETWORK"
docker run --rm --network "$COMPOSE_NETWORK" \
	-v "$ROOT/migrations:/migrations" \
	-e GOOSE_DRIVER=postgres \
	-e GOOSE_MIGRATION_DIR=/migrations \
	-e GOOSE_DBSTRING="$GOOSE_DBSTRING" \
	-e GOOSE_COMMAND="$GOOSE_COMMAND" \
	"$GOOSE_IMAGE"

log "done"
