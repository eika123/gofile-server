#!/usr/bin/env sh
set -eu

# Build the Go application.
# Run this from the repository root.

go test ./...
go build -o bin/file-server ./cmd/hello

echo "Built bin/file-server"
