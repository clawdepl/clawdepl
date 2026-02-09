# Lovable Prompt: Create `list-bots` Edge Function

## Goal

Create a new Supabase edge function `list-bots` that returns all bots/sandboxes owned by the authenticated user.

## Endpoint

`POST /list-bots`

## Auth

- Requires `Authorization: Bearer <token>` header
- Validate the token the same way `verify-token` does
- Return 401 if invalid

## Request Body

```json
{}
```

No parameters needed — the user is identified from the token.

## Response

```json
{
  "bots": [
    {
      "id": "bot-uuid",
      "name": "my-agent",
      "sandbox_id": "sandbox-abc123",
      "state": "running",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Bot record ID |
| `name` | string | User-given name for the bot |
| `sandbox_id` | string | E2B sandbox identifier |
| `state` | string | Current state: `running`, `stopped`, `provisioning`, `error` |
| `created_at` | string | ISO-8601 creation timestamp |

## Implementation Notes

- Query the `bots` table (or equivalent) filtering by `user_id` from the validated token
- Sort by `created_at` descending (newest first)
- Return empty array if user has no bots
- Follow the same patterns used in `create-sandbox` and `sandbox-exec` for auth validation and error handling
