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
IMAGE_TAG=""
CLI_IMAGE_TAG=""
ROLLBACK=false
WITH_PROXY=false
SKIP_PULL=false

log() {
	echo "[deploy] $*"
}

die() {
	echo "[deploy] error: $*" >&2
	exit 1
}

usage() {
	cat <<EOF
usage: $0 [--tag TAG] [--rollback] [--env-file PATH] [--with-proxy] [--skip-pull]

  --tag TAG        Deploy specific image tag (default: IMAGE_TAG from .env)
  --rollback       Deploy previous successful tag from ${DEPLOY_STATE_DIR}/previous
  --env-file PATH  Env file path (default: .env)
  --with-proxy     Start Caddy reverse proxy profile
  --skip-pull      Skip docker pull (use local image)
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

wait_healthy() {
	log "waiting for healthy bot container ($CONTAINER_NAME)"
	local attempts=30
	for _ in $(seq 1 "$attempts"); do
		local status
		status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$CONTAINER_NAME" 2>/dev/null || echo "")"
		if [[ "$status" == "healthy" ]]; then
			log "health check passed"
			compose exec -T bot wget -qO- http://127.0.0.1:8080/health 2>/dev/null || true
			return 0
		fi
		sleep 2
	done
	die "bot did not become healthy within $((attempts * 2))s"
}

save_deploy_state() {
	mkdir -p "$DEPLOY_STATE_DIR"
	local current_file="${DEPLOY_STATE_DIR}/current"
	local previous_file="${DEPLOY_STATE_DIR}/previous"

	if [[ -f "$current_file" ]] && [[ "$ROLLBACK" != true ]]; then
		cp "$current_file" "$previous_file"
	fi
	echo "$IMAGE_TAG" >"$current_file"
	log "saved deploy tag to ${current_file}"
	if [[ -f "$previous_file" ]]; then
		log "rollback tag available: $(tr -d '[:space:]' <"$previous_file")"
	fi
}

show_status() {
	compose ps
}

main() {
	parse_args "$@"
	load_env
	resolve_tag
	registry_login
	pull_images
	cleanup_stale_endpoints
	compose_up
	wait_healthy
	save_deploy_state
	show_status
	log "deploy complete: ${REGISTRY_URL}/*:${IMAGE_TAG}"
}

main "$@"
