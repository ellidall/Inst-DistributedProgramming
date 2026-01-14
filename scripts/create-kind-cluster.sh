#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

mkdir -p \
  "${PROJECT_ROOT}/data/master" \
  "${PROJECT_ROOT}/data/worker1" \
  "${PROJECT_ROOT}/data/worker2" \
  "${PROJECT_ROOT}/data/worker3"

sed "s#__PROJECT_ROOT__#${PROJECT_ROOT}#g" "${PROJECT_ROOT}/kind-config.yaml" | kind create cluster --config -
