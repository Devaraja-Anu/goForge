# GoForge

GoForge generates the infrastructure a Go backend service needs before real
development can start — config loading, structured logging, graceful
shutdown, PostgreSQL + sqlc wiring, Docker, and CI — so you can skip the
first day of setup and start writing your actual API.

The generated project has zero dependency on GoForge itself. It's ordinary,
idiomatic Go you're free to edit, restructure, or rip apart from line one.

## Install

```bash
go install github.com/devaraja-anu/goforge/cmd/goforge@latest
```

Requires Go to already be installed — GoForge targets developers who already
have a working Go setup, not first-time Go users.

## Usage

```bash
goforge new myapi --module github.com/you/myapi
```

`--module` is required and explicit — GoForge doesn't infer it from git
config or prompt interactively. This is deliberate: generation stays
scriptable and predictable.

That's it. The generated project is a complete, standalone Go module —
`cd myapi` and it's yours from there.

## What you get

- [Chi](https://github.com/go-chi/chi) router, with request ID, recovery,
  timeout, and CORS middleware wired in
- Structured logging via `slog`, JSON output
- Graceful shutdown on SIGINT/SIGTERM
- PostgreSQL via `pgx`, connection pool sizing via env vars
- `sqlc` wired and ready — write SQL, get typed Go
- `golang-migrate` migrations, with an empty starter migration establishing
  the file convention
- Request validation via `go-playground/validator`
- A multi-stage Dockerfile (distroless runtime, non-root) and
  `compose.yaml` for local dev (app + Postgres)
- A GitHub Actions workflow (build/vet/test) generated fresh for your new
  repo
- A `Makefile` with `build`, `run`, `test`, `migrate`, `sqlc-generate`, and
  more — run `make help` inside the generated project for the full list

## What you don't get, on purpose

GoForge generates infrastructure, not application code. It won't generate
users, products, CRUD resources, service layers, or any particular
application architecture (no imposed Clean/Hexagonal/DDD structure) — that's
entirely yours to design. See the design philosophy below for the reasoning.

No README ships inside generated projects either — the generated code and
its own `make help` output are meant to speak for themselves. This document
is about using the GoForge CLI itself, not about the projects it produces.

## Stack (V1, not configurable)

Chi, slog, PostgreSQL, pgx, sqlc, golang-migrate, Docker, GitHub Actions,
go-playground/validator. GoForge intentionally picks one stack rather than
supporting alternatives — if you need a materially different stack, GoForge
may not be the right tool.

## Contributing / local development

```bash
git clone https://github.com/devaraja-anu/goforge
cd goforge
make install-hooks   # sets up a pre-commit hook that runs gofmt/goimports
make build
```

`blueprint/` (the actual source every generated project comes from) is a
separate, independently buildable Go module living in this repo. If you
change anything under `blueprint/`, run `make sync-blueprint` afterward and
commit the result — `blueprintsrc/` is a derived mirror used for embedding,
and CI will fail if it's out of sync.

## Design philosophy

GoForge is a generator, not a framework — it should never become a runtime
dependency of anything it generates, and a GoForge-generated project should
be indistinguishable from one written by hand. If you're curious about the
reasoning behind specific choices, most of it is documented in this
repo's `Decisions.md`.