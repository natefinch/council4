package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDrawLimitVariants proves the per-turn draw-limit static family lowers
// to a RuleEffectDrawLimitPerTurn with the expected affected-player relation and
// limit (Narset, Parter of Veils; Spirit of the Labyrinth).
func TestLowerDrawLimitVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text     string
		affected game.PlayerRelation
		limit    int
	}{
		{"Each opponent can't draw more than one card each turn.", game.PlayerOpponent, 1},
		{"Your opponents can't draw more than one card each turn.", game.PlayerOpponent, 1},
		{"Each player can't draw more than one card each turn.", game.PlayerAny, 1},
		{"Players can't draw more than one card each turn.", game.PlayerAny, 1},
		{"You can't draw more than one card each turn.", game.PlayerYou, 1},
	}
	for _, tc := range cases {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Draw Cap",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: tc.text,
		})
		if len(face.StaticAbilities) != 1 {
			t.Fatalf("%q: static abilities = %d, want 1", tc.text, len(face.StaticAbilities))
		}
		effects := face.StaticAbilities[0].Body.RuleEffects
		if len(effects) != 1 {
			t.Fatalf("%q: rule effects = %#v, want one", tc.text, effects)
		}
		effect := effects[0]
		if effect.Kind != game.RuleEffectDrawLimitPerTurn ||
			effect.AffectedPlayer != tc.affected ||
			effect.DrawLimitPerTurn != tc.limit {
			t.Fatalf("%q: draw limit = %#v", tc.text, effect)
		}
	}
}

// TestLowerDrawLimitNarset proves the Narset, Parter of Veils static line lowers
// to an opponent-scoped per-turn draw cap.
func TestLowerDrawLimitNarset(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spirit of the Labyrinth",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "Each player can't draw more than one card each turn.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effect := face.StaticAbilities[0].Body.RuleEffects[0]
	if effect.Kind != game.RuleEffectDrawLimitPerTurn ||
		effect.AffectedPlayer != game.PlayerAny ||
		effect.DrawLimitPerTurn != 1 {
		t.Fatalf("draw limit = %#v", effect)
	}
}
