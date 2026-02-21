# Juno — Home Inspection SaaS: Application Specification

## 1. Product Overview

**Juno** is a mobile-first SaaS platform for licensed home inspectors, compliant with the InterNACHI Standards of Practice (SOP). Inspectors use Juno to schedule inspections, conduct walkthrough checklists on phone or tablet, generate InterNACHI-compliant PDF reports, and deliver reports to clients — all without requiring an internet connection during the inspection itself.

### Core Principles

- **Offline-first**: All inspection and report generation functionality works without internet access. Data syncs to the cloud when connectivity is restored.
- **Mobile-first**: UI designed primarily for phone and tablet use in the field, with full desktop support for office workflows.
- **InterNACHI-compliant**: Report output conforms to InterNACHI Standards of Practice (current revision).
- **DDD-architected**: Domain logic is isolated from infrastructure concerns. All persistence is behind repository interfaces enabling swappable storage backends.
- **Go backend**: The entire server-side is written in Go, embedding a compiled frontend for single-binary deployments.

---

## 2. InterNACHI Compliance Requirements

### 2.1 The Ten Inspection Systems (Mandatory)

Every inspection must cover all ten systems defined by InterNACHI SOP Section 3. Each system has three inspector obligation types that map directly to UI elements:

| SOP Verb | UI Element |
|---|---|
| `shall inspect` | Status selector: `Inspected` / `Not Inspected` / `Not Present` |
| `shall describe` | Structured description field (type, method, location, energy source, etc.) |
| `shall report as in need of correction` | Deficiency flag with narrative text and optional photo |

**The Ten Systems and their required "Shall Describe" fields:**

| # | System | Required Descriptions |
|---|---|---|
| 3.1 | Roof | Roof-covering material type |
| 3.2 | Exterior | Exterior wall-covering material type |
| 3.3 | Foundation / Basement / Crawl Space | Foundation type; under-floor access location |
| 3.4 | Heating | Thermostat location; energy source; heating method |
| 3.5 | Cooling | Thermostat location; cooling method |
| 3.6 | Plumbing | Water supply type (public/private); main water shutoff location; main fuel shutoff location; fuel storage location; water heater capacity |
| 3.7 | Electrical | Main service amperage (if labeled); wiring type observed |
| 3.8 | Fireplace | Fireplace type |
| 3.9 | Attic / Insulation / Ventilation | Insulation type; approximate average depth at unfinished attic floor |
| 3.10 | Doors / Windows / Interior | Garage vehicle door type (manual or opener-equipped) |

### 2.2 Required Report Header Fields

- Client name
- Property address (street, city, state, zip)
- Inspection date and time
- Weather conditions at time of inspection
- Approximate outdoor air temperature
- People present
- Estimated year built
- Structure type
- Inspector name
- Company name, address, phone, email
- Inspector license number (state-specific)

### 2.3 Condition Rating Vocabulary

Every inspectable item uses these four statuses (I/NI/NP/D):

- **Inspected (I)** — Visually examined, no deficiency noted
- **Not Inspected (NI)** — Could not be inspected; reason required
- **Not Present (NP)** — Item does not exist at this property
- **Deficient (D)** — Condition adversely and materially affects performance or constitutes a hazard

### 2.4 Required Report Sections

1. Report header with all required fields
2. Ten system sections, each containing:
   - Status per item (I / NI / NP / D)
   - Required description fields
   - Deficiency narratives (if D)
   - Not-inspected reasons (if NI)
   - Photos attached to items
3. Summary of all deficiencies aggregated across all systems
4. Limitations and disclaimer block (Section 2 exclusions)

---

## 3. Domain Model

### 3.1 Bounded Contexts

The application is organized into four bounded contexts, each a discrete package with its own domain model:

```
scheduling   — inspector calendar, appointments, property metadata
inspection   — walkthrough session, checklist items, findings, photos
reporting    — report assembly, PDF generation, delivery
identity     — inspector accounts, companies, clients
```

### 3.2 Ubiquitous Language

| Term | Definition |
|---|---|
| **Inspector** | A licensed home inspector using the platform |
| **Company** | The inspection company an inspector belongs to |
| **Client** | The person commissioning the inspection |
| **Appointment** | A scheduled inspection at a specific property and time |
| **Inspection** | An active or completed walkthrough session bound to an Appointment |
| **System** | One of the ten InterNACHI SOP inspection systems |
| **Item** | A specific component or sub-area within a System |
| **Finding** | The inspector's documented observation for an Item |
| **Deficiency** | A Finding where the item is rated Deficient (D) |
| **Report** | The InterNACHI-compliant document assembled from a completed Inspection |
| **Delivery** | The act of sending a Report to a Client |

### 3.3 Aggregate Definitions

#### Aggregate: `Appointment` (scheduling context)
- **Root Entity**: `Appointment`
  - `AppointmentID` (UUID)
  - `InspectorID` (ref)
  - `ClientID` (ref)
  - `PropertyAddress` (value object)
  - `ScheduledAt` (time)
  - `EstimatedDuration` (duration)
  - `Status`: `Scheduled | InProgress | Completed | Cancelled`
  - `Notes` (string)
- **Value Object**: `PropertyAddress`
  - Street, City, State, Zip, Country
- **Value Object**: `ContactInfo`
  - Name, Email, Phone

#### Aggregate: `Inspection` (inspection context)
- **Root Entity**: `Inspection`
  - `InspectionID` (UUID)
  - `AppointmentID` (ref, correlationID across contexts)
  - `InspectorID` (ref)
  - `StartedAt` (time)
  - `CompletedAt` (time, nullable)
  - `Status`: `InProgress | Completed | Voided`
  - `HeaderData` (value object — weather, temp, attendees, year built, structure type)
  - `Systems` (collection of `SystemSection`)

- **Entity**: `SystemSection`
  - `SystemSectionID` (UUID)
  - `SystemType` (enum: Roof, Exterior, Foundation, Heating, Cooling, Plumbing, Electrical, Fireplace, Attic, Interior)
  - `Items` (collection of `InspectionItem`)
  - `Descriptions` (map[DescriptionKey]string — system-specific required fields)
  - `InspectorNotes` (string — section-level narrative)

- **Entity**: `InspectionItem`
  - `ItemID` (UUID)
  - `ItemKey` (enum — canonical item identifier, e.g., `RoofCoveringMaterial`, `GuttersDownspouts`)
  - `Label` (string — display name)
  - `Status`: `Inspected | NotInspected | NotPresent | Deficient`
  - `NotInspectedReason` (string, required when Status = NI)
  - `Findings` (collection of `Finding`)

- **Entity**: `Finding`
  - `FindingID` (UUID)
  - `Narrative` (string — inspector's written observation)
  - `IsDeficiency` (bool)
  - `Photos` (collection of `PhotoRef`)
  - `CreatedAt` (time)

- **Value Object**: `PhotoRef`
  - `PhotoID` (UUID)
  - `StoragePath` (string — local or remote path)
  - `CapturedAt` (time)

- **Domain Events**:
  - `InspectionStarted`
  - `ItemStatusChanged`
  - `DeficiencyAdded`
  - `InspectionCompleted`

#### Aggregate: `Report` (reporting context)
- **Root Entity**: `Report`
  - `ReportID` (UUID)
  - `InspectionID` (ref)
  - `GeneratedAt` (time)
  - `Status`: `Draft | Finalized | Delivered`
  - `PDFStoragePath` (string, set after generation)
  - `Deliveries` (collection of `Delivery`)

- **Entity**: `Delivery`
  - `DeliveryID` (UUID)
  - `RecipientEmail` (string)
  - `SentAt` (time, nullable)
  - `Status`: `Pending | Sent | Failed`
  - `FailureReason` (string, nullable)

- **Domain Events**:
  - `ReportGenerated`
  - `ReportDelivered`
  - `DeliveryFailed`

#### Aggregate: `Inspector` (identity context)
- **Root Entity**: `Inspector`
  - `InspectorID` (UUID)
  - `CompanyID` (ref)
  - `Name` (value object: FirstName, LastName)
  - `Email` (string)
  - `PasswordHash` (string)
  - `LicenseNumbers` (collection: State + LicenseNumber)
  - `Role`: `Owner | Member`
  - `CreatedAt` (time)

- **Aggregate**: `Company`
  - `CompanyID` (UUID)
  - `Name` (string)
  - `Address` (value object)
  - `Phone` (string)
  - `Email` (string)
  - `LogoStoragePath` (string, nullable)
  - `InspectorIDs` (collection of refs)

- **Aggregate**: `Client`
  - `ClientID` (UUID)
  - `CompanyID` (ref)
  - `Name` (value object: FirstName, LastName)
  - `Email` (string)
  - `Phone` (string)
  - `Inspections` (collection of InspectionID refs)

---

## 4. Go Architecture

### 4.1 Directory Structure

```
juno/
├── cmd/
│   └── server/
│       └── main.go               # Binary entrypoint; wires all dependencies
│
├── internal/
│   ├── scheduling/
│   │   ├── domain/
│   │   │   ├── appointment.go    # Aggregate root + entities
│   │   │   ├── repository.go     # AppointmentRepository interface
│   │   │   └── events.go         # Domain events
│   │   ├── application/
│   │   │   ├── commands/         # ScheduleInspection, CancelAppointment, etc.
│   │   │   └── queries/          # GetAppointment, ListAppointments, etc.
│   │   └── infrastructure/
│   │       ├── sqlite/           # SQLite repository impl
│   │       └── postgres/         # PostgreSQL repository impl
│   │
│   ├── inspection/
│   │   ├── domain/
│   │   │   ├── inspection.go     # Aggregate root
│   │   │   ├── system.go         # SystemSection entity + SystemType enum
│   │   │   ├── item.go           # InspectionItem entity + ItemKey catalog
│   │   │   ├── finding.go        # Finding entity + PhotoRef value object
│   │   │   ├── repository.go     # InspectionRepository interface
│   │   │   └── events.go
│   │   ├── application/
│   │   │   ├── commands/         # StartInspection, SetItemStatus, AddFinding, CompleteInspection, etc.
│   │   │   └── queries/          # GetInspection, ListInspections, GetSystemSection, etc.
│   │   └── infrastructure/
│   │       ├── sqlite/
│   │       └── postgres/
│   │
│   ├── reporting/
│   │   ├── domain/
│   │   │   ├── report.go         # Aggregate root
│   │   │   ├── delivery.go       # Delivery entity
│   │   │   ├── repository.go     # ReportRepository interface
│   │   │   └── events.go
│   │   ├── application/
│   │   │   ├── commands/         # GenerateReport, FinalizeReport, QueueDelivery, etc.
│   │   │   └── queries/          # GetReport, ListReports, etc.
│   │   └── infrastructure/
│   │       ├── sqlite/
│   │       ├── postgres/
│   │       ├── pdf/              # PDF generator (uses gofpdf or wkhtmltopdf)
│   │       └── email/            # Email delivery (SMTP, SendGrid, etc.)
│   │
│   ├── identity/
│   │   ├── domain/
│   │   │   ├── inspector.go
│   │   │   ├── company.go
│   │   │   ├── client.go
│   │   │   └── repository.go     # InspectorRepository, CompanyRepository, ClientRepository
│   │   ├── application/
│   │   │   ├── commands/
│   │   │   └── queries/
│   │   └── infrastructure/
│   │       ├── sqlite/
│   │       └── postgres/
│   │
│   └── sync/
│       ├── domain/               # SyncRecord aggregate (tracks local mutations for upload)
│       ├── application/          # SyncService: push local changes, pull remote updates
│       └── infrastructure/
│
├── api/
│   ├── rest/
│   │   ├── router.go             # Route registration (chi or stdlib)
│   │   ├── middleware/           # Auth, logging, CORS, offline-mode detection
│   │   └── handlers/             # HTTP handlers per bounded context
│   └── openapi/
│       └── spec.yaml             # OpenAPI 3 spec
│
├── web/                          # Compiled frontend assets (embedded via go:embed)
│   ├── static/
│   └── templates/
│
├── migrations/
│   ├── sqlite/
│   └── postgres/
│
└── pkg/
    ├── storage/                  # PhotoStorage interface + local/S3 implementations
    ├── clock/                    # Time abstraction (testability)
    ├── id/                       # UUID generation
    └── config/                   # Config struct + loader (env, file)
```

### 4.2 Repository Interface Pattern

Every bounded context exposes its persistence requirements as an interface in the `domain` package. Infrastructure implementations live in `infrastructure/sqlite/` and `infrastructure/postgres/` subpackages. The application layer only imports the domain interface.

```go
// internal/inspection/domain/repository.go
package domain

import "context"

type InspectionRepository interface {
    Save(ctx context.Context, inspection *Inspection) error
    FindByID(ctx context.Context, id InspectionID) (*Inspection, error)
    FindByAppointmentID(ctx context.Context, apptID AppointmentID) (*Inspection, error)
    FindByInspector(ctx context.Context, inspectorID InspectorID, filter InspectionFilter) ([]*Inspection, error)
    Delete(ctx context.Context, id InspectionID) error
}

type InspectionFilter struct {
    Status   *InspectionStatus
    FromDate *time.Time
    ToDate   *time.Time
    Limit    int
    Offset   int
}
```

All repository implementations are registered at startup via dependency injection in `cmd/server/main.go`. Switching backends requires only changing which implementation is wired in.

### 4.3 Application Commands and Queries (CQRS-lite)

Commands mutate state; queries return read models. Each is a standalone struct with a handler:

```go
// internal/inspection/application/commands/start_inspection.go
package commands

type StartInspectionCommand struct {
    AppointmentID string
    InspectorID   string
    HeaderData    InspectionHeaderInput
}

type StartInspectionHandler struct {
    repo        domain.InspectionRepository
    appointments scheduling_domain.AppointmentRepository
    eventBus    EventBus
}

func (h *StartInspectionHandler) Handle(ctx context.Context, cmd StartInspectionCommand) (string, error)
```

### 4.4 Domain Events

Domain events are published after aggregate mutations. The event bus is an in-process interface (synchronous or async). Key events drive cross-context reactions:

- `InspectionCompleted` → triggers `ReportService.InitializeDraftReport()`
- `ReportGenerated` → triggers notification to inspector
- `DeliveryFailed` → queues retry in sync context

### 4.5 Photo Storage Interface

```go
// pkg/storage/photo.go
package storage

type PhotoStorage interface {
    Save(ctx context.Context, photoID string, data io.Reader, mimeType string) (storagePath string, err error)
    Get(ctx context.Context, storagePath string) (io.ReadCloser, error)
    Delete(ctx context.Context, storagePath string) error
}
```

Implementations: `LocalDiskStorage` (offline / single-binary mode), `S3Storage` (cloud mode).

---

## 5. Offline-First Architecture

### 5.1 Storage Strategy

The application runs in one of two modes, selected at startup via config:

| Mode | Primary DB | Photo Storage | Sync |
|---|---|---|---|
| **Local** | SQLite (embedded) | Local disk | Manual or automatic when online |
| **Cloud** | PostgreSQL | S3 / compatible | N/A (always connected) |

In Local mode, the binary is entirely self-contained: it embeds migrations, runs SQLite via CGo-free `modernc.org/sqlite`, and stores photos on disk. Inspectors can run this on a laptop, tablet (via OS app wrapper), or a self-hosted server.

### 5.2 Sync Context

The `sync` bounded context manages bidirectional data synchronization:

- A **change log** table (`sync_records`) records every mutation (insert/update/delete) with a vector clock or lamport timestamp.
- When connectivity is detected, `SyncService.Push()` sends pending local mutations to the cloud API.
- `SyncService.Pull()` fetches mutations from the cloud that occurred on other devices (e.g., office admin scheduling from desktop while inspector is in the field).
- **Conflict resolution**: Last-writer-wins by timestamp for most fields. Inspection `Findings` are append-only — conflicts append rather than overwrite.
- The sync service is a background goroutine that polls connectivity and processes the queue.

### 5.3 Offline Delivery Queue

Report deliveries that cannot be sent immediately (no internet / no SMTP) are persisted with `Status = Pending`. When connectivity returns, the sync service processes pending deliveries in order.

---

## 6. API Design

### 6.1 Transport

REST over HTTP/1.1. All endpoints return `application/json`. The single Go binary serves both the API and the embedded frontend assets.

Authentication: JWT bearer tokens. Local mode uses a simplified single-inspector auth. Cloud (SaaS) mode uses company-scoped multi-tenant auth.

### 6.2 Core Endpoints

**Identity**
```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
GET    /api/v1/me
PUT    /api/v1/me
POST   /api/v1/companies
GET    /api/v1/companies/:id
POST   /api/v1/clients
GET    /api/v1/clients
GET    /api/v1/clients/:id
PUT    /api/v1/clients/:id
```

**Scheduling**
```
GET    /api/v1/appointments
POST   /api/v1/appointments
GET    /api/v1/appointments/:id
PUT    /api/v1/appointments/:id
DELETE /api/v1/appointments/:id
```

**Inspection**
```
POST   /api/v1/inspections                          # Start inspection for an appointment
GET    /api/v1/inspections/:id
GET    /api/v1/inspections/:id/systems
GET    /api/v1/inspections/:id/systems/:systemType
PUT    /api/v1/inspections/:id/systems/:systemType/descriptions
PUT    /api/v1/inspections/:id/systems/:systemType/items/:itemKey/status
POST   /api/v1/inspections/:id/systems/:systemType/items/:itemKey/findings
PUT    /api/v1/inspections/:id/systems/:systemType/items/:itemKey/findings/:findingID
DELETE /api/v1/inspections/:id/systems/:systemType/items/:itemKey/findings/:findingID
POST   /api/v1/inspections/:id/systems/:systemType/items/:itemKey/photos
DELETE /api/v1/inspections/:id/systems/:systemType/items/:itemKey/photos/:photoID
POST   /api/v1/inspections/:id/complete
GET    /api/v1/inspections/:id/summary              # Aggregated deficiency summary
```

**Reporting**
```
POST   /api/v1/reports                              # Generate report from completed inspection
GET    /api/v1/reports/:id
GET    /api/v1/reports/:id/pdf                      # Stream PDF file
PUT    /api/v1/reports/:id/finalize
POST   /api/v1/reports/:id/deliver                  # Queue email delivery
GET    /api/v1/reports/:id/deliveries
```

**Sync (Local mode only)**
```
POST   /api/v1/sync/push
POST   /api/v1/sync/pull
GET    /api/v1/sync/status
```

### 6.3 Photo Upload

Photos are uploaded as `multipart/form-data`. The server validates MIME type (JPEG, PNG, HEIC), stores via `PhotoStorage`, and returns a `PhotoRef`. Max photo size: 20 MB. Photos are stored locally in offline mode and uploaded to object storage on sync.

---

## 7. Inspection Checklist Item Catalog

The item catalog is compiled into the application (not user-editable). Each `ItemKey` maps to a system, label, SOP obligation type, and — for "shall describe" items — an expected value format.

### 7.1 Item Keys by System

**Roof (3.1)**
- `RoofCoveringMaterial` — describe (type)
- `GuttersDownspouts` — inspect
- `Vents` — inspect
- `Flashing` — inspect
- `Skylights` — inspect
- `Chimneys` — inspect
- `RoofPenetrations` — inspect
- `GeneralRoofStructure` — inspect

**Exterior (3.2)**
- `ExteriorWallCovering` — describe (type)
- `EavesSoffitsFascia` — inspect
- `Windows` — inspect (representative number)
- `ExteriorDoors` — inspect
- `FlashingAndTrim` — inspect
- `WalkwaysDriveways` — inspect
- `StairsStepsRamps` — inspect
- `PorchesPatiosDecks` — inspect
- `RailingsGuards` — inspect (deficiency: improper baluster spacing)
- `VegetationDrainageGrading` — inspect

**Foundation / Basement / Crawl Space (3.3)**
- `FoundationType` — describe (type)
- `CrawlSpaceAccess` — describe (location)
- `Foundation` — inspect
- `Basement` — inspect
- `CrawlSpace` — inspect
- `StructuralComponents` — inspect

**Heating (3.4)**
- `HeatingThermostatLocation` — describe
- `HeatingEnergySource` — describe
- `HeatingMethod` — describe
- `HeatingSystem` — inspect (using normal operating controls)

**Cooling (3.5)**
- `CoolingThermostatLocation` — describe
- `CoolingMethod` — describe
- `CoolingSystem` — inspect (using normal operating controls)

**Plumbing (3.6)**
- `WaterSupplyType` — describe (public/private)
- `MainWaterShutoffLocation` — describe
- `MainFuelShutoffLocation` — describe
- `FuelStorageLocation` — describe
- `WaterHeaterCapacity` — describe (if labeled)
- `MainWaterShutoff` — inspect
- `MainFuelShutoff` — inspect
- `WaterHeater` — inspect
- `InteriorWaterSupply` — inspect
- `Toilets` — inspect
- `SinksTubsShowers` — inspect (functional drainage)
- `DrainWasteVent` — inspect
- `SumpPumps` — inspect

**Electrical (3.7)**
- `MainServiceAmperage` — describe (if labeled)
- `WiringType` — describe
- `ServiceDrop` — inspect
- `ServiceHead` — inspect
- `ElectricMeter` — inspect
- `MainServiceDisconnect` — inspect
- `Panelboards` — inspect
- `ServiceGroundingBonding` — inspect
- `SwitchesLightingReceptacles` — inspect (representative number)
- `AFCIProtection` — inspect
- `GFCIProtection` — inspect
- `SmokeDetectors` — inspect (deficiency: absent)
- `CarbonMonoxideDetectors` — inspect (deficiency: absent)

**Fireplace (3.8)**
- `FireplaceType` — describe
- `FireplaceChimney` — inspect
- `Lintels` — inspect
- `DamperDoors` — inspect
- `CleanoutDoors` — inspect

**Attic / Insulation / Ventilation (3.9)**
- `InsulationType` — describe
- `InsulationDepth` — describe (approximate average depth in inches)
- `UnfinishedSpaceInsulation` — inspect
- `UnfinishedSpaceVentilation` — inspect
- `KitchenExhaust` — inspect
- `BathroomExhaust` — inspect
- `LaundryExhaust` — inspect

**Doors / Windows / Interior (3.10)**
- `GarageDoorType` — describe (manual or opener-equipped)
- `InteriorDoors` — inspect (representative number)
- `InteriorWindows` — inspect (representative number)
- `Floors` — inspect
- `Walls` — inspect
- `Ceilings` — inspect
- `InteriorStairs` — inspect
- `InteriorRailings` — inspect (deficiency: improper baluster spacing)
- `GarageVehicleDoors` — inspect
- `GarageDoorOpeners` — inspect (deficiency: missing photo-electric sensors)

---

## 8. Report Generation

### 8.1 PDF Structure

The generated PDF follows InterNACHI-preferred layout:

1. **Cover Page** — Company logo, inspector name, property address, inspection date, client name
2. **Table of Contents**
3. **Report Summary** — Deficiency count by system; all Deficient items listed with brief descriptions
4. **Inspection Details** — One section per system (3.1–3.10), containing:
   - Status table for all items (I / NI / NP / D)
   - Embedded photos adjacent to their findings
   - Required description fields
   - Deficiency narratives with "in need of correction" language
   - Section-level notes
5. **Limitations and Disclaimer** — Full InterNACHI Section 2 exclusions text
6. **Inspector Credentials** — License numbers, certifications, company info
7. **Appendix** — All photos in full size (optional, if not embedded inline)

### 8.2 PDF Engine

Use `github.com/signintech/gopdf` or `github.com/go-pdf/fpdf` for pure-Go PDF generation. The report template is defined in Go code (not external template files) for single-binary portability. Photo embedding is handled natively via JPEG/PNG inclusion.

### 8.3 Deficiency Language

When an item is marked Deficient, the generated report uses the phrasing: *"[Item label] was observed to be in need of correction."* followed by the inspector's narrative. Inspectors may write their own narrative text; the app does not auto-generate AI narratives.

---

## 9. Mobile-First UI/UX

### 9.1 Frontend Stack

The frontend is compiled to static assets embedded in the Go binary via `go:embed`. Recommended stack:

- **Framework**: SvelteKit or Vue 3 (lightweight, fast on mobile)
- **Styling**: Tailwind CSS with mobile-first breakpoints
- **Offline**: Service Worker with Cache API for asset caching; IndexedDB (via Dexie.js) for local read model caching when API is temporarily unreachable
- **PWA**: `manifest.json` enabling "Add to Home Screen" on iOS and Android; inspectors access Juno through the browser like a native app

### 9.2 Key Screens

**Dashboard / Home**
- Today's appointments at a glance
- Active in-progress inspections
- Recent completed reports
- Sync status indicator (online / offline / pending changes count)

**Appointment Calendar**
- Month/week/day view
- Tap appointment to view details or start inspection
- Quick-create appointment form (client, address, date/time)

**Inspection Walkthrough**
- System tab bar (10 tabs, one per SOP system)
- Each tab shows a scrollable checklist of items
- Per-item status selector (large tap targets: I / NI / NP / D)
- "NI" tapped → inline reason input
- "D" tapped → deficiency narrative drawer slides up; photo capture button
- Progress indicator: X / Y items addressed per system; overall completion %
- "Mark All As Inspected" quick action per system (for systems with no issues)
- Floating "Complete Inspection" button appears when all systems are addressed

**Item Detail View**
- Full-screen view for a single item
- Status selector at top
- Description fields (system-specific required fields)
- Findings list with edit/delete
- Photo grid with tap-to-expand and camera/library capture
- Inspector notes field

**Photo Capture**
- Native camera integration via `<input type="file" accept="image/*" capture="environment">`
- Thumbnail preview immediately after capture
- Photos stored locally; uploaded during sync

**Report Review**
- After completing inspection, inspector reviews a summary screen
- List of all deficiencies with edit capability
- Tap to navigate back to any item to edit
- "Generate Report" button produces PDF
- PDF preview in-browser

**Report Delivery**
- Client email field (pre-filled from appointment)
- Optional CC addresses
- Custom message body
- "Send Now" (if online) or "Queue for Send" (if offline)
- Delivery status tracking

**Client Management**
- Client list with search
- Client detail: contact info, inspection history

**Settings**
- Company profile (name, logo, address, phone, email)
- Inspector profile (name, license numbers per state, signature)
- Default report disclaimers (editable)
- Database mode (local SQLite / cloud PostgreSQL)
- Sync configuration (cloud API URL, credentials)
- Email settings (SMTP config or API key)

### 9.3 UX Principles for Field Use

- All interactive elements minimum 44×44pt touch target
- Swipe-to-navigate between systems during walkthrough
- Auto-save every field change (no "Save" buttons; changes persist immediately via API call)
- Optimistic UI updates: item status changes immediately in the UI while the API call resolves in the background
- Graceful degradation: if the API call fails (offline), the change is queued in IndexedDB and retried
- Dark mode support (inspectors often work in dark attics and basements)
- Large, high-contrast text

---

## 10. Feature Specifications

### 10.1 Scheduling

- Create, edit, and cancel appointments
- Appointment statuses: `Scheduled → InProgress → Completed | Cancelled`
- Calendar view (month/week/day) with appointments displayed
- Start inspection directly from an appointment (creates Inspection aggregate)
- Appointment notes field for pre-inspection information

### 10.2 Inspection Walkthrough

- One active Inspection per Appointment
- Inspection progress auto-saved; inspectors can close the app and resume
- Each item defaults to `NotInspected`; inspector must actively set status
- Inspector can add multiple findings per item
- Findings can include photos (multiple per finding)
- Descriptions are saved per system (not per item) for SOP "shall describe" items
- Inspector can add free-text notes at the item level and section level
- "Complete Inspection" validates that all items have a status other than the default placeholder and that all required description fields are filled; if not, it lists what's missing

### 10.3 Photo Management

- Photos captured via camera or selected from device library
- Photos are associated with a specific Finding on a specific Item
- Photos are stored locally; displayed in report adjacent to the finding
- Photos are included in PDF as embedded images

### 10.4 Report Generation and Review

- Reports are generated from completed Inspections
- Inspector can re-generate a report if they edit findings after initial generation
- Inspector can preview the PDF before finalizing
- Finalized reports are locked (items can no longer be edited)
- Draft reports can be edited freely

### 10.5 Report Delivery

- Deliver by email; PDF attached
- Support multiple recipient addresses (client + buyer's agent, for example)
- Email delivery queued offline and sent when connectivity restored
- Delivery history with timestamp and status per recipient

### 10.6 Multi-Inspector Company Support (Cloud Mode)

- Company account with multiple inspector members
- Owner role: can manage inspectors, view all company inspections/reports
- Member role: can only access their own inspections
- Shared client database within a company

### 10.7 Sync

- Sync is initiated manually (pull-to-refresh gesture or "Sync" button) or automatically on connectivity change
- Sync status displayed prominently: "All changes saved", "X changes pending sync", "Offline"
- Inspector can view the pending change queue
- Sync log for debugging

---

## 11. Data Storage Schema (SQLite / Logical)

The following tables apply to the SQLite schema. PostgreSQL uses equivalent schemas with different DDL syntax (UUID types, etc.).

```
appointments
  id TEXT PK, inspector_id TEXT, client_id TEXT, street TEXT, city TEXT,
  state TEXT, zip TEXT, scheduled_at INTEGER, estimated_duration_min INTEGER,
  status TEXT, notes TEXT, created_at INTEGER, updated_at INTEGER

inspections
  id TEXT PK, appointment_id TEXT, inspector_id TEXT, started_at INTEGER,
  completed_at INTEGER, status TEXT, weather TEXT, temperature_f INTEGER,
  attendees TEXT, year_built INTEGER, structure_type TEXT,
  created_at INTEGER, updated_at INTEGER

system_sections
  id TEXT PK, inspection_id TEXT, system_type TEXT, inspector_notes TEXT,
  updated_at INTEGER

system_descriptions
  system_section_id TEXT, description_key TEXT, value TEXT,
  PRIMARY KEY (system_section_id, description_key)

inspection_items
  id TEXT PK, system_section_id TEXT, item_key TEXT, status TEXT,
  not_inspected_reason TEXT, updated_at INTEGER

findings
  id TEXT PK, inspection_item_id TEXT, narrative TEXT, is_deficiency INTEGER,
  created_at INTEGER, updated_at INTEGER

photos
  id TEXT PK, finding_id TEXT, storage_path TEXT, mime_type TEXT,
  captured_at INTEGER, uploaded INTEGER DEFAULT 0

reports
  id TEXT PK, inspection_id TEXT, generated_at INTEGER, status TEXT,
  pdf_storage_path TEXT, created_at INTEGER, updated_at INTEGER

deliveries
  id TEXT PK, report_id TEXT, recipient_email TEXT, sent_at INTEGER,
  status TEXT, failure_reason TEXT

inspectors
  id TEXT PK, company_id TEXT, first_name TEXT, last_name TEXT,
  email TEXT, password_hash TEXT, role TEXT, created_at INTEGER

inspector_licenses
  inspector_id TEXT, state TEXT, license_number TEXT,
  PRIMARY KEY (inspector_id, state)

companies
  id TEXT PK, name TEXT, street TEXT, city TEXT, state TEXT, zip TEXT,
  phone TEXT, email TEXT, logo_storage_path TEXT

clients
  id TEXT PK, company_id TEXT, first_name TEXT, last_name TEXT,
  email TEXT, phone TEXT, created_at INTEGER

sync_records
  id TEXT PK, table_name TEXT, record_id TEXT, operation TEXT,
  payload TEXT, lamport_clock INTEGER, synced INTEGER DEFAULT 0,
  created_at INTEGER
```

---

## 12. Configuration

The application is configured via environment variables (12-factor) or a YAML config file. Key config groups:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "local"            # "local" | "cloud"

database:
  driver: "sqlite"         # "sqlite" | "postgres"
  dsn: "./juno.db"         # SQLite path or Postgres DSN

storage:
  driver: "local"          # "local" | "s3"
  local_path: "./data/photos"
  s3_bucket: ""
  s3_region: ""

sync:
  cloud_api_url: ""
  enabled: false

email:
  driver: "smtp"           # "smtp" | "sendgrid" | "queue_only"
  smtp_host: ""
  smtp_port: 587
  smtp_user: ""
  smtp_pass: ""

auth:
  jwt_secret: ""
  token_ttl_hours: 24
```

---

## 13. Non-Functional Requirements

| Requirement | Target |
|---|---|
| PDF generation time | < 5 seconds for a complete 100-item inspection |
| API response time (local SQLite) | < 100ms p95 |
| Offline operation | 100% of inspection and report generation without network |
| Photo storage per inspection | Up to 200 photos at up to 20 MB each |
| Concurrent inspections per company | No hard limit in local mode; cloud mode scales horizontally |
| Mobile browser support | iOS Safari 15+, Chrome for Android 100+ |
| Accessibility | WCAG 2.1 AA for all core flows |
| Binary size | < 50 MB including embedded frontend |

---

## 14. Future Considerations (Out of Scope for MVP)

- **Template customization**: Inspectors can create their own item catalogs beyond the InterNACHI baseline
- **AI-assisted narratives**: Suggest deficiency descriptions based on item type (requires network)
- **E-signature**: Client signs the inspection agreement in-app
- **Billing / Subscription management**: Stripe integration for SaaS billing
- **State-specific compliance**: TREC (Texas), DPOR (Virginia), and other state-specific SOP overlays
- **Ancillary inspections**: Pool/spa, sewer scope, radon — extend the item catalog
- **Inspector marketplace**: Client-facing booking portal
- **Analytics dashboard**: Deficiency frequency reports, inspection volume trends
