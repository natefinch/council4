package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// findAddCounter returns the first AddCounter primitive in an ability's
// resolution sequence.
func findAddCounter(t *testing.T, content game.AbilityContent) game.AddCounter {
	t.Helper()
	for _, mode := range content.Modes {
		for i := range mode.Sequence {
			if add, ok := mode.Sequence[i].Primitive.(game.AddCounter); ok {
				return add
			}
		}
	}
	t.Fatalf("no AddCounter in %#v", content)
	return game.AddCounter{}
}

// TestLowerWillowduskDynamicCounterTarget proves Willowdusk, Essence Seer's
// "{1}, {T}: Choose another target creature. Put a number of +1/+1 counters on
// it equal to the amount of life you gained this turn or the amount of life you
// lost this turn, whichever is greater." lowers to an activated ability that
// places +1/+1 counters on the chosen target, counted by a DynamicAmountMaxOf
// combinator over the controller's life gained and lost this turn.
func TestLowerWillowduskDynamicCounterTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Willowdusk, Essence Seer",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Dryad Druid",
		OracleText: "{1}, {T}: Choose another target creature. Put a number of +1/+1 counters on it equal to the amount of life you gained this turn or the amount of life you lost this turn, whichever is greater. Activate only as a sorcery.",
		Power:      new("0"),
		Toughness:  new("3"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	add := findAddCounter(t, face.ActivatedAbilities[0].Content)
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want PlusOnePlusOne", add.CounterKind)
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v, want TargetPermanentReference(0)", add.Object)
	}
	if add.Group.Valid() {
		t.Fatalf("group = %#v, want none for a single target", add.Group)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountMaxOf {
		t.Fatalf("amount = %#v, want DynamicAmountMaxOf", add.Amount)
	}
	wantOperands := []game.DynamicAmountKind{
		game.DynamicAmountLifeGainedThisTurn,
		game.DynamicAmountLifeLostThisTurn,
	}
	got := make([]game.DynamicAmountKind, 0, len(dynamic.Val.Operands))
	for _, operand := range dynamic.Val.Operands {
		got = append(got, operand.Kind)
	}
	if !slices.Equal(got, wantOperands) {
		t.Fatalf("operands = %v, want %v", got, wantOperands)
	}
}

// TestLowerAerithDynamicCounterGroup proves Aerith Gainsborough's "When Aerith
// Gainsborough dies, put X +1/+1 counters on each legendary creature you
// control, where X is the number of +1/+1 counters on Aerith Gainsborough."
// lowers to a dies-triggered ability that places +1/+1 counters on every
// legendary creature the controller controls, counted by a
// DynamicAmountObjectCounters reading the source's +1/+1 counters.
func TestLowerAerithDynamicCounterGroup(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Aerith Gainsborough",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Hero",
		OracleText: "Lifelink\nWhenever you gain life, put a +1/+1 counter on Aerith Gainsborough.\nWhen Aerith Gainsborough dies, put X +1/+1 counters on each legendary creature you control, where X is the number of +1/+1 counters on Aerith Gainsborough.",
		Power:      new("2"),
		Toughness:  new("4"),
	})
	var groupAdd game.AddCounter
	var found bool
	for _, ability := range face.TriggeredAbilities {
		add := findAddCounter(t, ability.Content)
		if add.Group.Valid() {
			groupAdd = add
			found = true
		}
	}
	if !found {
		t.Fatal("no group AddCounter among triggered abilities")
	}
	if groupAdd.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want PlusOnePlusOne", groupAdd.CounterKind)
	}
	selection := groupAdd.Group.Selection()
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("required types = %v, want [Creature]", selection.RequiredTypes)
	}
	if !slices.Equal(selection.Supertypes, []types.Super{types.Legendary}) {
		t.Fatalf("supertypes = %v, want [Legendary]", selection.Supertypes)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	dynamic := groupAdd.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectCounters {
		t.Fatalf("amount = %#v, want DynamicAmountObjectCounters", groupAdd.Amount)
	}
	if dynamic.Val.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counted kind = %v, want PlusOnePlusOne", dynamic.Val.CounterKind)
	}
}
