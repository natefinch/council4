package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// TestLowerOtherwiseBranchKeyedOnEventPower proves that an entering-creature
// trigger whose body is "draw a card if its power is 3 or greater. Otherwise,
// put two +1/+1 counters on it." (Tribute to the World Tree) lowers to two
// mutually-exclusive instructions: the draw gated on the event permanent's
// power being at least 3, and the counter placement gated on the negation of
// that same condition, with the counters addressed to the event permanent.
func TestLowerOtherwiseBranchKeyedOnEventPower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Otherwise",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control enters, draw a card if its power is 3 or greater. Otherwise, put two +1/+1 counters on it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}

	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("first primitive = %+v, want draw one", mode.Sequence[0].Primitive)
	}
	assertEventPowerGate(t, mode.Sequence[0].Condition, 3, false)

	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddCounter", mode.Sequence[1].Primitive)
	}
	if add.Object != game.EventPermanentReference() {
		t.Fatalf("counter object = %#v, want EventPermanentReference", add.Object)
	}
	if add.CounterKind != counter.PlusOnePlusOne || add.Amount != game.Fixed(2) {
		t.Fatalf("counter = %+v, want two +1/+1", add)
	}
	assertEventPowerGate(t, mode.Sequence[1].Condition, 3, true)
}

func assertEventPowerGate(t *testing.T, gate opt.V[game.EffectCondition], value int, negated bool) {
	t.Helper()
	if !gate.Exists {
		t.Fatal("instruction has no condition gate, want event-power gate")
	}
	cond := gate.Val.Condition
	if !cond.Exists {
		t.Fatal("effect condition has no wrapped condition")
	}
	if cond.Val.Negate != negated {
		t.Fatalf("gate negate = %v, want %v", cond.Val.Negate, negated)
	}
	if !cond.Val.Object.Exists || cond.Val.Object.Val != game.EventPermanentReference() {
		t.Fatalf("wrapped object = %#v, want EventPermanentReference", cond.Val.Object)
	}
	if !cond.Val.ObjectMatches.Exists {
		t.Fatal("wrapped condition lacks ObjectMatches selection")
	}
	power := cond.Val.ObjectMatches.Val.Power
	if !power.Exists || power.Val.Op != compare.GreaterOrEqual || power.Val.Value != value {
		t.Fatalf("power selection = %#v, want >= %d", power, value)
	}
}

// TestLowerOtherwiseFailsClosedWithoutPrecedingCondition proves that an
// "Otherwise," branch whose preceding effect carries no gate condition fails
// closed: with nothing to negate, the else branch cannot be gated, so the
// sequence is rejected rather than silently running both effects.
func TestLowerOtherwiseFailsClosedWithoutPrecedingCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Bare Otherwise",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control enters, you gain 1 life. Otherwise, draw a card.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("bare Otherwise lowered without diagnostics; expected fail-closed")
	}
}
