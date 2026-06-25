package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerTapStunCounterTriggeredSequence verifies the modern tap-and-stun
// trigger ("When this creature enters, tap target creature an opponent controls
// and put a stun counter on it.") lowers to a two-instruction sequence — tap
// the opponent's creature, then add one stun counter to that same target.
func TestLowerTapStunCounterTriggeredSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Snowballer",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "When Test Snowballer enters, tap target creature an opponent controls and put a stun counter on it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if mode.Targets[0].Selection.Val.Controller != game.ControllerOpponent {
		t.Fatalf("target controller = %v, want opponent", mode.Targets[0].Selection.Val.Controller)
	}
	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok || tap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first primitive = %+v, want tap of target 0", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok || add.Object != game.TargetPermanentReference(0) || add.CounterKind != counter.Stun {
		t.Fatalf("second primitive = %+v, want one stun counter on target 0", mode.Sequence[1].Primitive)
	}
	if mode.Sequence[1].Condition.Exists {
		t.Fatalf("unconditional stun placement gained a condition: %+v", mode.Sequence[1].Condition)
	}
}

// TestLowerStunCounterPlacement verifies a standalone stun-counter placement
// ("Put a stun counter on target creature.") lowers to a single AddCounter
// instruction now that the untap step models stun counters.
func TestLowerStunCounterPlacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Stun",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put a stun counter on target creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	add, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.Stun {
		t.Fatalf("primitive = %+v, want stun AddCounter", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
}

// TestLowerTapStunCounterConditionalGate verifies the conditional tap-and-stun
// trigger (Protocol Knight) gates only the stun placement on the intervening
// condition, leaving the tap unconditional.
func TestLowerTapStunCounterConditionalGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Knight",
		Layout:     "normal",
		TypeLine:   "Creature — Knight",
		OracleText: "When Test Knight enters, tap target creature an opponent controls. Put a stun counter on that creature if you control another Knight.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want two instructions", mode)
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("tap instruction was gated, want unconditional tap")
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.Stun {
		t.Fatalf("second primitive = %+v, want stun AddCounter", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].Condition.Exists {
		t.Fatal("conditional stun placement was not gated on its intervening condition")
	}
}

// TestLowerTapFinalityCounterFailsClosed verifies that finality counters, whose
// death-replacement semantics remain unmodeled, keep a tap-and-counter sequence
// unsupported even though the parallel stun wording now lowers.
func TestLowerTapFinalityCounterFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Finality",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "When Test Finality enters, tap target creature an opponent controls and put a finality counter on it.",
	})
	for _, face := range faces {
		if len(face.TriggeredAbilities) != 0 {
			t.Error("finality tap-and-counter lowered a triggered ability, want fail closed")
		}
	}
}
