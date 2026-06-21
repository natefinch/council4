package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func soulshiftSourcePermanent(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Hundred-Talon Kami",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Sub("Spirit")},
	}})
}

func spiritCardDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Sub("Spirit")},
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}}
}

// TestSoulshiftReturnsTargetedSpiritFromGraveyardToHand covers Soulshift N (CR
// 702.46): the dies trigger returns the targeted Spirit card from the
// controller's graveyard to their hand.
func TestSoulshiftReturnsTargetedSpiritFromGraveyardToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := soulshiftSourcePermanent(g)
	obj := triggeredObjFor(source)

	spirit := addCardToGraveyard(g, game.Player1, spiritCardDef("Kami", 3))
	obj.Targets = []game.Target{currentCardTarget(t, g, spirit)}

	instr := game.SoulshiftTriggeredAbility(4).Content.Modes[0].Sequence[0]
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, &instr, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(spirit) {
		t.Fatal("Soulshift did not return the targeted Spirit card to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(spirit) {
		t.Fatal("returned Spirit card still in graveyard")
	}
}

// TestSoulshiftTargetSpecBoundsSpiritByManaValue confirms the canonical template
// targets a Spirit card the controller owns with mana value N or less, the
// filter the generic target machinery enforces during target selection.
func TestSoulshiftTargetSpecBoundsSpiritByManaValue(t *testing.T) {
	t.Parallel()
	spec := game.SoulshiftTriggeredAbility(4).Content.Modes[0].Targets[0]
	if spec.Allow != game.TargetAllowCard {
		t.Fatalf("Allow = %v, want TargetAllowCard", spec.Allow)
	}
	if !spec.Selection.Exists {
		t.Fatal("target spec has no Selection")
	}
	selection := spec.Selection.Val
	if selection.Controller != game.ControllerYou {
		t.Fatalf("Controller = %v, want ControllerYou", selection.Controller)
	}
	if len(selection.SubtypesAny) != 1 || selection.SubtypesAny[0] != types.Sub("Spirit") {
		t.Fatalf("SubtypesAny = %v, want [Spirit]", selection.SubtypesAny)
	}
	wantBound := compare.Int{Op: compare.LessOrEqual, Value: 4}
	if !selection.ManaValue.Exists || selection.ManaValue.Val != wantBound {
		t.Fatalf("ManaValue = %+v, want %+v", selection.ManaValue, wantBound)
	}
}
