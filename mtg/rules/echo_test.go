package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// addEchoPermanent puts an Echo permanent controlled by controller onto the
// battlefield with the generic Echo triggered ability from the game template.
func addEchoPermanent(g *game.Game, controller game.PlayerID, manaCost cost.Mana) *game.Permanent {
	ability := game.EchoTriggeredAbility(manaCost)
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Echo Permanent",
		Types:              []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{ability},
	}}
	return addCombatPermanent(g, controller, def)
}

func TestEchoTriggersFirstUpkeepAndPaysToKeep(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
	land := addBasicLandPermanent(g, game.Player1, types.Island)

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo trigger was not put on the stack on the first upkeep")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := g.PermanentByID(source.ObjectID); !ok {
		t.Fatal("paid echo permanent was sacrificed")
	}
	if !land.Tapped {
		t.Fatal("echo payment did not tap a land")
	}
	if !source.EchoResolvedController.Exists || source.EchoResolvedController.Val != game.Player1 {
		t.Fatalf("EchoResolvedController = %+v; want Player1 recorded", source.EchoResolvedController)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay resolution cost?" {
		t.Fatalf("choices = %+v; want single payment prompt", log.Choices)
	}
}

func TestEchoDoesNotTriggerOnLaterUpkeeps(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
	addBasicLandPermanent(g, game.Player1, types.Island)

	// First upkeep: pay to keep, recording the resolved controller.
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo trigger was not put on the stack on the first upkeep")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})
	if _, ok := g.PermanentByID(source.ObjectID); !ok {
		t.Fatal("source should survive the first paid upkeep")
	}

	// Second upkeep: the obligation is resolved for Player1, so nothing triggers.
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo re-triggered on a later upkeep without a control change")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d; want empty on later upkeep", g.Stack.Size())
	}
}

func TestEchoSacrificesOnDeclineOrInability(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		lands       int
		choices     [][]int
		wantChoices int
	}{
		{name: "controller declines", lands: 1, choices: [][]int{{0}}, wantChoices: 1},
		{name: "controller cannot pay", lands: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
			for i := 0; i < test.lands; i++ {
				addBasicLandPermanent(g, game.Player1, types.Island)
			}

			emitBeginningOfStepEvent(g, game.StepUpkeep)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("echo trigger was not put on the stack")
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: test.choices},
			}, &log)

			if _, ok := g.PermanentByID(source.ObjectID); ok {
				t.Fatal("unpaid echo permanent remains on the battlefield")
			}
			if got := g.Players[game.Player1].Graveyard.Size(); got != 1 {
				t.Fatalf("graveyard size = %d; want 1", got)
			}
			if len(log.Choices) != test.wantChoices {
				t.Fatalf("choices = %+v; want %d", log.Choices, test.wantChoices)
			}
			assertEvent(t, g.Events, game.EventPermanentSacrificed, func(event game.Event) bool {
				return event.PermanentID == source.ObjectID
			})
		})
	}
}

func TestEchoRetriggersAfterControlChange(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
	// The obligation has already been resolved for its original controller.
	source.EchoResolvedController = opt.Val(game.Player1)

	// Player2 steals the permanent and reaches their upkeep.
	source.Controller = game.Player2
	g.Turn.ActivePlayer = game.Player2
	addBasicLandPermanent(g, game.Player2, types.Island)

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo did not re-trigger for the new controller after a control change")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Controller != game.Player2 {
		t.Fatalf("echo trigger controller = %+v; want Player2", top)
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})

	if _, ok := g.PermanentByID(source.ObjectID); !ok {
		t.Fatal("new controller paid echo but permanent was sacrificed")
	}
	if source.EchoResolvedController.Val != game.Player2 {
		t.Fatalf("EchoResolvedController = %+v; want Player2 recorded after resolution", source.EchoResolvedController)
	}
}

func TestEchoFizzlesWhenSourceLeavesBeforeResolution(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
	addBasicLandPermanent(g, game.Player1, types.Island)

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo trigger was not put on the stack")
	}
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("failed to move echo source out of play")
	}

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v; want no payment prompt for a missing source", log.Choices)
	}
}

func TestEchoBlinkYieldsFreshObligation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo trigger was not put on the stack")
	}
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("failed to exile echo source")
	}
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok {
		t.Fatal("failed to return echo source")
	}
	if returned.ObjectID == source.ObjectID {
		t.Fatal("blink preserved object identity")
	}

	// The old object's trigger fizzles: its source is gone, so the returned
	// permanent is not sacrificed by it.
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if _, ok := g.PermanentByID(returned.ObjectID); !ok {
		t.Fatal("returned permanent was sacrificed by the old object's trigger")
	}
	if returned.EchoResolvedController.Exists {
		t.Fatalf("returned permanent already has a recorded echo controller: %+v", returned.EchoResolvedController)
	}

	// The freshly returned object owes echo again on the next upkeep.
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("returned permanent did not owe a fresh echo on its next upkeep")
	}
}

func TestEchoMultiplePermanentsEachTrigger(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})
	second := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(1)})

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d; want 2 echo triggers", got)
	}

	log := TurnLog{}
	// Neither controller pays; both are sacrificed.
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if _, ok := g.PermanentByID(first.ObjectID); ok {
		t.Fatal("first echo permanent should have been sacrificed")
	}
	if _, ok := g.PermanentByID(second.ObjectID); ok {
		t.Fatal("second echo permanent should have been sacrificed")
	}
}

func TestEchoZeroCostIsPayableAndKeepsPermanent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEchoPermanent(g, game.Player1, cost.Mana{cost.O(0)})

	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("echo trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}, &TurnLog{})

	if _, ok := g.PermanentByID(source.ObjectID); !ok {
		t.Fatal("echo {0} permanent was sacrificed despite a free, payable cost")
	}
}
