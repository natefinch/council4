package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// leylineOfTheVoidPermanent puts a real Leyline of the Void onto the battlefield
// under controller and registers its continuous graveyard-redirect replacement.
func leylineOfTheVoidPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.LeylineOfTheVoid())
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

// TestLeylineOfTheVoidExilesOpponentCardsFromAnyZone proves the real card's
// "If a card would be put into an opponent's graveyard from anywhere, exile it
// instead." replacement redirects an opponent's card from hand, library, or the
// battlefield to exile, while sparing the controller's own graveyard.
func TestLeylineOfTheVoidExilesOpponentCardsFromAnyZone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	leylineOfTheVoidPermanent(g, game.Player1)

	// An opponent's card discarded from hand is exiled instead of hitting the
	// graveyard.
	fromHand := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	if !moveCardBetweenZones(g, game.Player2, fromHand, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones(hand->graveyard) = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(fromHand) {
		t.Fatal("opponent's discarded card reached the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(fromHand) {
		t.Fatal("opponent's discarded card was not exiled")
	}

	// "From anywhere" includes a permanent dying from the battlefield.
	dying := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Doomed Beast", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, dying, zone.Graveyard) {
		t.Fatal("movePermanentToZone(battlefield->graveyard) = false, want true")
	}
	if !g.Players[game.Player2].Exile.Contains(dying.CardInstanceID) {
		t.Fatal("opponent's dying permanent was not exiled")
	}

	// A milled card from the library is likewise exiled.
	fromLibrary := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Milled"}})
	if !moveCardBetweenZones(g, game.Player2, fromLibrary, zone.Library, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones(library->graveyard) = false, want true")
	}
	if !g.Players[game.Player2].Exile.Contains(fromLibrary) {
		t.Fatal("opponent's milled card was not exiled")
	}
}

// TestLeylineOfTheVoidSparesControllerGraveyard proves the opponent scope: the
// controller's own cards still reach their graveyard normally.
func TestLeylineOfTheVoidSparesControllerGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	leylineOfTheVoidPermanent(g, game.Player1)

	ownCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "My Card"}})
	if !moveCardBetweenZones(g, game.Player1, ownCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones(hand->graveyard) = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCard) {
		t.Fatal("controller's own card was wrongly exiled by its own Leyline of the Void")
	}
	if g.Players[game.Player1].Exile.Contains(ownCard) {
		t.Fatal("controller's own card was exiled")
	}
}

// TestLeylineOfTheVoidAffectsEveryOpponentInMultiplayer proves the opponent
// scope is relative to the controller, so every other player's graveyard is
// watched, not just one.
func TestLeylineOfTheVoidAffectsEveryOpponentInMultiplayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	leylineOfTheVoidPermanent(g, game.Player1)

	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		card := addCardToHand(g, opponent, &game.CardDef{CardFace: game.CardFace{Name: "Doomed"}})
		if !moveCardBetweenZones(g, opponent, card, zone.Hand, zone.Graveyard) {
			t.Fatalf("moveCardBetweenZones for %v = false, want true", opponent)
		}
		if !g.Players[opponent].Exile.Contains(card) {
			t.Fatalf("opponent %v card was not exiled", opponent)
		}
	}
}
