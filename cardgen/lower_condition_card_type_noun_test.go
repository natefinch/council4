package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerClingToDustCardTypeNoun guards parseConditionCardTypeNoun with the
// real card Cling to Dust: "Exile target card from a graveyard. If it was a
// creature card, you gain 3 life. Otherwise, you draw a card." Before this
// slice, "creature card" failed to parse at all (parseConditionNoun required
// every remaining token to itself be a card-type atom, and the generic
// "card" noun is not one), so the whole per-effect gate -- and therefore the
// whole ordered sequence -- failed closed with a structural "condition
// unrecognized" diagnostic regardless of the no-trigger Target fallback
// slice 4 added.
func TestLowerClingToDustCardTypeNoun(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Cling to Dust Test",
		"Exile target card from a graveyard. If it was a creature card, you gain 3 life. Otherwise, you draw a card.")
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions (exile, gated gain, gated draw)", sequence)
	}
	gainLife := sequence[1]
	if _, ok := gainLife.Primitive.(game.GainLife); !ok {
		t.Fatalf("instruction[1] = %T, want game.GainLife", gainLife.Primitive)
	}
	gate := effectConditionMatch(t, gainLife)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetCard || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target card 0", gate.Object)
	}
	if !slices.Contains(gate.ObjectMatches.Val.RequiredTypes, types.Creature) {
		t.Fatalf("gate selection = %#v, want required type Creature", gate.ObjectMatches.Val)
	}
	if gate.Negate {
		t.Fatalf("gain-life gate = %#v, want the affirmative (non-negated) branch", gate)
	}
	draw := sequence[2]
	if _, ok := draw.Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[2] = %T, want game.Draw", draw.Primitive)
	}
	drawGate := effectConditionMatch(t, draw)
	if !drawGate.Negate {
		t.Fatalf("otherwise-branch draw gate = %#v, want the negated branch", drawGate)
	}
}

// TestLowerActivatedAbilityCardTypeNounBindsTarget guards the
// activationConditionOwnedByBody fix (cardgen/lower_activated.go) with the
// real card Kaseto, Orochi Archmage: "{G}{U}: Target creature can't be
// blocked this turn. If that creature is a Snake, it gets +2/+2 until end of
// turn." Before this slice's fix, prepareActivationCondition unconditionally
// tried to extract this ability's single condition as an "Activate only if"
// activation-time gate (the only recognized alternative was a narrower,
// Source-bound negated rider for a different card shape), which failed
// because the condition binds to Target, not Source, and Target references
// do not exist until the ability actually resolves -- exactly the kind of
// gate lowerCondition's activation context correctly refuses. The fix
// recognizes a Target-bound ObjectMatches condition contained within one of
// the body's own effect clauses as a per-effect resolution gate instead,
// leaving it in place for the ordinary ordered-sequence lowerer.
func TestLowerActivatedAbilityCardTypeNounBindsTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Kaseto, Orochi Archmage",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Snake Wizard",
		ManaCost:   "{1}{G}{U}",
		OracleText: "{G}{U}: Target creature can't be blocked this turn. If that creature is a Snake, it gets +2/+2 until end of turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want 1", face.ActivatedAbilities)
	}
	modes := face.ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 2 {
		t.Fatalf("modes = %#v, want one mode with two instructions", modes)
	}
	pump := modes[0].Sequence[1]
	if _, ok := pump.Primitive.(game.ModifyPT); !ok {
		t.Fatalf("instruction[1] = %T, want game.ModifyPT", pump.Primitive)
	}
	gate := effectConditionMatch(t, pump)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target permanent 0", gate.Object)
	}
	if !slices.Contains(gate.ObjectMatches.Val.SubtypesAny, types.Sub("Snake")) {
		t.Fatalf("gate selection = %#v, want subtype Snake", gate.ObjectMatches.Val)
	}
}

// TestLowerActivatedAbilitySourceBoundGateStaysActivationCondition guards
// against activationConditionOwnedByBody over-broadening: a Source-bound
// ObjectMatches condition on an activated add-mana ability must still route
// through isSemanticManaAbility (one of this function's two callers) to the
// mana-ability lowerer, which strips and reinterprets that same condition
// itself, rather than being treated as a body-owned per-effect gate here.
// Restricting the new Target-bound case to Target specifically is what keeps
// this working; broadening it to Source during development flipped
// isSemanticManaAbility to false for this exact shape and regressed
// TestLowerIncubationDruidCounterMultiplierMana (still unmodified, the
// primary regression guard for this).
func TestLowerActivatedAbilitySourceBoundGateStaysActivationCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Source Bound Gate Test",
		Layout:     "normal",
		TypeLine:   "Creature — Ooze",
		OracleText: "{G}: Add one mana of any color. If this creature has a +1/+1 counter on it, add two mana of that color instead.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %#v, want 1", face.ManaAbilities)
	}
}
