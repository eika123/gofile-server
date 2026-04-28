#!/usr/bin/env sh
set -eu

# Clean build artifacts and temporary files.
# Run this from the repository root.

rm -rf bin
find . -name '*.test' -type f -delete
find . -name '*.out' -type f -delete

echo "Cleaned build artifacts"
