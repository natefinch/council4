package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDynamicAmountTriggeringAttackerCountFiltersDeclarationBatch(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addSubtypedCreaturePermanent(g, game.Player1, types.Dinosaur)
	second := addSubtypedCreaturePermanent(g, game.Player1, types.Dinosaur)
	other := addSubtypedCreaturePermanent(g, game.Player1, types.Human)
	opponent := addSubtypedCreaturePermanent(g, game.Player2, types.Dinosaur)
	batch := g.IDGen.Next()
	for _, permanent := range []*game.Permanent{first, second, other, opponent} {
		g.AppendEvent(game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    permanent.ObjectID,
			Controller:     permanent.Controller,
			SimultaneousID: batch,
		})
	}
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    first.ObjectID,
			Controller:     game.Player1,
			SimultaneousID: batch,
		},
	}
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountTriggeringAttackerCount,
		Selection: &game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypesAny:   []types.Sub{types.Dinosaur},
		},
	}); got != 2 {
		t.Fatalf("Dinosaur attacker count = %d, want 2", got)
	}
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:      game.DynamicAmountTriggeringAttackerCount,
		Selection: &game.Selection{},
	}); got != 3 {
		t.Fatalf("controller attacker count = %d, want 3", got)
	}
}
