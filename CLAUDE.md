# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**Juno** — Mobile-first SaaS for licensed home inspectors. Inspectors schedule inspections, conduct InterNACHI-compliant walkthroughs on phone/tablet, generate PDF reports, and deliver them to clients. The app must work fully offline (local SQLite + disk storage) with optional cloud sync to PostgreSQL/S3.

Full specification: `SPEC.md`

## Tech Stack

- **Backend**: Go (single binary, serves API + embedded frontend)
- **Frontend**: SvelteKit + Tailwind CSS, compiled and embedded via `go:embed`
- **Local DB**: SQLite via `modernc.org/sqlite` (CGo-free)
- **Cloud DB**: PostgreSQL
- **PDF**: `github.com/go-pdf/fpdf` or `github.com/signintech/gopdf`
- **Auth**: JWT bearer tokens

## Commands

```bash
make run                        # go run ./cmd/server (local SQLite, port 8080)
make build                      # produces bin/juno
make test                       # go test ./...
make test-v                     # go test -v ./...
make test-pkg PKG=./internal/inspection/...   # single package
make lint                       # golangci-lint run
make clean                      # remove bin/, data/, *.db files
```

## Architecture

The app uses **Domain-Driven Design** organized into four bounded contexts, each with `domain/`, `application/`, and `infrastructure/` layers:

| Context | Package | Responsibility |
|---|---|---|
| `scheduling` | `internal/scheduling` | Appointments, calendar |
| `inspection` | `internal/inspection` | Walkthrough session, items, findings, photos |
| `reporting` | `internal/reporting` | Report assembly, PDF generation, email delivery |
| `identity` | `internal/identity` | Inspectors, companies, clients |
| `sync` | `internal/sync` | Offline change log, cloud sync |

**Critical rule**: The `domain/` layer has zero external dependencies — no database drivers, no HTTP, no infrastructure packages. Application and infrastructure layers import domain; domain imports nothing from infrastructure.

## Repository Interface Pattern

Every bounded context defines its persistence contract in `domain/repository.go`. Infrastructure implementations live in `infrastructure/sqlite/` and `infrastructure/postgres/`. Application layer code only imports the domain interface — never the implementation directly.

All dependencies are wired in `cmd/server/main.go`.

## InterNACHI Domain Language

Ten mandatory inspection systems (3.1–3.10): Roof, Exterior, Foundation/Basement/Crawl Space, Heating, Cooling, Plumbing, Electrical, Fireplace, Attic/Insulation/Ventilation, Doors/Windows/Interior.

Item statuses: `Inspected (I)` | `NotInspected (NI)` | `NotPresent (NP)` | `Deficient (D)`

SOP obligation types map to distinct data fields:
- `shall inspect` → `InspectionItem.Status`
- `shall describe` → `SystemSection.Descriptions` (map keyed by `DescriptionKey`)
- `shall report as in need of correction` → `Finding.IsDeficiency = true` + `Finding.Narrative`

The item catalog (`internal/inspection/domain/item.go`) is compiled-in (not DB-stored). See `SPEC.md §7` for the full list of `ItemKey` values per system.

## Offline-First Rules

- All write operations hit local SQLite first; the UI must never block on network
- Every mutation is recorded in `sync_records` with a Lamport clock for eventual sync
- `sync` bounded context handles push/pull; conflict resolution is last-writer-wins except `findings` (append-only)
- Photos are stored on local disk first (`pkg/storage.LocalDiskStorage`); marked `uploaded=false` until synced
- Report PDF generation must work entirely from local data — no network calls during generation

## Key Value Objects and Enums

```go
// SystemType — 10 values matching InterNACHI SOP sections
type SystemType string
const (
    SystemRoof       SystemType = "roof"
    SystemExterior   SystemType = "exterior"
    SystemFoundation SystemType = "foundation"
    SystemHeating    SystemType = "heating"
    SystemCooling    SystemType = "cooling"
    SystemPlumbing   SystemType = "plumbing"
    SystemElectrical SystemType = "electrical"
    SystemFireplace  SystemType = "fireplace"
    SystemAttic      SystemType = "attic"
    SystemInterior   SystemType = "interior"
)

// ItemStatus
type ItemStatus string
const (
    StatusInspected    ItemStatus = "I"
    StatusNotInspected ItemStatus = "NI"
    StatusNotPresent   ItemStatus = "NP"
    StatusDeficient    ItemStatus = "D"
)
```

## Configuration

All config is via environment variables. Defaults are suitable for local dev with no setup.

```
SERVER_MODE=local|cloud          default: local
SERVER_PORT=8080                 default: 8080
DATABASE_DRIVER=sqlite|postgres  default: sqlite
DATABASE_DSN=./juno.db           default: ./juno.db
STORAGE_DRIVER=local|s3          default: local
STORAGE_LOCAL_PATH=./data/photos default: ./data/photos
JWT_SECRET=...                   required in cloud mode; defaults to dev value in local mode
EMAIL_DRIVER=smtp|queue_only     default: queue_only
```

## What Exists Now

**Phase 1 is complete.** The following is implemented and compiles:

```
cmd/server/main.go                          entrypoint; loads config, starts HTTP server
api/rest/server.go                          Server struct, chi router, middleware wiring
api/rest/routes.go                          GET /health registered; /api/v1 stub ready
api/rest/health.go                          GET /health → {status, timestamp, mode}
api/rest/respond.go                         respond() / respondError() helpers
api/rest/middleware/logging.go              structured request logger
pkg/config/config.go                        env-var config with validation
pkg/id/id.go                                UUID generation + validation (google/uuid)
pkg/clock/clock.go                          Clock interface + Real() + Fixed() for tests
internal/inspection/domain/                 Inspection aggregate, 10-system catalog, findings
internal/scheduling/domain/                 Appointment aggregate
internal/identity/domain/                   Inspector, Company, Client aggregates
internal/reporting/domain/                  Report aggregate, Delivery entity
internal/platform/db/                       DB wrapper, WithTx, embedded migration runner
internal/platform/db/migrations/sqlite/     001_initial_schema.sql (all tables + indexes)
Makefile
.gitignore
```

**Next up: Phase 4 — Identity Context.** See `ROADMAP.md` for full phase status.

## Testing Approach

- Domain logic: pure unit tests, no DB, no HTTP
- Application handlers: in-memory repository fakes
- Repository implementations: real SQLite in-process; testcontainers for Postgres
- HTTP handlers: `net/http/httptest`
- Inject `pkg/clock.Fixed()` wherever time is needed in tests
