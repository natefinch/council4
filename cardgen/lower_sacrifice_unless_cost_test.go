package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
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

// TestLowerSacrificeSourceUnlessReturnToHandForms proves the broadened
// return-to-hand controller payment recognizes the untapped-subtype (Karoo
// cycle), "another" self-excluding (Faerie Impostor), and non-Lair excluded
// subtype (Lair cycle) wordings, each lowering to a return-to-hand additional
// cost that carries the matching object constraint.
func TestLowerSacrificeSourceUnlessReturnToHandForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
		check      func(t *testing.T, additional cost.Additional)
	}{
		{
			name:       "Karoo Bounce",
			typeLine:   "Land",
			oracleText: "When this land enters, sacrifice it unless you return an untapped Plains you control to its owner's hand.",
			check: func(t *testing.T, a cost.Additional) {
				if !a.RequireUntapped {
					t.Error("RequireUntapped = false, want true")
				}
				if a.SubtypesAny != (cost.SubtypeSet{types.Plains}) {
					t.Errorf("SubtypesAny = %v, want {Plains}", a.SubtypesAny)
				}
			},
		},
		{
			name:       "Lair Bounce",
			typeLine:   "Land — Lair",
			oracleText: "When this land enters, sacrifice it unless you return a non-Lair land you control to its owner's hand.",
			check: func(t *testing.T, a cost.Additional) {
				if a.ExcludeSubtype != types.Lair {
					t.Errorf("ExcludeSubtype = %v, want Lair", a.ExcludeSubtype)
				}
				if !a.MatchPermanentType || a.PermanentType != types.Land {
					t.Errorf("permanent type = (%v, %v), want (true, Land)", a.MatchPermanentType, a.PermanentType)
				}
			},
		},
		{
			// "return another creature you control" excludes the source itself;
			// the cost carries ExcludeSource so the runtime drops the source from
			// the eligible set (Faerie Impostor).
			name:       "Faerie Impostor",
			typeLine:   "Creature — Faerie",
			oracleText: "When this creature enters, sacrifice it unless you return another creature you control to its owner's hand.",
			check: func(t *testing.T, a cost.Additional) {
				if !a.ExcludeSource {
					t.Error("ExcludeSource = false, want true for the \"another\" return cost")
				}
				if !a.MatchPermanentType || a.PermanentType != types.Creature {
					t.Errorf("permanent type = (%v, %v), want (true, Creature)", a.MatchPermanentType, a.PermanentType)
				}
			},
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
			if additional.Kind != cost.AdditionalReturnToHand {
				t.Fatalf("additional cost kind = %v, want AdditionalReturnToHand", additional.Kind)
			}
			tc.check(t, additional)
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

// TestLowerSacrificeSourceUnlessReturnAnother proves the source-excluding
// ("another") return cost now lowers with ExcludeSource set, so the runtime
// drops the ability's own source permanent from the eligible set rather than
// letting the payer return the source itself to satisfy "return another creature
// you control." (Faerie Impostor, Quickling.)
func TestLowerSacrificeSourceUnlessReturnAnother(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bounce Faerie",
		Layout:     "normal",
		TypeLine:   "Creature — Faerie",
		OracleText: "When this creature enters, sacrifice it unless you return another creature you control to its owner's hand.",
	})
	additional := sacrificeUnlessCostSequence(t, face)
	if additional.Kind != cost.AdditionalReturnToHand {
		t.Fatalf("additional cost kind = %v, want AdditionalReturnToHand", additional.Kind)
	}
	if !additional.ExcludeSource {
		t.Error("ExcludeSource = false, want true")
	}
}
