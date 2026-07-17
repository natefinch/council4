package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestLowerCrewPowerContributionIsTextBlind(t *testing.T) {
	t.Parallel()
	lowered, matched, diagnostic := lowerStaticDeclarations(compiler.CompiledAbility{
		Kind: compiler.AbilityStatic,
		Text: "unrelated metadata",
		Static: &compiler.CompiledStaticSemantics{Declarations: []compiler.StaticDeclaration{{
			Kind: compiler.StaticDeclarationCrewPowerContribution,
			CrewPowerContribution: &compiler.StaticCrewPowerContributionDeclaration{
				Bonus: 2,
			},
		}}},
	}, &parser.Ability{})
	if !matched || diagnostic != nil {
		t.Fatalf("matched=%v diagnostic=%#v", matched, diagnostic)
	}
	if len(lowered.staticAbilities) != 1 ||
		lowered.staticAbilities[0].Body.CrewPowerBonus != 2 {
		t.Fatalf("lowered static abilities = %#v", lowered.staticAbilities)
	}
}
