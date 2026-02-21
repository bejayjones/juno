# Juno — Implementation Roadmap

Phases are sequential unless noted. Each phase ends with a compilable, testable state.

---

## ✅ Phase 1 — Project Foundation
**Status: Complete**

- Go module (`github.com/bejayjones/juno`)
- Directory structure following DDD layout
- `pkg/config` — env-var config with local/cloud mode validation
- `pkg/id` — UUID generation (`google/uuid`)
- `pkg/clock` — `Clock` interface + `Real()` + `Fixed()` for deterministic tests
- `api/rest` — chi router, structured request logger, `respond`/`respondError` helpers
- `GET /health` endpoint
- `Makefile` with `run`, `build`, `test`, `lint`, `clean` targets

---

## ✅ Phase 2 — Domain Models
**Status: Complete**
**Depends on: Phase 1**

Pure Go domain types. Zero external dependencies. No database, no HTTP.

- `internal/identity/domain` — `Inspector`, `Company`, `Client` aggregates; value objects (`Name`, `Address`, `LicenseNumber`); `InspectorRepository`, `CompanyRepository`, `ClientRepository` interfaces
- `internal/scheduling/domain` — `Appointment` aggregate; `PropertyAddress` value object; `AppointmentRepository` interface; domain events (`AppointmentScheduled`, `AppointmentCancelled`)
- `internal/inspection/domain` — `Inspection` aggregate; `SystemSection` entity; `InspectionItem` entity; `Finding` entity; `PhotoRef` value object; `SystemType` and `ItemStatus` enums; compiled-in item catalog (`item_catalog.go`) covering all 10 InterNACHI systems and their `ItemKey` values; `InspectionRepository` interface; domain events (`InspectionStarted`, `ItemStatusChanged`, `DeficiencyAdded`, `InspectionCompleted`)
- `internal/reporting/domain` — `Report` aggregate; `Delivery` entity; `ReportRepository` interface; domain events (`ReportGenerated`, `ReportDelivered`, `DeliveryFailed`)

---

## ✅ Phase 3 — Database Infrastructure
**Status: Complete**
**Depends on: Phase 2**

- Add `modernc.org/sqlite` (CGo-free SQLite driver)
- `internal/platform/db` — `DB` wrapper, `BeginTx`, `WithTx` helper for transactional handlers
- Migration runner (`cmd/migrate` or embedded on server start)
- SQLite migrations for all 4 contexts (see `SPEC.md §11` for full schema)
- `migrations/sqlite/` — numbered `.sql` files

---

## ⬜ Phase 4 — Identity Context
**Status: Not started**
**Depends on: Phase 3**

- `internal/identity/infrastructure/sqlite` — SQLite implementations of all three repository interfaces
- `internal/identity/application/commands` — `RegisterInspector`, `UpdateInspector`, `CreateCompany`, `CreateClient`
- `internal/identity/application/queries` — `GetInspector`, `GetCompany`, `ListClients`
- `internal/identity/infrastructure/auth` — JWT issue + verify using `golang-jwt/jwt`
- REST handlers wired into `api/rest/routes.go`:
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `GET  /api/v1/me`
  - `PUT  /api/v1/me`
  - `POST /api/v1/companies`
  - `GET/PUT /api/v1/companies/:id`
  - `POST /api/v1/clients`
  - `GET  /api/v1/clients`
  - `GET/PUT /api/v1/clients/:id`
- Auth middleware added to all `/api/v1` routes except `/auth/*`

---

## ⬜ Phase 5 — Scheduling Context
**Status: Not started**
**Depends on: Phase 4**

- `internal/scheduling/infrastructure/sqlite` — SQLite `AppointmentRepository`
- `internal/scheduling/application/commands` — `ScheduleAppointment`, `UpdateAppointment`, `CancelAppointment`
- `internal/scheduling/application/queries` — `GetAppointment`, `ListAppointments` (with date-range filter)
- REST handlers:
  - `GET    /api/v1/appointments`
  - `POST   /api/v1/appointments`
  - `GET    /api/v1/appointments/:id`
  - `PUT    /api/v1/appointments/:id`
  - `DELETE /api/v1/appointments/:id`

---

## ⬜ Phase 6 — Inspection Context
**Status: Not started**
**Depends on: Phase 5**

Core domain logic of the application.

- `internal/inspection/infrastructure/sqlite` — SQLite `InspectionRepository`
- `internal/inspection/application/commands`:
  - `StartInspection` (creates Inspection from Appointment, initializes all 10 SystemSections with the compiled-in item catalog)
  - `SetItemStatus` (I/NI/NP/D; requires NI reason)
  - `SetSystemDescriptions` (required "shall describe" fields per system)
  - `AddFinding` (narrative + isDeficiency flag)
  - `UpdateFinding`
  - `DeleteFinding`
  - `CompleteInspection` (validates all items have status + all required descriptions filled)
- `internal/inspection/application/queries`:
  - `GetInspection`, `ListInspections`, `GetSystemSection`, `GetDeficiencySummary`
- REST handlers for all walkthrough endpoints (see `SPEC.md §6.2`)

---

## ⬜ Phase 7 — Photo Storage
**Status: Not started**
**Depends on: Phase 6**

- `pkg/storage` — `PhotoStorage` interface (`Save`, `Get`, `Delete`)
- `pkg/storage/local` — `LocalDiskStorage` (stores under `STORAGE_LOCAL_PATH`)
- `pkg/storage/s3` — `S3Storage` stub (returns `ErrNotImplemented` until Phase 9)
- Photo upload/delete endpoints attached to findings:
  - `POST   /api/v1/inspections/:id/systems/:systemType/items/:itemKey/photos`
  - `DELETE /api/v1/inspections/:id/systems/:systemType/items/:itemKey/photos/:photoID`
  - `GET    /api/v1/photos/:photoID` (streams file from storage)
- MIME validation (JPEG, PNG, HEIC); 20 MB limit

---

## ⬜ Phase 8 — Reporting Context
**Status: Not started**
**Depends on: Phase 7**

- `internal/reporting/infrastructure/sqlite` — SQLite `ReportRepository`
- `internal/reporting/infrastructure/pdf` — PDF generator using `github.com/go-pdf/fpdf`; InterNACHI report layout (cover page, TOC, deficiency summary, 10 system sections with embedded photos, limitations block)
- `internal/reporting/infrastructure/email` — SMTP sender + `queue_only` no-op (deliveries persisted as `Pending` for later send)
- `internal/reporting/application/commands` — `GenerateReport`, `FinalizeReport`, `QueueDelivery`, `RetryFailedDeliveries`
- `internal/reporting/application/queries` — `GetReport`, `ListReports`, `GetDeliveries`
- REST handlers:
  - `POST /api/v1/reports`
  - `GET  /api/v1/reports/:id`
  - `GET  /api/v1/reports/:id/pdf`
  - `PUT  /api/v1/reports/:id/finalize`
  - `POST /api/v1/reports/:id/deliver`
  - `GET  /api/v1/reports/:id/deliveries`

---

## ⬜ Phase 9 — Sync Context
**Status: Not started**
**Depends on: Phase 8**

- `sync_records` table: tracks every mutation (table, record_id, operation, payload, lamport_clock, synced)
- Mutation hooks added to all SQLite repository `Save`/`Delete` calls
- `internal/sync/application` — `SyncService`: background goroutine, connectivity check, push local unsynced records, pull remote mutations
- Conflict resolution: last-writer-wins by Lamport clock; `findings` are append-only
- `S3Storage` fully implemented for photo sync
- Sync REST endpoints:
  - `POST /api/v1/sync/push`
  - `POST /api/v1/sync/pull`
  - `GET  /api/v1/sync/status`

---

## ⬜ Phase 10 — Frontend Foundation
**Status: Not started**
**Depends on: Phase 4** (auth endpoints must exist)

- SvelteKit project in `/web`
- Tailwind CSS
- PWA: `manifest.json`, service worker (cache app shell + API GET responses)
- Auth: login page, JWT storage in `localStorage`, `Authorization` header on all API calls, token refresh
- App shell: bottom tab nav (Dashboard / Appointments / Reports / Settings), dark mode support
- `go:embed web/build` wired into the binary; `api/rest` serves frontend assets at `/`

---

## ⬜ Phase 11 — Frontend: Scheduling
**Status: Not started**
**Depends on: Phase 10**

- Calendar view (month/week/day toggle)
- Appointment card list
- Create/edit appointment form (client picker, address, date/time, notes)
- Appointment detail page with "Start Inspection" action
- IndexedDB cache of appointments (Dexie.js) for offline viewing

---

## ⬜ Phase 12 — Frontend: Inspection Walkthrough
**Status: Not started**
**Depends on: Phase 11**

- 10-system tab bar with per-system progress indicator (X/Y items addressed)
- Per-item row: large status selector (I / NI / NP / D), tap to expand detail
- NI status → inline reason text input
- D status → deficiency drawer (narrative textarea + photo capture button)
- Photo capture via `<input type="file" accept="image/*" capture="environment">`
- System description fields (required "shall describe" inputs per system)
- Optimistic UI: status/findings written to IndexedDB immediately; API call fires in background; queued and retried if offline
- "Complete Inspection" button with pre-flight validation modal listing any missing required fields
- Dark mode support (critical for attic/basement work)

---

## ⬜ Phase 13 — Frontend: Reporting
**Status: Not started**
**Depends on: Phase 12**

- Post-completion report review screen: deficiency count by system, full deficiency list
- Tap any deficiency to navigate back to the item and edit
- "Generate Report" action → polls until PDF ready
- PDF preview in browser (iframe or PDF.js)
- "Finalize" confirmation (locks the inspection)
- Delivery form: recipient email(s), custom message body, "Send Now" / "Queue for Send"
- Delivery history list with status badges (Pending / Sent / Failed) and retry action
