package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// playLandsFromGraveyardPermanent gives playerID a battlefield permanent whose
// static ability lets that player play lands from their graveyard (Ramunap
// Excavator, Crucible of Worlds).
func playLandsFromGraveyardPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name:            "Test Excavator",
		StaticAbilities: []game.StaticAbility{game.PlayLandsFromGraveyardStaticBody},
	}})
}

func TestCanPlayLandFromGraveyardRequiresStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	landID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	if canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Graveyard) {
		t.Fatal("graveyard land is playable without the static permission")
	}

	playLandsFromGraveyardPermanent(g, game.Player1)

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Graveyard) {
		t.Fatal("graveyard land is not playable despite the static permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player2, landID, zone.Graveyard) {
		t.Fatal("opponent may play a land from a graveyard they do not control the static for")
	}
}

func TestApplyPlayLandFromGraveyardWithStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playLandsFromGraveyardPermanent(g, game.Player1)
	landID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Graveyard, game.FaceFront)) {
		t.Fatal("playing a land from the graveyard was rejected despite the static permission")
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want 1", g.Turn.LandsPlayedThisTurn)
	}
	if g.Players[game.Player1].Graveyard.Contains(landID) {
		t.Fatal("land remained in the graveyard after being played")
	}
}
