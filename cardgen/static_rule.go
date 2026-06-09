package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
)

func lowerStaticRuleDeclaration(
	ability oracle.CompiledAbility,
) (abilityLowering, bool, *oracle.Diagnostic) {
	if len(ability.Effects) != 1 || ability.Effects[0].Kind != oracle.EffectCantBlock {
		return abilityLowering{}, false, nil
	}
	if ability.Kind != oracle.AbilityStatic ||
		ability.Text != game.CantBlockStaticBody.Text ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Modes) != 0 ||
		len(ability.Targets) != 0 ||
		len(ability.Conditions) != 0 ||
		len(ability.Keywords) != 0 ||
		len(ability.References) != 1 ||
		ability.References[0].Kind != oracle.ReferenceThisObject ||
		ability.AbilityWord != "" {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported static rule declaration",
			"the executable source backend supports only exact self cannot-block text",
		)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body:    game.CantBlockStaticBody,
			VarName: "game.CantBlockStaticBody",
		}},
		consumed: semanticConsumption{
			effects:    1,
			references: 1,
		},
		sourceSpans: []oracle.Span{ability.Effects[0].Span},
	}, true, nil
}
