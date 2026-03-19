# Operations Guide

## Daily Operations

### Morning Check (Optional)
1. Verify last night's deploy succeeded
2. Check `/status` page for any failures

### Evening Check
1. Verify 21:35 digest generation completed
2. Confirm Pages reflects latest digest

## Manual Triggers

### Collect Single Batch
```bash
gh workflow run collect.yml -f date=2026-03-19 -f batch=09
```

### Regenerate Digest
```bash
gh workflow run digest.yml -f date=2026-03-19 --force
```

### Full Redeploy
```bash
gh workflow run deploy.yml -f date=2026-03-19
```

## Maintenance

### Weekly Tasks
1. Review source performance
2. Check disk usage in `data/`
3. Verify indexes are current

### Monthly Tasks
1. Archive old data if needed
2. Review and update source list
3. Check for design document updates

## Troubleshooting

See `incident-runbook.md` for detailed issue resolution.
