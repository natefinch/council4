package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerProgenitorsIconNextChosenTypeFlash(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Progenitor's Icon",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: "As this artifact enters, choose a creature type.\n{T}: Add one mana of any color.\n{T}: The next spell of the chosen type you cast this turn can be cast as though it had flash.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one nonmana activation", face.ActivatedAbilities)
	}
	apply, ok := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.ApplyRule)
	if !ok || len(apply.RuleEffects) != 1 {
		t.Fatalf("primitive = %#v, want apply rule", face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash ||
		!effect.AppliesToNextSpellOnly ||
		effect.SpellChosenSubtypeFrom != game.EntryTypeChoiceKey {
		t.Fatalf("effect = %#v, want next chosen-type flash", effect)
	}
}
