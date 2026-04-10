# AGENTS.md

## Overview
- Go-based POS REST API built with **Gin**, **MongoDB**, and **Redis**.
- Entry point: `main.go` → `app/init.go` (`Routes.StartGin`).
- API base path: `/api/pos/v1`.
- HTTP server exposes configurable timeouts via env vars (`HTTP_READ_TIMEOUT_SEC`, `HTTP_WRITE_TIMEOUT_SEC`, `HTTP_IDLE_TIMEOUT_SEC`, `HTTP_READ_HEADER_TIMEOUT_SEC`).

## Project Layout

### `app/core` — shared utilities
- `constant/role.go` — role constants (`SUPER`, `ADMIN`, `USER`).
- `errcode/` — structured error codes (`codes.go`) and abort helper (`errcode.go`).
- `pdf/` — PDF generation utilities and embedded Thai fonts.
- `utils/` — Gin context helpers, time helpers.

### `app/data/entities` — persistence models
- Key entities: `product`, `order`, `customer`, `patient`, `receive`, `branch`, `employee`, `supplier`, `category`, `promotion`, `stock_transfer`, `customer_history`, `sequence`, `setting`, `payment`.
- `Product` has `DrugInfo` (metadata: genericName, dosageForm, strength, etc.) and `DrugRegistrations []string` (values: `"KHY9"` – `"KHY13"`).
- `Order` stores compliance fields at order level: `PharmacistName`, `LicenseNo`, `PrescriberName`, `BuyerName`, `BuyerIdCard`, `PatientId`.

### `app/data/repositories` — Mongo/Redis repository logic
- One interface + implementation per entity (e.g. `IOrder` / `NewOrderEntity`).
- Extensive unit tests co-located here (`*_test.go` files for pipelines, filters, validation).

### `app/domain` — DI, constants, request DTOs
- `init.go` — `Repository` struct aggregating 15 repository interfaces; `InitRepository` wires them from `db.Resource`.
- `constant/` — customer types, payment types, product-history types, sequence constants, status constants.
- `request/` — one DTO file per feature (matches entity names); includes `flexible_time.go` for date parsing.

### `app/featues` — feature HTTP handlers (grouped by domain)
- **Keep the existing `featues` / `catagory` directory names** — the typos are wired through the whole import graph.
- Feature packages: `branch`, `catagory`, `customer`, `customer_history`, `dashboard`, `dispensing` (empty placeholder), `employee`, `order`, `patient`, `product`, `promotion`, `receive`, `report`, `setting`, `stock_transfer`, `supplier`.
- Each package has `api.go` (route registration via `Apply*API`) and `usecase/` sub-package with handler functions.

### `app/featues/report` — KHY pharmacy reports
- KHY9 (purchase report) pulls from **receives**; KHY10–KHY13 pull from **orders** joined with products.
- Reports filter products by `product.DrugRegistrations` array.
- Each report exposes both `/data` (JSON) and `/csv` endpoints.
- Also includes sales-report Excel, stock-report Excel, barcode PDF, and PromptPay QR/payload endpoints.

### `middlewares` — HTTP middleware chain
- `RequireAuthenticated()` — JWT validation (SECRET_KEY, CLIENT_ID, SYSTEM).
- `RequireSession(session)` — Redis session check.
- `RequireBranch(employee, branch)` — resolves `BranchId` and `EmployeeRole` into context; falls back to default `HQ` branch.
- `RequireAuthorization(roles...)` — role-based access control.
- `NewRecovery()`, `NewCors()`, `NoRoute()` — panic recovery, CORS, 404 handler.

### Other top-level directories
- `db/` — `init.go` initialises Mongo client and Redis client into `db.Resource`.
- `docs/business-logic/` — rich documentation: numbered markdown chapters (01–08), `api-contracts/`, `flows/`, `lifecycle/`, `screens/`, `feature-flow-matrix.md`, `requirement.md`.

## Common Commands
- Run app: `go run main.go`
- Run all tests: `go test ./...`
- Run a package's tests: `go test ./app/data/repositories`
- Format code: `gofmt -w <file>`

## Environment Notes
- Local development expects `.env` in repo root.
- **Required** env vars (validated at startup in `validateStartupConfig`): `SECRET_KEY`, `CLIENT_ID`, `SYSTEM`, `MONGO_HOST`, `MONGO_POS_DB_NAME`, `REDIS_HOST`.
- Optional: `PORT` (default `8080`), `GIN_MODE`, `AUTO_INIT_DEFAULT_BRANCH`, HTTP timeout vars.
- On Cloud Run (`K_SERVICE` / `K_REVISION` set), `.env` is not loaded, Gin runs in release mode, and default-branch auto-init is skipped unless overridden.
- Startup config validation is covered by tests in `app/init_test.go`.

## Working Conventions
- Prefer small, targeted changes that preserve current package names and imports.
- Follow existing layering: handlers in `app/featues`, DTOs in `app/domain/request`, persistence in `app/data/repositories` and `app/data/entities`.
- Add or update tests near repository and middleware changes when practical.
- Avoid broad refactors unless explicitly requested, especially around route/package naming typos that are already wired through the codebase.
- When changing API behavior, check related business docs under `docs/business-logic` (chapters, api-contracts, flows) and keep docs aligned if needed.

## Pharmacy / Drug Registration Notes
- Products may carry `DrugRegistrations` values `KHY9`–`KHY13`. These drive both report filtering and compliance dialog triggering on the client.
- `KHY10`–`KHY13` trigger compliance (pharmacist name, license, prescriber, buyer info stored at order level). `KHY9` is purchase-report only — no compliance trigger.
- `DrugInfo` stores pharmaceutical metadata (genericName, dosageForm, strength, etc.) used for display in reports.

## Validation Guidance
- Start with focused package tests for the files you changed.
- Run `go test ./...` only after targeted checks pass or when making broader changes.
- If you touch formatting-sensitive Go files, run `gofmt -w` on the edited files.

## Cautions
- Database/resource initialization happens in `db/init.go`; avoid introducing startup behavior that requires external services in unit tests.
- Some tests are pure unit tests around repository helpers and validation; preserve that style unless integration coverage is explicitly needed.
- The `dispensing` feature package exists but is currently empty — do not wire routes for it without explicit direction.
