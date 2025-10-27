# waf-admin AI Agent Guide
## Architecture
- `cmd/waf-admin/main.go` loads YAML config (`-config` flag), wires filesystem storage, the Caddy/Coraza render driver, the Caddy Admin reloader, and daily backup scheduler before launching the HTTP API.
- `internal/api` hosts the chi router (`router.go`), request middleware, and handlers that read/write Caddy site snippets and Coraza rule files via the storage interface; protected routes sit under `/v1/*` and require the bearer token from config.
- `internal/render/caddy_coraza.go` implements `render.Driver` by calling `caddy validate --config <Caddyfile>`; ensure the binary is available or mock the command when testing.
- `internal/reload/caddy_admin.go` posts the rendered Caddyfile to the Caddy Admin UNIX socket using a custom byte reader and returns a typed error when reload fails (HTTP status !2xx).
- `internal/storage/fs.go` provides the default `storage.Storage` backed by the host filesystem with `util.AtomicWrite` to avoid partial writes; any new storage implementation must respect this contract.

## Workflows
- Use `go build ./...` for verification and `make run` to launch against `configs/config.example.yaml`; ensure the example paths exist or override via flags.
- API schema lives in `internal/api/openapi.yaml`; update it whenever endpoints change so clients and docs remain accurate.
- Backups (`scheduler.RunBackup`) rely on the AWS CLI and S3-compatible credentials in config; local runs without those tools should disable `backup.enabled`.

## Patterns & Conventions
- Site resources map to files ending with `.caddy` in `CaddyOptions.SitesDir`; rule files must match `^[a-zA-Z0-9._-]+\.conf(?:\.disabled)?$` inside `<RulesRoot>/<site>/rules` as enforced in `handlers.go`.
- Mutations call `applyNow`, which validates via the render driver before invoking the reloader; preserve this ordering when adding new write paths.
- `domain.ListSites` is the single source for aggregating site metadata; prefer extending it over re-listing directories elsewhere.
- Auth is a simple bearer token (`Authorization: Bearer <token>`); remember to keep health and metrics endpoints public when adjusting middleware.

## Integration Notes
- Logging uses zerolog configured in `internal/util/log.go` to emit human-readable console output; stick with zerolog for consistency.
- Rate limiting is applied globally via `httprate.LimitByIP(100, 1*time.Minute)`; consider adjustments here when exposing new long-running routes.
- Scheduler jobs run in background goroutines; long operations should honor context cancellation and avoid panics to protect the job loop.

## When Extending
- Prefer injecting new dependencies through `api.Config` and the main wiring in `cmd/waf-admin/main.go` to keep boot sequence explicit.
- For alternative renderers/reloaders/storage backends, implement the respective interfaces and configure them at composition time.
- Include sample snippets/configs under `examples/` if you introduce new runtime expectations to aid users and tests.
