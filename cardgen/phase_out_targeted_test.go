package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func spellMode(t *testing.T, card *ScryfallCard) game.Mode {
	t.Helper()
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability = %+v, want one mode", face.SpellAbility)
	}
	return face.SpellAbility.Val.Modes[0]
}

// TestLowerAnyNumberTargetPhaseOut covers the unbounded "any number of target"
// phase-out lowering (Clever Concealment shape): a single target spec with the
// any-number cardinality and one PhaseOut over every chosen target permanent.
func TestLowerAnyNumberTargetPhaseOut(t *testing.T) {
	t.Parallel()
	mode := spellMode(t, &ScryfallCard{
		Name:       "Phase Out Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Any number of target nonland permanents you control phase out.",
	})
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %+v, want one spec", mode.Targets)
	}
	if mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 99 {
		t.Fatalf("target spec = %+v, want any number (0..99)", mode.Targets[0])
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one instruction", mode.Sequence)
	}
	phase, ok := mode.Sequence[0].Primitive.(game.PhaseOut)
	if !ok || phase.Object.Kind() != game.ObjectReferenceAllTargetPermanents {
		t.Fatalf("instruction = %+v, want all-target-permanents PhaseOut", mode.Sequence[0])
	}
}

// TestLowerSingleTargetPhaseOut covers the exact single-target phase-out lowering
// (Reality Ripple / Vodalian Illusionist shape): one 1..1 target spec and one
// PhaseOut over that chosen target permanent.
func TestLowerSingleTargetPhaseOut(t *testing.T) {
	t.Parallel()
	mode := spellMode(t, &ScryfallCard{
		Name:       "Phase Out Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature phases out.",
	})
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %+v, want one 1..1 spec", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one instruction", mode.Sequence)
	}
	phase, ok := mode.Sequence[0].Primitive.(game.PhaseOut)
	if !ok || phase.Object.Kind() != game.ObjectReferenceTargetPermanent {
		t.Fatalf("instruction = %+v, want single-target PhaseOut", mode.Sequence[0])
	}
}

// TestLowerUpToOneTargetPhaseOut covers the bounded "up to one target" phase-out
// lowering (Talon Gates of Madara's phase-out clause): a 0..1 target spec unrolled
// to one PhaseOut per slot targeting a chosen permanent reference.
func TestLowerUpToOneTargetPhaseOut(t *testing.T) {
	t.Parallel()
	mode := spellMode(t, &ScryfallCard{
		Name:       "Phase Out Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Up to one target creature phases out.",
	})
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %+v, want one 0..1 spec", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one instruction per slot", mode.Sequence)
	}
	phase, ok := mode.Sequence[0].Primitive.(game.PhaseOut)
	if !ok || phase.Object.Kind() != game.ObjectReferenceTargetPermanent {
		t.Fatalf("instruction = %+v, want single-target PhaseOut", mode.Sequence[0])
	}
}

// TestGenerateAnyNumberTargetPhaseOutSource confirms the generated source for the
// unbounded any-number phase-out carries the any-number spec and all-target
// reference.
func TestGenerateAnyNumberTargetPhaseOutSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Phase Out Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Any number of target nonland permanents you control phase out.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	for _, want := range []string{
		"MinTargets: 0,",
		"MaxTargets: 99,",
		"game.PhaseOut{",
		"Object: game.AllTargetPermanentsReference(0),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
