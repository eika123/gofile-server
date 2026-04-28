#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$ROOT_DIR"

usage() {
  cat <<EOF
Usage: $0 [--follow] [--file] [service ...]

Show logs for the Podman compose stack.
If no service names are given, logs for all services are shown.

Options:
  -f, --follow   Follow log output.
  --file         Read logs from the persisted host log files instead of Podman logs.
  -h, --help     Show this help message.
EOF
}

FOLLOW=false
FILE_MODE=false
SERVICES=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    -f|--follow)
      FOLLOW=true
      shift
      ;;
    --file)
      FILE_MODE=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    *)
      SERVICES="$SERVICES $1"
      shift
      ;;
  esac
done

if [ "$FILE_MODE" = true ]; then
  if [ -z "$SERVICES" ]; then
    tail -n 100 -f ./logs/file-server/app.log ./logs/nginx/error.log || true
    exit 0
  fi

  for service in $SERVICES; do
    case "$service" in
      file-server)
        tail -n 100 -f ./logs/file-server/app.log || true
        ;;
      reverse-proxy)
        tail -n 100 -f ./logs/nginx/error.log || true
        ;;
      *)
        echo "Unknown service: $service"
        exit 1
        ;;
    esac
  done
  exit 0
fi

COMMAND_ARGS="logs"
if [ "$FOLLOW" = true ]; then
  COMMAND_ARGS="$COMMAND_ARGS -f"
fi
if [ -n "$SERVICES" ]; then
  COMMAND_ARGS="$COMMAND_ARGS$SERVICES"
fi

if command -v podman-compose >/dev/null 2>&1; then
  podman-compose -f podman-compose.yml $COMMAND_ARGS
elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
  podman compose -f podman-compose.yml $COMMAND_ARGS
else
  echo "ERROR: podman-compose or podman compose is required."
  echo "Install the Podman compose plugin or podman-compose package."
  exit 1
fi
