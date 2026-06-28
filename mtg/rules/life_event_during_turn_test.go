package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func lifeDuringTurnPattern(event game.EventKind, player game.TriggerPlayerFilter, relation game.TriggerTurnRelation) *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:          event,
		Player:         player,
		CastDuringTurn: relation,
	}
}

func lifeCounterInstruction() []game.Instruction {
	return []game.Instruction{{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		},
	}}
}

func TestLifeGainedDuringYourTurnTriggerFiresOnlyOnYourTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1,
		lifeDuringTurnPattern(game.EventLifeGained, game.TriggerPlayerYou, game.TriggerTurnYours),
		lifeCounterInstruction(), nil)

	// Gaining life during an opponent's turn must not fire.
	g.Turn.ActivePlayer = game.Player2
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-your-turn life-gain trigger fired on an opponent's turn")
	}

	// Gaining life on the controller's own turn fires.
	g.Turn.ActivePlayer = game.Player1
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-your-turn life-gain trigger did not fire on the controller's own turn")
	}
}

func TestLifeLostDuringTheirTurnTriggerFiresOnlyOnThatOpponentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1,
		lifeDuringTurnPattern(game.EventLifeLost, game.TriggerPlayerOpponent, game.TriggerTurnNotYours),
		lifeCounterInstruction(), nil)

	// An opponent losing life during the controller's own turn must not fire.
	g.Turn.ActivePlayer = game.Player1
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-their-turn life-loss trigger fired on the controller's own turn")
	}

	// An opponent losing life during an opponent's turn fires.
	g.Turn.ActivePlayer = game.Player2
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-their-turn life-loss trigger did not fire on an opponent's turn")
	}
}
