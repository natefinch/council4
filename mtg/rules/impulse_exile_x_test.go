package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestImpulseExileVariableXExilesXCards proves that an impulse exile whose
// amount is the spell's {X} ("Exile the top X cards of your library…", Commune
// with Lava, Hugs) exiles exactly the chosen value of X off the top of the
// controller's library, leaving the rest in place.
func TestImpulseExileVariableXExilesXCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Library order: first added is on the bottom, last added is on top.
	bottomID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bottom Card",
		Types: []types.Card{types.Creature},
	}})
	midID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Middle Card",
		Types: []types.Card{types.Creature},
	}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Top Card",
		Types: []types.Card{types.Creature},
	}})

	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		XValue:       2,
	}, game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
		Duration: game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	exile := g.Players[game.Player1].Exile
	if !exile.Contains(topID) || !exile.Contains(midID) {
		t.Fatalf("top two cards were not exiled (exile size %d)", exile.Size())
	}
	if exile.Contains(bottomID) {
		t.Fatal("third card was exiled; only X=2 cards should leave the library")
	}
	if !g.Players[game.Player1].Library.Contains(bottomID) {
		t.Fatal("bottom card should remain in the library")
	}
}
