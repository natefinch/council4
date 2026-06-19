package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func scenarioCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

func scenarioOnBattlefield(g *game.Game, objectID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.ObjectID == objectID {
			return true
		}
	}
	return false
}

// TestScenarioLethalDamageDestroysCreature shows the builder placing a damaged
// creature and asserting the lethal-damage state-based action destroys it.
func TestScenarioLethalDamageDestroysCreature(t *testing.T) {
	s := newScenario(t)
	bear := s.permanent(game.Player1, scenarioCreature("Grizzly Bears", 2, 2)).damage(2)

	s.applyStateBasedActions()

	if scenarioOnBattlefield(s.game(), bear.permanent().ObjectID) {
		t.Error("a 2/2 with 2 marked damage should be destroyed by state-based actions")
	}
}

// TestScenarioCountersRaiseToughness shows counters keeping a creature alive: a
// +1/+1 counter lifts effective toughness above the marked damage.
func TestScenarioCountersRaiseToughness(t *testing.T) {
	s := newScenario(t)
	bear := s.permanent(game.Player1, scenarioCreature("Grizzly Bears", 2, 2)).
		counter(counter.PlusOnePlusOne, 1).
		damage(2)

	s.applyStateBasedActions()

	if !scenarioOnBattlefield(s.game(), bear.permanent().ObjectID) {
		t.Error("a 3/3 (after a +1/+1 counter) with 2 marked damage should survive")
	}
}

// TestScenarioZeroLifeLoses shows setting life and asserting the loss outcome.
func TestScenarioZeroLifeLoses(t *testing.T) {
	s := newScenario(t)
	s.life(game.Player2, 0)

	losses := s.applyStateBasedActions()

	found := false
	for _, loss := range losses {
		if loss.Player == game.Player2 {
			found = true
		}
	}
	if !found {
		t.Errorf("a player at 0 life should lose; losses = %+v", losses)
	}
}
