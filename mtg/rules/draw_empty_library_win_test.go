package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestDrawFromEmptyLibraryWinsInsteadOfFailing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:                 "Lab Wizard",
			ReplacementAbilities: []game.ReplacementAbility{game.DrawFromEmptyLibraryWinReplacement("If you would draw a card while your library has no cards in it, you win the game instead.")},
		},
	}
	permanent := addCombatPermanent(g, game.Player1, def)
	registerPermanentReplacementEffects(g, permanent)

	if _, ok := engine.drawCard(g, game.Player1, false); ok {
		t.Fatal("drawCard() ok = true, want false for empty library")
	}
	if g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag set despite draw-from-empty win replacement")
	}
	for _, player := range g.Players {
		if player.ID == game.Player1 {
			if g.MarkedToLoseGame[player.ID] {
				t.Fatal("controller marked to lose instead of winning")
			}
			continue
		}
		if !g.MarkedToLoseGame[player.ID] {
			t.Fatalf("opponent %v not marked to lose when controller wins", player.ID)
		}
	}
}

func TestDrawFromEmptyLibraryWinDoesNotHelpOtherPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:                 "Lab Wizard",
			ReplacementAbilities: []game.ReplacementAbility{game.DrawFromEmptyLibraryWinReplacement("If you would draw a card while your library has no cards in it, you win the game instead.")},
		},
	}
	permanent := addCombatPermanent(g, game.Player1, def)
	registerPermanentReplacementEffects(g, permanent)

	if _, ok := engine.drawCard(g, game.Player2, false); ok {
		t.Fatal("drawCard() ok = true, want false for empty library")
	}
	if !g.FailedDraws[game.Player2] {
		t.Fatal("Player2 should fail the draw; the replacement is controlled by Player1")
	}
	if g.MarkedToLoseGame[game.Player1] {
		t.Fatal("Player1 marked to lose from an opponent's failed draw")
	}
}
