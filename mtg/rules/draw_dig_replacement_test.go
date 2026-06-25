package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func drawDigPermanent(g *game.Game, controller game.PlayerID, look, take int, remainder game.DigRemainder) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Digger",
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardDigReplacement("digger", look, take, remainder),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

// TestDrawCardDigReplacementReplacesDraw verifies that the controller's draw is
// replaced by looking at the top N cards, taking the chosen card into hand, and
// sending the rest to the graveyard.
func TestDrawCardDigReplacementReplacesDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Added bottom-to-top: c1 deepest, c3 top. peekLibrary sees c3, c2, c1.
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	drawDigPermanent(g, game.Player1, 3, 1, game.DigRemainderGraveyard)

	log := TurnLog{}
	// Seen order is [c3, c2, c1]; choosing index 1 takes the middle card c2.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.drawCards(g, game.Player1, 1, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) {
		t.Fatal("dig replacement did not put the chosen card into hand")
	}
	if player.Hand.Size() != 1 {
		t.Fatalf("hand size = %d, want 1 (only the dug card)", player.Hand.Size())
	}
	if !player.Graveyard.Contains(c1) || !player.Graveyard.Contains(c3) {
		t.Fatal("dig replacement did not send the unchosen cards to the graveyard")
	}
	if player.Library.Size() != 0 {
		t.Fatalf("library size = %d, want 0 (all three looked-at cards left the library)", player.Library.Size())
	}
}

// TestDrawCardDigReplacementOnlyHelpsController verifies the dig replacement does
// not alter an opponent's draw.
func TestDrawCardDigReplacementOnlyHelpsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 3 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawDigPermanent(g, game.Player1, 3, 1, game.DigRemainderGraveyard)

	log := TurnLog{}
	engine.drawCards(g, game.Player2, 1, [game.NumPlayers]PlayerAgent{}, &log)

	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("opponent hand size = %d, want 1 (controller-only replacement)", got)
	}
	if got := g.Players[game.Player2].Graveyard.Size(); got != 0 {
		t.Fatalf("opponent graveyard size = %d, want 0 (draw not replaced)", got)
	}
}
