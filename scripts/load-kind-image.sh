#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <image>" >&2
  echo "Example: $0 my-app:latest" >&2
  exit 1
fi

IMAGE="$1"

kind load docker-image "$IMAGE" --name rp-practice
