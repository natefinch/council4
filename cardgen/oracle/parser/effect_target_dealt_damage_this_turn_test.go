package parser

import "testing"

// firstDestroyTargets parses a single destroy sentence and returns its targets,
// requiring the parse to be diagnostic-free and shaped as one destroy effect.
func firstDestroyTargets(t *testing.T, source string) []TargetSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: "Test"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDestroy {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Targets
}

// TestParseDealtDamageThisTurnTargetIsExact verifies that the trailing
// "that was dealt damage this turn" qualifier is captured onto the target's
// selection rather than leaking into the effect, and that the target still
// round-trips to an exact, lowerable production (Fatal Blow).
func TestParseDealtDamageThisTurnTargetIsExact(t *testing.T) {
	t.Parallel()
	targets := firstDestroyTargets(t, "Destroy target creature that was dealt damage this turn.")
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	target := targets[0]
	if target.Selection.Kind != SelectionCreature {
		t.Fatalf("target kind = %v, want SelectionCreature", target.Selection.Kind)
	}
	if !target.Selection.DealtDamageThisTurn {
		t.Fatal("target DealtDamageThisTurn = false, want true")
	}
	if !target.Exact {
		t.Fatal("target Exact = false, want true")
	}
}

// TestParseDealtDamageThisTurnBareTargetHasNoFilter verifies that an ordinary
// destroy target without the qualifier does not spuriously gain the filter.
func TestParseDealtDamageThisTurnBareTargetHasNoFilter(t *testing.T) {
	t.Parallel()
	targets := firstDestroyTargets(t, "Destroy target creature.")
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].Selection.DealtDamageThisTurn {
		t.Fatal("bare target DealtDamageThisTurn = true, want false")
	}
}
