package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func lifeLossReplacementCardDef(multiplier, addend int, recipientOpponent, duringControllerTurn bool) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Bloodletter of Aclazotz",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.LifeLossReplacement(
				"If an opponent would lose life during your turn, they lose twice that much life instead.",
				multiplier,
				addend,
				recipientOpponent,
				duringControllerTurn,
			),
		},
	}}
}

func TestLifeLossReplacementDoublesOpponentLossDuringYourTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	addReplacementPermanent(t, g, game.Player1, lifeLossReplacementCardDef(2, 0, true, true))
	before := g.Players[game.Player2].Life

	if got := loseLife(g, game.Player2, 3); got != 6 {
		t.Fatalf("loseLife() = %d, want 6", got)
	}
	if got := before - g.Players[game.Player2].Life; got != 6 {
		t.Fatalf("life lost = %d, want 6", got)
	}
}

func TestLifeLossReplacementNotDoubledOnOpponentTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player2
	addReplacementPermanent(t, g, game.Player1, lifeLossReplacementCardDef(2, 0, true, true))

	if got := loseLife(g, game.Player2, 3); got != 3 {
		t.Fatalf("loseLife() on opponent turn = %d, want 3", got)
	}
}

func TestLifeLossReplacementSkipsControllerWhenOpponentOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	addReplacementPermanent(t, g, game.Player1, lifeLossReplacementCardDef(2, 0, true, false))

	if got := loseLife(g, game.Player1, 3); got != 3 {
		t.Fatalf("controller loseLife() = %d, want 3", got)
	}
}

func TestLifeLossReplacementAnyPlayerDoublesController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player2
	addReplacementPermanent(t, g, game.Player1, lifeLossReplacementCardDef(2, 0, false, false))

	if got := loseLife(g, game.Player1, 3); got != 6 {
		t.Fatalf("any-player loseLife() = %d, want 6", got)
	}
}
