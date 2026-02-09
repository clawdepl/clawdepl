#!/bin/bash
# Test script for local development and CI
# This script builds the Go binary and tests both npm and Python wrappers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== Building Go binary ==="
go build -o clawdepl .
./clawdepl --version

echo ""
echo "=== Testing npm package ==="
cp clawdepl npm/bin/
cd npm
node bin/clawdepl.js --version
rm -f bin/clawdepl
cd "$PROJECT_ROOT"

echo ""
echo "=== Testing Python package ==="
cp clawdepl python/clawdepl/
cd python
python -m clawdepl.cli --version
rm -f clawdepl/clawdepl
cd "$PROJECT_ROOT"

echo ""
echo "=== All tests passed! ==="
