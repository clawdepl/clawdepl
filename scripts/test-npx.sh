#!/bin/bash
# Test npx execution with local package
# This simulates what users will experience with `npx clawdepl`

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== Building Go binary ==="
go build -o clawdepl .

echo ""
echo "=== Preparing npm package ==="
cp clawdepl npm/bin/

echo ""
echo "=== Creating npm tarball ==="
cd npm
npm pack

echo ""
echo "=== Testing npx with local tarball ==="
TARBALL=$(ls clawdepl-*.tgz | head -1)
npx "./$TARBALL" --version

echo ""
echo "=== Cleanup ==="
rm -f "$TARBALL" bin/clawdepl

echo ""
echo "=== npx test passed! ==="
