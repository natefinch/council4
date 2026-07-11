package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerAlchemistsRefugeFlashGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Alchemist's Refuge",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{G}{U}, {T}: You may cast spells this turn as though they had flash.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one nonmana ability", face.ActivatedAbilities)
	}
	primitive := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive
	apply, ok := primitive.(game.ApplyRule)
	if !ok || len(apply.RuleEffects) != 1 ||
		apply.RuleEffects[0].Kind != game.RuleEffectCastSpellsAsThoughFlash {
		t.Fatalf("primitive = %#v, want cast-as-flash rule", primitive)
	}
}
