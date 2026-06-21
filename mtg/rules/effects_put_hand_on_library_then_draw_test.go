package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestPutHandOnLibraryThenDrawDrawsCountPlusOffset proves that putting K cards
// from hand on the library draws K plus the offset, modeling Valakut Awakening's
// "put any number of cards from your hand on the bottom of your library, then
// draw that many cards plus one" effect.
func TestPutHandOnLibraryThenDrawDrawsCountPlusOffset(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old B"}})
	for _, name := range []string{"Top 1", "Top 2", "Top 3", "Top 4"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.PutHandOnLibraryThenDraw{
			Player:     game.ControllerReference(),
			Bottom:     true,
			DrawOffset: 1,
		}},
	})

	agent := &orderedHandChoiceAgent{order: []string{"Old A", "Old B"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Put 2 cards on the bottom (hand 2 -> 0), then draw 2+1 = 3.
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("hand size = %d, want 3", got)
	}
	// Library started with 4, gained 2 from hand, lost 3 to the draw.
	if got := g.Players[game.Player1].Library.Size(); got != 3 {
		t.Fatalf("library size = %d, want 3", got)
	}
}

// TestPutHandOnLibraryThenDrawWithNoCardsDrawsOffset proves that choosing no
// hand cards still draws the fixed offset.
func TestPutHandOnLibraryThenDrawWithNoCardsDrawsOffset(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Keep"}})
	for _, name := range []string{"Top 1", "Top 2"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.PutHandOnLibraryThenDraw{
			Player:     game.ControllerReference(),
			Bottom:     true,
			DrawOffset: 1,
		}},
	})

	agent := &orderedHandChoiceAgent{answer: []int{}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Kept the one card, put 0 on the library, then drew 0+1 = 1.
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
}
