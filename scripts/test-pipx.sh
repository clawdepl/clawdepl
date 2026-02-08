#!/bin/bash
# Test pipx execution with local package
# This simulates what users will experience with `pipx run clawdpl`

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== Building Go binary ==="
go build -o clawdpl .

echo ""
echo "=== Preparing Python package ==="
cp clawdpl python/clawdpl/

echo ""
echo "=== Testing pipx run with local package ==="
cd python

# Use pipx run with --spec to point to local directory
pipx run --spec . clawdpl --version

echo ""
echo "=== Cleanup ==="
rm -f clawdpl/clawdpl

echo ""
echo "=== pipx test passed! ==="
