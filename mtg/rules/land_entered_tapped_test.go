package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestRecordLandEnteredTappedFlagsTappedLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	tapped := addLandPermanent(g, game.Player1, "Tapland")
	tapped.Tapped = true
	untapped := addLandPermanent(g, game.Player1, "Untapland")

	tappedLog := &ActionLog{Player: game.Player1}
	recordLandEnteredTapped(g, tappedLog, action.PlayLandFace(tapped.CardInstanceID, game.FaceFront))
	if !tappedLog.LandEnteredTapped {
		t.Fatal("a land that entered tapped should set LandEnteredTapped")
	}

	untappedLog := &ActionLog{Player: game.Player1}
	recordLandEnteredTapped(g, untappedLog, action.PlayLandFace(untapped.CardInstanceID, game.FaceFront))
	if untappedLog.LandEnteredTapped {
		t.Fatal("a land that entered untapped should not set LandEnteredTapped")
	}
}

func TestRecordLandEnteredTappedIgnoresNonLandActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	log := &ActionLog{Player: game.Player1}
	recordLandEnteredTapped(g, log, action.Pass())
	if log.LandEnteredTapped {
		t.Fatal("a non-play-land action should not set LandEnteredTapped")
	}
}
