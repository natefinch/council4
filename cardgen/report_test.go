package cardgen

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestBuildUnsupportedReportIncludesProblemRows(t *testing.T) {
	manifest := Manifest{Version: ManifestVersion, Cards: []ManifestCard{
		{CanonicalName: "Good Card", Quantity: 1, Status: BatchStatusFetched, FileStatus: BatchFileStatusExisting, Validation: BatchValidationStatusValid},
		{InputName: "Bad Fetch", Quantity: 1, Status: BatchStatusFetchError, FetchError: "unsupported layout"},
		{CanonicalName: "Missing Card", Quantity: 1, Status: BatchStatusFetched, FileStatus: BatchFileStatusMissing, FilePath: "mtg/cards/m/missing_card.go"},
		{CanonicalName: "Invalid Card", Quantity: 1, Status: BatchStatusFetched, FileStatus: BatchFileStatusExisting, Validation: BatchValidationStatusInvalid, Issues: []ValidationIssue{{Code: IssueOracleWithoutAbilities, Message: "unfinished"}}},
	}}

	report := BuildUnsupportedReport(manifest)

	if report.Summary.ManifestTotal != 4 || report.Summary.UnsupportedTotal != 3 || report.Summary.FetchError != 1 || report.Summary.MissingGeneratedFile != 1 || report.Summary.Invalid != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if len(report.Cards) != 3 {
		t.Fatalf("len(report.Cards) = %d, want 3", len(report.Cards))
	}
	for _, card := range report.Cards {
		if card.Name == "Good Card" {
			t.Fatalf("valid card included in report: %+v", card)
		}
		if len(card.NextWork) == 0 {
			t.Fatalf("card missing next work: %+v", card)
		}
	}
}

func TestWriteUnsupportedReportMarkdown(t *testing.T) {
	report := BuildUnsupportedReport(Manifest{Version: ManifestVersion, Cards: []ManifestCard{{
		CanonicalName: "Invalid Card",
		Quantity:      1,
		FirstLine:     7,
		Section:       "Main",
		FilePath:      "mtg/cards/i/invalid_card.go",
		OracleText:    "Draw a card.\nDiscard a card.",
		Validation:    BatchValidationStatusInvalid,
		Issues:        []ValidationIssue{{Code: IssueOracleWithoutAbilities, Message: "unfinished"}},
	}}})

	var buf bytes.Buffer
	if err := WriteUnsupportedReportMarkdown(&buf, report); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	for _, want := range []string{
		"# Unsupported card report",
		"## Invalid Card",
		"`oracle-without-abilities`",
		"> Draw a card.\n> Discard a card.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("markdown missing %q:\n%s", want, output)
		}
	}
}

func TestWriteUnsupportedReportJSON(t *testing.T) {
	report := BuildUnsupportedReport(Manifest{Version: ManifestVersion, Cards: []ManifestCard{{
		CanonicalName: "Missing Card",
		Quantity:      1,
		FileStatus:    BatchFileStatusMissing,
	}}})

	var buf bytes.Buffer
	if err := WriteUnsupportedReportJSON(&buf, report); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"missing_generated_file": 1`) {
		t.Fatalf("json report = %s", buf.String())
	}
}

func TestBuildUnsupportedReportIncludesValidationPending(t *testing.T) {
	report := BuildUnsupportedReport(Manifest{Version: ManifestVersion, Cards: []ManifestCard{{
		CanonicalName: "Pending Card",
		Quantity:      1,
		FileStatus:    BatchFileStatusExisting,
		Validation:    BatchValidationStatusUnvalidated,
	}}})

	if report.Summary.ValidationPending != 1 || len(report.Cards) != 1 {
		t.Fatalf("report = %+v", report)
	}
	if !strings.Contains(report.Cards[0].NextWork[0], "validate") {
		t.Fatalf("next work = %+v", report.Cards[0].NextWork)
	}
}

func TestBuildUnsupportedReportWithSourceIncludesMissingFunctionality(t *testing.T) {
	repoRoot := t.TempDir()
	cardPath := filepath.Join("mtg", "cards", "b", "blocked_card.go")
	if err := os.MkdirAll(filepath.Join(repoRoot, filepath.Dir(cardPath)), 0o755); err != nil {
		t.Fatal(err)
	}
	source := `package b

// Blocked Card
//
// Missing primitives:
//   - EffectSelectorEquippedCreature does not exist; equipped creature cannot
//     be selected declaratively.
//   - DynamicAmountCountOpponents does not exist.
var BlockedCard = nil
`
	if err := os.WriteFile(filepath.Join(repoRoot, cardPath), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{Version: ManifestVersion, Cards: []ManifestCard{{
		CanonicalName: "Blocked Card",
		Quantity:      1,
		FileStatus:    BatchFileStatusExisting,
		Validation:    BatchValidationStatusValid,
		FilePath:      cardPath,
	}}}

	report := BuildUnsupportedReportWithSource(manifest, repoRoot)

	if report.Summary.UnsupportedTotal != 1 || report.Summary.FunctionalityBlocked != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if got := report.Cards[0].Functionality; len(got) != 2 {
		t.Fatalf("functionality = %+v", got)
	}
	if len(report.Functionality) != 2 {
		t.Fatalf("rollup = %+v", report.Functionality)
	}
	var buf bytes.Buffer
	if err := WriteUnsupportedReportMarkdown(&buf, report); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	for _, want := range []string{
		"functionality-blocked cards: 1",
		"- Missing functionality:",
		"## Missing functionality rollup",
		"### EffectSelectorEquippedCreature",
		"- Cards: Blocked Card",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("markdown missing %q:\n%s", want, output)
		}
	}
}

func TestCardFunctionalityPlan1UnsupportedBaseline(t *testing.T) {
	repoRoot := t.TempDir()
	manifest := cardFunctionalityPlan1Manifest()
	writePlan1SourceStubs(t, repoRoot, manifest)

	report := BuildUnsupportedReportWithSource(manifest, repoRoot)
	got := snapshotUnsupportedReport(report)
	gotJSON := marshalSnapshot(t, got)
	want, err := os.ReadFile(filepath.Join("testdata", "card_functionality_plan_1_unsupported.json"))
	if err != nil {
		t.Fatal(err)
	}
	wantJSON := normalizeSnapshotJSON(t, want)
	if gotJSON != wantJSON {
		t.Fatalf("unsupported baseline mismatch\nwant:\n%s\n\ngot:\n%s", wantJSON, gotJSON)
	}
}

type unsupportedReportSnapshot struct {
	Summary       UnsupportedReportSummary        `json:"summary"`
	Cards         []unsupportedCardSnapshot       `json:"cards"`
	Functionality []unsupportedFunctionalityEntry `json:"functionality"`
}

type unsupportedCardSnapshot struct {
	Name          string   `json:"name"`
	Issues        []string `json:"issues,omitempty"`
	Functionality []string `json:"functionality,omitempty"`
}

type unsupportedFunctionalityEntry struct {
	Capability string   `json:"capability"`
	Cards      []string `json:"cards"`
}

func snapshotUnsupportedReport(report UnsupportedReport) unsupportedReportSnapshot {
	snapshot := unsupportedReportSnapshot{
		Summary: report.Summary,
	}
	for _, card := range report.Cards {
		snapshot.Cards = append(snapshot.Cards, unsupportedCardSnapshot{
			Name:          card.Name,
			Issues:        issueCodes(card.Issues),
			Functionality: functionalityCapabilities(card.Functionality),
		})
	}
	for _, functionality := range report.Functionality {
		snapshot.Functionality = append(snapshot.Functionality, unsupportedFunctionalityEntry{
			Capability: functionality.Capability,
			Cards:      functionality.Cards,
		})
	}
	return snapshot
}

func issueCodes(issues []ValidationIssue) []string {
	codes := make([]string, 0, len(issues))
	for _, issue := range issues {
		codes = append(codes, string(issue.Code))
	}
	sort.Strings(codes)
	return codes
}

func functionalityCapabilities(details []string) []string {
	seen := map[string]bool{}
	for _, detail := range details {
		seen[functionalityCapability(detail)] = true
	}
	capabilities := make([]string, 0, len(seen))
	for capability := range seen {
		capabilities = append(capabilities, capability)
	}
	sort.Strings(capabilities)
	return capabilities
}

func marshalSnapshot(t *testing.T, snapshot unsupportedReportSnapshot) string {
	t.Helper()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data) + "\n"
}

func normalizeSnapshotJSON(t *testing.T, data []byte) string {
	t.Helper()
	var snapshot unsupportedReportSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatal(err)
	}
	return marshalSnapshot(t, snapshot)
}

func cardFunctionalityPlan1Manifest() Manifest {
	card := func(name, path, validation string, issueCodes ...ValidationCode) ManifestCard {
		issues := make([]ValidationIssue, 0, len(issueCodes))
		for _, code := range issueCodes {
			issues = append(issues, ValidationIssue{
				CardName: name,
				FaceName: name,
				Code:     code,
				Message:  string(code),
			})
		}
		return ManifestCard{
			InputName:     name,
			CanonicalName: name,
			Quantity:      1,
			FirstLine:     len(issueCodes) + 1,
			Status:        BatchStatusFetched,
			FileStatus:    BatchFileStatusExisting,
			FilePath:      path,
			Validation:    validation,
			Issues:        issues,
		}
	}
	return Manifest{Version: ManifestVersion, Cards: []ManifestCard{
		card("Anger", filepath.Join("mtg", "cards", "a", "anger.go"), BatchValidationStatusValid),
		card("Arena", filepath.Join("mtg", "cards", "a", "arena.go"), BatchValidationStatusInvalid, IssueImplementationRequired),
		card("Basilisk Collar", filepath.Join("mtg", "cards", "b", "basilisk_collar.go"), BatchValidationStatusValid),
		card("Beast Within", filepath.Join("mtg", "cards", "b", "beast_within.go"), BatchValidationStatusValid),
		card("Birds of Paradise", filepath.Join("mtg", "cards", "b", "birds_of_paradise.go"), BatchValidationStatusValid),
		card("Bite Down", filepath.Join("mtg", "cards", "b", "bite_down.go"), BatchValidationStatusValid),
		card("Blazemire Verge", filepath.Join("mtg", "cards", "b", "blazemire_verge.go"), BatchValidationStatusValid),
		card("Blazing Sunsteel", filepath.Join("mtg", "cards", "b", "blazing_sunsteel.go"), BatchValidationStatusValid),
		card("Bridgeworks Battle // Tanglespan Bridgeworks", filepath.Join("mtg", "cards", "b", "bridgeworks_battle_tanglespan_bridgeworks.go"), BatchValidationStatusValid),
		card("Bugenhagen, Wise Elder", filepath.Join("mtg", "cards", "b", "bugenhagen_wise_elder.go"), BatchValidationStatusValid),
		card("Bushwhack", filepath.Join("mtg", "cards", "b", "bushwhack.go"), BatchValidationStatusValid),
		card("Chandra's Ignition", filepath.Join("mtg", "cards", "c", "chandra_s_ignition.go"), BatchValidationStatusValid),
		card("Chaos Warp", filepath.Join("mtg", "cards", "c", "chaos_warp.go"), BatchValidationStatusValid),
		card("Cinder Glade", filepath.Join("mtg", "cards", "c", "cinder_glade.go"), BatchValidationStatusValid),
		card("Command Tower", filepath.Join("mtg", "cards", "c", "command_tower.go"), BatchValidationStatusInvalid, IssueImplementationRequired),
	}}
}

func writePlan1SourceStubs(t *testing.T, repoRoot string, manifest Manifest) {
	t.Helper()
	comments := plan1MissingFunctionalityComments()
	for _, card := range manifest.Cards {
		source := "package cardstub\n"
		if comment := comments[card.CanonicalName]; comment != "" {
			source += "\n" + comment
		}
		path := filepath.Join(repoRoot, card.FilePath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func plan1MissingFunctionalityComments() map[string]string {
	return map[string]string{
		"Arena": `// Missing primitives:
	//   - TargetSpec has no "chooser" field; there is no way to declare that the
	//     second target is chosen by an opponent rather than the active player.
	//     ImplementationID "arena" is set so a hand-written rules handler can prompt
	//     the correct player to choose the opponent-controlled creature.
	`,
		"Bushwhack": `// Missing primitives:
	//   - SearchSpec has no MatchSupertype field; "basic" cannot be enforced
	//     declaratively -- the search allows any land card.
	`,
		"Command Tower": `// Missing primitives:
	//   - ResolutionChoice.Colors is a static slice; it cannot express "the colors in your
	//     commander's color identity," which is a dynamic game-state query. The approximation
	//     below offers all five colors; ImplementationID "command-tower" must restrict the
	//     choice to the controller's commander's color identity at activation time.
	`,
	}
}
