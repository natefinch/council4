package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// counterUnlessPaysPayment lowers a "Counter target spell unless its controller
// pays <cost>." spell and returns its Pay primitive.
func counterUnlessPaysPayment(t *testing.T, oracleText string) game.Pay {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Unless Pays",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2 (pay then counter)", len(mode.Sequence))
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.Pay", mode.Sequence[0].Primitive)
	}
	if counter, ok := mode.Sequence[1].Primitive.(game.CounterObject); !ok ||
		counter.Object != game.TargetStackObjectReference(0) ||
		!mode.Sequence[1].ResultGate.Exists ||
		mode.Sequence[1].ResultGate.Val.Succeeded != game.TriFalse {
		t.Fatalf("counter instruction = %#v, want CounterObject gated on the unpaid tax", mode.Sequence[1])
	}
	return pay
}

// TestLowerCounterUnlessPaysVariableX lowers "Counter target spell unless its
// controller pays {X}." (Clash of Wills, Martyr of Frost, Logic Knot) to a Pay
// whose generic cost is the resolving spell's X, evaluated as the payment
// resolves, rather than a fixed mana cost.
func TestLowerCounterUnlessPaysVariableX(t *testing.T) {
	t.Parallel()
	pay := counterUnlessPaysPayment(t, "Counter target spell unless its controller pays {X}.")
	if pay.Payment.ManaCost.Exists {
		t.Fatalf("variable {X} tax should not set a fixed ManaCost: %#v", pay.Payment)
	}
	dynamic := pay.Payment.DynamicGenericManaCost
	if !dynamic.Exists || dynamic.Val == nil || dynamic.Val.Kind != game.DynamicAmountX {
		t.Fatalf("dynamic generic cost = %#v, want DynamicAmountX", pay.Payment.DynamicGenericManaCost)
	}
}

// TestLowerCounterUnlessPaysFixed confirms the fixed "{N}" tax still lowers to a
// fixed ManaCost with no dynamic generic cost, unchanged by the {X} extension.
func TestLowerCounterUnlessPaysFixed(t *testing.T) {
	t.Parallel()
	pay := counterUnlessPaysPayment(t, "Counter target spell unless its controller pays {4}.")
	if !pay.Payment.ManaCost.Exists {
		t.Fatalf("fixed {4} tax missing ManaCost: %#v", pay.Payment)
	}
	if pay.Payment.DynamicGenericManaCost.Exists {
		t.Fatalf("fixed {4} tax should not set a dynamic generic cost: %#v", pay.Payment)
	}
}
