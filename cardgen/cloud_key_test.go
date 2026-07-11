package cardgen

import "testing"

func TestLowerCloudKeyChosenCardTypeReduction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Cloud Key",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: "As this artifact enters, choose artifact, creature, enchantment, instant, or sorcery.\nSpells you cast of the chosen type cost {1} less to cast.",
	})
	if len(face.ReplacementAbilities) != 1 ||
		!face.ReplacementAbilities[0].Replacement.EntryCardTypeChoice {
		t.Fatalf("replacement abilities = %#v, want card-type choice", face.ReplacementAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.GenericReduction != 1 || !modifier.ChosenCardTypeFromEntryChoice {
		t.Fatalf("modifier = %#v, want chosen card-type reduction", modifier)
	}
}
