package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerTargetedGraveyardReturnPowerFilter covers the printed-power graveyard
// return filter ("Return target creature card with power 2 or less from your
// graveyard to your hand.", Graceful Restoration / Reveillark / Vesperlark). The
// fixed power bound rides the card-zone selection's Power comparison, which the
// runtime evaluates against the card's printed power in the graveyard.
func TestLowerTargetedGraveyardReturnPowerFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Power Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card with power 2 or less from your graveyard to your hand.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	selection := target.Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		selection.Controller != game.ControllerYou ||
		!selection.Power.Exists ||
		selection.Power.Val.Op != compare.LessOrEqual ||
		selection.Power.Val.Value != 2 ||
		selection.Toughness.Exists {
		t.Fatalf("selection = %#v", selection)
	}
}

// TestLowerTargetedGraveyardReturnToughnessFilter covers the printed-toughness
// graveyard return filter ("... with toughness 3 or greater ..."), the toughness
// analogue of the power bound.
func TestLowerTargetedGraveyardReturnToughnessFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Toughness Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target creature card with toughness 3 or greater from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if !selection.Toughness.Exists ||
		selection.Toughness.Val.Op != compare.GreaterOrEqual ||
		selection.Toughness.Val.Value != 3 ||
		selection.Power.Exists {
		t.Fatalf("selection = %#v", selection)
	}
}

// TestLowerTargetedGraveyardReturnSupertypeFilter covers the supertype graveyard
// return filter ("Return target legendary creature card from your graveyard to
// your hand."). The supertype rides the selection's Supertypes set, which the
// runtime requires the returned card's printed supertypes to contain.
func TestLowerTargetedGraveyardReturnSupertypeFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Legendary Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target legendary creature card from your graveyard to your hand.",
	})
	selection := face.SpellAbility.Val.Modes[0].Targets[0].Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		!slices.Equal(selection.Supertypes, []types.Super{types.Legendary}) {
		t.Fatalf("selection = %#v", selection)
	}
}

// TestLowerChosenGraveyardReturnSupertypeFilter confirms the same supertype rider
// rides the non-target chosen reanimation form ("Return a legendary creature card
// from your graveyard to the battlefield.").
func TestLowerChosenGraveyardReturnSupertypeFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chosen Legendary Reanimate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return a legendary creature card from your graveyard to the battlefield.",
	})
	if len(face.SpellAbility.Val.Modes[0].Targets) != 0 {
		t.Fatalf("chosen reanimation must not target: %#v", face.SpellAbility.Val.Modes[0].Targets)
	}
}

// TestLowerTargetedGraveyardReturnVariableXToHand covers the variable-count
// graveyard return "Return X target creature cards from your graveyard to your
// hand." (Death Denied). The spell's chosen X fixes how many cards return: the
// target spec binds its count to X with CountEqualsX over a 0..max range, and the
// sequence unrolls one per-index return instruction for every legal X.
func TestLowerTargetedGraveyardReturnVariableXToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Death Denied",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return X target creature cards from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one variable target spec", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != maxVariableRemovalTargets ||
		!target.CountEqualsX ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypes, []types.Card{types.Creature}) ||
		target.Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != maxVariableRemovalTargets {
		t.Fatalf("sequence length = %d, want %d", len(mode.Sequence), maxVariableRemovalTargets)
	}
	for i, instruction := range mode.Sequence {
		move, ok := instruction.Primitive.(game.MoveCard)
		if !ok {
			t.Fatalf("primitive %d = %T, want game.MoveCard", i, instruction.Primitive)
		}
		if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != i ||
			move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
			t.Fatalf("move %d = %#v", i, move)
		}
	}
}

// TestLowerTargetedGraveyardReturnVariableXToBattlefield confirms the variable-X
// form also reanimates to the battlefield ("Return X target creature cards with
// mana value 2 or less from your graveyard to the battlefield.", Return to the
// Ranks), keeping the per-card mana-value bound on the X-bound target spec.
func TestLowerTargetedGraveyardReturnVariableXToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Return To The Ranks",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return X target creature cards with mana value 2 or less from your graveyard to the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	selection := target.Selection.Val
	if target.MinTargets != 0 || target.MaxTargets != maxVariableRemovalTargets ||
		!target.CountEqualsX ||
		!slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		!selection.ManaValue.Exists ||
		selection.ManaValue.Val.Op != compare.LessOrEqual ||
		selection.ManaValue.Val.Value != 2 {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != maxVariableRemovalTargets {
		t.Fatalf("sequence length = %d, want %d", len(mode.Sequence), maxVariableRemovalTargets)
	}
	for i, instruction := range mode.Sequence {
		put, ok := instruction.Primitive.(game.PutOnBattlefield)
		if !ok {
			t.Fatalf("primitive %d = %T, want game.PutOnBattlefield", i, instruction.Primitive)
		}
		ref, ok := put.Source.CardRef()
		if !ok || ref.Kind != game.CardReferenceTarget || ref.TargetIndex != i {
			t.Fatalf("put %d source = %#v", i, put.Source)
		}
	}
}
