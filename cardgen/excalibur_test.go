package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerExcaliburHistoricTotalManaValueReduction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Excalibur, Sword of Eden",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact — Equipment",
		ManaCost:   "{12}",
		OracleText: "This spell costs {X} less to cast, where X is the total mana value of historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)\nEquipped creature gets +10/+0 and has vigilance.\nEquip legendary creature {2}",
	})
	var modifier game.CostModifier
	found := false
	for _, static := range face.StaticAbilities {
		for _, effect := range static.Body.RuleEffects {
			if effect.Kind == game.RuleEffectCostModifier {
				modifier = effect.CostModifier
				found = true
			}
		}
	}
	if !found {
		t.Fatal("source cost modifier not found")
	}
	if modifier.DynamicReduction == nil ||
		modifier.DynamicReduction.Kind != game.DynamicAmountTotalManaValueInGroup {
		t.Fatalf("modifier = %#v, want total mana value reduction", modifier)
	}
	selection := modifier.DynamicReduction.Group.Selection()
	if selection.Controller != game.ControllerYou || len(selection.AnyOf) != 3 {
		t.Fatalf("historic group selection = %#v", selection)
	}
}
