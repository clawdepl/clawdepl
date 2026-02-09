# clawdepl Test Plan - Provisioning Endpoints

## Overview
Test plan for validating the new Convex HTTP endpoints that allow the CLI to provision and manage instances through proper auth validation.

## Changes Being Tested
- **Backend**: New HTTP endpoints in Convex (`/api/provision/*`)
- **CLI**: Updated to call Convex instead of provisioner directly
- **Architecture**: CLI → Convex (validates JWT) → Provisioner (service token)

## Prerequisites

### 1. Backend Deployed
- [ ] Backend changes deployed to Convex (commit `6b69090`)
- [ ] Verify endpoints exist: `curl -I https://colorless-gull-839.convex.site/api/provision`
- [ ] Check Convex dashboard for deployment status

### 2. CLI Built
- [ ] Latest CLI built from commit `5f66aed`
- [ ] Run: `go build -o clawdepl`
- [ ] Verify version: `./clawdepl --version`

### 3. Authentication
- [ ] User logged in: `./clawdepl login`
- [ ] Verify credentials: `./clawdepl list` should work
- [ ] Have a valid Claude API key ready for testing

## Test Cases

### TC1: Provision New Instance (Happy Path)

**Steps:**
```bash
./clawdepl new test-instance
```

**Interactive prompts:**
1. Enter Claude API key: `sk-ant-api03-...`
2. Enter purpose/vibe: `A test instance for validation`

**Expected Result:**
- ✅ TUI wizard starts without panic
- ✅ No "index out of range" error when name is provided
- ✅ Provisioning starts successfully
- ✅ Returns sandboxId and status "provisioning"
- ✅ Success message displayed

**Error Cases to Verify:**
- ❌ Not logged in → "Not logged in. Run 'clawdepl login' first."
- ❌ Invalid API key → Clear error from provisioner
- ❌ Reached instance limit → "You've reached your limit" message

### TC2: Provision Without Name Argument

**Steps:**
```bash
./clawdepl new
```

**Expected Result:**
- ✅ TUI wizard prompts for name first
- ✅ Then prompts for API key
- ✅ Then prompts for vibe
- ✅ Completes successfully

### TC3: List Instances

**Steps:**
```bash
./clawdepl list
```

**Expected Result:**
- ✅ Shows newly created instance
- ✅ Displays: name, status, sandboxId
- ✅ Status shows "provisioning" or "running"

### TC4: Check Instance Status

**Steps:**
```bash
./clawdepl status <instance-name>
```

**Expected Result:**
- ✅ Shows detailed status
- ✅ Displays provisioning progress if still provisioning
- ✅ Shows "running" when complete
- ✅ Includes gatewayUrl when ready

**Error Cases:**
- ❌ Non-existent instance → "Instance not found"
- ❌ Other user's instance → Access denied

### TC5: Stop Instance

**Steps:**
```bash
./clawdepl stop <instance-name>
```

**Expected Result:**
- ✅ Instance stops successfully
- ✅ Status changes to "stopped"
- ✅ Confirmation message displayed

**Error Cases:**
- ❌ Already stopped → Appropriate message
- ❌ Not owner → Access denied

### TC6: Start Instance

**Steps:**
```bash
./clawdepl start <instance-name>
```

**Expected Result:**
- ✅ Instance starts successfully
- ✅ Status changes to "running"
- ✅ Confirmation message displayed

**Error Cases:**
- ❌ Already running → Appropriate message
- ❌ Not owner → Access denied

### TC7: Delete Instance

**Steps:**
```bash
./clawdepl delete <instance-name>
```

**Expected Result:**
- ✅ Instance deleted successfully
- ✅ No longer appears in `clawdepl list`
- ✅ Confirmation message displayed

**Error Cases:**
- ❌ Not owner → Access denied
- ❌ Non-existent instance → "Instance not found"

## Security Validation

### SV1: Authentication Required
**Test:** Try operations without login
```bash
./clawdepl logout
./clawdepl new test
```
**Expected:** "Not logged in" error

### SV2: Authorization Enforced
**Test:** Try to access another user's instance
- Get sandboxId of instance you don't own
- Try: `./clawdepl stop <other-sandbox-id>`
**Expected:** "Access denied" or "Molty not found"

### SV3: JWT Validation
**Test:** Use expired or invalid token
```bash
./clawdepl --unsafe-token "invalid-token" list
```
**Expected:** 401 Unauthorized error

## Performance Tests

### PT1: Provisioning Time
**Test:** Time from `new` command to "running" status
**Expected:** < 5 minutes for full provisioning
**Measure:**
- Time to get sandboxId: < 5 seconds
- Time to reach "running": < 5 minutes

### PT2: API Response Time
**Test:** Measure response time for operations
```bash
time ./clawdepl list
time ./clawdepl status <instance>
```
**Expected:** < 2 seconds for list/status operations

## Edge Cases

### EC1: Network Errors
**Test:** Simulate network issues
- Disconnect network during provisioning
- Check if error is handled gracefully

### EC2: Long Instance Names
**Test:** Create instance with 64-char name
**Expected:** Should work (char limit is 64)

### EC3: Special Characters in Vibe
**Test:** Use emojis/special chars in vibe field
**Expected:** Should be accepted and stored correctly

### EC4: Concurrent Operations
**Test:** Run multiple operations simultaneously
```bash
./clawdepl new instance1 &
./clawdepl new instance2 &
```
**Expected:** Both should succeed (different sandboxIds)

## Regression Tests

### RT1: List Command Still Works
**Test:** Ensure existing functionality unchanged
```bash
./clawdepl list
```
**Expected:** Same behavior as before

### RT2: Login/Logout Unchanged
**Test:** Auth flow still works
```bash
./clawdepl logout
./clawdepl login
```
**Expected:** OAuth flow works as before

## Backend Verification

### BV1: Convex Logs
**Check:** Convex dashboard for request logs
- Verify requests hitting `/api/provision` endpoints
- Check for errors in function execution
- Verify auth validation logs

### BV2: Provisioner Calls
**Check:** Verify Convex is calling provisioner correctly
- Look for provisioner logs showing requests from Convex
- Verify service token is being used
- Check for any 401 errors from provisioner

### BV3: Database State
**Check:** Convex database
- Verify moltys are created with correct ownerId
- Check subscription status is being validated
- Verify instance limits are enforced

## Rollback Plan

If critical issues found:

1. **Immediate:** Revert backend deployment
   ```bash
   # Burry can revert in Convex dashboard
   ```

2. **CLI users:** Can continue using old version
   - Old CLI won't break (will just get 404s on new endpoints)
   - No data corruption risk

3. **Investigation:** Check logs for specific errors
   - Convex function logs
   - Provisioner logs
   - CLI error output

## Success Criteria

- [ ] All happy path tests pass (TC1-TC7)
- [ ] Security validation passes (SV1-SV3)
- [ ] No performance regressions (PT1-PT2)
- [ ] Edge cases handled gracefully (EC1-EC4)
- [ ] No errors in Convex logs for 1 hour post-deployment
- [ ] At least 3 successful provisions with different users

## Test Results Template

```markdown
## Test Execution Results - [Date]

### Environment
- Backend Version: [commit hash]
- CLI Version: [commit hash]
- Tester: [name]

### Results
| Test Case | Status | Notes |
|-----------|--------|-------|
| TC1 | ✅/❌ | |
| TC2 | ✅/❌ | |
| TC3 | ✅/❌ | |
| ... | | |

### Issues Found
1. [Issue description]
   - Severity: High/Medium/Low
   - Steps to reproduce:
   - Expected vs Actual:

### Recommendations
- [Action items]
```

## Notes

- Run tests in **staging environment first** if available
- Test with both **free and pro users** to verify limits
- Keep Convex dashboard open to monitor real-time logs
- Have rollback access ready before starting tests
- Document any unexpected behavior immediately
