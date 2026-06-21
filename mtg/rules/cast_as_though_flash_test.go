package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// flashTimingStaticPermanent gives playerID a battlefield permanent whose static
// ability lets them cast spells as though they had flash.
func flashTimingStaticPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Flash Permission",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer: game.PlayerYou,
			}},
		}},
	}})
}

func TestPlayerCanCastAsThoughFlashHonorsRuleEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if playerCanCastAsThoughFlash(g, game.Player1) {
		t.Fatal("player has flash timing without any rule effect")
	}
	flashTimingStaticPermanent(g, game.Player1)
	if !playerCanCastAsThoughFlash(g, game.Player1) {
		t.Fatal("player lacks flash timing despite an active rule effect")
	}
	if playerCanCastAsThoughFlash(g, game.Player2) {
		t.Fatal("opponent gained flash timing from the controller's rule effect")
	}
}

// TestCanCastAtCurrentTimingAllowsSorceryWithFlashPermission proves the timing
// permission lets a sorcery-speed card be cast at instant speed, while leaving
// it sorcery-speed for a player without the permission.
func TestCanCastAtCurrentTimingAllowsSorceryWithFlashPermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Make it not the active player's main phase with an empty stack, so sorcery
	// speed is unavailable to anyone by default.
	g.Turn.ActivePlayer = game.Player2
	sorcery := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Sorcery",
		Types: []types.Card{types.Sorcery},
	}}
	if canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("non-active player may cast a sorcery without flash timing")
	}
	flashTimingStaticPermanent(g, game.Player1)
	if !canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("flash timing permission did not allow casting a sorcery at instant speed")
	}
}
