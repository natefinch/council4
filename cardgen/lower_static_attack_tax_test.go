package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerStaticAttackTax(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Propaganda Tester",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if len(body.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", body.RuleEffects)
	}
	effect := body.RuleEffects[0]
	if effect.Kind != game.RuleEffectAttackTax ||
		effect.AffectedPlayer != game.PlayerYou ||
		effect.AttackTaxGeneric != 2 {
		t.Fatalf("rule effect = %#v, want controller-scoped {2} attack tax", effect)
	}
}
