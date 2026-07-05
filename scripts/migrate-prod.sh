#!/usr/bin/env bash
# Run goose in Docker against prod postgres (internal docker network).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="${ENV_FILE:-.env}"
GOOSE_IMAGE="${GOOSE_IMAGE:-ghcr.io/kukymbr/goose-docker:3.27.1}"
COMPOSE_NETWORK="${COMPOSE_NETWORK:-anonimus-prod_default}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
COMPOSE_PROD_FILE="${COMPOSE_PROD_FILE:-docker-compose.prod.yml}"
SKIP_PULL=false
GOOSE_COMMAND=""
GOOSE_COMMAND_ARG=""

log() {
	echo "[migrate] $*"
}

die() {
	echo "[migrate] error: $*" >&2
	exit 1
}

usage() {
	cat <<EOF
usage: $0 [--no-pull] <command> [arg]

  up              apply pending migrations
  status          show migration status
  down            roll back one migration
  down-to VERSION roll back to goose version (e.g. 2 for 00002_*.sql)
  version         print current max applied version_id (0 if none)
EOF
}

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--no-pull)
			SKIP_PULL=true
			shift
			;;
		-h | --help)
			usage
			exit 0
			;;
		-*)
			die "unknown option: $1"
			;;
		*)
			break
			;;
		esac
	done

	GOOSE_COMMAND="${1:-up}"
	case "$GOOSE_COMMAND" in
	up | status | down | version)
		;;
	down-to)
		GOOSE_COMMAND_ARG="${2:-}"
		[[ -n "$GOOSE_COMMAND_ARG" ]] || die "down-to requires VERSION (e.g. 2)"
		;;
	*)
		die "unknown command: $GOOSE_COMMAND (try --help)"
		;;
	esac
}

load_env() {
	[[ -f "$ENV_FILE" ]] || die "env file not found: $ENV_FILE"
	set -a
	# shellcheck disable=SC1090
	source "$ENV_FILE"
	set +a
	POSTGRES_USER="${POSTGRES_USER:-anonimus}"
	POSTGRES_DB="${POSTGRES_DB:-anonimus}"
	[[ -n "${POSTGRES_PASSWORD:-}" ]] || die "POSTGRES_PASSWORD is not set in $ENV_FILE"
}

compose() {
	docker compose -f "$COMPOSE_FILE" -f "$COMPOSE_PROD_FILE" "$@"
}

migration_version() {
	local version
	version="$(compose exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc \
		"SELECT COALESCE(MAX(version_id), 0) FROM goose_db_version;" 2>/dev/null || true)"
	version="$(echo "$version" | tr -d '[:space:]')"
	if [[ -z "$version" || ! "$version" =~ ^[0-9]+$ ]]; then
		echo 0
		return
	fi
	echo "$version"
}

run_goose() {
	local -a env_args=(
		-e GOOSE_DRIVER=postgres
		-e GOOSE_MIGRATION_DIR=/migrations
		-e "GOOSE_DBSTRING=host=postgres port=5432 user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable"
		-e "GOOSE_COMMAND=${GOOSE_COMMAND}"
	)
	if [[ -n "$GOOSE_COMMAND_ARG" ]]; then
		env_args+=(-e "GOOSE_COMMAND_ARG=${GOOSE_COMMAND_ARG}")
	fi

	if [[ "$SKIP_PULL" != true ]]; then
		log "pulling $GOOSE_IMAGE"
		docker pull "$GOOSE_IMAGE"
	fi

	log "goose ${GOOSE_COMMAND}${GOOSE_COMMAND_ARG:+ ${GOOSE_COMMAND_ARG}} on network $COMPOSE_NETWORK"
	docker run --rm --network "$COMPOSE_NETWORK" \
		-v "$ROOT/migrations:/migrations" \
		"${env_args[@]}" \
		"$GOOSE_IMAGE"
}

main() {
	parse_args "$@"
	load_env

	if [[ "$GOOSE_COMMAND" == version ]]; then
		migration_version
		exit 0
	fi

	run_goose
	log "done"
}

main "$@"
