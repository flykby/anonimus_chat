#!/usr/bin/env bash
# Deploy or rollback production stack on VM: pull image → compose up → health check.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
COMPOSE_PROD_FILE="${COMPOSE_PROD_FILE:-docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-.env}"
DEPLOY_STATE_DIR="${DEPLOY_STATE_DIR:-.deploy}"
CONTAINER_NAME="${CONTAINER_NAME:-anonimus-bot}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-anonimus-postgres}"
MIGRATION_SKIP=false
MIGRATION_BEFORE=""
IMAGE_TAG=""
CLI_IMAGE_TAG=""
LAST_GOOD_TAG=""
ROLLBACK=false
WITH_PROXY=false
SKIP_PULL=false
IN_AUTO_ROLLBACK=false

log() {
	echo "[deploy] $*"
}

die() {
	echo "[deploy] error: $*" >&2
	exit 1
}

usage() {
	cat <<EOF
usage: $0 [--tag TAG] [--rollback] [--env-file PATH] [--with-proxy] [--skip-pull] [--skip-migrate]

  --tag TAG        Deploy specific image tag (default: IMAGE_TAG from .env)
  --rollback       Deploy previous successful tag from ${DEPLOY_STATE_DIR}/previous
  --env-file PATH  Env file path (default: .env)
  --with-proxy     Start Caddy reverse proxy profile
  --skip-pull      Skip docker pull (use local image)
  --skip-migrate   Skip goose migrations (emergency only)
EOF
}

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--tag)
			CLI_IMAGE_TAG="$2"
			shift 2
			;;
		--rollback)
			ROLLBACK=true
			shift
			;;
		--env-file)
			ENV_FILE="$2"
			shift 2
			;;
		--with-proxy)
			WITH_PROXY=true
			shift
			;;
		--skip-pull)
			SKIP_PULL=true
			shift
			;;
		--skip-migrate)
			MIGRATION_SKIP=true
			shift
			;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			die "unknown argument: $1 (try --help)"
			;;
		esac
	done
}

load_env() {
	[[ -f "$ENV_FILE" ]] || die "env file not found: $ENV_FILE (copy .env.prod.example)"
	set -a
	# shellcheck disable=SC1090
	source "$ENV_FILE"
	set +a
	if [[ -n "$CLI_IMAGE_TAG" ]]; then
		IMAGE_TAG="$CLI_IMAGE_TAG"
	fi
	export ENV_FILE
	export IMAGE_TAG
}

cleanup_stale_endpoints() {
	log "cleaning stale container endpoints"
	local -a names=(anonimus-postgres anonimus-redis anonimus-api anonimus-ai anonimus-bot)
	for name in "${names[@]}"; do
		local state
		state=$(docker inspect -f '{{.State.Status}}' "$name" 2>/dev/null || continue)
		if [[ "$state" == "created" ]]; then
			log "removing $name (state=created)"
			docker rm -f "$name" >/dev/null 2>&1 || true
		fi
	done
}

resolve_tag() {
	if [[ "$ROLLBACK" == true ]]; then
		local prev_file="${DEPLOY_STATE_DIR}/previous"
		[[ -f "$prev_file" ]] || die "no previous deploy tag at $prev_file"
		IMAGE_TAG="$(tr -d '[:space:]' <"$prev_file")"
		log "rollback to tag: $IMAGE_TAG"
		return
	fi

	if [[ -z "$IMAGE_TAG" ]]; then
		die "IMAGE_TAG is not set (use --tag or set in $ENV_FILE)"
	fi
}

capture_last_good_tag() {
	LAST_GOOD_TAG=""
	if [[ "$ROLLBACK" == true ]]; then
		return 0
	fi
	local current_file="${DEPLOY_STATE_DIR}/current"
	if [[ -f "$current_file" ]]; then
		LAST_GOOD_TAG="$(tr -d '[:space:]' <"$current_file")"
		if [[ -n "$LAST_GOOD_TAG" ]]; then
			log "last good deploy tag: $LAST_GOOD_TAG"
		fi
	fi
}

registry_login() {
	[[ -n "${REGISTRY_URL:-}" ]] || die "REGISTRY_URL is not set"
	[[ -n "${REGISTRY_HOST:-}" ]] || REGISTRY_HOST="${REGISTRY_URL%%/*}"
	if [[ -n "${REGISTRY_PASSWORD:-}" ]]; then
		log "logging in to $REGISTRY_HOST"
		echo "$REGISTRY_PASSWORD" | docker login "$REGISTRY_HOST" -u "${REGISTRY_USER:-}" --password-stdin
	fi
}

pull_images() {
	if [[ "$SKIP_PULL" == true ]]; then
		log "skip pull"
		return
	fi
	for svc in bot api ai; do
		local image="${REGISTRY_URL}/${svc}:${IMAGE_TAG}"
		log "pulling $image"
		docker pull "$image"
	done
}

compose() {
	docker compose -f "$COMPOSE_FILE" -f "$COMPOSE_PROD_FILE" "$@"
}

compose_up() {
	export IMAGE_TAG
	local -a cmd=(compose up -d --remove-orphans)
	if [[ "$WITH_PROXY" == true ]]; then
		cmd+=(--profile proxy)
	fi
	log "starting stack (tag=$IMAGE_TAG)"
	"${cmd[@]}"
}

wait_postgres_healthy() {
	log "waiting for healthy postgres ($POSTGRES_CONTAINER)"
	local attempts=30
	for _ in $(seq 1 "$attempts"); do
		local status
		status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$POSTGRES_CONTAINER" 2>/dev/null || echo "")"
		if [[ "$status" == "healthy" ]]; then
			log "postgres is healthy"
			return 0
		fi
		sleep 2
	done
	die "postgres did not become healthy within $((attempts * 2))s"
}

run_migrations() {
	if [[ "$MIGRATION_SKIP" == true ]]; then
		log "skip migrations"
		return 0
	fi

	if [[ "$ROLLBACK" == true ]]; then
		local target_file="${DEPLOY_STATE_DIR}/migration_previous"
		if [[ ! -f "$target_file" ]]; then
			log "no migration rollback target (${target_file}) — skip"
			return 0
		fi
		local target
		target="$(tr -d '[:space:]' <"$target_file")"
		[[ "$target" =~ ^[0-9]+$ ]] || die "invalid migration rollback target: $target"
		local current
		current="$(bash "$ROOT/scripts/migrate-prod.sh" --no-pull version)"
		if [[ "$current" -le "$target" ]]; then
			log "migrations already at version $current (rollback target $target)"
			return 0
		fi
		log "rolling back migrations: $current -> $target"
		bash "$ROOT/scripts/migrate-prod.sh" --no-pull down-to "$target"
		return 0
	fi

	MIGRATION_BEFORE="$(bash "$ROOT/scripts/migrate-prod.sh" --no-pull version)"
	log "applying migrations (current version=$MIGRATION_BEFORE)"
	bash "$ROOT/scripts/migrate-prod.sh" --no-pull up
	local after
	after="$(bash "$ROOT/scripts/migrate-prod.sh" --no-pull version)"
	log "migrations applied: $MIGRATION_BEFORE -> $after"
}

rollback_migrations_on_failure() {
	if [[ "$MIGRATION_SKIP" == true || "$ROLLBACK" == true ]]; then
		return 0
	fi
	if [[ -z "$MIGRATION_BEFORE" ]]; then
		return 0
	fi
	local current
	current="$(bash "$ROOT/scripts/migrate-prod.sh" --no-pull version 2>/dev/null || echo 0)"
	if [[ "$current" -le "$MIGRATION_BEFORE" ]]; then
		return 0
	fi
	log "deploy failed — rolling back migrations to version $MIGRATION_BEFORE"
	bash "$ROOT/scripts/migrate-prod.sh" --no-pull down-to "$MIGRATION_BEFORE" || \
		log "warning: migration rollback failed (check postgres manually)"
}

on_deploy_failure() {
	if [[ "$IN_AUTO_ROLLBACK" == true ]]; then
		return 0
	fi
	rollback_migrations_on_failure
}

wait_healthy() {
	log "waiting for healthy bot container ($CONTAINER_NAME)"
	local attempts=30
	for _ in $(seq 1 "$attempts"); do
		local status
		status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$CONTAINER_NAME" 2>/dev/null || echo "")"
		if [[ "$status" == "healthy" ]]; then
			log "health check passed"
			if [[ -n "${WEBHOOK_URL:-}" ]]; then
				compose exec -T bot wget -qO- --no-check-certificate https://127.0.0.1:8080/health 2>/dev/null || true
			else
				compose exec -T bot wget -qO- http://127.0.0.1:8080/health 2>/dev/null || true
			fi
			return 0
		fi
		sleep 2
	done
	log "bot did not become healthy within $((attempts * 2))s"
	return 1
}

# After a failed health check: restore DB schema and last known-good images.
auto_rollback_after_health_failure() {
	local failed_tag="$IMAGE_TAG"
	log "health check failed for tag=$failed_tag — starting auto-rollback"

	IN_AUTO_ROLLBACK=true
	rollback_migrations_on_failure

	if [[ "$ROLLBACK" == true ]]; then
		log "already in --rollback mode; skipping image restore"
		return 1
	fi
	if [[ -z "$LAST_GOOD_TAG" ]]; then
		log "no previous successful tag in ${DEPLOY_STATE_DIR}/current — cannot restore images"
		return 1
	fi
	if [[ "$LAST_GOOD_TAG" == "$failed_tag" ]]; then
		log "last good tag equals failed tag ($failed_tag) — cannot restore images"
		return 1
	fi

	IMAGE_TAG="$LAST_GOOD_TAG"
	export IMAGE_TAG
	log "restoring previous images: $IMAGE_TAG"
	if ! pull_images; then
		log "warning: pull of rollback images failed"
		return 1
	fi
	if ! compose_up; then
		log "warning: compose up of rollback images failed"
		return 1
	fi
	if wait_healthy; then
		log "auto-rollback succeeded: ${REGISTRY_URL}/*:${IMAGE_TAG}"
		show_status
		return 0
	fi
	log "error: auto-rollback to $IMAGE_TAG also failed health check"
	return 1
}

save_deploy_state() {
	mkdir -p "$DEPLOY_STATE_DIR"
	local current_file="${DEPLOY_STATE_DIR}/current"
	local previous_file="${DEPLOY_STATE_DIR}/previous"
	local migration_current_file="${DEPLOY_STATE_DIR}/migration_current"
	local migration_previous_file="${DEPLOY_STATE_DIR}/migration_previous"

	if [[ -f "$current_file" ]] && [[ "$ROLLBACK" != true ]]; then
		cp "$current_file" "$previous_file"
	fi
	echo "$IMAGE_TAG" >"$current_file"
	log "saved deploy tag to ${current_file}"
	if [[ -f "$previous_file" ]]; then
		log "rollback tag available: $(tr -d '[:space:]' <"$previous_file")"
	fi

	if [[ "$MIGRATION_SKIP" == true ]]; then
		return 0
	fi

	if [[ "$ROLLBACK" == true ]]; then
		if [[ -f "$migration_previous_file" ]]; then
			local target
			target="$(tr -d '[:space:]' <"$migration_previous_file")"
			echo "$target" >"$migration_current_file"
			log "migration version after rollback: $target"
		fi
		return 0
	fi

	if [[ -n "$MIGRATION_BEFORE" ]]; then
		echo "$MIGRATION_BEFORE" >"$migration_previous_file"
	fi
	local migration_after
	migration_after="$(bash "$ROOT/scripts/migrate-prod.sh" --no-pull version)"
	echo "$migration_after" >"$migration_current_file"
	log "saved migration version $migration_after (rollback target: $MIGRATION_BEFORE)"
}

show_status() {
	compose ps
}

main() {
	parse_args "$@"
	load_env
	resolve_tag
	capture_last_good_tag
	registry_login
	pull_images
	cleanup_stale_endpoints
	trap on_deploy_failure ERR
	compose up -d postgres redis
	wait_postgres_healthy
	run_migrations
	compose_up
	if ! wait_healthy; then
		trap - ERR
		local failed_tag="$IMAGE_TAG"
		if auto_rollback_after_health_failure; then
			die "deploy of ${failed_tag} failed health check; rolled back to ${IMAGE_TAG}"
		fi
		die "deploy of ${failed_tag} failed health check; auto-rollback could not restore previous version"
	fi
	trap - ERR
	save_deploy_state
	show_status
	log "deploy complete: ${REGISTRY_URL}/*:${IMAGE_TAG}"
}

main "$@"
