package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestRagavanCombatDamageCreatesTreasureAndImpulseExiles runs the real generated
// Ragavan, Nimble Pilferer card through the engine and proves its combat-damage
// trigger both creates a Treasure token under Ragavan's controller and exiles the
// top card of the damaged player's library with a cast-until-end-of-turn
// permission for Ragavan's controller.
func TestRagavanCombatDamageCreatesTreasureAndImpulseExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ragavan := addCombatPermanent(g, game.Player1, cards.RagavanNimblePilferer())

	// Player2 is the damaged player whose top library card Ragavan exiles.
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Damaged Player Spell",
		Types: []types.Card{types.Sorcery},
	}})

	dealPlayerDamage(g, ragavan.ObjectID, ragavan.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Ragavan combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	treasures := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 && permanentTokenName(permanent) == "Treasure" {
			treasures++
		}
	}
	if treasures != 1 {
		t.Fatalf("Treasure tokens under Ragavan's controller = %d, want 1", treasures)
	}

	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("the damaged player's top library card was not exiled")
	}
	if g.Players[game.Player2].Library.Contains(topID) {
		t.Fatal("the exiled card is still in the damaged player's library")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("Ragavan's controller was not granted permission to cast the exiled card")
	}
}
