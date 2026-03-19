# Incident Runbook

## Common Issues and Resolution

### Collection Failures

**Symptom**: `collect.yml` fails to fetch from a source

**Resolution**:
1. Check source URL availability
2. Verify rate limiting hasn't been triggered
3. Check for format changes in source
4. Run with `--verbose` for detailed logs

### Openclaw Failures

**Symptom**: `digestor` fails with JSON parse error

**Resolution**:
1. Check Openclaw service status
2. Verify `OPENCLAW_TOKEN` is valid
3. Retry with `--force` flag
4. Maximum 2 automatic retries

### Deploy Failures

**Symptom**: `deploy.yml` fails at Pages deployment

**Resolution**:
1. Check Pages settings in repository
2. Verify `pages: write` permission
3. Check artifact size (limit: 10GB)
4. Manual retry via `workflow_dispatch`

## Escalation

For persistent issues:
1. Check GitHub Actions logs
2. Review state files in `data/state/`
3. Verify GitHub Pages configuration
4. Contact system maintainer
