package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// additionalLandStaticPermanent gives playerID a battlefield permanent whose
// static ability grants `count` extra land plays each turn (Exploration/Azusa).
func additionalLandStaticPermanent(g *game.Game, playerID game.PlayerID, count int) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Exploration",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                game.RuleEffectAdditionalLandPlays,
				AffectedPlayer:      game.PlayerYou,
				AdditionalLandPlays: count,
			}},
		}},
	}})
}

func TestAdditionalLandPlaysForCountsStaticGrants(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if got := additionalLandPlaysFor(g, game.Player1); got != 0 {
		t.Fatalf("baseline additional land plays = %d, want 0", got)
	}
	additionalLandStaticPermanent(g, game.Player1, 1)
	additionalLandStaticPermanent(g, game.Player1, 2)
	if got := additionalLandPlaysFor(g, game.Player1); got != 3 {
		t.Fatalf("granted additional land plays = %d, want 3", got)
	}
	if got := additionalLandPlaysFor(g, game.Player2); got != 0 {
		t.Fatalf("opponent additional land plays = %d, want 0", got)
	}
}

// eachPlayerAdditionalLandStaticPermanent gives playerID a permanent whose
// static grants `count` extra land plays to EVERY player (PlayerAny scope), as
// produced by the symmetric "Each player may play..." wording.
func eachPlayerAdditionalLandStaticPermanent(g *game.Game, playerID game.PlayerID, count int) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Rites of Flourishing",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                game.RuleEffectAdditionalLandPlays,
				AffectedPlayer:      game.PlayerAny,
				AdditionalLandPlays: count,
			}},
		}},
	}})
}

func TestAdditionalLandPlaysForEachPlayerGrantsEveryPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	eachPlayerAdditionalLandStaticPermanent(g, game.Player1, 1)
	if got := additionalLandPlaysFor(g, game.Player1); got != 1 {
		t.Fatalf("controller additional land plays = %d, want 1", got)
	}
	if got := additionalLandPlaysFor(g, game.Player2); got != 1 {
		t.Fatalf("opponent additional land plays = %d, want 1", got)
	}
}

func TestPlayerCanPlayLandRespectsAdditionalAllowance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.LandsAllowedThisTurn = 1
	g.Turn.LandsPlayedThisTurn = 1
	if playerCanPlayLand(g, game.Player1) {
		t.Fatal("player may play a second land without an allowance")
	}
	additionalLandStaticPermanent(g, game.Player1, 1)
	if !playerCanPlayLand(g, game.Player1) {
		t.Fatal("player cannot play a second land despite an additional-land grant")
	}
	g.Turn.LandsPlayedThisTurn = 2
	if playerCanPlayLand(g, game.Player1) {
		t.Fatal("player may exceed the granted allowance")
	}
}

func TestApplyActionPlaySecondLandWithStaticGrant(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	additionalLandStaticPermanent(g, game.Player1, 1)
	first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}}})
	second := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLand(first)) {
		t.Fatal("first land play was rejected")
	}
	if !engine.applyAction(g, game.Player1, action.PlayLand(second)) {
		t.Fatal("second land play was rejected despite an additional-land grant")
	}
	if g.Turn.LandsPlayedThisTurn != 2 {
		t.Fatalf("lands played = %d, want 2", g.Turn.LandsPlayedThisTurn)
	}
}
