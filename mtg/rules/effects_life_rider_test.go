package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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

// addPermanentWithManaCost adds a vanilla permanent whose printed mana cost
// fixes its mana value, so a characteristic life rider that reads mana value is
// observable after the permanent leaves the battlefield.
func addPermanentWithManaCost(g *game.Game, controller game.PlayerID, manaValue int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "MV Permanent",
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}})
}

// TestFeedTheSwarmLifeRiderUsesLastKnownManaValue proves the Feed the Swarm
// runtime shape: after the target permanent is destroyed, its destroyer's
// controller loses life equal to the destroyed permanent's last-known mana
// value, read through the published target reference and the spell's controller.
func TestFeedTheSwarmLifeRiderUsesLastKnownManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addPermanentWithManaCost(g, game.Player2, 5)
	before := g.Players[game.Player1].Life

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Feed the Swarm"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	instrs := []game.Instruction{
		{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.LoseLife{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectManaValue,
				Multiplier: 1,
				Object:     game.TargetPermanentReference(0),
			}),
			Player: game.ControllerReference(),
		}},
	}
	log := TurnLog{}
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &log)
	}

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target permanent remained on the battlefield after destroy")
	}
	if got := before - g.Players[game.Player1].Life; got != 5 {
		t.Fatalf("controller life loss = %d, want 5 (destroyed permanent's last-known mana value)", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("target controller life = %d, want unchanged 40", got)
	}
}
