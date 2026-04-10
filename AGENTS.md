# AGENTS.md

## Overview
- This repository is a Go-based POS REST API built with Gin, MongoDB, and Redis.
- Entry point: `main.go`.
- Main HTTP bootstrap and route wiring live in `app/init.go`.
- API base path is `/api/pos/v1`.

## Project Layout
- `app/core` — shared constants, error helpers, PDF utilities, common helpers.
- `app/data/entities` — persistence models.
- `app/data/repositories` — Mongo/Redis repository logic and many unit tests.
- `app/domain` — repository initialization, constants, request DTOs.
- `app/featues` — feature HTTP handlers grouped by domain. Keep the existing directory/package naming unless explicitly requested; it intentionally uses `featues` and `catagory` in current imports.
- `db` — database/resource initialization.
- `middlewares` — auth, recovery, CORS, no-route middleware.
- `docs/business-logic` — product/business documentation.

## Common Commands
- Run app: `go run main.go`
- Run all tests: `go test ./...`
- Run a package tests: `go test ./app/data/repositories`
- Format code: `gofmt -w <file>`

## Environment Notes
- Local development expects `.env` in repo root.
- Important vars include `PORT`, `MONGO_HOST`, `MONGO_POS_DB_NAME`, `REDIS_HOST`, `SECRET_KEY`.
- On Cloud Run, `.env` is not loaded and release-friendly defaults are applied.
- Default branch auto-init is controlled by `AUTO_INIT_DEFAULT_BRANCH`.

## Working Conventions
- Prefer small, targeted changes that preserve current package names and imports.
- Follow existing layering: handlers in `app/featues`, DTOs in `app/domain/request`, persistence in `app/data/repositories` and `app/data/entities`.
- Add or update tests near repository and middleware changes when practical.
- Avoid broad refactors unless explicitly requested, especially around route/package naming typos that are already wired through the codebase.
- When changing API behavior, check related business docs under `docs/business-logic` and keep docs aligned if needed.

## Validation Guidance
- Start with focused package tests for the files you changed.
- Run `go test ./...` only after targeted checks pass or when making broader changes.
- If you touch formatting-sensitive Go files, run `gofmt -w` on the edited files.

## Cautions
- Database/resource initialization happens in `db/init.go`; avoid introducing startup behavior that requires external services in unit tests.
- Some tests are pure unit tests around repository helpers and validation; preserve that style unless integration coverage is explicitly needed.
