#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$ROOT_DIR"

# Load .env values into the environment from the repo root.
ENV_FILE="$ROOT_DIR/.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  . "$ENV_FILE"
  set +a
fi

# Override the runtime root path for the container.
export ROOT_PATH=/app/shared

DEV_MODE=false
if [ "$#" -ge 1 ] && [ "$1" = "--dev" ]; then
  DEV_MODE=true
fi

if [ "$DEV_MODE" = "true" ]; then
  echo "Starting in development mode; using non-privileged ports."
  export PROXY_PORT=${PROXY_PORT:-8080}
  export PROXY_SSL_PORT=${PROXY_SSL_PORT:-8443}
  export APP_PORT=${APP_PORT:-8081}
fi

if command -v podman-compose >/dev/null 2>&1; then
  podman-compose -f podman-compose.yml up -d --build
elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
  podman compose -f podman-compose.yml up -d --build
else
  echo "ERROR: podman-compose or podman compose is required."
  echo "Install the Podman compose plugin or podman-compose package."
  exit 1
fi
