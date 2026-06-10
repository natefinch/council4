package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

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
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
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
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.CardID == cyclingID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	})
	assertEvent(t, g.Events, game.EventCycled, func(event game.Event) bool {
		return event.Controller == game.Player1 &&
			event.Player == game.Player1 &&
			event.CardID == cyclingID &&
			event.SourceID == cyclingID
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("cycling did not draw the top card")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0", g.Stack.Size())
	}
}

func TestCyclingEventTriggerFiresOnlyOnCycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycling Trigger Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCycled,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !discardCardFromHand(g, game.Player1, discardID) {
		t.Fatal("discardCardFromHand() = false")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycling trigger fired for ordinary discard")
	}

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycling trigger did not fire for cycling")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want cycling triggered ability", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want cycling trigger to draw one card", got)
	}
}

func TestCycleOrDiscardTriggerFiresForCyclingAndOrdinaryDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 1"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 2"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDiscarded,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !discardCardFromHand(g, game.Player1, discardID) {
		t.Fatal("discardCardFromHand() = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycle-or-discard trigger did not fire for ordinary discard")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycle-or-discard trigger did not fire for cycling discard")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want two trigger draws", got)
	}
}

func cyclingCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cycling Test Card",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		}},
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
