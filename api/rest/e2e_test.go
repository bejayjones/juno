// Package rest_test contains end-to-end tests that wire the full server stack
// (real SQLite, real services, real JWT auth) and exercise the HTTP API top to
// bottom, covering the complete inspector workflow.
package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/bejayjones/juno/api/rest"
	inspectionapp "github.com/bejayjones/juno/internal/inspection/application"
	inspectiondomain "github.com/bejayjones/juno/internal/inspection/domain"
	inspectionsqlite "github.com/bejayjones/juno/internal/inspection/infrastructure/sqlite"
	identityapp "github.com/bejayjones/juno/internal/identity/application"
	identityauth "github.com/bejayjones/juno/internal/identity/infrastructure/auth"
	identitysqlite "github.com/bejayjones/juno/internal/identity/infrastructure/sqlite"
	"github.com/bejayjones/juno/internal/platform/db"
	reportingapp "github.com/bejayjones/juno/internal/reporting/application"
	reportingemail "github.com/bejayjones/juno/internal/reporting/infrastructure/email"
	reportingpdf "github.com/bejayjones/juno/internal/reporting/infrastructure/pdf"
	reportingsqlite "github.com/bejayjones/juno/internal/reporting/infrastructure/sqlite"
	schedulingapp "github.com/bejayjones/juno/internal/scheduling/application"
	schedulingsqlite "github.com/bejayjones/juno/internal/scheduling/infrastructure/sqlite"
	syncapp "github.com/bejayjones/juno/internal/sync/application"
	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
	syncsqlite "github.com/bejayjones/juno/internal/sync/infrastructure/sqlite"
	"github.com/bejayjones/juno/internal/sync/recorder"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/storage/local"
)

// ── Stack setup ──────────────────────────────────────────────────────────────

// e2eStack wires the full server stack using a real SQLite database in a
// temporary directory. The returned httptest.Server is closed by t.Cleanup.
func e2eStack(t *testing.T) *httptest.Server {
	t.Helper()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "e2e.db")

	database, err := db.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Sync infrastructure.
	syncRepo := syncsqlite.NewSyncRepository(database)
	maxClock, err := syncRepo.MaxClock(context.Background())
	if err != nil {
		t.Fatalf("max clock: %v", err)
	}
	lamportClock := syncdomain.NewLamportClock(maxClock)
	syncRecorder := recorder.New(lamportClock)
	syncSvc := syncapp.NewSyncService(syncRepo, lamportClock)

	// Identity infrastructure.
	const jwtSecret = "e2e-test-secret"
	jwtSvc := identityauth.NewJWTService(jwtSecret, 24)
	hasher := identityauth.BcryptHasher{}
	companyRepo := identitysqlite.NewCompanyRepository(database).WithRecorder(syncRecorder)
	inspectorRepo := identitysqlite.NewInspectorRepository(database).WithRecorder(syncRecorder)
	clientRepo := identitysqlite.NewClientRepository(database).WithRecorder(syncRecorder)

	clk := clock.Real()
	inspectorSvc := identityapp.NewInspectorService(inspectorRepo, companyRepo, hasher, jwtSvc, clk)
	companySvc := identityapp.NewCompanyService(companyRepo, clk)
	clientSvc := identityapp.NewClientService(clientRepo, clk)

	// Scheduling.
	appointmentRepo := schedulingsqlite.NewAppointmentRepository(database).WithRecorder(syncRecorder)
	appointmentSvc := schedulingapp.NewAppointmentService(appointmentRepo, clk)

	// Photo storage (local disk under temp dir).
	photoStore := local.New(filepath.Join(tmp, "photos"))

	// Inspection.
	inspectionRepo := inspectionsqlite.NewInspectionRepository(database).WithRecorder(syncRecorder)
	inspectionSvc := inspectionapp.NewInspectionService(inspectionRepo, photoStore, clk)

	// Reporting.
	reportRepo := reportingsqlite.NewReportRepository(database).WithRecorder(syncRecorder)
	pdfGen := reportingpdf.NewGenerator()
	reportsDir := filepath.Join(tmp, "reports")
	emailSvc := reportingemail.NewQueueOnlySender()
	reportSvc := reportingapp.NewReportService(reportRepo, inspectionRepo, pdfGen, emailSvc, reportsDir, clk)

	tokenVerifier := rest.NewJWTAdapter(jwtSvc)

	srv := rest.NewServer(nil, database, inspectorSvc, companySvc, clientSvc,
		appointmentSvc, inspectionSvc, reportSvc, syncSvc, tokenVerifier)

	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts
}

// ── HTTP helpers ─────────────────────────────────────────────────────────────

// do sends an HTTP request and decodes the JSON response body into out (if not nil).
// Returns the HTTP status code.
func do(t *testing.T, client *http.Client, method, url, token string, body, out any) int {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, url, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response from %s %s: %v", method, url, err)
		}
	} else {
		io.Copy(io.Discard, resp.Body)
	}
	return resp.StatusCode
}

// mustDo calls do and fatals if the status code is not want.
func mustDo(t *testing.T, client *http.Client, method, url, token string, body, out any, want int) {
	t.Helper()
	got := do(t, client, method, url, token, body, out)
	if got != want {
		t.Fatalf("%s %s: status %d, want %d", method, url, got, want)
	}
}

// ── Tests ────────────────────────────────────────────────────────────────────

// TestE2E_HappyPath exercises the complete inspector workflow end-to-end:
//
//  1. Register + login
//  2. Create client
//  3. Create appointment
//  4. Start inspection
//  5. Set every item in every system to "I"
//  6. Fill every required description
//  7. Add one deficiency finding
//  8. Complete inspection
//  9. Generate report (PDF runs synchronously in queue_only mode)
//  10. Finalize report
//  11. Queue delivery to a recipient
//  12. Retry deliveries
//  13. Verify sync records were written
func TestE2E_HappyPath(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	// ── 1. Register ──────────────────────────────────────────────────────────
	var regOut identityapp.RegisterOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/register", "", map[string]any{
		"first_name":   "Alice",
		"last_name":    "Inspector",
		"email":        "alice@juno.test",
		"password":     "correcthorsebatterystaple",
		"company_name": "Acme Inspections LLC",
	}, &regOut, http.StatusCreated)

	if regOut.Token == "" {
		t.Fatal("register: empty token")
	}
	if regOut.Inspector.Email != "alice@juno.test" {
		t.Errorf("register: email = %q", regOut.Inspector.Email)
	}
	token := regOut.Token

	// ── 2. GET /me ───────────────────────────────────────────────────────────
	var meOut identityapp.InspectorView
	mustDo(t, client, http.MethodGet, base+"/api/v1/me", token, nil, &meOut, http.StatusOK)
	if meOut.ID != regOut.Inspector.ID {
		t.Errorf("GET /me: id = %q, want %q", meOut.ID, regOut.Inspector.ID)
	}

	// ── 3. Login (verify credentials round-trip) ──────────────────────────────
	var loginOut identityapp.LoginOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/login", "", map[string]any{
		"email":    "alice@juno.test",
		"password": "correcthorsebatterystaple",
	}, &loginOut, http.StatusOK)
	if loginOut.Token == "" {
		t.Fatal("login: empty token")
	}

	// ── 4. Create client ─────────────────────────────────────────────────────
	var clientOut identityapp.ClientView
	mustDo(t, client, http.MethodPost, base+"/api/v1/clients", token, map[string]any{
		"first_name": "Bob",
		"last_name":  "Buyer",
		"email":      "bob@buyer.test",
		"phone":      "555-0100",
	}, &clientOut, http.StatusCreated)
	if clientOut.ID == "" {
		t.Fatal("create client: empty id")
	}

	// List clients — should have exactly one.
	var clients []identityapp.ClientView
	mustDo(t, client, http.MethodGet, base+"/api/v1/clients", token, nil, &clients, http.StatusOK)
	if len(clients) != 1 {
		t.Errorf("list clients: got %d, want 1", len(clients))
	}

	// ── 5. Create appointment ────────────────────────────────────────────────
	scheduledAt := time.Now().Add(24 * time.Hour).Unix()
	var apptOut schedulingapp.AppointmentView
	mustDo(t, client, http.MethodPost, base+"/api/v1/appointments", token, map[string]any{
		"client_id":    clientOut.ID,
		"street":       "123 Main St",
		"city":         "Springfield",
		"state":        "IL",
		"zip":          "62701",
		"country":      "US",
		"scheduled_at": scheduledAt,
		"duration_min": 120,
		"notes":        "Buyer present at inspection.",
	}, &apptOut, http.StatusCreated)
	if apptOut.ID == "" {
		t.Fatal("create appointment: empty id")
	}
	if apptOut.Status != "scheduled" {
		t.Errorf("appointment status = %q, want scheduled", apptOut.Status)
	}

	// GET appointment.
	var apptGet schedulingapp.AppointmentView
	mustDo(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/appointments/%s", base, apptOut.ID), token, nil, &apptGet, http.StatusOK)
	if apptGet.ID != apptOut.ID {
		t.Errorf("GET appointment: id mismatch")
	}

	// ── 6. Start inspection ──────────────────────────────────────────────────
	var inspOut inspectionapp.InspectionView
	mustDo(t, client, http.MethodPost, base+"/api/v1/inspections", token, map[string]any{
		"appointment_id": apptOut.ID,
	}, &inspOut, http.StatusCreated)
	if inspOut.ID == "" {
		t.Fatal("start inspection: empty id")
	}
	if inspOut.Status != "in_progress" {
		t.Errorf("inspection status = %q, want in_progress", inspOut.Status)
	}
	if len(inspOut.Systems) != 10 {
		t.Errorf("inspection systems = %d, want 10", len(inspOut.Systems))
	}
	inspID := inspOut.ID

	// ── 7. Set all items to "I" and fill required descriptions ───────────────
	// Iterate the catalog (authoritative) — not the API response — to ensure
	// every item and description key is hit.
	for _, sysDef := range inspectiondomain.Catalog {
		systemType := string(sysDef.Type)

		// Set all items to "Inspected".
		for _, itemDef := range sysDef.Items {
			itemKey := string(itemDef.Key)
			var sysOut inspectionapp.SystemSectionView
			mustDo(t, client,
				http.MethodPut,
				fmt.Sprintf("%s/api/v1/inspections/%s/systems/%s/items/%s/status",
					base, inspID, systemType, itemKey),
				token,
				map[string]any{"status": "I"},
				&sysOut,
				http.StatusOK,
			)
		}

		// Fill all required descriptions.
		if len(sysDef.RequiredDescriptions) > 0 {
			descs := make(map[string]string, len(sysDef.RequiredDescriptions))
			for _, req := range sysDef.RequiredDescriptions {
				descs[string(req.Key)] = fmt.Sprintf("Test value for %s", req.Key)
			}
			var sysOut inspectionapp.SystemSectionView
			mustDo(t, client,
				http.MethodPut,
				fmt.Sprintf("%s/api/v1/inspections/%s/systems/%s/descriptions",
					base, inspID, systemType),
				token,
				descs,
				&sysOut,
				http.StatusOK,
			)
		}
	}

	// ── 8. Add a deficiency finding on the first roof item ───────────────────
	var findingOut inspectionapp.FindingView
	mustDo(t, client,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/inspections/%s/systems/roof/items/roof.gutters_downspouts/findings",
			base, inspID),
		token,
		map[string]any{
			"narrative":      "Gutters show rust and separation at east corner.",
			"is_deficiency":  true,
		},
		&findingOut,
		http.StatusCreated,
	)
	if findingOut.ID == "" {
		t.Fatal("add finding: empty id")
	}
	if !findingOut.IsDeficiency {
		t.Error("add finding: is_deficiency should be true")
	}
	findingID := findingOut.ID

	// Update the finding.
	var updatedFinding inspectionapp.FindingView
	mustDo(t, client,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/inspections/%s/systems/roof/items/roof.gutters_downspouts/findings/%s",
			base, inspID, findingID),
		token,
		map[string]any{
			"narrative":     "Gutters show rust and separation at east corner. Recommend cleaning and resealing.",
			"is_deficiency": true,
		},
		&updatedFinding,
		http.StatusOK,
	)

	// GET the system to verify the finding is persisted.
	var roofSys inspectionapp.SystemSectionView
	mustDo(t, client,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/inspections/%s/systems/roof", base, inspID),
		token, nil, &roofSys, http.StatusOK,
	)
	foundFinding := false
	for _, item := range roofSys.Items {
		for _, f := range item.Findings {
			if f.ID == findingID {
				foundFinding = true
			}
		}
	}
	if !foundFinding {
		t.Error("finding not found in GET system response")
	}

	// ── 9. Complete inspection ────────────────────────────────────────────────
	var completedInsp inspectionapp.InspectionView
	mustDo(t, client,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/inspections/%s/complete", base, inspID),
		token, map[string]any{}, &completedInsp, http.StatusOK,
	)
	if completedInsp.Status != "completed" {
		t.Errorf("complete inspection: status = %q, want completed", completedInsp.Status)
	}

	// GET inspection summary (deficiency list).
	var deficiencies []inspectionapp.DeficiencyView
	mustDo(t, client,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/inspections/%s/summary", base, inspID),
		token, nil, &deficiencies, http.StatusOK,
	)
	if len(deficiencies) != 1 {
		t.Errorf("deficiency summary: got %d, want 1", len(deficiencies))
	}

	// ── 10. Generate report ───────────────────────────────────────────────────
	var reportOut reportingapp.ReportView
	mustDo(t, client, http.MethodPost, base+"/api/v1/reports", token, map[string]any{
		"inspection_id": inspID,
	}, &reportOut, http.StatusCreated)
	if reportOut.ID == "" {
		t.Fatal("generate report: empty id")
	}
	if reportOut.InspectionID != inspID {
		t.Errorf("report inspection_id = %q, want %q", reportOut.InspectionID, inspID)
	}
	if reportOut.Status != "draft" {
		t.Errorf("report status = %q, want draft", reportOut.Status)
	}
	reportID := reportOut.ID

	// Poll for PDF generation (synchronous in queue_only mode; should be fast).
	var pdfReady bool
	for range 20 {
		var r reportingapp.ReportView
		mustDo(t, client, http.MethodGet,
			fmt.Sprintf("%s/api/v1/reports/%s", base, reportID),
			token, nil, &r, http.StatusOK,
		)
		if r.GeneratedAt != nil {
			pdfReady = true
			reportOut = r
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !pdfReady {
		t.Fatal("PDF not generated within 2 seconds")
	}

	// Download PDF — should return 200 with application/pdf content-type.
	pdfResp, err := func() (*http.Response, error) {
		req, _ := http.NewRequest(http.MethodGet,
			fmt.Sprintf("%s/api/v1/reports/%s/pdf", base, reportID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		return client.Do(req)
	}()
	if err != nil {
		t.Fatalf("GET PDF: %v", err)
	}
	defer pdfResp.Body.Close()
	io.Copy(io.Discard, pdfResp.Body)
	if pdfResp.StatusCode != http.StatusOK {
		t.Errorf("GET PDF: status %d, want 200", pdfResp.StatusCode)
	}

	// List reports — should have 1.
	var reportList []reportingapp.ReportView
	mustDo(t, client, http.MethodGet, base+"/api/v1/reports", token, nil, &reportList, http.StatusOK)
	if len(reportList) != 1 {
		t.Errorf("list reports: got %d, want 1", len(reportList))
	}

	// ── 11. Finalize report ───────────────────────────────────────────────────
	var finalizedReport reportingapp.ReportView
	mustDo(t, client,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/reports/%s/finalize", base, reportID),
		token, map[string]any{}, &finalizedReport, http.StatusOK,
	)
	if finalizedReport.Status != "finalized" {
		t.Errorf("finalize report: status = %q, want finalized", finalizedReport.Status)
	}

	// Finalizing again should conflict.
	status := do(t, client,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/reports/%s/finalize", base, reportID),
		token, map[string]any{}, nil,
	)
	if status != http.StatusConflict {
		t.Errorf("re-finalize: status %d, want 409", status)
	}

	// ── 12. Queue delivery ────────────────────────────────────────────────────
	var deliveryOut reportingapp.DeliveryView
	mustDo(t, client,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/reports/%s/deliver", base, reportID),
		token,
		map[string]any{"recipient_email": "bob@buyer.test"},
		&deliveryOut,
		http.StatusCreated,
	)
	if deliveryOut.ID == "" {
		t.Fatal("deliver: empty id")
	}
	if deliveryOut.RecipientEmail != "bob@buyer.test" {
		t.Errorf("delivery recipient = %q, want bob@buyer.test", deliveryOut.RecipientEmail)
	}

	// GET report to confirm delivery appears in list.
	var reportWithDelivery reportingapp.ReportView
	mustDo(t, client, http.MethodGet,
		fmt.Sprintf("%s/api/v1/reports/%s", base, reportID),
		token, nil, &reportWithDelivery, http.StatusOK,
	)
	if len(reportWithDelivery.Deliveries) != 1 {
		t.Errorf("report deliveries: got %d, want 1", len(reportWithDelivery.Deliveries))
	}

	// ── 13. Retry deliveries ──────────────────────────────────────────────────
	var retryOut reportingapp.ReportView
	mustDo(t, client,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/reports/%s/deliveries/retry", base, reportID),
		token, map[string]any{}, &retryOut, http.StatusOK,
	)

	// ── 14. Sync status — Lamport clock should have advanced ─────────────────
	var syncStatus struct {
		PendingCount int   `json:"pending_count"`
		CurrentClock int64 `json:"current_clock"`
	}
	mustDo(t, client, http.MethodGet, base+"/api/v1/sync/status", token, nil, &syncStatus, http.StatusOK)
	if syncStatus.CurrentClock == 0 {
		t.Error("sync status: current_clock should be > 0 after all mutations")
	}
}

// TestE2E_Validation_CompleteIncomplete verifies that completing an inspection
// with un-addressed items or missing descriptions returns 422.
func TestE2E_Validation_CompleteIncomplete(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	// Register & get token.
	var regOut identityapp.RegisterOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/register", "", map[string]any{
		"first_name":   "Bob",
		"last_name":    "Tester",
		"email":        "bob@juno.test",
		"password":     "password123",
		"company_name": "Bob's Inspections",
	}, &regOut, http.StatusCreated)
	token := regOut.Token

	// Create client and appointment.
	var clientOut identityapp.ClientView
	mustDo(t, client, http.MethodPost, base+"/api/v1/clients", token, map[string]any{
		"first_name": "Carol",
		"last_name":  "Client",
		"email":      "carol@client.test",
	}, &clientOut, http.StatusCreated)

	var apptOut schedulingapp.AppointmentView
	mustDo(t, client, http.MethodPost, base+"/api/v1/appointments", token, map[string]any{
		"client_id":    clientOut.ID,
		"street":       "1 Test Ave",
		"city":         "Chicago",
		"state":        "IL",
		"zip":          "60601",
		"scheduled_at": time.Now().Add(time.Hour).Unix(),
		"duration_min": 60,
	}, &apptOut, http.StatusCreated)

	// Start inspection.
	var inspOut inspectionapp.InspectionView
	mustDo(t, client, http.MethodPost, base+"/api/v1/inspections", token, map[string]any{
		"appointment_id": apptOut.ID,
	}, &inspOut, http.StatusCreated)
	inspID := inspOut.ID

	// Attempt to complete with no items addressed — expect 422.
	status := do(t, client,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/inspections/%s/complete", base, inspID),
		token, map[string]any{}, nil,
	)
	if status != http.StatusUnprocessableEntity {
		t.Errorf("complete (incomplete): status %d, want 422", status)
	}
}

// TestE2E_Conflict_DuplicateReport verifies that generating a second report
// for the same completed inspection returns 409.
func TestE2E_Conflict_DuplicateReport(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	// Register.
	var regOut identityapp.RegisterOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/register", "", map[string]any{
		"first_name":   "Eve",
		"last_name":    "Tester",
		"email":        "eve@juno.test",
		"password":     "password123",
		"company_name": "Eve's Inspections",
	}, &regOut, http.StatusCreated)
	token := regOut.Token

	// Create client + appointment.
	var clientOut identityapp.ClientView
	mustDo(t, client, http.MethodPost, base+"/api/v1/clients", token, map[string]any{
		"first_name": "Frank",
		"last_name":  "Buyer",
		"email":      "frank@buyer.test",
	}, &clientOut, http.StatusCreated)

	var apptOut schedulingapp.AppointmentView
	mustDo(t, client, http.MethodPost, base+"/api/v1/appointments", token, map[string]any{
		"client_id":    clientOut.ID,
		"street":       "2 Duplicate Rd",
		"city":         "Peoria",
		"state":        "IL",
		"zip":          "61602",
		"scheduled_at": time.Now().Add(time.Hour).Unix(),
		"duration_min": 60,
	}, &apptOut, http.StatusCreated)

	// Start inspection.
	var inspOut inspectionapp.InspectionView
	mustDo(t, client, http.MethodPost, base+"/api/v1/inspections", token, map[string]any{
		"appointment_id": apptOut.ID,
	}, &inspOut, http.StatusCreated)
	inspID := inspOut.ID

	// Complete inspection: address all items and descriptions.
	for _, sysDef := range inspectiondomain.Catalog {
		systemType := string(sysDef.Type)
		for _, itemDef := range sysDef.Items {
			mustDo(t, client, http.MethodPut,
				fmt.Sprintf("%s/api/v1/inspections/%s/systems/%s/items/%s/status",
					base, inspID, systemType, string(itemDef.Key)),
				token, map[string]any{"status": "I"}, nil, http.StatusOK,
			)
		}
		if len(sysDef.RequiredDescriptions) > 0 {
			descs := make(map[string]string)
			for _, req := range sysDef.RequiredDescriptions {
				descs[string(req.Key)] = "test"
			}
			mustDo(t, client, http.MethodPut,
				fmt.Sprintf("%s/api/v1/inspections/%s/systems/%s/descriptions",
					base, inspID, systemType),
				token, descs, nil, http.StatusOK,
			)
		}
	}
	mustDo(t, client, http.MethodPost,
		fmt.Sprintf("%s/api/v1/inspections/%s/complete", base, inspID),
		token, map[string]any{}, nil, http.StatusOK,
	)

	// Generate report once — should succeed.
	mustDo(t, client, http.MethodPost, base+"/api/v1/reports", token, map[string]any{
		"inspection_id": inspID,
	}, nil, http.StatusCreated)

	// Generate report again — should conflict.
	status := do(t, client, http.MethodPost, base+"/api/v1/reports", token, map[string]any{
		"inspection_id": inspID,
	}, nil)
	if status != http.StatusConflict {
		t.Errorf("duplicate report: status %d, want 409", status)
	}
}

// TestE2E_Auth_Unauthorized verifies protected routes reject missing tokens.
func TestE2E_Auth_Unauthorized(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	protectedRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/me"},
		{http.MethodGet, "/api/v1/clients"},
		{http.MethodGet, "/api/v1/appointments"},
		{http.MethodGet, "/api/v1/inspections"},
		{http.MethodGet, "/api/v1/reports"},
	}

	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			status := do(t, client, route.method, base+route.path, "", nil, nil)
			if status != http.StatusUnauthorized {
				t.Errorf("status %d, want 401", status)
			}
		})
	}
}

// TestE2E_AppointmentCRUD verifies appointment update and cancel operations.
func TestE2E_AppointmentCRUD(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	// Register.
	var regOut identityapp.RegisterOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/register", "", map[string]any{
		"first_name":   "Grace",
		"last_name":    "Tester",
		"email":        "grace@juno.test",
		"password":     "password123",
		"company_name": "Grace Inspects",
	}, &regOut, http.StatusCreated)
	token := regOut.Token

	// Create client.
	var clientOut identityapp.ClientView
	mustDo(t, client, http.MethodPost, base+"/api/v1/clients", token, map[string]any{
		"first_name": "Henry",
		"last_name":  "Homebuyer",
		"email":      "henry@buyer.test",
	}, &clientOut, http.StatusCreated)

	// Create appointment.
	scheduledAt := time.Now().Add(48 * time.Hour).Unix()
	var apptOut schedulingapp.AppointmentView
	mustDo(t, client, http.MethodPost, base+"/api/v1/appointments", token, map[string]any{
		"client_id":    clientOut.ID,
		"street":       "10 Oak Ln",
		"city":         "Decatur",
		"state":        "IL",
		"zip":          "62521",
		"scheduled_at": scheduledAt,
		"duration_min": 90,
	}, &apptOut, http.StatusCreated)

	// Update the appointment.
	var updatedAppt schedulingapp.AppointmentView
	mustDo(t, client,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/appointments/%s", base, apptOut.ID),
		token,
		map[string]any{
			"duration_min": 120,
			"notes":        "Updated notes.",
		},
		&updatedAppt,
		http.StatusOK,
	)
	if updatedAppt.DurationMin != 120 {
		t.Errorf("updated duration = %d, want 120", updatedAppt.DurationMin)
	}

	// List appointments — should include the updated one.
	var apptList []schedulingapp.AppointmentView
	mustDo(t, client, http.MethodGet, base+"/api/v1/appointments", token, nil, &apptList, http.StatusOK)
	if len(apptList) != 1 {
		t.Errorf("list appointments: got %d, want 1", len(apptList))
	}

	// Cancel the appointment.
	mustDo(t, client,
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/appointments/%s", base, apptOut.ID),
		token, nil, nil, http.StatusNoContent,
	)

	// GET should now show cancelled status.
	var cancelledAppt schedulingapp.AppointmentView
	mustDo(t, client,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/appointments/%s", base, apptOut.ID),
		token, nil, &cancelledAppt, http.StatusOK,
	)
	if cancelledAppt.Status != "cancelled" {
		t.Errorf("cancelled appointment status = %q, want cancelled", cancelledAppt.Status)
	}
}

// TestE2E_ClientCRUD verifies client update and delete operations.
func TestE2E_ClientCRUD(t *testing.T) {
	ts := e2eStack(t)
	client := ts.Client()
	base := ts.URL

	// Register.
	var regOut identityapp.RegisterOutput
	mustDo(t, client, http.MethodPost, base+"/api/v1/auth/register", "", map[string]any{
		"first_name":   "Iris",
		"last_name":    "Tester",
		"email":        "iris@juno.test",
		"password":     "password123",
		"company_name": "Iris Inspects",
	}, &regOut, http.StatusCreated)
	token := regOut.Token

	// Create client.
	var clientOut identityapp.ClientView
	mustDo(t, client, http.MethodPost, base+"/api/v1/clients", token, map[string]any{
		"first_name": "Jake",
		"last_name":  "Jones",
		"email":      "jake@jones.test",
		"phone":      "555-0200",
	}, &clientOut, http.StatusCreated)

	// Update client.
	var updatedClient identityapp.ClientView
	mustDo(t, client, http.MethodPut,
		fmt.Sprintf("%s/api/v1/clients/%s", base, clientOut.ID),
		token,
		map[string]any{"first_name": "Jacob", "last_name": "Jones", "email": "jake@jones.test"},
		&updatedClient, http.StatusOK,
	)
	if updatedClient.FirstName != "Jacob" {
		t.Errorf("updated client first_name = %q, want Jacob", updatedClient.FirstName)
	}

	// Delete client.
	mustDo(t, client, http.MethodDelete,
		fmt.Sprintf("%s/api/v1/clients/%s", base, clientOut.ID),
		token, nil, nil, http.StatusNoContent,
	)

	// GET should 404.
	status := do(t, client, http.MethodGet,
		fmt.Sprintf("%s/api/v1/clients/%s", base, clientOut.ID),
		token, nil, nil,
	)
	if status != http.StatusNotFound {
		t.Errorf("GET deleted client: status %d, want 404", status)
	}
}
