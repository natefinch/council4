package parser

import "testing"

// TestKickerScaledTargetPreambleFolds verifies that the two-target "Choose any
// target, then choose another target for each time this spell was kicked."
// preamble (Comet Storm) folds into a single target flagged KickerScaledCount,
// dropping the second "another target" slot whose count the Multikicker count
// supplies.
func TestKickerScaledTargetPreambleFolds(t *testing.T) {
	document, diags := Parse(
		"Choose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.",
		Context{CardName: "Comet Storm", InstantOrSorcery: true},
	)
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("preamble targets = %d, want 1 folded target", len(targets))
	}
	if !targets[0].KickerScaledCount {
		t.Fatal("target[0].KickerScaledCount = false, want true")
	}
	if targets[0].Text != "any target" {
		t.Fatalf("target[0].Text = %q, want %q", targets[0].Text, "any target")
	}
}

// TestKickerScaledTargetPreambleBareItWording verifies the "for each time it was
// kicked" spelling folds the same way as the explicit "this spell" spelling.
func TestKickerScaledTargetPreambleBareItWording(t *testing.T) {
	document, _ := Parse(
		"Choose any target, then choose another target for each time it was kicked. Comet Storm deals X damage to each of them.",
		Context{CardName: "Comet Storm", InstantOrSorcery: true},
	)
	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 || !targets[0].KickerScaledCount {
		t.Fatalf("targets = %#v, want a single KickerScaledCount target", targets)
	}
}

// TestNonKickerTwoTargetPreambleUnchanged verifies a two-target preamble whose
// second slot is not the kicker-scaled phrase keeps its ordinary two-target parse
// rather than folding.
func TestNonKickerTwoTargetPreambleUnchanged(t *testing.T) {
	document, _ := Parse(
		"Choose target creature, then choose target player. Test Spell deals 2 damage to each of them.",
		Context{CardName: "Test Spell", InstantOrSorcery: true},
	)
	targets := document.Abilities[0].Sentences[0].Targets
	for i := range targets {
		if targets[i].KickerScaledCount {
			t.Fatalf("target[%d].KickerScaledCount = true, want false for a non-kicker preamble", i)
		}
	}
}
