package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addCreatureWithPowerToughness adds a vanilla creature permanent with distinct
// power and toughness so that a characteristic life rider's choice of
// characteristic (power vs toughness) is observable.
func addCreatureWithPowerToughness(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "PT Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
}

// TestSwordsToPlowsharesLifeRiderUsesLastKnownPower proves the Swords to
// Plowshares runtime shape: after the target creature is exiled, its controller
// gains life equal to the exiled creature's last-known power, read through the
// published linked object and the target's last-known controller.
func TestSwordsToPlowsharesLifeRiderUsesLastKnownPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreatureWithPowerToughness(g, game.Player2, 4, 7)
	before := g.Players[game.Player2].Life

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Swords"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	instrs := []game.Instruction{
		{Primitive: game.Exile{Object: game.TargetPermanentReference(0), ExileLinkedKey: "life-rider"}},
		{Primitive: game.GainLife{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectPower,
				Multiplier: 1,
				Object:     game.LinkedObjectReference("life-rider"),
			}),
			Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
		}},
	}
	log := TurnLog{}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target creature remained on the battlefield after exile")
	}
	if got := g.Players[game.Player2].Life - before; got != 4 {
		t.Fatalf("controller life gain = %d, want 4 (exiled creature's last-known power)", got)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("caster life = %d, want unchanged 40", got)
	}
}

// TestToughnessLifeRiderUsesLastKnownToughness proves the toughness sibling
// (Avenger en-Dal characteristic): the gained life equals the exiled creature's
// last-known toughness, not its power.
func TestToughnessLifeRiderUsesLastKnownToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreatureWithPowerToughness(g, game.Player2, 4, 7)
	before := g.Players[game.Player2].Life

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Avenger"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	instrs := []game.Instruction{
		{Primitive: game.Exile{Object: game.TargetPermanentReference(0), ExileLinkedKey: "life-rider"}},
		{Primitive: game.GainLife{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectToughness,
				Multiplier: 1,
				Object:     game.LinkedObjectReference("life-rider"),
			}),
			Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
		}},
	}
	log := TurnLog{}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}

	if got := g.Players[game.Player2].Life - before; got != 7 {
		t.Fatalf("controller life gain = %d, want 7 (exiled creature's last-known toughness)", got)
	}
}
