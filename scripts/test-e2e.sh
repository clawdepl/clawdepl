#!/usr/bin/env bash
# End-to-end integration tests against real backend and Anthropic API.
# Requires repository secrets/environment variables:
#   CLAWDEPL_E2E_TOKEN
#   CLAWDEPL_E2E_ANTHROPIC_KEY
#   CLAWDEPL_E2E_INVALID_ANTHROPIC_KEY
# Optional:
#   CLAWDEPL_E2E_AUTH_CHOICE (api-key|setup-token, default: api-key)
#   CLAWDEPL_E2E_MODEL (default: anthropic/claude-opus-4-6)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

require_env() {
  local var_name="$1"
  if [[ -z "${!var_name:-}" ]]; then
    echo "Missing required environment variable: $var_name" >&2
    exit 1
  fi
}

require_env "CLAWDEPL_E2E_TOKEN"
require_env "CLAWDEPL_E2E_ANTHROPIC_KEY"
require_env "CLAWDEPL_E2E_INVALID_ANTHROPIC_KEY"

BIN_PATH="${BIN_PATH:-./clawdepl}"
if [[ ! -x "$BIN_PATH" ]]; then
  echo "Binary not found or not executable: $BIN_PATH" >&2
  exit 1
fi

TEST_HOME="$(mktemp -d)"
export HOME="$TEST_HOME"

TEST_ID="$(date +%s)"
INSTANCE_NAME="e2e-${TEST_ID}"
INSTANCE_TO_CLEAN=""
AUTH_CHOICE="${CLAWDEPL_E2E_AUTH_CHOICE:-api-key}"
MODEL_REF="${CLAWDEPL_E2E_MODEL:-anthropic/claude-opus-4-6}"

cleanup() {
  set +e
  if [[ -n "$INSTANCE_TO_CLEAN" ]]; then
    sandbox_id="$("$BIN_PATH" list 2>/dev/null | grep -E "^${INSTANCE_TO_CLEAN}[[:space:]]" | grep -oE 'sandbox_[A-Za-z0-9_-]+' | head -n1)"
    if [[ -n "$sandbox_id" ]]; then
      "$BIN_PATH" delete "$sandbox_id" --yes >/dev/null 2>&1 || true
    fi
  fi
  rm -rf "$TEST_HOME"
}
trap cleanup EXIT

expect_fail() {
  local description="$1"
  shift
  if "$@"; then
    echo "Expected failure but command succeeded: $description" >&2
    exit 1
  fi
}

echo "=== E2E: invalid login token should fail ==="
expect_fail "invalid login token" "$BIN_PATH" login --token "invalid-token-for-e2e"

echo "=== E2E: valid login token should succeed ==="
"$BIN_PATH" login --token "$CLAWDEPL_E2E_TOKEN"

echo "=== E2E: invalid Anthropic token should fail before provisioning ==="
expect_fail \
  "invalid Anthropic token pre-validation" \
  "$BIN_PATH" new "e2e-invalid-${TEST_ID}" \
    --claude-token "$CLAWDEPL_E2E_INVALID_ANTHROPIC_KEY" \
    --auth-choice "$AUTH_CHOICE" \
    --model "$MODEL_REF" \
    --purpose "E2E invalid token validation"

echo "=== E2E: valid Anthropic token should provision successfully ==="
"$BIN_PATH" new "$INSTANCE_NAME" \
  --claude-token "$CLAWDEPL_E2E_ANTHROPIC_KEY" \
  --auth-choice "$AUTH_CHOICE" \
  --model "$MODEL_REF" \
  --purpose "E2E validation instance"
INSTANCE_TO_CLEAN="$INSTANCE_NAME"

echo "=== E2E: verify instance appears in list ==="
list_output="$("$BIN_PATH" list)"
echo "$list_output"
if ! grep -qE "^${INSTANCE_NAME}[[:space:]]" <<<"$list_output"; then
  echo "Provisioned instance not found in list output" >&2
  exit 1
fi

echo "=== E2E completed successfully ==="
