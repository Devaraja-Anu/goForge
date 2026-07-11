# GoForge — Decisions Log

Format: each entry is a settled decision. Don't silently override — flag and discuss if revisiting.

---

## Core stack (V1)

Chi, slog, PostgreSQL, pgx, sqlc, golang-migrate, Docker, GitHub Actions, go-playground/validator.
No alternatives supported in V1. See Product Philosophy doc — "opinionated but not configurable."

## Blueprint project lives inside the GoForge repo, as a separate Go module

- Directory: `blueprint/` (nested `go.mod`, own module path, e.g. `github.com/<org>/blueprint`)
- Rejected: pointing generation at an external repo (network dependency at generate-time,
  version coupling to another repo's commits).
- Rejected: `template/` as the directory name — collides conceptually with Go's `text/template`,
  and we are explicitly NOT using a template language.
- Blueprint builds, vets, and tests independently of GoForge's own module.
- GoForge embeds it via `go:embed all:blueprint` at build time.

## Generation mechanism: copy + token substitution, not a template language

- V1 has no conditional/optional file inclusion (no feature flags), so no templating engine
  is needed — Blueprint's Go source _is_ the generated source.
- `goforge new <name>` flow:
  1. Copy embedded blueprint files into target dir
  2. `go mod init <module-path>`
  3. Rewrite internal import paths from blueprint's module path to the new one (exact-match
     string replace on the module prefix)
  4. Copy `go.mod`/`go.sum` version pins verbatim (module line updated) — do NOT run
     `go mod tidy` at generation time, to guarantee reproducible, tested dependency versions
     per GoForge release.

## Module path input

- OPEN — see "Open questions" below.

## Validator (go-playground/validator) is part of the V1 stack

- Justification: request validation is infra-adjacent boilerplate every API needs; the
  integration work (JSON-tag error names, tag→message mapping) is non-trivial and reusable.
- Blueprint ships the validator wiring (tag name func, error mapping, `app.validate` on
  `application`) but no validated structs — those are the user's domain code.

## No auth in the blueprint

- Auth strategy (JWT vs sessions vs OAuth) is a product decision, not infrastructure.
- Deferred to a future `goforge add auth` command, to be designed separately — not assumed
  to be as simple as `add redis`/`add otel`.

## No Redis, no OTel in V1

- Both are real future features (`goforge add redis`, `goforge add otel`) per the roadmap,
  but per "generate only what exists today," no placeholder config/env vars/commented code
  for them ships in V1.

## Formatting standard

- `gofmt`/`goimports` clean, enforced in blueprint CI. No manual spacing deviations.

## Testing / CI enforcement

- Blueprint has its own test suite and its own CI job (build + vet + test), independent of
  GoForge's CI job.
- GoForge CI additionally does a golden-file generation smoke test: run `goforge new` into a
  temp dir, then `go build ./...`, `go vet ./...`, `go test ./...`, `docker build` against
  the _generated_ output.

## CLI: Cobra

- Chosen over stdlib `flag` despite only one command in V1, because `add redis`/`add otel`/
  `add auth` subcommands are already on the roadmap and Cobra's subcommand ergonomics pay for
  themselves. Reconsidered from an earlier stdlib-only recommendation.

## No README generation

- `goforge new` does not generate a README. Left entirely to the project author.

## `/v1/healthcheck` URL versioning

- OPEN — see below.

---

## Open questions (not yet decided)

- How is the target module path supplied to `goforge new`? (required arg/flag vs. inferred
  from git config vs. GoForge-level config file) — leaning required flag, not yet confirmed.
- Is `/v1/...` prefix on routes a blueprint opinion we commit to, or should health check be
  unprefixed (`/healthcheck`)?
- IP strategy config (remote_addr / xff_trusted / cloudflare / custom_header) — is 4-way
  configurability appropriate for V1, or should this collapse to one default behavior?
- Lint (golangci-lint) in blueprint CI — in scope for V1 or later?

## Module path input: required flag

- `goforge new <name> --module <path>` — no interactive prompt, no inference from git config.
  Keeps generation scriptable and explicit ("no magic").

## golangci-lint is in scope for V1

- Runs in blueprint CI alongside build/vet/test.

## IP strategy: collapsed to two modes

- `remote_addr` (default) and `trust_proxy_headers` (on/off).
- Rejected: 4-way strategy enum with per-vendor headers and CIDR trust lists — over-built for
  the current use (the value is only ever logged; no rate limiting or IP-based logic exists
  yet in V1). Revisit if/when a feature that actually depends on precise client IP (e.g.
  rate limiting) is added.
- `remote_addr` is the safe default — avoids trusting a spoofable header when no proxy is
  actually in front of the service.
