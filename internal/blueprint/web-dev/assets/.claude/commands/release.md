# /release - Release a Feature

Prepare and execute a feature release.

## Usage

```
/release [feature-id]
```

## What This Command Does

1. **Verifies Gates**: Ensures all quality gates pass
2. **Creates Release**: Prepares release artifacts
3. **Deployment Steps**: Provides deployment checklist
4. **Updates State**: Marks feature as released

## Prerequisites

Before using this command:
- Feature must be in "approved" status
- All tests must pass
- Code review must be complete

## Process

### Step 1: Verify Release Gates

{{if .GatesEnabled}}
**Gates are ENABLED for this workflow.**

Required gates:
- [ ] All tests passing
- [ ] Build succeeds
- [ ] Code review approved
- [ ] No blocking security issues
- [ ] Documentation updated
{{else}}
**Gates are DISABLED for this workflow.**
Proceeding without gate verification.
{{end}}

### Step 2: Pre-Release Checks

```bash
# Verify clean state
git status

# Run full test suite
npm test

# Build for production
npm run build

# Check for vulnerabilities
npm audit --production
```

### Step 3: Release Checklist

#### Before Deployment
- [ ] All automated checks pass
- [ ] Manual testing complete
- [ ] Rollback plan documented
- [ ] Monitoring dashboards ready
- [ ] On-call notified (if needed)

#### Deployment Steps
1. Merge PR to main branch
2. Tag release: `git tag v<version>`
3. Deploy to staging environment
4. Verify staging deployment
5. Deploy to production
6. Verify production deployment

#### After Deployment
- [ ] Smoke tests pass
- [ ] Monitoring shows healthy metrics
- [ ] No error spikes in logs
- [ ] Feature functions as expected

### Step 4: Update State
```json
{
  "status": "released",
  "release_completed_at": "<ISO timestamp>",
  "version": "<version>",
  "deployed_environments": ["staging", "production"],
  "release_notes": "<summary>"
}
```

### Step 5: Output Summary
- Release status
- Version number
- Deployed environments
- Post-release monitoring tips
- Rollback instructions (if needed)

## Rollback Procedure

If issues are discovered:
1. Revert the merge commit: `git revert <commit>`
2. Push revert to main
3. Redeploy previous version
4. Update state to "rollback"
5. Investigate and fix issues
6. Re-release when ready

## Guidelines

- Never skip quality gates for "urgent" releases
- Always have a rollback plan
- Monitor closely after release
- Document any incidents
- Celebrate successful releases! 🎉
