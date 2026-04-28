#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
cd "$ROOT_DIR"

usage() {
  cat <<EOF
Usage: $0 [--clean] [source-directory]

Copy the contents of a host directory into the Podman-managed shared volume.
If no source directory is provided, the repository-local ./shared directory is used.

Options:
  --clean   Remove existing shared volume contents before copying.
  -h, --help  Show this help message.
EOF
}

CLEAN=false
SOURCE_DIR="./shared"
VOLUME_NAME=${VOLUME_NAME:-shared_data}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --clean)
      CLEAN=true
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
      SOURCE_DIR="$1"
      shift
      ;;
  esac
done

if [ ! -d "$SOURCE_DIR" ]; then
  echo "ERROR: source directory does not exist or is not a directory: $SOURCE_DIR"
  exit 1
fi

SOURCE_DIR=$(cd "$SOURCE_DIR" && pwd)

if ! command -v podman >/dev/null 2>&1; then
  echo "ERROR: podman is required to copy files into the shared volume."
  exit 1
fi

cat <<EOF
Copying from host directory: $SOURCE_DIR
Into Podman volume: $VOLUME_NAME
EOF

podman run --rm \
  -v "$VOLUME_NAME":/mnt/shared:Z \
  sh -euxc '
    mkdir -p /mnt/shared
    if [ "'$CLEAN'" = "true" ]; then
      find /mnt/shared -mindepth 1 -delete
    fi
    cp -a /tmp/source/. /mnt/shared/
  '

echo "Shared volume update complete."
