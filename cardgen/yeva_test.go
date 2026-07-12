package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerYevaGreenCreatureFlash(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Yeva, Nature's Herald",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Elf Shaman",
		ManaCost:   "{2}{G}{G}",
		OracleText: "Flash\nYou may cast green creature spells as though they had flash.",
	})
	var found bool
	for i := range face.StaticAbilities {
		for _, effect := range face.StaticAbilities[i].Body.RuleEffects {
			if effect.Kind == game.RuleEffectCastSpellsAsThoughFlash &&
				len(effect.SpellTypes) == 1 && effect.SpellTypes[0] == types.Creature &&
				len(effect.SpellColors) == 1 && effect.SpellColors[0] == color.Green {
				found = true
			}

		}
	}
	if !found {
		t.Fatalf("static abilities = %#v, want green creature flash permission", face.StaticAbilities)
	}
}

func TestRenderSpellColorFilterWithoutType(t *testing.T) {
	rendered, err := (Renderer{}).renderRuleEffect(newRenderCtx(), &game.RuleEffect{
		Kind:        game.RuleEffectCastSpellsAsThoughFlash,
		SpellColors: []color.Color{color.Green},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "SpellColors: []color.Color{color.Green}") {
		t.Fatalf("rendered = %q, want independent spell color filter", rendered)
	}
}
