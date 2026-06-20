package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func TestLowerStaticGrantedAnyColorManaAbility(t *testing.T) {
	declaration := compiler.StaticDeclaration{
		Group: compiler.StaticGroupReference{
			Domain: compiler.StaticGroupSourceControllerPermanents,
			Selection: compiler.StaticSelection{
				RequiredTypes: []compiler.StaticCardType{compiler.StaticCardTypeLand},
			},
		},
		Continuous: &compiler.StaticContinuousDeclaration{
			Layer:     compiler.StaticLayerAbility,
			Operation: compiler.StaticContinuousGrantManaAbility,
			GrantedMana: &compiler.StaticGrantedManaAbility{
				TapCost: true, Amount: 1, AnyColor: true,
			},
		},
	}

	effect, ok := lowerStaticContinuousDeclaration(declaration)
	if !ok {
		t.Fatal("lowerStaticContinuousDeclaration() = false")
	}
	if effect.Layer != game.LayerAbility || len(effect.AddAbilities) != 1 {
		t.Fatalf("effect = %#v, want one granted ability in ability layer", effect)
	}
	body, ok := effect.AddAbilities[0].(*game.ManaAbility)
	if !ok || !game.IsTapAnyColorManaAbility(body) {
		t.Fatalf("granted ability = %#v, want canonical tap-any-color mana ability", effect.AddAbilities[0])
	}
}
