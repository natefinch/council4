package cardgen

import "testing"

func TestLowerFiendishDuoOpponentDamageReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fiendish Duo",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "First strike\nIf a source would deal damage to an opponent, it deals double that damage to that player instead.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %#v, want one", face.ReplacementAbilities)
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.DamageMultiplier != 2 ||
		!replacement.DamageRecipientOpponent ||
		!replacement.DamageRecipientOpponentPlayerOnly {
		t.Fatalf("replacement = %#v, want double damage to opponents", replacement)
	}
}
