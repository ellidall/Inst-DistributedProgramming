#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

kind delete cluster --name rp-practice
rm -rf "${PROJECT_ROOT}/data"
