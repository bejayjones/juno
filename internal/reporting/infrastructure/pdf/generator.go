// Package pdf provides the PDF generator for the reporting context.
// It uses github.com/go-pdf/fpdf to produce InterNACHI-style inspection reports.
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	inspectiondomain "github.com/bejayjones/juno/internal/inspection/domain"
)

// Generator implements application.PDFGenerator using fpdf.
type Generator struct{}

func NewGenerator() *Generator { return &Generator{} }

// Generate builds the inspection report PDF and writes it to outputPath.
func (g *Generator) Generate(_ context.Context, insp *inspectiondomain.Inspection, outputPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	g.addCoverPage(pdf, insp)
	g.addDeficiencySummary(pdf, insp)
	g.addSystemSections(pdf, insp)
	g.addLimitations(pdf)

	if pdf.Err() {
		return fmt.Errorf("pdf generation error: %w", pdf.Error())
	}
	return pdf.OutputFileAndClose(outputPath)
}

// avail returns the usable page width (A4 − left − right margin).
func avail(pdf *fpdf.Fpdf) float64 {
	w, _ := pdf.GetPageSize()
	return w - 30
}

// pageH returns the usable page height (A4 − top − bottom margin).
func pageH(pdf *fpdf.Fpdf) float64 {
	_, h := pdf.GetPageSize()
	return h - 30
}

// ── Cover page ────────────────────────────────────────────────────────────────

func (g *Generator) addCoverPage(pdf *fpdf.Fpdf, insp *inspectiondomain.Inspection) {
	pdf.AddPage()
	w := avail(pdf)

	// Title.
	pdf.SetFont("Helvetica", "B", 22)
	pdf.CellFormat(w, 14, "HOME INSPECTION REPORT", "", 1, "C", false, 0, "")
	pw, _ := pdf.GetPageSize()
	pdf.SetDrawColor(80, 80, 80)
	pdf.Line(15, pdf.GetY(), pw-15, pdf.GetY())
	pdf.Ln(6)

	inspDate := insp.StartedAt.Format("January 2, 2006")
	if insp.CompletedAt != nil {
		inspDate = insp.CompletedAt.Format("January 2, 2006")
	}

	rows := []struct{ label, value string }{
		{"Inspection ID", string(insp.ID)},
		{"Inspector ID", string(insp.InspectorID)},
		{"Date", inspDate},
		{"Structure Type", insp.Header.StructureType},
		{"Year Built", fmt.Sprintf("%d", insp.Header.YearBuilt)},
		{"Weather Conditions", insp.Header.Weather},
		{"Temperature", fmt.Sprintf("%d deg F", insp.Header.TemperatureF)},
		{"Attendees", strings.Join(insp.Header.Attendees, ", ")},
	}
	for _, row := range rows {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(55, 7, row.label+":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(w-55, 7, row.value, "", 1, "L", false, 0, "")
	}

	// Status banner.
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(w, 8, fmt.Sprintf("Status: %s", strings.ToUpper(string(insp.Status))), "1", 1, "C", true, 0, "")

	// Deficiency count.
	deficiencies := insp.Deficiencies()
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 11)
	if len(deficiencies) == 0 {
		pdf.SetTextColor(0, 120, 0)
		pdf.CellFormat(w, 7, "No deficiencies identified.", "", 1, "C", false, 0, "")
	} else {
		pdf.SetTextColor(180, 0, 0)
		pdf.CellFormat(w, 7,
			fmt.Sprintf("%d deficien%s identified — see Deficiency Summary.",
				len(deficiencies), plural(len(deficiencies), "cy", "cies")),
			"", 1, "C", false, 0, "")
	}
	pdf.SetTextColor(0, 0, 0)

	pdf.Ln(8)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(w, 6, "Report generated: "+time.Now().Format("January 2, 2006 15:04 MST"), "", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// ── Deficiency summary ────────────────────────────────────────────────────────

func (g *Generator) addDeficiencySummary(pdf *fpdf.Fpdf, insp *inspectiondomain.Inspection) {
	deficiencies := insp.Deficiencies()
	if len(deficiencies) == 0 {
		return
	}

	pdf.AddPage()
	w := avail(pdf)

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(w, 10, "DEFICIENCY SUMMARY", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Table header.
	pdf.SetFillColor(50, 50, 50)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(42, 7, "System", "1", 0, "L", true, 0, "")
	pdf.CellFormat(52, 7, "Item", "1", 0, "L", true, 0, "")
	pdf.CellFormat(w-94, 7, "Finding", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Helvetica", "", 10)
	for i, d := range deficiencies {
		if i%2 == 0 {
			pdf.SetFillColor(255, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		narrative := d.Narrative
		if len(narrative) > 80 {
			narrative = narrative[:77] + "..."
		}
		pdf.CellFormat(42, 6, d.SystemLabel, "1", 0, "L", true, 0, "")
		pdf.CellFormat(52, 6, d.ItemLabel, "1", 0, "L", true, 0, "")
		pdf.CellFormat(w-94, 6, narrative, "1", 1, "L", true, 0, "")
	}
}

// ── System sections ───────────────────────────────────────────────────────────

func (g *Generator) addSystemSections(pdf *fpdf.Fpdf, insp *inspectiondomain.Inspection) {
	for _, sysDef := range inspectiondomain.Catalog {
		section, ok := insp.Systems[sysDef.Type]
		if !ok {
			continue
		}
		g.addSystemSection(pdf, section, sysDef)
	}
}

func (g *Generator) addSystemSection(pdf *fpdf.Fpdf, section *inspectiondomain.SystemSection, sysDef *inspectiondomain.SystemDefinition) {
	pdf.AddPage()
	w := avail(pdf)

	// Section header banner.
	pdf.SetFillColor(30, 60, 120)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(w, 10, "  "+sysDef.Label, "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)

	// Descriptions.
	if len(section.Descriptions) > 0 {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(w, 7, "Descriptions", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		for k, v := range section.Descriptions {
			pdf.CellFormat(65, 6, "  "+string(k)+":", "", 0, "L", false, 0, "")
			pdf.MultiCell(w-65, 6, v, "", "L", false)
		}
		pdf.Ln(2)
	}

	// Inspector notes.
	if section.InspectorNotes != "" {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetFillColor(250, 250, 220)
		pdf.MultiCell(w, 6, "Inspector Notes: "+section.InspectorNotes, "1", "L", true)
		pdf.Ln(3)
		pdf.SetFillColor(255, 255, 255)
	}

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(w, 7, "Inspection Items", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	for _, item := range section.Items {
		g.addItem(pdf, item)
	}
}

func (g *Generator) addItem(pdf *fpdf.Fpdf, item inspectiondomain.InspectionItem) {
	w := avail(pdf)

	statusBadge := map[inspectiondomain.ItemStatus]string{
		inspectiondomain.StatusInspected:    "[I] ",
		inspectiondomain.StatusNotInspected: "[NI]",
		inspectiondomain.StatusNotPresent:   "[NP]",
		inspectiondomain.StatusDeficient:    "[D] ",
	}[item.Status]

	switch item.Status {
	case inspectiondomain.StatusDeficient:
		pdf.SetTextColor(180, 0, 0)
		pdf.SetFont("Helvetica", "B", 10)
	case inspectiondomain.StatusNotPresent, inspectiondomain.StatusNotInspected:
		pdf.SetTextColor(120, 120, 120)
		pdf.SetFont("Helvetica", "", 10)
	default:
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Helvetica", "", 10)
	}

	pdf.CellFormat(12, 6, statusBadge, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(w-12, 6, item.Label, "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	if item.NotInspectedReason != "" {
		pdf.SetFont("Helvetica", "I", 9)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(12, 5, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(w-12, 5, "  Reason: "+item.NotInspectedReason, "", 1, "L", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	for _, f := range item.Findings {
		g.addFinding(pdf, f)
	}
	pdf.Ln(1)
}

func (g *Generator) addFinding(pdf *fpdf.Fpdf, f inspectiondomain.Finding) {
	w := avail(pdf)

	if f.IsDeficiency {
		pdf.SetTextColor(180, 0, 0)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(10, 5, "", "", 0, "L", false, 0, "")
		pdf.MultiCell(w-10, 5, "[DEFICIENCY] "+f.Narrative, "", "L", false)
	} else {
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(10, 5, "", "", 0, "L", false, 0, "")
		pdf.MultiCell(w-10, 5, "- "+f.Narrative, "", "L", false)
	}
	pdf.SetTextColor(0, 0, 0)

	for _, photo := range f.Photos {
		g.addPhoto(pdf, photo)
	}
}

func (g *Generator) addPhoto(pdf *fpdf.Fpdf, photo inspectiondomain.PhotoRef) {
	imgType := imageTypeFromPath(photo.StoragePath)
	if imgType == "" {
		return
	}

	data, err := os.ReadFile(photo.StoragePath)
	if err != nil {
		return // file missing or unreadable; skip silently
	}

	if pdf.GetY() > pageH(pdf)-75 {
		pdf.AddPage()
	}

	opts := fpdf.ImageOptions{ImageType: imgType}
	reader := bytes.NewReader(data)
	name := string(photo.ID)

	pdf.RegisterImageOptionsReader(name, opts, reader)
	if pdf.Err() {
		pdf.ClearError() // image could not be parsed; skip and continue
		return
	}

	pdf.ImageOptions(name, pdf.GetX()+10, pdf.GetY(), 75, 55, true, opts, 0, "")
	pdf.Ln(2)
}

// ── Limitations ───────────────────────────────────────────────────────────────

func (g *Generator) addLimitations(pdf *fpdf.Fpdf) {
	pdf.AddPage()
	w := avail(pdf)

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(w, 10, "LIMITATIONS AND SCOPE", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "", 10)
	items := []string{
		"This inspection was performed in accordance with the InterNACHI Standards of Practice.",
		"This report is not a guarantee or warranty of any kind, expressed or implied.",
		"The inspection was limited to visually accessible areas and components of the property.",
		"Concealed, inaccessible, or obstructed areas were not inspected and are excluded from this report.",
		"This report reflects the condition of the property at the time of inspection only.",
		"Latent defects or conditions that develop after the inspection date are beyond the scope of this report.",
		"This report is provided for the exclusive use of the client named above and is not transferable.",
		"The inspector is not required to determine the cause of deficiencies, predict future conditions, or report on non-deficient items.",
	}
	for _, text := range items {
		pdf.CellFormat(6, 6, "o", "", 0, "L", false, 0, "")
		pdf.MultiCell(w-6, 6, text, "", "L", false)
		pdf.Ln(1)
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(w, 7, "InterNACHI Standards of Practice Reference", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(60, 60, 120)
	pdf.CellFormat(w, 6, "https://www.nachi.org/sop.htm", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func imageTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "JPG"
	case ".png":
		return "PNG"
	default:
		return ""
	}
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
