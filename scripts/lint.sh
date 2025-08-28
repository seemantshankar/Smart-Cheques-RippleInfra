#!/bin/bash

# Script to run golangci-lint with different options

set -e

cd "$(dirname "$0")/.."

case "${1:-default}" in
  "fix")
    echo "Running golangci-lint with auto-fix..."
    golangci-lint run --fix
    ;;
  "short")
    echo "Running golangci-lint with short timeout..."
    golangci-lint run --timeout=2m
    ;;
  "verbose")
    echo "Running golangci-lint with verbose output..."
    golangci-lint run --verbose
    ;;
  "default"|*)
    echo "Running golangci-lint..."
    golangci-lint run
    ;;
esac

echo "Linting completed successfully!"