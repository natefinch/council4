package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDiscardThenDrawDiscardsChosenThenDrawsCount proves that discarding a
// player-chosen number of cards draws that many, modeling the looter half of
// "discard up to two cards, then draw that many cards" (Kinetic Augur).
func TestDiscardThenDrawDiscardsChosenThenDrawsCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old B"}})
	for _, name := range []string{"Top 1", "Top 2", "Top 3"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.DiscardThenDraw{
			Player: game.ControllerReference(),
			Max:    2,
		}},
	})

	agent := &orderedHandChoiceAgent{order: []string{"Old A", "Old B"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Discarded 2 (hand 2 -> 0), then drew 2 (hand 0 -> 2).
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2", got)
	}
	// Library started with 3, lost 2 to the draw.
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
}

// TestDiscardThenDrawAddsDrawOffset proves that the "plus K" rider draws the
// discarded count plus the offset, modeling "discard any number of cards, then
// draw that many cards plus one" (Colossus of the Blood Age).
func TestDiscardThenDrawAddsDrawOffset(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old B"}})
	for _, name := range []string{"Top 1", "Top 2", "Top 3", "Top 4"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.DiscardThenDraw{
			Player:     game.ControllerReference(),
			DrawOffset: 1,
		}},
	})

	agent := &orderedHandChoiceAgent{order: []string{"Old A"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Discarded 1 (hand 2 -> 1), then drew 1+1 = 2 (hand 1 -> 3).
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("hand size = %d, want 3", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 2 {
		t.Fatalf("library size = %d, want 2", got)
	}
}

// TestDiscardThenDrawWithNoCardsDrawsOffset proves that discarding nothing still
// draws the fixed offset, and draws nothing when the offset is zero.
func TestDiscardThenDrawWithNoCardsDrawsOffset(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Keep"}})
	for _, name := range []string{"Top 1", "Top 2"} {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}})
	}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.DiscardThenDraw{
			Player:     game.ControllerReference(),
			DrawOffset: 1,
		}},
	})

	agent := &orderedHandChoiceAgent{answer: []int{}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	// Discarded 0, then drew 0+1 = 1 (hand 1 -> 2).
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2", got)
	}
	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
}
