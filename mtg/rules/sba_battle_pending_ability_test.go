package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestBattleZeroDefenseWaitsForPendingAbility verifies CR 704.5v: a battle with
// 0 defense is not put into its graveyard while it is the source of a triggered
// ability still on the stack, and is removed once that ability leaves.
func TestBattleZeroDefenseWaitsForPendingAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	battle := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Battle",
		Types:   []types.Card{types.Battle},
		Defense: opt.Val(0),
	}})
	// Battle is at 0 defense but is the source of a triggered ability on the stack.
	g.Stack.Push(&game.StackObject{
		ID:       g.IDGen.Next(),
		Kind:     game.StackTriggeredAbility,
		SourceID: battle.ObjectID,
	})

	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, battle.ObjectID); !ok {
		t.Fatal("battle was put into graveyard while its triggered ability was still on the stack")
	}

	// Remove the pending ability; now the SBA applies.
	g.Stack.RemoveByID(g.Stack.Objects()[0].ID)
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, battle.ObjectID); ok {
		t.Fatal("zero-defense battle remained after its triggered ability left the stack")
	}
}
