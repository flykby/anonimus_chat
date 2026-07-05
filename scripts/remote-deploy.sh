#!/usr/bin/env bash
# Run on the production VM: sync repo and deploy a specific image tag.
set -euo pipefail

TAG="${1:-}"
if [[ -z "$TAG" ]]; then
	echo "usage: $0 <image-tag>" >&2
	echo "example: $0 $(git rev-parse --short HEAD 2>/dev/null || echo latest)" >&2
	exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

git fetch origin main
git checkout main
git pull --ff-only origin main

exec bash scripts/deploy.sh --tag "$TAG"
