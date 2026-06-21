package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func lifeGainReplacementCardDef(multiplier, addend int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Boon Reflection",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.LifeGainReplacement(
				"If you would gain life, you gain twice that much life instead.",
				multiplier,
				addend,
			),
		},
	}}
}

func TestLifeGainReplacementDoublesGainedLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, lifeGainReplacementCardDef(2, 0))
	before := g.Players[game.Player1].Life

	if got := gainLife(g, game.Player1, 3); got != 6 {
		t.Fatalf("gainLife() = %d, want 6", got)
	}
	if got := g.Players[game.Player1].Life - before; got != 6 {
		t.Fatalf("life gained = %d, want 6", got)
	}
}

func TestLifeGainReplacementAddsBonus(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, lifeGainReplacementCardDef(1, 1))

	if got := gainLife(g, game.Player1, 3); got != 4 {
		t.Fatalf("gainLife() = %d, want 4", got)
	}
}

func TestLifeGainReplacementOnlyHelpsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, lifeGainReplacementCardDef(2, 0))

	if got := gainLife(g, game.Player2, 3); got != 3 {
		t.Fatalf("opponent gainLife() = %d, want 3", got)
	}
}
