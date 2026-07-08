package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerResolutionOrdinalGate proves the "if this is the Nth time this
// ability has resolved this turn" clause lowers onto the convert instruction as
// an effect-gate Condition carrying the ordinal, and flags the enclosing
// triggered ability so the runtime tallies its resolutions. This backs Prowl,
// Pursuit Vehicle's back-face enters trigger. The leading "has resolved this
// turn" wording must not seed a spurious keyword-grant effect, so the ability
// lowers to exactly the counter and the gated convert.
func TestLowerResolutionOrdinalGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Resolver",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact — Vehicle",
		OracleText: "Whenever another creature or Vehicle you control enters, put a +1/+1 counter on Test Resolver. If this is the second time this ability has resolved this turn, convert Test Resolver.",
		Power:      new("2"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if !ability.CountsResolutionsThisTurn {
		t.Fatal("CountsResolutionsThisTurn = false, want true so the runtime tallies resolutions")
	}
	sequence := ability.Content.Modes[0].Sequence
	var gatedTransforms int
	var sawTransform bool
	for i := range sequence {
		if _, ok := sequence[i].Primitive.(game.Transform); !ok {
			continue
		}
		sawTransform = true
		gate := sequence[i].Condition
		if !gate.Exists || !gate.Val.Condition.Exists {
			t.Fatalf("convert instruction %d has no effect-gate condition", i)
		}
		if got := gate.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn; got != 2 {
			t.Fatalf("convert gate ordinal = %d, want 2", got)
		}
		gatedTransforms++
	}
	if !sawTransform {
		t.Fatal("no convert (Transform) instruction lowered")
	}
	if gatedTransforms != 1 {
		t.Fatalf("gated convert instructions = %d, want 1", gatedTransforms)
	}
}
