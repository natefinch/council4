// Command corpusdelta compiles the Oracle corpus and prepares review artifacts
// for a compiler expansion.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen"
)

// Config describes one corpus comparison workflow.
type Config struct {
	CorpusPath     string
	BaselineReport string
	CurrentReport  string
	GeneratedRoot  string
	SupportedPath  string
	ManifestPath   string
	Compile        bool
	Validate       bool
}

// Engine runs corpus compilation, comparison, documentation, and validation.
type Engine struct {
	Config Config
	Runner func(name string, args ...string) error
}

// Manifest is the deterministic review packet emitted by Engine.
type Manifest struct {
	CardCount              int                    `json:"card_count"`
	BaselineGeneratedCount int                    `json:"baseline_generated_count"`
	CurrentGeneratedCount  int                    `json:"current_generated_count"`
	GeneratedDelta         int                    `json:"generated_delta"`
	NewlySupported         []InspectionCard       `json:"newly_supported"`
	NewlyUnsupported       []InspectionCard       `json:"newly_unsupported"`
	ChangedDiagnostics     []DiagnosticChangeCard `json:"changed_diagnostics"`
	DiagnosticChanges      []DiagnosticChange     `json:"diagnostic_changes"`
	GeneratedValidated     bool                   `json:"generated_validated"`
}

// InspectionCard contains the corpus and source details needed to review a delta.
type InspectionCard struct {
	ID            string             `json:"id"`
	OracleID      string             `json:"oracle_id,omitempty"`
	Name          string             `json:"name"`
	Layout        string             `json:"layout,omitempty"`
	TypeLine      string             `json:"type_line,omitempty"`
	OracleText    string             `json:"oracle_text,omitempty"`
	Faces         []InspectionFace   `json:"faces,omitempty"`
	GeneratedPath string             `json:"generated_path,omitempty"`
	Diagnostics   []ReportDiagnostic `json:"diagnostics,omitempty"`
}

// InspectionFace preserves review-relevant fields from a multi-face card.
type InspectionFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line,omitempty"`
	OracleText string `json:"oracle_text,omitempty"`
}

// DiagnosticChange reports how often a diagnostic summary occurs before and after.
type DiagnosticChange struct {
	Summary  string `json:"summary"`
	Baseline int    `json:"baseline"`
	Current  int    `json:"current"`
	Delta    int    `json:"delta"`
}

// DiagnosticChangeCard records a diagnostic transition for a card that remains unsupported.
type DiagnosticChangeCard struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Before []ReportDiagnostic `json:"before"`
	After  []ReportDiagnostic `json:"after"`
}

type compileReport struct {
	CardCount        int                 `json:"card_count"`
	GeneratedCount   int                 `json:"generated_count"`
	UnsupportedCount int                 `json:"unsupported_count"`
	Unsupported      []unsupportedReport `json:"unsupported"`
}

type unsupportedReport struct {
	ID          string             `json:"id"`
	OracleID    string             `json:"oracle_id,omitempty"`
	Name        string             `json:"name"`
	Layout      string             `json:"layout,omitempty"`
	Diagnostics []ReportDiagnostic `json:"diagnostics"`
}

// ReportDiagnostic is the review-relevant portion of a compiler diagnostic.
type ReportDiagnostic struct {
	Severity string `json:"severity,omitempty"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail,omitempty"`
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	engine := Engine{Config: cfg}
	if err := engine.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func parseFlags(args []string) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("corpusdelta", flag.ContinueOnError)
	flags.StringVar(&cfg.CorpusPath, "in", "", "Scryfall Oracle Cards bulk-data JSON file")
	flags.StringVar(&cfg.BaselineReport, "baseline", "", "previous compilecards JSON report")
	flags.StringVar(&cfg.CurrentReport, "report", ".cardwork/current-report.json", "current compilecards JSON report")
	flags.StringVar(&cfg.GeneratedRoot, "out", ".cardwork/current-generated", "generated cards root")
	flags.StringVar(&cfg.SupportedPath, "supported", "docs/supported.md", "supported-card Markdown path")
	flags.StringVar(&cfg.ManifestPath, "manifest", ".cardwork/current-delta.json", "inspection manifest JSON path")
	flags.BoolVar(&cfg.Compile, "compile", true, "run compilecards before comparison")
	flags.BoolVar(&cfg.Validate, "validate", true, "run go test and go vet on generated packages")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.CorpusPath == "" {
		return Config{}, errors.New("-in is required")
	}
	if cfg.BaselineReport == "" {
		return Config{}, errors.New("-baseline is required")
	}
	return cfg, nil
}

// Run executes the configured workflow.
func (e *Engine) Run() error {
	cfg := e.Config
	if cfg.CorpusPath == "" || cfg.BaselineReport == "" || cfg.CurrentReport == "" ||
		cfg.GeneratedRoot == "" || cfg.SupportedPath == "" || cfg.ManifestPath == "" {
		return errors.New("corpus, reports, generated root, supported list, and manifest paths are required")
	}
	if e.Runner == nil {
		e.Runner = runCommand
	}
	if cfg.Compile {
		if err := safeToReplace(cfg.GeneratedRoot); err != nil {
			return err
		}
		if err := os.RemoveAll(cfg.GeneratedRoot); err != nil {
			return fmt.Errorf("removing generated root: %w", err)
		}
		if err := e.Runner(
			"go", "run", "./cardgen/oracle/cmd/compilecards",
			"-in", cfg.CorpusPath,
			"-out", cfg.GeneratedRoot,
			"-report", cfg.CurrentReport,
		); err != nil {
			return fmt.Errorf("compiling corpus: %w", err)
		}
	}

	cards, err := readCorpus(cfg.CorpusPath)
	if err != nil {
		return err
	}
	baseline, err := readReport(cfg.BaselineReport)
	if err != nil {
		return fmt.Errorf("reading baseline: %w", err)
	}
	current, err := readReport(cfg.CurrentReport)
	if err != nil {
		return fmt.Errorf("reading current report: %w", err)
	}
	if err := validateInputs(cards, baseline, current); err != nil {
		return err
	}

	manifest, supported, err := buildManifest(cards, baseline, current, cfg.GeneratedRoot)
	if err != nil {
		return err
	}
	if err := writeSupported(cfg.SupportedPath, supported); err != nil {
		return err
	}
	if cfg.Validate {
		if err := e.validateGenerated(); err != nil {
			return err
		}
		manifest.GeneratedValidated = true
	}
	if err := writeJSON(cfg.ManifestPath, manifest); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(
		os.Stdout,
		"Cards supported: %s / %s (%+d)\nNewly supported: %d\nNewly unsupported: %d\nManifest: %s\n",
		formatCount(current.GeneratedCount),
		formatCount(current.CardCount),
		manifest.GeneratedDelta,
		len(manifest.NewlySupported),
		len(manifest.NewlyUnsupported),
		cfg.ManifestPath,
	)
	return nil
}

func safeToReplace(root string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolving generated root: %w", err)
	}
	absoluteWork, err := filepath.Abs(".cardwork")
	if err != nil {
		return fmt.Errorf("resolving .cardwork: %w", err)
	}
	relative, err := filepath.Rel(absoluteWork, absoluteRoot)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing to replace generated root outside .cardwork: %s", root)
	}
	return nil
}

func readCorpus(path string) (map[string]cardgen.ScryfallCard, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening corpus: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading corpus: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("corpus must be a JSON array")
	}
	cards := make(map[string]cardgen.ScryfallCard)
	for decoder.More() {
		var card cardgen.ScryfallCard
		if err := decoder.Decode(&card); err != nil {
			return nil, fmt.Errorf("decoding corpus card: %w", err)
		}
		if card.ID == "" {
			return nil, fmt.Errorf("corpus card %q has no stable ID", card.Name)
		}
		if _, exists := cards[card.ID]; exists {
			return nil, fmt.Errorf("corpus contains duplicate stable ID %q", card.ID)
		}
		cards[card.ID] = card
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("closing corpus array: %w", err)
	}
	return cards, nil
}

func readReport(path string) (compileReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return compileReport{}, err
	}
	defer file.Close()
	var report compileReport
	if err := json.NewDecoder(file).Decode(&report); err != nil {
		return compileReport{}, err
	}
	return report, nil
}

func validateInputs(cards map[string]cardgen.ScryfallCard, baseline, current compileReport) error {
	if baseline.CardCount != len(cards) || current.CardCount != len(cards) {
		return fmt.Errorf(
			"corpus/report card counts differ: corpus=%d baseline=%d current=%d",
			len(cards), baseline.CardCount, current.CardCount,
		)
	}
	for name, report := range map[string]compileReport{"baseline": baseline, "current": current} {
		if report.UnsupportedCount != len(report.Unsupported) ||
			report.GeneratedCount+report.UnsupportedCount != report.CardCount {
			return fmt.Errorf("%s report counts are inconsistent", name)
		}
		seen := make(map[string]bool)
		for _, card := range report.Unsupported {
			if _, ok := cards[card.ID]; !ok {
				return fmt.Errorf("%s report contains unknown stable ID %q", name, card.ID)
			}
			if seen[card.ID] {
				return fmt.Errorf("%s report repeats stable ID %q", name, card.ID)
			}
			seen[card.ID] = true
		}
	}
	return nil
}

func buildManifest(
	cards map[string]cardgen.ScryfallCard,
	baseline, current compileReport,
	generatedRoot string,
) (Manifest, []cardgen.ScryfallCard, error) {
	baselineUnsupported := unsupportedByID(baseline.Unsupported)
	currentUnsupported := unsupportedByID(current.Unsupported)
	manifest := Manifest{
		CardCount:              current.CardCount,
		BaselineGeneratedCount: baseline.GeneratedCount,
		CurrentGeneratedCount:  current.GeneratedCount,
		GeneratedDelta:         current.GeneratedCount - baseline.GeneratedCount,
		DiagnosticChanges:      diagnosticChanges(baseline.Unsupported, current.Unsupported),
	}
	supported := make([]cardgen.ScryfallCard, 0, current.GeneratedCount)
	for id, card := range cards {
		currentFailure, unsupported := currentUnsupported[id]
		if !unsupported {
			supported = append(supported, card)
		}
		if _, wasUnsupported := baselineUnsupported[id]; wasUnsupported && !unsupported {
			inspection := inspectCard(card)
			inspection.GeneratedPath = filepath.Join(
				generatedRoot,
				cardgen.CardNameToPackageLetter(card.Name),
				cardgen.CardNameToSafeFileName(card.Name)+".go",
			)
			if _, err := os.Stat(inspection.GeneratedPath); err != nil {
				return Manifest{}, nil, fmt.Errorf("generated source for %s: %w", card.Name, err)
			}
			manifest.NewlySupported = append(manifest.NewlySupported, inspection)
		}
		if _, wasUnsupported := baselineUnsupported[id]; !wasUnsupported && unsupported {
			inspection := inspectCard(card)
			inspection.Diagnostics = currentFailure.Diagnostics
			manifest.NewlyUnsupported = append(manifest.NewlyUnsupported, inspection)
		}
		if baselineFailure, wasUnsupported := baselineUnsupported[id]; wasUnsupported && unsupported &&
			!slices.Equal(baselineFailure.Diagnostics, currentFailure.Diagnostics) {
			manifest.ChangedDiagnostics = append(manifest.ChangedDiagnostics, DiagnosticChangeCard{
				ID:     id,
				Name:   card.Name,
				Before: baselineFailure.Diagnostics,
				After:  currentFailure.Diagnostics,
			})
		}
	}
	sortCards(supported)
	slices.SortFunc(manifest.NewlySupported, compareInspectionCards)
	slices.SortFunc(manifest.NewlyUnsupported, compareInspectionCards)
	slices.SortFunc(manifest.ChangedDiagnostics, func(a, b DiagnosticChangeCard) int {
		if byName := strings.Compare(a.Name, b.Name); byName != 0 {
			return byName
		}
		return strings.Compare(a.ID, b.ID)
	})
	if len(supported) != current.GeneratedCount {
		return Manifest{}, nil, fmt.Errorf(
			"derived supported count %d differs from report count %d",
			len(supported), current.GeneratedCount,
		)
	}
	return manifest, supported, nil
}

func unsupportedByID(cards []unsupportedReport) map[string]unsupportedReport {
	output := make(map[string]unsupportedReport, len(cards))
	for _, card := range cards {
		output[card.ID] = card
	}
	return output
}

func inspectCard(card cardgen.ScryfallCard) InspectionCard {
	output := InspectionCard{
		ID:         card.ID,
		OracleID:   card.OracleID,
		Name:       card.Name,
		Layout:     card.Layout,
		TypeLine:   card.TypeLine,
		OracleText: card.OracleText,
	}
	for _, face := range card.CardFaces {
		output.Faces = append(output.Faces, InspectionFace{
			Name:       face.Name,
			TypeLine:   face.TypeLine,
			OracleText: face.OracleText,
		})
	}
	return output
}

func diagnosticChanges(baseline, current []unsupportedReport) []DiagnosticChange {
	before := diagnosticCounts(baseline)
	after := diagnosticCounts(current)
	summaries := make(map[string]bool)
	for summary := range before {
		summaries[summary] = true
	}
	for summary := range after {
		summaries[summary] = true
	}
	var output []DiagnosticChange
	for summary := range summaries {
		if before[summary] == after[summary] {
			continue
		}
		output = append(output, DiagnosticChange{
			Summary:  summary,
			Baseline: before[summary],
			Current:  after[summary],
			Delta:    after[summary] - before[summary],
		})
	}
	slices.SortFunc(output, func(a, b DiagnosticChange) int {
		return strings.Compare(a.Summary, b.Summary)
	})
	return output
}

func diagnosticCounts(cards []unsupportedReport) map[string]int {
	counts := make(map[string]int)
	for _, card := range cards {
		for _, diagnostic := range card.Diagnostics {
			counts[diagnostic.Summary]++
		}
	}
	return counts
}

func sortCards(cards []cardgen.ScryfallCard) {
	slices.SortFunc(cards, func(a, b cardgen.ScryfallCard) int {
		if byName := strings.Compare(a.Name, b.Name); byName != 0 {
			return byName
		}
		return strings.Compare(a.ID, b.ID)
	})
}

func compareInspectionCards(a, b InspectionCard) int {
	if byName := strings.Compare(a.Name, b.Name); byName != 0 {
		return byName
	}
	return strings.Compare(a.ID, b.ID)
}

func writeSupported(path string, cards []cardgen.ScryfallCard) error {
	var builder strings.Builder
	_, _ = fmt.Fprintf(&builder, "# Supported Cards\n\nCards supported: %s\n\n", formatCount(len(cards)))
	for _, card := range cards {
		_, _ = fmt.Fprintf(&builder, "- %s\n", card.Name)
	}
	return writeFile(path, []byte(builder.String()))
}

func formatCount(count int) string {
	digits := fmt.Sprintf("%d", count)
	for i := len(digits) - 3; i > 0; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	data = append(data, '\n')
	return writeFile(path, data)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", path, err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("creating temporary file for %s: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		closeErr := temporary.Close()
		return fmt.Errorf("writing %s: %w", path, errors.Join(err, closeErr))
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

func (e *Engine) validateGenerated() error {
	temporary, err := os.MkdirTemp("cardgen", "corpusdelta")
	if err != nil {
		return fmt.Errorf("creating generated validation package: %w", err)
	}
	defer os.RemoveAll(temporary)
	if err := os.CopyFS(temporary, os.DirFS(e.Config.GeneratedRoot)); err != nil {
		return fmt.Errorf("copying generated packages: %w", err)
	}
	packagePattern := "./" + filepath.ToSlash(temporary) + "/..."
	if err := e.Runner("go", "test", packagePattern, "-count=1"); err != nil {
		return fmt.Errorf("testing generated packages: %w", err)
	}
	if err := e.Runner("go", "vet", packagePattern); err != nil {
		return fmt.Errorf("vetting generated packages: %w", err)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	command := exec.CommandContext(context.Background(), name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, output)
	}
	return nil
}
