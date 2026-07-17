package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
)

// firstSentenceTarget parses a single instant/sorcery sentence and returns its
// lone target, failing the test on any diagnostic or unexpected shape.
func firstSentenceTarget(t *testing.T, source string) TargetSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("Parse(%q) targets = %#v, want one", source, targets)
	}
	return targets[0]
}

// TestParseManaValueXTargetRoundTrips proves "target creature with mana value X
// or less" survives the text-blind target round-trip as an exact creature target
// carrying the spell-X-derived mana value bound (ManaValueX). The bare X token
// must not wipe the selection, and the reconstruction must reproduce the "X or
// less" spelling rather than the placeholder value 0.
func TestParseManaValueXTargetRoundTrips(t *testing.T) {
	t.Parallel()
	target := firstSentenceTarget(t, "Gain control of target creature with mana value X or less.")
	if !target.Exact {
		t.Fatal("\"target creature with mana value X or less\" must round-trip as an exact target")
	}
	sel := target.Selection
	if sel.Kind != SelectionCreature {
		t.Fatalf("selection kind = %v, want SelectionCreature", sel.Kind)
	}
	if !sel.MatchManaValue || !sel.ManaValueX {
		t.Fatalf("selection MatchManaValue=%v ManaValueX=%v, want both true", sel.MatchManaValue, sel.ManaValueX)
	}
	if sel.ManaValue.Op != compare.LessOrEqual {
		t.Fatalf("selection mana value op = %v, want LessOrEqual", sel.ManaValue.Op)
	}
	got, ok := exactPermanentTargetText(sel)
	if !ok || got != "target creature with mana value X or less" {
		t.Fatalf("reconstruction = %q ok=%v, want \"target creature with mana value X or less\"", got, ok)
	}
}

// TestParseManaValueXExactTargetRoundTrips proves the bare "with mana value X"
// bound (no "or less") survives the text-blind target round-trip as an exact
// nontoken-artifact-you-control target carrying the exact X comparison
// (compare.Equal). The bare X token must not wipe the selection, and the
// reconstruction must reproduce "mana value X" rather than "mana value X or less"
// or the placeholder value 0, so the exactness fails closed if reworded.
func TestParseManaValueXExactTargetRoundTrips(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("Parse diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse shape = %#v", document.Abilities)
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("targets = %#v, want one", targets)
	}
	target := targets[0]
	if !target.Exact {
		t.Fatal("\"target nontoken artifact you control with mana value X\" must round-trip as an exact target")
	}
	sel := target.Selection
	if sel.Kind != SelectionArtifact {
		t.Fatalf("selection kind = %v, want SelectionArtifact", sel.Kind)
	}
	if !sel.NonToken {
		t.Fatal("selection must record the nontoken qualifier")
	}
	if sel.Controller != SelectionControllerYou {
		t.Fatalf("selection controller = %v, want SelectionControllerYou", sel.Controller)
	}
	if !sel.MatchManaValue || !sel.ManaValueX {
		t.Fatalf("selection MatchManaValue=%v ManaValueX=%v, want both true", sel.MatchManaValue, sel.ManaValueX)
	}
	if sel.ManaValue.Op != compare.Equal {
		t.Fatalf("selection mana value op = %v, want Equal", sel.ManaValue.Op)
	}
	got, ok := exactPermanentTargetText(sel)
	if !ok || got != "target nontoken artifact you control with mana value X" {
		t.Fatalf("reconstruction = %q ok=%v, want \"target nontoken artifact you control with mana value X\"", got, ok)
	}
}

// TestParseManaValueFixedTargetStillRoundTrips guards the fixed "mana value N or
// less" target: it must keep round-tripping as an exact creature target with the
// concrete comparison and without the X-derived flag, so the X handling does not
// disturb the fixed form.
func TestParseManaValueFixedTargetStillRoundTrips(t *testing.T) {
	t.Parallel()
	target := firstSentenceTarget(t, "Gain control of target creature with mana value 3 or less.")
	if !target.Exact {
		t.Fatal("\"target creature with mana value 3 or less\" must round-trip as an exact target")
	}
	sel := target.Selection
	if sel.ManaValueX {
		t.Fatal("a fixed mana-value target must not set ManaValueX")
	}
	if !sel.MatchManaValue || sel.ManaValue.Op != compare.LessOrEqual || sel.ManaValue.Value != 3 {
		t.Fatalf("selection mana value = %+v, want <=3", sel.ManaValue)
	}
}
