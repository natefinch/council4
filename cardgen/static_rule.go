package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
)

func lowerStaticRuleDeclaration(
	ability oracle.CompiledAbility,
) (abilityLowering, bool, *oracle.Diagnostic) {
	if len(ability.Content.Effects) != 1 {
		return abilityLowering{}, false, nil
	}
	var body game.StaticAbility
	var varName string
	var detail string
	switch ability.Content.Effects[0].Kind {
	case oracle.EffectCantBeBlocked:
		body = game.CantBeBlockedStaticBody
		varName = "game.CantBeBlockedStaticBody"
		detail = "the executable source backend supports only exact self cannot-be-blocked text"
	case oracle.EffectCantBlock:
		body = game.CantBlockStaticBody
		varName = "game.CantBlockStaticBody"
		detail = "the executable source backend supports only exact self cannot-block text"
	case oracle.EffectCantBeCountered:
		body = game.CantBeCounteredStaticBody
		varName = "game.CantBeCounteredStaticBody"
		detail = "the executable source backend supports only exact self uncounterable text"
	case oracle.EffectMustAttack:
		body = game.MustAttackStaticBody
		varName = "game.MustAttackStaticBody"
		detail = "the executable source backend supports only exact self must-attack text"
	default:
		return abilityLowering{}, false, nil
	}
	if ability.Kind != oracle.AbilityStatic ||
		ability.Text != body.Text ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != oracle.ReferenceThisObject ||
		ability.AbilityWord != "" {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported static rule declaration",
			detail,
		)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body:    body,
			VarName: varName,
		}},
		consumed: semanticConsumption{
			effects:    1,
			references: 1,
		},
		sourceSpans: []oracle.Span{ability.Content.Effects[0].Span},
	}, true, nil
}
