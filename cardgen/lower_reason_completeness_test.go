package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestDedupeReasonsKeepsFirstOfDuplicates(t *testing.T) {
	t.Parallel()
	got := dedupeReasons([]shared.Diagnostic{
		{Summary: "a", Detail: "x"},
		{Summary: "a", Detail: "x"},
		{Summary: "a", Detail: "y"},
		{Summary: "b", Detail: "x"},
	})
	if len(got) != 3 {
		t.Fatalf("dedupeReasons returned %d reasons, want 3: %#v", len(got), got)
	}
	if got[0].Detail != "x" || got[1].Detail != "y" || got[2].Summary != "b" {
		t.Fatalf("dedupeReasons did not preserve first-seen order: %#v", got)
	}
}

func TestCombineReasonsKeepsFirstPrimaryAndDedupes(t *testing.T) {
	t.Parallel()
	got := combineReasons([]shared.Diagnostic{
		{Summary: "primary", Detail: "1"},
		{Summary: "second", Detail: "2"},
		{Summary: "primary", Detail: "1"},
	})
	if got.Summary != "primary" || got.Detail != "1" {
		t.Fatalf("combineReasons primary = %q/%q, want primary/1", got.Summary, got.Detail)
	}
	if len(got.Additional) != 1 || got.Additional[0].Summary != "second" {
		t.Fatalf("combineReasons Additional = %#v, want one 'second' reason", got.Additional)
	}
}

func TestFlattenAdditionalReasonsExpandsSiblings(t *testing.T) {
	t.Parallel()
	got := flattenAdditionalReasons([]shared.Diagnostic{
		{
			Summary:    "primary",
			Additional: []shared.Diagnostic{{Summary: "extra-a"}, {Summary: "extra-b"}},
		},
		{Summary: "standalone"},
	})
	if len(got) != 4 {
		t.Fatalf("flattenAdditionalReasons returned %d entries, want 4: %#v", len(got), got)
	}
	if got[0].Summary != "primary" || got[0].Additional != nil {
		t.Fatalf("primary should keep summary and clear Additional: %#v", got[0])
	}
	if got[1].Summary != "extra-a" || got[2].Summary != "extra-b" || got[3].Summary != "standalone" {
		t.Fatalf("flattenAdditionalReasons did not expand siblings in order: %#v", got)
	}
}

// TestOptionalProbeSurfacesInnerBlocker proves the optional wrapper no longer
// hides the blockers that would remain even if optional resolving effects were
// supported: Chain of Plasma is blocked by both an optional effect and an
// unsupported ordered sequence, and both must be reported so support planning does
// not overcount how many cards supporting optional effects alone would unblock.
func TestOptionalProbeSurfacesInnerBlocker(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Chain of Plasma",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Chain of Plasma deals 3 damage to any target. Then that player or that permanent's controller may discard a card. If the player does, they may copy this spell and may choose a new target for that copy.",
	})
	if !hasDiagnosticSummary(diagnostics, "unsupported optional effect") {
		t.Errorf("expected the optional blocker to be reported; got %#v", diagnostics)
	}
	if !hasDiagnosticSummary(diagnostics, "unsupported ordered effect sequence") {
		t.Errorf("expected the inner ordered-sequence blocker to be reported alongside the optional blocker; got %#v", diagnostics)
	}
}

// TestSequenceCollectsAllClauseBlockers proves an ordered effect sequence reports
// blockers from more than one clause rather than bailing at the first. Rashmi's
// trigger fans out into an optional cast plus an ordered sequence whose clauses are
// unsupported, so the card must carry the ordered-sequence blocker.
func TestSequenceCollectsAllClauseBlockers(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Rashmi, Eternities Crafter",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elf Druid",
		Power:      new("2"),
		Toughness:  new("3"),
		OracleText: "Whenever you cast your first spell each turn, reveal the top card of your library. You may cast it without paying its mana cost if it's a spell with lesser mana value. If you don't cast it, put it into your hand.",
	})
	if len(diagnostics) == 0 {
		t.Fatalf("expected diagnostics for an unsupported card")
	}
	if !hasDiagnosticSummary(diagnostics, "unsupported ordered effect sequence") {
		t.Errorf("expected the ordered-sequence blocker to be reported; got %#v", diagnostics)
	}
}
