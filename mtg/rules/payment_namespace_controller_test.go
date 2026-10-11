package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledPaymentDoesNotStealControllerResult(t *testing.T) {
	def := compileCounterTaxCard(t, "Controller Result Beside Payment",
		"Counter target spell unless its controller pays {3}. You gain 1 life. If you do, draw a card.", "Instant")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibrary(g, game.Player1, vanillaCreature("Draw", 1, 1))
	target := addStackSpell(g, game.Player2, "Target", []types.Card{types.Sorcery})
	pushCounterTaxSpell(g, def, target)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("controller result was incorrectly gated on another player's failed payment")
	}
	if _, remains := stackObjectByID(g, target.ID); remains {
		t.Fatal("unpayable target was not countered")
	}
}
