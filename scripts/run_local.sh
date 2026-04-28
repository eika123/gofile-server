#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$ROOT_DIR"

ENV_FILE="$ROOT_DIR/.env"
if [ -f "$ENV_FILE" ]; then
  set -a
  . "$ENV_FILE"
  set +a
fi

if [ -z "${BASIC_AUTH_USER:-}" ] || [ -z "${BASIC_AUTH_PASS:-}" ]; then
  echo "ERROR: BASIC_AUTH_USER and BASIC_AUTH_PASS must be set in .env or the environment."
  exit 1
fi

exec ./bin/file-server "$@"
