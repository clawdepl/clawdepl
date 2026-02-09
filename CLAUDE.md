## Project Architecture

This is a **multi-repo workspace** with three interconnected repositories:

### Repository Structure

```
clawdepl-repos/
├── clawdepl/          # THIS REPO - Go CLI tool for managing OpenClaw instances
├── backend/           # Convex-based backend (provisioning, data storage)
└── quick-claw-cloud/  # Lovable-based React frontend (OAuth, dashboard UI)
```

### 1. clawdepl (Current Repo)
- **Tech**: Go CLI using Cobra
- **Purpose**: Command-line tool for provisioning and managing OpenClaw instances
- **Key paths**:
  - `cmd/` - CLI commands (new, list, delete, start, stop, status, login, logout)
  - `internal/api/` - Backend API client
  - `internal/auth/` - Authentication handling
  - `hidden-docs/` - **API documentation and integration specs** (check here for backend API details)
    - `api-endpoints.md` - Backend API endpoint reference
    - `clawdpl-backend-integration-brief-v2.md` - Integration guide
    - `api-1.json` - API schema

### 2. backend (Sibling: `../backend/`)
- **Tech**: Convex (serverless backend), TypeScript
- **Purpose**: Stores all provisioning data, user data, instance management
- **Key paths**:
  - `packages/platform/convex/` - Convex functions and schema
    - `schema.ts` - Database schema
    - `moltys.ts`, `users.ts`, `verses.ts` - Core data models
    - `activation.ts` - Instance activation logic
    - `http.ts` - HTTP endpoints
  - `packages/provisioner/` - Instance provisioning service
  - `packages/switchboard/` - Message routing/Discord adapter
  - `docs/` - Architecture docs, PRDs, API specs
    - `api/openapi.yaml` - OpenAPI specification
    - `ARCHITECTURE-*.md` - System architecture docs
    - `PRD-*.md` - Product requirement docs

### 3. quick-claw-cloud (Sibling: `../quick-claw-cloud/`)
- **Tech**: React + Vite + Tailwind (Lovable-generated), Supabase
- **Purpose**: Web dashboard and **OAuth authentication layer**
- **Key paths**:
  - `src/pages/` - Main pages (Auth.tsx, CLIAuth.tsx, AppDashboard.tsx)
  - `src/integrations/supabase/` - Supabase client and types
  - `supabase/functions/` - Edge functions (webhooks)

### Cross-Repo Dependencies

```
┌─────────────────┐     OAuth      ┌───────────────────┐
│  quick-claw-    │◄──────────────►│     clawdepl      │
│     cloud       │                │      (CLI)        │
│   (Frontend)    │                └─────────┬─────────┘
└────────┬────────┘                          │
         │                                   │ API calls
         │ Supabase Auth                     │
         │                                   ▼
         │                          ┌─────────────────┐
         └─────────────────────────►│    backend      │
                                    │   (Convex)      │
                                    └─────────────────┘
```

### When to Look Where

| Need to understand...          | Look in...                                      |
|--------------------------------|-------------------------------------------------|
| CLI commands & flags           | `clawdepl/cmd/`                                 |
| Backend API endpoints          | `clawdepl/hidden-docs/`, `backend/docs/api/`    |
| Database schema                | `backend/packages/platform/convex/schema.ts`   |
| OAuth/Auth flow                | `quick-claw-cloud/src/pages/Auth.tsx`           |
| CLI auth integration           | `quick-claw-cloud/src/pages/CLIAuth.tsx`        |
| Provisioning logic             | `backend/packages/provisioner/`                 |
| Architecture decisions         | `backend/docs/ARCHITECTURE-*.md`                |

---

## Workflow Orchestration

### 1. Plan Mode Default
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately — don’t keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### 2. Subagent Strategy
- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution

### 3. Self-Improvement Loop
- After ANY correction from the user, update `tasks/lessons.md` with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until mistake rate drops
- Review lessons at session start for relevant project

### 4. Verification Before Done
- Never mark a task complete without proving it works
- Diff your behavior between main and your changes when relevant
- Ask yourself: “Would a staff engineer approve this?”
- Run tests, check logs, demonstrate correctness

### 5. Demand Elegance (Balanced)
- For non-trivial changes: pause and ask “is there a more elegant way?”
- If a fix feels hacky: “Knowing everything I know now, implement the elegant solution”
- Skip this for simple, obvious fixes — don’t over-engineer
- Challenge your own work before presenting it

### 6. Autonomous Bug Fixing
- When given a bug report: just fix it. Don’t ask for hand-holding
- Point at logs, errors, failing tests — then resolve them
- Zero context switching required from the user
- Go fix failing CI tests without being told how

## Task Management

1. **Plan First**: Write plan to `tasks/todo.md` with checkable items  
2. **Verify Plan**: Check in before starting implementation  
3. **Track Progress**: Mark items complete as you go  
4. **Explain Changes**: High-level summary at each step  
5. **Document Results**: Add review section to `tasks/todo.md`  
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections  

## Documentation Requirements

### CLI Documentation Updates
When making ANY changes to the CLI interface (commands, subcommands, flags, aliases, help text), you MUST update:

1. **`docs/cli-reference.md`** - Full documentation with examples
2. **`llms.txt`** - Concise LLM-friendly reference

Changes that require documentation updates:
- Adding/removing/renaming commands or subcommands
- Adding/removing/changing flags
- Changing command aliases
- Modifying help text or descriptions
- Changing command behavior

## Commit Message Rules

- **Never mention AI tools** in commit messages - no references to Cursor, Claude, Copilot, ChatGPT, or any AI assistant
- Commits should read as if written by a human developer
- Follow conventional commit format as specified in `.cursor/rules/commit-messages.mdc`
- **Post-commit cleanup**: After every commit, immediately amend to remove any auto-added Co-authored-by lines:
  ```bash
  git log -1 --format='%B' | grep -v '^Co-authored-by:' | git commit --amend -F -
  ```

## Core Principles
- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Hackiness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what’s necessary. Avoid introducing bugs.
