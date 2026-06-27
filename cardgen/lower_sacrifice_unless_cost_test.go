package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

// sacrificeUnlessCostSequence asserts a single triggered ability whose body is
// the gated "Pay <additional cost>, otherwise sacrifice the source" sequence the
// "sacrifice <source> unless you <non-mana cost>" wording lowers to, and returns
// the resolution payment's lone additional cost for cost-specific assertions.
func sacrificeUnlessCostSequence(t *testing.T, face loweredFaceAbilities) cost.Additional {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(content.Modes))
	}
	seq := content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first instruction primitive = %T, want game.Pay", seq[0].Primitive)
	}
	if pay.Payment.ManaCost.Exists {
		t.Fatal("Pay instruction has a mana cost, want only an additional cost")
	}
	if len(pay.Payment.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %d, want 1", len(pay.Payment.AdditionalCosts))
	}
	if seq[0].PublishResult == "" {
		t.Error("Pay instruction does not publish a result")
	}
	if _, ok := seq[1].Primitive.(game.Sacrifice); !ok {
		t.Fatalf("second instruction primitive = %T, want game.Sacrifice", seq[1].Primitive)
	}
	gate := seq[1].ResultGate
	if !gate.Exists {
		t.Fatal("Sacrifice instruction is not gated on the payment result")
	}
	if gate.Val.Key != seq[0].PublishResult {
		t.Errorf("gate key = %v, want %v (the Pay result)", gate.Val.Key, seq[0].PublishResult)
	}
	if gate.Val.Succeeded != game.TriFalse {
		t.Errorf("gate Succeeded = %v, want TriFalse", gate.Val.Succeeded)
	}
	return pay.Payment.AdditionalCosts[0]
}

// TestLowerSacrificeSourceUnlessNonManaCost proves the non-mana "unless you
// <cost>" controller payment forms lower to the gated Pay/Sacrifice sequence,
// each carrying the matching additional cost. These broaden the previously
// mana-only "sacrifice <source> unless you pay {mana}" path to the discard,
// sacrifice-another, exile, and return-to-hand resolution payments.
func TestLowerSacrificeSourceUnlessNonManaCost(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
		wantKind   cost.AdditionalKind
	}{
		{
			name:       "Discard Imp",
			typeLine:   "Creature — Imp",
			oracleText: "At the beginning of your upkeep, sacrifice this creature unless you discard a card.",
			wantKind:   cost.AdditionalDiscard,
		},
		{
			name:       "Sacrifice Hound",
			typeLine:   "Creature — Hound",
			oracleText: "When this creature enters, sacrifice it unless you sacrifice another creature.",
			wantKind:   cost.AdditionalSacrifice,
		},
		{
			name:       "Graveyard Giant",
			typeLine:   "Creature — Giant",
			oracleText: "Whenever this creature attacks or blocks, sacrifice it unless you exile a card from your graveyard.",
			wantKind:   cost.AdditionalExile,
		},
		{
			name:       "Bounce Elephant",
			typeLine:   "Creature — Elephant",
			oracleText: "When this creature enters, sacrifice it unless you return two Forests you control to their owner's hand.",
			wantKind:   cost.AdditionalReturnToHand,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracleText,
			})
			additional := sacrificeUnlessCostSequence(t, face)
			if additional.Kind != tc.wantKind {
				t.Errorf("additional cost kind = %v, want %v", additional.Kind, tc.wantKind)
			}
		})
	}
}

// TestLowerSacrificeSourceUnlessNonManaCostFailsClosed keeps the gate strict:
// the trailing payment must be a single non-mana controller cost. A second
// genuine effect after the gated sacrifice must not be folded into the cost, so
// the multi-effect wording still falls through to ordered-sequence lowering
// rather than silently dropping the trailing effect.
func TestLowerSacrificeSourceUnlessNonManaCostFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Overreach Ogre",
		Layout:     "normal",
		TypeLine:   "Creature — Ogre",
		OracleText: "When this creature enters, sacrifice it unless you discard a card. Draw a card.",
	})
}
