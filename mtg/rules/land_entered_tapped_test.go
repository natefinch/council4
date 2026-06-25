package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestRecordLandEnteredTappedFlagsTappedLand checks the flag is written to the
// log entry (not the pre-copy local action log).
func TestRecordLandEnteredTappedFlagsTappedLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	tapped := addLandPermanent(g, game.Player1, "Tapland")
	tapped.Tapped = true
	untapped := addLandPermanent(g, game.Player1, "Untapland")

	log := &TurnLog{}
	log.addAction(&ActionLog{Player: game.Player1, Action: action.PlayLandFace(tapped.CardInstanceID, game.FaceFront)})
	recordLandEnteredTapped(g, log, lastEntryIndex(log), action.PlayLandFace(tapped.CardInstanceID, game.FaceFront))
	if !log.Entries[0].Action.LandEnteredTapped {
		t.Fatal("a land that entered tapped should set LandEnteredTapped on its log entry")
	}

	log.addAction(&ActionLog{Player: game.Player1, Action: action.PlayLandFace(untapped.CardInstanceID, game.FaceFront)})
	recordLandEnteredTapped(g, log, lastEntryIndex(log), action.PlayLandFace(untapped.CardInstanceID, game.FaceFront))
	if log.Entries[1].Action.LandEnteredTapped {
		t.Fatal("a land that entered untapped should not set LandEnteredTapped")
	}
}

// TestGoldfishFlagsTaplandDropEndToEnd plays a deck of real enters-tapped lands
// through the engine and asserts a play-land action entry is flagged as having
// entered tapped, proving the flag survives addAction's copy.
func TestGoldfishFlagsTaplandDropEndToEnd(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	tapland := &game.CardDef{CardFace: game.CardFace{
		Name:  "Tapped Land",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("Tapped Land enters the battlefield tapped."),
		},
	}}
	config := game.PlayerConfig{
		Name:      "Goldfish",
		Commander: commander,
		Deck:      repeatedCard(tapland, 99),
	}

	engine := NewEngine(rand.New(rand.NewPCG(1, 2)))
	g := engine.NewGoldfishGame(config)
	result := engine.RunGoldfish(g, landPlayingAgent{}, 5)

	sawTappedLandDrop := false
	for _, turn := range result.Turns {
		for _, entry := range turn.Entries {
			if entry.Kind != TurnLogEntryAction || entry.Action.Action.Kind != action.ActionPlayLand {
				continue
			}
			if !entry.Action.LandEnteredTapped {
				t.Fatal("a tapland drop was not flagged as entering tapped")
			}
			sawTappedLandDrop = true
		}
	}
	if !sawTappedLandDrop {
		t.Fatal("no land was played across the goldfish run")
	}
}
