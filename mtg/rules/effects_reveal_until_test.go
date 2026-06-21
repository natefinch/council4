package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestRevealUntilTargetPlayerLandToGraveyard proves the Undercity Informer shape
// reveals from the top of a target player's library until a land is revealed,
// then puts every revealed card (including the land) into that player's
// graveyard, leaving the rest of the library intact.
func TestRevealUntilTargetPlayerLandToGraveyard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.RevealUntil{
		Player:      game.TargetPlayerReference(0),
		Until:       game.Selection{RequiredTypes: []types.Card{types.Land}},
		Destination: zone.Graveyard,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	// Library top-to-bottom: Spell, Spell, Forest (land), Keep.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Keep"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Spell2"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Spell1"}})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
	if got := g.Players[game.Player2].Graveyard.Size(); got != 3 {
		t.Fatalf("graveyard size = %d, want 3", got)
	}
}

// TestRevealUntilControllerLandToHand proves the Treasure Hunt shape reveals
// from the top of the controller's own library until a land is revealed, then
// puts every revealed card into the controller's hand.
func TestRevealUntilControllerLandToHand(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.RevealUntil{
		Player:      game.ControllerReference(),
		Until:       game.Selection{RequiredTypes: []types.Card{types.Land}},
		Destination: zone.Hand,
	}, nil)

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Keep"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spell"}})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Library.Size(); got != 1 {
		t.Fatalf("library size = %d, want 1", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2", got)
	}
}

// TestRevealUntilEmptyLibraryMovesAll proves a reveal-until with no matching
// card empties the library into the destination instead of looping forever.
func TestRevealUntilEmptyLibraryMovesAll(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.RevealUntil{
		Player:      game.TargetPlayerReference(0),
		Until:       game.Selection{RequiredTypes: []types.Card{types.Land}},
		Destination: zone.Graveyard,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Spell2"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Spell1"}})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Library.Size(); got != 0 {
		t.Fatalf("library size = %d, want 0", got)
	}
	if got := g.Players[game.Player2].Graveyard.Size(); got != 2 {
		t.Fatalf("graveyard size = %d, want 2", got)
	}
}
