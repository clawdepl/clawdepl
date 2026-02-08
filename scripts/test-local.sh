#!/bin/bash
# Test script for local development and CI
# This script builds the Go binary and tests both npm and Python wrappers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== Building Go binary ==="
go build -o clawdpl .
./clawdpl --version

echo ""
echo "=== Testing npm package ==="
cp clawdpl npm/bin/
cd npm
node bin/clawdpl.js --version
rm -f bin/clawdpl
cd "$PROJECT_ROOT"

echo ""
echo "=== Testing Python package ==="
cp clawdpl python/clawdpl/
cd python
python -m clawdpl.cli --version
rm -f clawdpl/clawdpl
cd "$PROJECT_ROOT"

echo ""
echo "=== All tests passed! ==="
