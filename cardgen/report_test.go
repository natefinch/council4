package cardgen

import (
	"bytes"
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
