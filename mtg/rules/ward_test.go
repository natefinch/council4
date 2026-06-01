package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestWardCountersSpellWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, mana.Cost{mana.GenericMana(1)})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter targeting spell", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestWardPaidLeavesSpellOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, mana.Cost{mana.ColoredMana(mana.Green)})
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !forest.Tapped {
		t.Fatal("ward cost did not tap mana source")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want targeting spell still on stack", g.Stack.Size())
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell moved to graveyard after ward was paid")
	}
}

func TestWardDoesNotTriggerForControllerTargetingOwnPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, mana.Cost{mana.GenericMana(1)})
	spellID := addCardToHand(g, game.Player2, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player2

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("self-targeting spell cast failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward triggered for controller's own spell")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only targeting spell", g.Stack.Size())
	}
}

func TestWardCountersActivatedAbilityWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addWardPermanent(g, game.Player2, mana.Cost{mana.GenericMana(1)})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Targeting Permanent",
		Types: []types.Card{types.Artifact},
		Abilities: []game.AbilityDef{{
			Kind:    game.ActivatedAbility,
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}},
		}},
	})
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0)) {
		t.Fatal("targeting activated ability failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want ward to counter activated ability", g.Stack.Size())
	}
}

func addWardPermanent(g *game.Game, controller game.PlayerID, cost mana.Cost) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:      "Ward Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(pt),
		Toughness: optPT(pt),
		Abilities: []game.AbilityDef{{
			Kind:     game.StaticAbility,
			Keywords: []game.Keyword{game.Ward},
			WardCost: optCost(cost),
		}},
	})
}
