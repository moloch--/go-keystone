#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="${ROOT_DIR}/dist"

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

export DOCKER_BUILDKIT=1

docker build \
  --target artifacts \
  --output "type=local,dest=${DIST_DIR}" \
  "${ROOT_DIR}"
