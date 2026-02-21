-- =============================================================================
-- 001_initial_schema.sql
-- Full schema for all bounded contexts: identity, scheduling, inspection,
-- reporting, and sync.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Identity context
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS companies (
    id                TEXT    NOT NULL PRIMARY KEY,
    name              TEXT    NOT NULL,
    street            TEXT    NOT NULL DEFAULT '',
    city              TEXT    NOT NULL DEFAULT '',
    state             TEXT    NOT NULL DEFAULT '',
    zip               TEXT    NOT NULL DEFAULT '',
    country           TEXT    NOT NULL DEFAULT '',
    phone             TEXT    NOT NULL DEFAULT '',
    email             TEXT    NOT NULL DEFAULT '',
    logo_storage_path TEXT    NOT NULL DEFAULT '',
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS inspectors (
    id            TEXT    NOT NULL PRIMARY KEY,
    company_id    TEXT    NOT NULL,
    first_name    TEXT    NOT NULL,
    last_name     TEXT    NOT NULL,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    role          TEXT    NOT NULL DEFAULT 'member',
    created_at    INTEGER NOT NULL,
    updated_at    INTEGER NOT NULL,
    FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inspector_licenses (
    inspector_id   TEXT NOT NULL,
    state          TEXT NOT NULL,
    license_number TEXT NOT NULL,
    PRIMARY KEY (inspector_id, state),
    FOREIGN KEY (inspector_id) REFERENCES inspectors (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS clients (
    id         TEXT    NOT NULL PRIMARY KEY,
    company_id TEXT    NOT NULL,
    first_name TEXT    NOT NULL,
    last_name  TEXT    NOT NULL,
    email      TEXT    NOT NULL,
    phone      TEXT    NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE
);

-- -----------------------------------------------------------------------------
-- Scheduling context
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS appointments (
    id                     TEXT    NOT NULL PRIMARY KEY,
    inspector_id           TEXT    NOT NULL,
    client_id              TEXT    NOT NULL,
    street                 TEXT    NOT NULL,
    city                   TEXT    NOT NULL,
    state                  TEXT    NOT NULL,
    zip                    TEXT    NOT NULL,
    country                TEXT    NOT NULL DEFAULT 'US',
    scheduled_at           INTEGER NOT NULL,
    estimated_duration_min INTEGER NOT NULL DEFAULT 120,
    status                 TEXT    NOT NULL DEFAULT 'scheduled',
    notes                  TEXT    NOT NULL DEFAULT '',
    created_at             INTEGER NOT NULL,
    updated_at             INTEGER NOT NULL,
    FOREIGN KEY (inspector_id) REFERENCES inspectors (id),
    FOREIGN KEY (client_id)    REFERENCES clients (id)
);

-- -----------------------------------------------------------------------------
-- Inspection context
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS inspections (
    id             TEXT    NOT NULL PRIMARY KEY,
    appointment_id TEXT    NOT NULL,
    inspector_id   TEXT    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'in_progress',
    weather        TEXT    NOT NULL DEFAULT '',
    temperature_f  INTEGER NOT NULL DEFAULT 0,
    attendees      TEXT    NOT NULL DEFAULT '[]', -- JSON array of strings
    year_built     INTEGER NOT NULL DEFAULT 0,
    structure_type TEXT    NOT NULL DEFAULT '',
    started_at     INTEGER NOT NULL,
    completed_at   INTEGER,                       -- NULL until completed
    created_at     INTEGER NOT NULL,
    updated_at     INTEGER NOT NULL,
    FOREIGN KEY (appointment_id) REFERENCES appointments (id),
    FOREIGN KEY (inspector_id)   REFERENCES inspectors (id)
);

CREATE TABLE IF NOT EXISTS system_sections (
    id              TEXT    NOT NULL PRIMARY KEY,
    inspection_id   TEXT    NOT NULL,
    system_type     TEXT    NOT NULL,             -- matches SystemType enum
    inspector_notes TEXT    NOT NULL DEFAULT '',
    updated_at      INTEGER NOT NULL,
    FOREIGN KEY (inspection_id) REFERENCES inspections (id) ON DELETE CASCADE,
    UNIQUE (inspection_id, system_type)
);

CREATE TABLE IF NOT EXISTS system_descriptions (
    system_section_id TEXT NOT NULL,
    description_key   TEXT NOT NULL,             -- matches DescriptionKey constant
    value             TEXT NOT NULL,
    PRIMARY KEY (system_section_id, description_key),
    FOREIGN KEY (system_section_id) REFERENCES system_sections (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inspection_items (
    id                   TEXT    NOT NULL PRIMARY KEY,
    system_section_id    TEXT    NOT NULL,
    item_key             TEXT    NOT NULL,        -- matches ItemKey constant
    label                TEXT    NOT NULL,        -- display label (stored for resilience)
    status               TEXT    NOT NULL DEFAULT 'NI',
    not_inspected_reason TEXT    NOT NULL DEFAULT '',
    updated_at           INTEGER NOT NULL,
    FOREIGN KEY (system_section_id) REFERENCES system_sections (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS findings (
    id                 TEXT    NOT NULL PRIMARY KEY,
    inspection_item_id TEXT    NOT NULL,
    narrative          TEXT    NOT NULL,
    is_deficiency      INTEGER NOT NULL DEFAULT 0, -- boolean: 0 or 1
    created_at         INTEGER NOT NULL,
    updated_at         INTEGER NOT NULL,
    FOREIGN KEY (inspection_item_id) REFERENCES inspection_items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS photos (
    id           TEXT    NOT NULL PRIMARY KEY,
    finding_id   TEXT    NOT NULL,
    storage_path TEXT    NOT NULL,
    mime_type    TEXT    NOT NULL,
    captured_at  INTEGER NOT NULL,
    uploaded     INTEGER NOT NULL DEFAULT 0,      -- boolean: 0 = local only, 1 = synced
    FOREIGN KEY (finding_id) REFERENCES findings (id) ON DELETE CASCADE
);

-- -----------------------------------------------------------------------------
-- Reporting context
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS reports (
    id               TEXT    NOT NULL PRIMARY KEY,
    inspection_id    TEXT    NOT NULL UNIQUE,
    inspector_id     TEXT    NOT NULL,
    status           TEXT    NOT NULL DEFAULT 'draft',
    pdf_storage_path TEXT    NOT NULL DEFAULT '',
    generated_at     INTEGER,                      -- NULL until PDF is generated
    created_at       INTEGER NOT NULL,
    updated_at       INTEGER NOT NULL,
    FOREIGN KEY (inspection_id) REFERENCES inspections (id),
    FOREIGN KEY (inspector_id)  REFERENCES inspectors (id)
);

CREATE TABLE IF NOT EXISTS deliveries (
    id              TEXT    NOT NULL PRIMARY KEY,
    report_id       TEXT    NOT NULL,
    recipient_email TEXT    NOT NULL,
    status          TEXT    NOT NULL DEFAULT 'pending',
    attempts        INTEGER NOT NULL DEFAULT 0,
    sent_at         INTEGER,                       -- NULL until successfully sent
    failure_reason  TEXT    NOT NULL DEFAULT '',
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    FOREIGN KEY (report_id) REFERENCES reports (id) ON DELETE CASCADE
);

-- -----------------------------------------------------------------------------
-- Sync context
-- -----------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sync_records (
    id            TEXT    NOT NULL PRIMARY KEY,
    table_name    TEXT    NOT NULL,
    record_id     TEXT    NOT NULL,
    operation     TEXT    NOT NULL,               -- 'insert' | 'update' | 'delete'
    payload       TEXT    NOT NULL DEFAULT '{}',  -- JSON snapshot of the record
    lamport_clock INTEGER NOT NULL DEFAULT 0,
    synced        INTEGER NOT NULL DEFAULT 0,     -- boolean: 0 = pending, 1 = pushed
    created_at    INTEGER NOT NULL
);

-- =============================================================================
-- Indexes for common query patterns
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_inspectors_company    ON inspectors (company_id);
CREATE INDEX IF NOT EXISTS idx_clients_company       ON clients (company_id);

CREATE INDEX IF NOT EXISTS idx_appointments_inspector  ON appointments (inspector_id);
CREATE INDEX IF NOT EXISTS idx_appointments_scheduled  ON appointments (scheduled_at);

CREATE INDEX IF NOT EXISTS idx_inspections_appointment ON inspections (appointment_id);
CREATE INDEX IF NOT EXISTS idx_inspections_inspector   ON inspections (inspector_id);
CREATE INDEX IF NOT EXISTS idx_inspections_status      ON inspections (status);

CREATE INDEX IF NOT EXISTS idx_system_sections_inspection ON system_sections (inspection_id);
CREATE INDEX IF NOT EXISTS idx_inspection_items_section   ON inspection_items (system_section_id);
CREATE INDEX IF NOT EXISTS idx_findings_item              ON findings (inspection_item_id);
CREATE INDEX IF NOT EXISTS idx_photos_finding             ON photos (finding_id);

CREATE INDEX IF NOT EXISTS idx_reports_inspector    ON reports (inspector_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_report    ON deliveries (report_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_status    ON deliveries (status);

CREATE INDEX IF NOT EXISTS idx_sync_records_synced  ON sync_records (synced);
CREATE INDEX IF NOT EXISTS idx_sync_records_table   ON sync_records (table_name, record_id);
