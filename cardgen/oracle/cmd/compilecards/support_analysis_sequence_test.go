package main

import (
	"strings"
	"testing"
)

func orderedSequenceCard(name string, summaries ...string) unsupported {
	diagnostics := make([]reportDiagnostic, len(summaries))
	for i, summary := range summaries {
		detail := ""
		if summary == orderedSequenceReasonSummary {
			detail = "sub-effect — unsupported life spell"
		}
		diagnostics[i] = reportDiagnostic{Summary: summary, Detail: detail}
	}
	return unsupported{Name: name, Diagnostics: diagnostics}
}

func TestAnalyzeOrderedSequenceCategoriesCountsSoleAndCoBlocked(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		// Sole blocker: ordered sequence is the only distinct summary.
		orderedSequenceCard("Sole", orderedSequenceReasonSummary),
		// Co-blocked: a second distinct summary excludes it from sole blockers.
		orderedSequenceCard("CoBlocked", orderedSequenceReasonSummary, "unsupported Oracle construct"),
		// Structural sole blocker under a different category.
		{Name: "Structural", Diagnostics: []reportDiagnostic{{
			Summary: orderedSequenceReasonSummary,
			Detail:  "structural — contains sacrifice effect",
		}}},
		// Unrelated card must not appear.
		{Name: "Other", Diagnostics: []reportDiagnostic{{Summary: "unsupported Oracle construct"}}},
	}}

	categories := analyzeOrderedSequenceCategories(output)
	got := map[string]orderedSequenceCategory{}
	for _, category := range categories {
		got[category.category] = category
	}

	sub := got["sub-effect — unsupported life spell"]
	if sub.affectedCards != 2 || sub.soleBlockerCards != 1 {
		t.Errorf("sub-effect counts = affected %d sole %d, want 2/1", sub.affectedCards, sub.soleBlockerCards)
	}
	structural := got["structural — contains sacrifice effect"]
	if structural.affectedCards != 1 || structural.soleBlockerCards != 1 {
		t.Errorf("structural counts = affected %d sole %d, want 1/1", structural.affectedCards, structural.soleBlockerCards)
	}
	if len(categories) != 2 {
		t.Fatalf("category count = %d, want 2: %+v", len(categories), categories)
	}
	// Sorted by sole blockers desc, then affected desc: sub-effect (sole 1,
	// affected 2) precedes structural (sole 1, affected 1).
	if categories[0].category != "sub-effect — unsupported life spell" {
		t.Errorf("first category = %q, want sub-effect", categories[0].category)
	}
}

func TestWriteOrderedSequenceCategoriesRendersTable(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		orderedSequenceCard("Sole", orderedSequenceReasonSummary),
	}}
	var builder strings.Builder
	writeOrderedSequenceCategories(&builder, output)
	rendered := builder.String()
	for _, wanted := range []string{
		"## Ordered effect sequence sub-categories",
		"| Category | Affected cards | Sole blockers |",
		"| sub-effect — unsupported life spell | 1 | 1 |",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Errorf("rendered output missing %q:\n%s", wanted, rendered)
		}
	}
}

func TestWriteOrderedSequenceCategoriesEmptyWhenNoSequences(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		{Name: "Other", Diagnostics: []reportDiagnostic{{Summary: "unsupported Oracle construct"}}},
	}}
	var builder strings.Builder
	writeOrderedSequenceCategories(&builder, output)
	if builder.Len() != 0 {
		t.Errorf("expected no output, got:\n%s", builder.String())
	}
}

func TestNormalizeOrderedSequenceCategory(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		orderedSequenceUnrecognizedConditionPrefix + "if this spell was kicked": orderedSequenceUnrecognizedConditionCategory,
		orderedSequenceUnrecognizedConditionPrefix + "if you win":               orderedSequenceUnrecognizedConditionCategory,
		"sub-effect — unsupported life spell":                                   "sub-effect — unsupported life spell",
		"structural — per-effect condition spans multiple clauses":              "structural — per-effect condition spans multiple clauses",
		"": "",
	}
	for detail, want := range cases {
		if got := normalizeOrderedSequenceCategory(detail); got != want {
			t.Errorf("normalizeOrderedSequenceCategory(%q) = %q, want %q", detail, got, want)
		}
	}
}

func unrecognizedConditionCard(name, wording string, extraSummaries ...string) unsupported {
	diagnostics := []reportDiagnostic{{
		Summary: orderedSequenceReasonSummary,
		Detail:  orderedSequenceUnrecognizedConditionPrefix + wording,
	}}
	for _, summary := range extraSummaries {
		diagnostics = append(diagnostics, reportDiagnostic{Summary: summary})
	}
	return unsupported{Name: name, Diagnostics: diagnostics}
}

func TestAnalyzeConditionRecognitionBacklogRanksWordings(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		unrecognizedConditionCard("Kicked1", "if this spell was kicked"),
		unrecognizedConditionCard("Kicked2", "if this spell was kicked"),
		// Co-blocked: a second distinct summary excludes it from sole blockers.
		unrecognizedConditionCard("Kicked3", "if this spell was kicked", "unsupported Oracle construct"),
		unrecognizedConditionCard("Win1", "if you win"),
		// A recognized (gateable) detail must not appear in the backlog.
		orderedSequenceCard("Recognized", orderedSequenceReasonSummary),
	}}

	backlog := analyzeConditionRecognitionBacklog(output)
	if len(backlog) != 2 {
		t.Fatalf("backlog length = %d, want 2: %+v", len(backlog), backlog)
	}
	kicked := backlog[0]
	if kicked.condition != "if this spell was kicked" || kicked.affectedCards != 3 || kicked.soleBlockerCards != 2 {
		t.Errorf("kicked entry = %+v, want condition kicked affected 3 sole 2", kicked)
	}
	win := backlog[1]
	if win.condition != "if you win" || win.affectedCards != 1 || win.soleBlockerCards != 1 {
		t.Errorf("win entry = %+v, want condition win affected 1 sole 1", win)
	}
}

func TestWriteConditionRecognitionBacklogRendersTable(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		unrecognizedConditionCard("Kicked", "if this spell was kicked"),
	}}
	var builder strings.Builder
	writeConditionRecognitionBacklog(&builder, output)
	rendered := builder.String()
	for _, wanted := range []string{
		"## Unrecognized per-effect conditions (recognition backlog)",
		"| Unrecognized condition | Affected cards | Sole blockers |",
		"| if this spell was kicked | 1 | 1 |",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Errorf("rendered output missing %q:\n%s", wanted, rendered)
		}
	}
}

func TestWriteConditionRecognitionBacklogEmptyWhenNoUnrecognized(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		orderedSequenceCard("Sole", orderedSequenceReasonSummary),
	}}
	var builder strings.Builder
	writeConditionRecognitionBacklog(&builder, output)
	if builder.Len() != 0 {
		t.Errorf("expected no output, got:\n%s", builder.String())
	}
}
