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
