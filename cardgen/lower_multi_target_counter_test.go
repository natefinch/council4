package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// spellCounterMode lowers a single-face sorcery whose only effect is a counter
// placement and returns the lowered spell-ability mode.
func spellCounterMode(t *testing.T, oracleText string) game.Mode {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Sorcery",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered for %q", oracleText)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %#v, want one mode", modes)
	}
	return modes[0]
}

func TestLowerMultiTargetCounterPlacementUpToTwo(t *testing.T) {
	t.Parallel()
	mode := spellCounterMode(t, "Put a +1/+1 counter on each of up to two target creatures.")
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	spec := mode.Targets[0]
	if spec.MinTargets != 0 || spec.MaxTargets != 2 {
		t.Fatalf("cardinality = {%d,%d}, want {0,2}", spec.MinTargets, spec.MaxTargets)
	}
	if spec.Allow != game.TargetAllowPermanent {
		t.Fatalf("allow = %v, want permanent", spec.Allow)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	for i := range mode.Sequence {
		add, ok := mode.Sequence[i].Primitive.(game.AddCounter)
		if !ok {
			t.Fatalf("sequence[%d] = %#v, want AddCounter", i, mode.Sequence[i].Primitive)
		}
		if add.CounterKind != counter.PlusOnePlusOne {
			t.Fatalf("sequence[%d] counter = %v, want +1/+1", i, add.CounterKind)
		}
		if add.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] object = %#v, want target %d", i, add.Object, i)
		}
	}
}

func TestLowerMultiTargetCounterPlacementUpToThree(t *testing.T) {
	t.Parallel()
	mode := spellCounterMode(t, "Put a +1/+1 counter on each of up to three target creatures.")
	if mode.Targets[0].MaxTargets != 3 {
		t.Fatalf("max targets = %d, want 3", mode.Targets[0].MaxTargets)
	}
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence length = %d, want 3", len(mode.Sequence))
	}
}

// TestLowerCounterPlacementUpToOneTarget proves the optional single permanent
// target ("Put a +1/+1 counter on up to one target creature.") lowers through
// the per-target fan-out as a single optional-target slot, so the counter is
// placed only when the optional target is chosen.
func TestLowerCounterPlacementUpToOneTarget(t *testing.T) {
	t.Parallel()
	mode := spellCounterMode(t, "Put two +1/+1 counters on up to one target creature.")
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	spec := mode.Targets[0]
	if spec.MinTargets != 0 || spec.MaxTargets != 1 {
		t.Fatalf("cardinality = {%d,%d}, want {0,1}", spec.MinTargets, spec.MaxTargets)
	}
	if spec.Allow != game.TargetAllowPermanent {
		t.Fatalf("allow = %v, want permanent", spec.Allow)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want AddCounter", mode.Sequence[0].Primitive)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter = %v, want +1/+1", add.CounterKind)
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v, want target 0", add.Object)
	}
	if got, want := add.Amount, game.Fixed(2); got != want {
		t.Fatalf("amount = %#v, want %#v", got, want)
	}
}

func TestLowerMultiTargetCounterPlacementOtherYouControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Savior",
		Layout:     "normal",
		TypeLine:   "Creature — Cat",
		OracleText: "When this creature enters, put a +1/+1 counter on each of up to two other target creatures you control.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	spec := mode.Targets[0]
	if !spec.Predicate.Another {
		t.Fatal("predicate Another = false, want true for \"other\"")
	}
	if spec.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want you", spec.Predicate.Controller)
	}
	if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("permanent types = %#v, want [creature]", spec.Predicate.PermanentTypes)
	}
}

func TestLowerMultiTargetCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []string{
		// Distribution among targets is not modeled.
		"Distribute three +1/+1 counters among any number of target creatures.",
		// A dynamic per-target amount stays unsupported.
		"Put X +1/+1 counters on each of up to two target creatures, where X is the number of cards in your hand.",
		// Subtype-restricted plural targets are not a plain permanent noun.
		"Put a +1/+1 counter on each of up to two target Merfolk you control.",
		// An unbounded "any number of" cardinality is not represented.
		"Put a +1/+1 counter on each of any number of target creatures.",
	}
	for _, text := range cases {
		faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Counter Sorcery",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: text,
		})
		for i := range faces {
			if faces[i].SpellAbility.Exists {
				t.Fatalf("%q unexpectedly lowered a spell ability", text)
			}
		}
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without any diagnostic, want fail-closed", text)
		}
	}
}
