# waf-admin (OSS)

Admin API to manage **Caddy + Coraza** WAF sites & rules.
Validates via `caddy validate` and hot-reloads via Caddy Admin API (UNIX socket).
Includes optional daily S3 backups.

## Quick start
```bash
make run
```

## API
See `internal/api/openapi.yaml`.
