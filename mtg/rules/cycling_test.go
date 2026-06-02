package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLegalActionsIncludesCyclingFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want cycling activation", legal)
	}
}

func TestCyclingDiscardsCardAndDrawsOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn Card"})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}

	if !forest.Tapped {
		t.Fatal("cycling mana cost did not tap available land")
	}
	if g.Players[game.Player1].Hand.Contains(cyclingID) {
		t.Fatal("cycled card remained in hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cyclingID) {
		t.Fatal("cycled card was not discarded to graveyard")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 cycling ability", g.Stack.Size())
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != cyclingID || obj.SourceCardID != cyclingID || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("cycling stack object = %+v, want activated ability sourced from cycled card", obj)
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.GameEvent) bool {
		return event.Player == game.Player1 &&
			event.CardID == cyclingID &&
			event.FromZone == game.ZoneHand &&
			event.ToZone == game.ZoneGraveyard
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("cycling did not draw the top card")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0", g.Stack.Size())
	}
}

func cyclingCard() *game.CardDef {
	cost := cost.Mana{cost.O(1)}
	return &game.CardDef{
		Name: "Cycling Test Card",
		Abilities: []game.AbilityDef{
			{
				Kind:     game.ActivatedAbility,
				Keywords: []game.Keyword{game.Cycling},
				ManaCost: optCost(cost),
				AdditionalCosts: []game.AdditionalCost{
					{Kind: game.AdditionalCostDiscard, Text: "Discard this card", Amount: 1, Zone: game.ZoneHand},
				},
				Effects: []game.Effect{
					{Type: game.EffectDraw, TargetIndex: game.TargetIndexController, Amount: 1},
				},
			},
		},
	}
}

func actionsContain(actions []action.Action, want action.Action) bool {
	for _, got := range actions {
		if actionsEqual(got, want) {
			return true
		}
	}
	return false
}
