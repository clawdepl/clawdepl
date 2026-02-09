# clawdepl API Reference

**Base URL:** `https://ghcvpnixedcjpokvbsep.supabase.co/functions/v1`

**Authentication:** All endpoints accept `Authorization: Bearer <TOKEN>` where `<TOKEN>` is either a clawdepl API token (from `api_tokens` table) or a Supabase JWT.

---

## 1. `POST /verify-token`

Validate an API token and get user identity.

**Request:**
```bash
curl -X POST ${BASE_URL}/verify-token \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json"
```

**Response (200):**
```json
{
  "valid": true,
  "userId": "uuid-here",
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Response (401):**
```json
{ "valid": false, "error": "Invalid or revoked token" }
```

---

## 2. `POST /create-sandbox`

Provision a new sandbox (bot) running the `diogoiwasaki/openclaw-base` Docker image. Enforces tier-based bot limits.

**Request:**
```bash
curl -X POST ${BASE_URL}/create-sandbox \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-agent",
    "openclaw_config": "{\"key\": \"value\"}",
    "molty_prompt": "You are a helpful assistant..."
  }'
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | No | Bot name (max 50 chars). Defaults to `bot-<timestamp>` |
| `openclaw_config` | string | No | Passed as `OPENCLAW_CONFIG` env var to the container |
| `molty_prompt` | string | No | Passed as `MOLTY_PROMPT` env var to the container |

**Container details:**
- Image: `diogoiwasaki/openclaw-base`
- Exposed ports: `18789`, `18790`
- Environment variables: `OPENCLAW_CONFIG`, `MOLTY_PROMPT`

**Response (200):**
```json
{
  "success": true,
  "bot_name": "my-agent",
  "sandbox": {
    "id": "sandbox-id-from-daytona",
    "state": "...",
    "...": "..."
  }
}
```

**Response (403) — limit reached:**
```json
{
  "error": "Bot limit reached. Your free plan allows 2 concurrent bots. Upgrade to Pro for up to 10 bots.",
  "tier": "free",
  "limit": 2,
  "active": 2
}
```

**Response (409) — duplicate name:**
```json
{ "error": "A bot named \"my-agent\" already exists. Please choose a different name." }
```

---

## 3. `POST /sandbox-exec`

Multi-action endpoint for sandbox lifecycle and command execution. All actions require `sandbox_id` in the body.

### 3a. `check-sandbox-status`

```json
// Request
{ "sandbox_id": "abc123", "action": "check-sandbox-status" }

// Response (200)
{ "state": "started", "ready": true }
```

### 3b. `start-sandbox`

```json
// Request
{ "sandbox_id": "abc123", "action": "start-sandbox" }

// Response (200)
{ "success": true }
```

### 3c. `stop-sandbox`

```json
// Request
{ "sandbox_id": "abc123", "action": "stop-sandbox" }

// Response (200)
{ "success": true }
```

### 3d. `delete-sandbox`

Stops and deletes the sandbox, marks bot as `"deleted"` in DB.

```json
// Request
{ "sandbox_id": "abc123", "action": "delete-sandbox" }

// Response (200)
{ "success": true }
```

### 3e. `create-session`

Create a terminal session inside a sandbox.

```json
// Request
{ "sandbox_id": "abc123", "action": "create-session", "session_id": "my-session" }

// Response (200)
{ "success": true, "session_id": "my-session" }
```

### 3f. `exec`

Run a command synchronously (for quick commands).

```json
// Request
{ "sandbox_id": "abc123", "action": "exec", "session_id": "my-session", "command": "pwd" }

// Response (200)
{ "output": "/home/daytona\n", "exit_code": 0 }
```

### 3g. `stream-exec`

Run a command asynchronously with SSE streaming output.

```json
// Request
{ "sandbox_id": "abc123", "action": "stream-exec", "session_id": "my-session", "command": "apt update", "timeout": 120 }
```

**Response:** SSE stream with events:
```
data: {"type":"started","cmd_id":"cmd-xyz"}

data: {"type":"output","content":"Hit:1 http://deb.debian.org/debian bookworm InRelease\n"}

data: {"type":"output","content":"Reading package lists..."}

data: {"type":"done","exit_code":0}
```

Error event: `{"type":"error","content":"Command timed out"}`

### 3h. `delete-session`

```json
// Request
{ "sandbox_id": "abc123", "action": "delete-session", "session_id": "my-session" }

// Response (200)
{ "success": true }
```

---

## 4. `POST /sandbox-terminal-url`

SSH access management for sandboxes.

### 4a. `create-ssh`

```json
// Request
{ "sandbox_id": "abc123", "action": "create-ssh", "expires_in_minutes": 60 }

// Response (200)
{
  "ssh_token": "token-string",
  "ssh_command": "ssh token-string@ssh.app.daytona.io",
  "expires_in_minutes": 60,
  "sandbox_id": "abc123"
}
```

### 4b. `revoke-ssh`

```json
// Request
{ "sandbox_id": "abc123", "action": "revoke-ssh", "ssh_token": "token-string" }

// Response (200)
{ "success": true, "sandbox_id": "abc123" }
```

---

## 5. `POST /check-subscription` *(JWT only)*

Check current user's subscription status and active bot count.

**Request:**
```bash
curl -X POST ${BASE_URL}/check-subscription \
  -H "Authorization: Bearer <JWT>"
```

**Response (200):**
```json
{
  "subscribed": true,
  "product_id": "prod_TwlLmFbDKg0nGt",
  "subscription_end": "2025-08-15T00:00:00.000Z",
  "active_bot_count": 3
}
```

**Response (200) — free tier:**
```json
{
  "subscribed": false,
  "product_id": null,
  "subscription_end": null,
  "active_bot_count": 1
}
```

---

## 6. `POST /create-checkout` *(JWT only)*

Create a Stripe checkout session for a subscription.

**Request:**
```json
{ "price_id": "price_xxx" }
```

**Response (200):**
```json
{ "url": "https://checkout.stripe.com/c/pay/..." }
```

---

## 7. `POST /customer-portal` *(JWT only)*

Get a Stripe billing portal URL.

**Request:** *(empty body)*

**Response (200):**
```json
{ "url": "https://billing.stripe.com/p/session/..." }
```

---

## Error Format (all endpoints)

```json
{ "error": "Human-readable error message" }
```

HTTP status codes: `400` (bad request), `401` (auth failure), `403` (limit/permission), `409` (conflict), `500` (server error), `502` (upstream Daytona failure).
