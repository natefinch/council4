package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// playLandsFromLibraryTopPermanent gives playerID a battlefield permanent whose
// static abilities let that player play lands from the top of their library and
// play with that top card revealed (Oracle of Mul Daya, Courser of Kruphix).
func playLandsFromLibraryTopPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Oracle",
		StaticAbilities: []game.StaticAbility{
			game.PlayWithTopCardRevealedStaticBody,
			game.PlayLandsFromLibraryTopStaticBody,
		},
	}})
}

func TestCanPlayLandFromLibraryTopRequiresStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	if canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Library) {
		t.Fatal("library land is playable without the static permission")
	}

	playLandsFromLibraryTopPermanent(g, game.Player1)

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Library) {
		t.Fatal("top library land is not playable despite the static permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player2, landID, zone.Library) {
		t.Fatal("opponent may play a land from a library they do not control the static for")
	}
}

func TestCanPlayLandFromLibraryTopOnlyTopCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	playLandsFromLibraryTopPermanent(g, game.Player1)
	// The first card added becomes the top; the second is added above it.
	buriedID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Mountain", Types: []types.Card{types.Land}}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, topID, zone.Library) {
		t.Fatal("top library land is not playable despite the static permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, buriedID, zone.Library) {
		t.Fatal("a non-top library land must not be playable")
	}
}

func TestApplyPlayLandFromLibraryTopWithStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playLandsFromLibraryTopPermanent(g, game.Player1)
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Library, game.FaceFront)) {
		t.Fatal("playing a land from the top of the library was rejected despite the static permission")
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want 1", g.Turn.LandsPlayedThisTurn)
	}
	if g.Players[game.Player1].Library.Contains(landID) {
		t.Fatal("land remained in the library after being played")
	}
}

func TestLibraryTopRevealedObservation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	obs := PlayerObservation{g: g, Player: game.Player2}
	if _, ok := obs.LibraryTopRevealed(game.Player1); ok {
		t.Fatal("top card revealed without the visibility static")
	}

	playLandsFromLibraryTopPermanent(g, game.Player1)

	view, ok := obs.LibraryTopRevealed(game.Player1)
	if !ok {
		t.Fatal("top card not revealed despite the visibility static")
	}
	if view.CardInstanceID != landID || view.Name != "Forest" {
		t.Fatalf("revealed top card = %+v, want Forest %d", view, landID)
	}
}
