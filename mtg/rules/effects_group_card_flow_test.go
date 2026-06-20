package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDrawGroupEffectDrawsForEveryPlayer proves "each player draws" gives a card
// to every player, mirroring the GainLife group path.
func TestDrawGroupEffectDrawsForEveryPlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount:      game.Fixed(1),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %d hand size = %d, want 1", playerID, got)
		}
	}
}

// TestDiscardGroupEffectDiscardsForOpponents proves "each opponent discards"
// only affects opponents, leaving the controller's hand intact.
func TestDiscardGroupEffectDiscardsForOpponents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Discard{
		Amount:      game.Fixed(1),
		PlayerGroup: game.OpponentsReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("controller hand size = %d, want 1 (unaffected)", got)
	}
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Hand.Size(); got != 0 {
			t.Fatalf("opponent %d hand size = %d, want 0", playerID, got)
		}
	}
}

// TestMillGroupEffectMillsForEveryPlayer proves "each player mills" moves cards
// from every player's library to their graveyard.
func TestMillGroupEffectMillsForEveryPlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Mill{
		Amount:      game.Fixed(2),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	}
	before := graveyardSizes(g)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Library.Size(); got != 0 {
			t.Fatalf("player %d library size = %d, want 0 (milled both)", playerID, got)
		}
	}
	// The controller's resolved sorcery also enters their graveyard, so the
	// exact milled-card count is asserted on the opponents.
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Graveyard.Size() - before[playerID]; got != 2 {
			t.Fatalf("player %d milled = %d, want 2", playerID, got)
		}
	}
}

func graveyardSizes(g *game.Game) [game.NumPlayers]int {
	var sizes [game.NumPlayers]int
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		sizes[playerID] = g.Players[playerID].Graveyard.Size()
	}
	return sizes
}

// TestGroupDrawThenDiscardSequenceResolvesInOrder proves an ordered group
// sequence ("each player draws two cards, then discards a card") resolves the
// draw before the discard: every player ends with one extra card and a
// graveyard card.
func TestGroupDrawThenDiscardSequenceResolvesInOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.Draw{Amount: game.Fixed(2), PlayerGroup: game.AllPlayersReference()}},
		{Primitive: game.Discard{Amount: game.Fixed(1), PlayerGroup: game.AllPlayersReference()}},
	})
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
		addCardToLibrary(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	}
	before := graveyardSizes(g)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %d hand size = %d, want 1 (drew 2, discarded 1)", playerID, got)
		}
	}
	// The controller's resolved sorcery also enters their graveyard, so the
	// exact discarded-card count is asserted on the opponents.
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Graveyard.Size() - before[playerID]; got != 1 {
			t.Fatalf("player %d discarded = %d, want 1", playerID, got)
		}
	}
}

// TestDiscardEntireHandEachPlayerEmptiesEveryHand proves "each player discards
// their hand" moves every card from each player's hand to their graveyard.
func TestDiscardEntireHandEachPlayerEmptiesEveryHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Discard{
		EntireHand:  true,
		PlayerGroup: game.AllPlayersReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "C"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Hand.Size(); got != 0 {
			t.Fatalf("player %d hand size = %d, want 0", playerID, got)
		}
	}
	// Players 2-4 graveyards hold exactly their three discarded cards (Player1's
	// graveyard additionally receives the resolved spell).
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Graveyard.Size(); got != 3 {
			t.Fatalf("player %d graveyard size = %d, want 3", playerID, got)
		}
	}
}

// TestDiscardEntireHandControllerEmptiesOnlyControllerHand proves "Discard your
// hand" empties the controller's hand and leaves opponents untouched.
func TestDiscardEntireHandControllerEmptiesOnlyControllerHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Discard{
		EntireHand: true,
		Player:     game.ControllerReference(),
	}, nil)
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2} {
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
		addCardToHand(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("controller hand size = %d, want 0", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 2 {
		t.Fatalf("opponent hand size = %d, want 2 (unaffected)", got)
	}
}
