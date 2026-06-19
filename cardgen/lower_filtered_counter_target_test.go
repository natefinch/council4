package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// spellCounterTarget lowers a single-face instant whose only effect is a
// filtered counter placement and returns the lowered target spec.
func spellCounterTarget(t *testing.T, oracleText string) game.TargetSpec {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Spell",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered for %q", oracleText)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Targets) != 1 {
		t.Fatalf("modes = %#v, want one mode with one target", modes)
	}
	return modes[0].Targets[0]
}

// expectUnsupportedCounterPlacement asserts that the card fails closed with an
// "unsupported counter placement" diagnostic and lowers no executable face.
func expectUnsupportedCounterPlacement(t *testing.T, oracleText string) {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Counter Spell",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	for i := range faces {
		if faces[i].SpellAbility.Exists {
			t.Fatalf("%q unexpectedly lowered a spell ability", oracleText)
		}
	}
	found := false
	for i := range diagnostics {
		if diagnostics[i].Summary == "unsupported counter placement" {
			found = true
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want unsupported counter placement", diagnostics)
	}
}

func TestLowerCounterPlacementSubtypeTarget(t *testing.T) {
	t.Parallel()
	target := spellCounterTarget(t, "Put a +1/+1 counter on target Beast creature you control.")
	if !slices.Contains(target.Predicate.PermanentTypes, types.Creature) {
		t.Fatalf("permanent types = %#v, want Creature", target.Predicate.PermanentTypes)
	}
	if !slices.Contains(target.Predicate.Subtypes, types.Sub("Beast")) {
		t.Fatalf("subtypes = %#v, want Beast", target.Predicate.Subtypes)
	}
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
	if target.Predicate.Another {
		t.Fatal("another = true, want false")
	}
}

func TestLowerCounterPlacementSubtypeAsNounTarget(t *testing.T) {
	t.Parallel()
	target := spellCounterTarget(t, "Put a +1/+1 counter on target Soldier you control.")
	if !slices.Contains(target.Predicate.Subtypes, types.Sub("Soldier")) {
		t.Fatalf("subtypes = %#v, want Soldier", target.Predicate.Subtypes)
	}
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
}

func TestLowerCounterPlacementAnotherTargetExcludesSelf(t *testing.T) {
	t.Parallel()
	target := spellCounterTarget(t, "Put a +1/+1 counter on another target creature.")
	if !target.Predicate.Another {
		t.Fatal("another = false, want true (must exclude source)")
	}
	if !slices.Contains(target.Predicate.PermanentTypes, types.Creature) {
		t.Fatalf("permanent types = %#v, want Creature", target.Predicate.PermanentTypes)
	}
	// The "another" exclusion must flow through to the runtime selection.
	if !target.Predicate.Selection().ExcludeSource {
		t.Fatal("selection.ExcludeSource = false, want true")
	}
}

func TestLowerCounterPlacementAnotherSubtypeTarget(t *testing.T) {
	t.Parallel()
	target := spellCounterTarget(t, "Put a +1/+1 counter on another target Soldier you control.")
	if !target.Predicate.Another {
		t.Fatal("another = false, want true")
	}
	if !slices.Contains(target.Predicate.Subtypes, types.Sub("Soldier")) {
		t.Fatalf("subtypes = %#v, want Soldier", target.Predicate.Subtypes)
	}
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
}

func TestLowerCounterPlacementGroupRecipient(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Boon",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a +1/+1 counter on each creature you control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0 (group recipient)", len(mode.Targets))
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Object.Kind() != game.ObjectReferenceNone {
		t.Fatalf("Object = %v, want none (group form)", add.Object.Kind())
	}
	if add.Group.Domain() == 0 {
		t.Fatal("Group not set on group counter placement")
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", add.CounterKind)
	}
}

// A keyword-filtered group ("each creature you control with flying") reconstructs
// exactly with the controller clause preceding the keyword qualifier, so group
// counter placement lowers with the keyword carried onto the group selection.
func TestLowerCounterPlacementKeywordGroupRecipient(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wingspan",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a +1/+1 counter on each creature you control with flying.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0 (group recipient)", len(mode.Targets))
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Group.Domain() == 0 {
		t.Fatal("Group not set on group counter placement")
	}
	selection := add.Group.Selection()
	if selection.Keyword != game.Flying {
		t.Fatalf("group keyword = %v, want flying", selection.Keyword)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("group controller = %v, want ControllerYou", selection.Controller)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", add.CounterKind)
	}
}

func TestLowerCounterPlacementUnrepresentableFilterFailsClosed(t *testing.T) {
	t.Parallel()
	// "other than that creature" is a reference exclusion the parser does not capture.
	expectUnsupportedCounterPlacement(t, "Put a +1/+1 counter on target creature other than that creature.")
}

// A keyword counter on a group ("Put a deathtouch counter on each creature you
// control") lowers despite the spurious semantic keyword that naming a keyword
// counter registers: the group lowerer tolerates only the keyword matching the
// placed counter and carries the runtime-modeled counter kind onto the group.
func TestLowerCounterPlacementKeywordCounterGroupRecipient(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vraska",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a deathtouch counter on each creature you control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Group.Domain() == 0 {
		t.Fatal("Group not set on keyword counter group placement")
	}
	if add.CounterKind != counter.Deathtouch {
		t.Fatalf("counter kind = %v, want deathtouch", add.CounterKind)
	}
}

// A single target whose controller clause precedes a "without flying" qualifier
// lowers with an excluded-keyword predicate ("target creature you control
// without flying"), the canonical Oracle ordering.
func TestLowerCounterPlacementControllerWithoutKeywordTarget(t *testing.T) {
	t.Parallel()
	target := spellCounterTarget(t, "Put a +1/+1 counter on target creature you control without flying.")
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
	if target.Predicate.ExcludedKeyword != game.Flying {
		t.Fatalf("excluded keyword = %v, want flying", target.Predicate.ExcludedKeyword)
	}
}
