package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLookAtHandLeavesTargetHandUnchanged(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	addEffectSpellToStack(
		g,
		game.Player1,
		game.LookAtHand{Player: game.TargetPlayerReference(0)},
		[]game.Target{game.PlayerTarget(game.Player2)},
	)

	engine.resolveTopOfStack(g, &TurnLog{})

	hand := g.Players[game.Player2].Hand
	if !hand.Contains(first) || !hand.Contains(second) {
		t.Fatal("looking at a hand must not move cards out of it")
	}
	if got := len(hand.All()); got != 2 {
		t.Fatalf("target hand size = %d, want 2 unchanged after look-at-hand", got)
	}
}
