package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func dashTestCreature() *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Dash Tester",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		AlternativeCosts: []cost.Alternative{{
			Label:    "Dash",
			Mechanic: cost.AlternativeMechanicDash,
			ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		}},
		TriggeredAbilities: []game.TriggeredAbility{
			game.DashTriggeredAbility(),
		},
	}}
}

// TestDashCastGrantsHasteAndReturnsAtEndStep proves casting for the Dash
// alternative cost marks the spell dashed, grants the creature haste on entry,
// and returns it to its owner's hand at the beginning of the next end step.
func TestDashCastGrantsHasteAndReturnsAtEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, dashTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Dash"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("dash cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Dashed {
		t.Fatalf("stack object = %#v, want Dashed spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("dashed creature is not on the battlefield after resolution")
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("dashed creature did not gain haste")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) != nil {
		t.Fatal("dashed creature stayed on the battlefield past the end step, want returned to hand")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("dashed creature was not returned to its owner's hand at the end step")
	}
}

// TestNormalCastDoesNotDash proves casting for the normal mana cost leaves the
// spell undashed, grants no haste, and never returns the creature to hand.
func TestNormalCastDoesNotDash(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, dashTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Normal cost"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Dashed {
		t.Fatalf("stack object = %#v, want non-dashed spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("normally cast creature is not on the battlefield after resolution")
	}
	if hasKeyword(g, permanent, game.Haste) {
		t.Fatal("normally cast creature gained haste, want summoning sick")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) == nil {
		t.Fatal("normally cast creature was returned to hand at the end step, want it to stay")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("normally cast creature was returned to its owner's hand, want it on the battlefield")
	}
}
