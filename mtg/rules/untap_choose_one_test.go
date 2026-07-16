package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestUntapChooseOneGroupPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	first.Tapped = true
	second.Tapped = true
	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.Untap{
		ChooseOne: true,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}}, [game.NumPlayers]PlayerAgent{
		game.Player1: scriptedChoiceAgent{answer: []int{1}},
	}, &TurnLog{})
	if !first.Tapped || second.Tapped {
		t.Fatalf("tapped states = %v/%v, want true/false", first.Tapped, second.Tapped)
	}
}
