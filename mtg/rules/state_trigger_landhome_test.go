package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// sacrificeWhenControlNoIsland builds a creature carrying the runtime shape
// produced by lowering "When you control no Islands, sacrifice this creature."
// — a state-triggered ability (CR 603.8) whose condition is satisfied while the
// controller controls zero Islands and whose effect sacrifices the source.
func sacrificeWhenControlNoIsland(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addTriggeredPermanent(g, controller, &game.TriggerPattern{},
		[]game.Instruction{{Primitive: game.Sacrifice{Object: game.SourceCardPermanentReference()}}}, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		panic("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger = game.TriggerCondition{
		Type: game.TriggerState,
		State: opt.Val(game.StateTriggerCondition{
			Condition: opt.Val(game.Condition{
				Negate: true,
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{SubtypesAny: []types.Sub{types.Island}},
					MinCount:  1,
				}),
			}),
		}),
	}
	return permanent
}

func TestLandhomeStateTriggerSacrificesWhenControllerControlsNoIsland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := sacrificeWhenControlNoIsland(g, game.Player1)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger did not fire while controller controlled no Island")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source permanent remained on the battlefield after sacrifice")
	}
	if !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("sacrificed source did not move to its owner's graveyard")
	}
}

func TestLandhomeStateTriggerDormantWhileControllerControlsIsland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := sacrificeWhenControlNoIsland(g, game.Player1)
	addIslandPermanent(g, game.Player1)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger fired while controller still controlled an Island")
	}
	if _, ok := permanentByObjectID(g, source.ObjectID); !ok {
		t.Fatal("source permanent was removed despite controlling an Island")
	}
}

func TestLandhomeStateTriggerChecksControllerNotOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := sacrificeWhenControlNoIsland(g, game.Player1)
	// Only the opponent controls an Island; the controller still controls none.
	addIslandPermanent(g, game.Player2)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger did not fire when only the opponent controlled an Island")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("source survived despite its controller controlling no Island")
	}
}
