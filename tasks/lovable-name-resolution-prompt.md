# Lovable Prompt: Add Bot Name Resolution to `sandbox-exec`

## Goal

Allow users to pass either a sandbox ID or a bot name to `sandbox-exec`. The function should resolve bot names to sandbox IDs automatically.

## Current Behavior

`sandbox-exec` receives `sandbox_id` in the request body and passes it directly to the upstream sandbox API. If a user passes a bot name like `"test"` instead of a real sandbox ID like `"sandbox_abc123"`, the request fails with a 404.

## Desired Behavior

When `sandbox-exec` receives a `sandbox_id` value:

1. First, try to use it as a literal sandbox ID (current behavior)
2. If that fails with a 404, look up the `bots` table for a bot with that `name` belonging to the authenticated user
3. If a matching bot is found, retry the operation using that bot's actual `sandbox_id`
4. If no match is found, return the original 404 error

Alternatively (simpler approach): always check the `bots` table first. If the value matches a bot name, resolve it to the sandbox ID before making the upstream call.

## Request Body (unchanged)

```json
{
  "sandbox_id": "test",
  "action": "check-sandbox-status"
}
```

Where `sandbox_id` can now be either:
- A real sandbox ID (e.g. `"sandbox_abc123"`)
- A bot name (e.g. `"test"`, `"my-agent"`)

## Resolution Logic

```
resolve(sandbox_id, user_id):
  bot = SELECT * FROM bots WHERE (sandbox_id = $1 OR name = $1) AND user_id = $2
  if bot found:
    return bot.sandbox_id
  else:
    return error "Bot not found"
```

## Important

- Resolution must be scoped to the authenticated user — a user should only resolve their own bot names
- Bot names should be case-insensitive for lookup (e.g. `"Test"` matches `"test"`)
- This applies to ALL actions in `sandbox-exec`: `check-sandbox-status`, `start-sandbox`, `stop-sandbox`, `delete-sandbox`
- The `create-sandbox` endpoint does NOT need this — it creates new bots by name already
