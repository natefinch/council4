package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func castDuringTurnPattern(relation game.TriggerTurnRelation) *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:          game.EventSpellCast,
		Controller:     game.TriggerControllerYou,
		CastDuringTurn: relation,
	}
}

func TestCastDuringOpponentTurnTriggerFiresOnlyOnOpponentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, castDuringTurnPattern(game.TriggerTurnNotYours), []game.Instruction{{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		},
	}}, nil)

	// The controller casting a spell on their own turn must not fire.
	g.Turn.ActivePlayer = game.Player1
	castSpellTargeting(g, game.Player1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-opponent-turn trigger fired on the controller's own turn")
	}

	// The controller casting a spell during an opponent's turn fires.
	g.Turn.ActivePlayer = game.Player2
	castSpellTargeting(g, game.Player1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-opponent-turn trigger did not fire on an opponent's turn")
	}
}

func TestCastDuringYourTurnTriggerFiresOnlyOnYourTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, castDuringTurnPattern(game.TriggerTurnYours), []game.Instruction{{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		},
	}}, nil)

	// The controller casting a spell during an opponent's turn must not fire.
	g.Turn.ActivePlayer = game.Player2
	castSpellTargeting(g, game.Player1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-your-turn trigger fired on an opponent's turn")
	}

	// The controller casting a spell on their own turn fires.
	g.Turn.ActivePlayer = game.Player1
	castSpellTargeting(g, game.Player1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("during-your-turn trigger did not fire on the controller's own turn")
	}
}
