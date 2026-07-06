package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerEquipCostReductionAbility lowers an Equip keyword ability that carries a
// self-referential flat cost reduction ("Equip {4}. This ability costs {3} less
// to activate if you're the monarch.", Crown of Gondor). The parser reports the
// reduction amount and its source span on SourceAbilityCostReduction, and the
// gating clause as a single ability condition. The reduction becomes a
// CostModifierAbility whose GenericReduction the runtime applies only while the
// condition holds (costModifierAppliesToAbility). Equip abilities without the
// reduction, or whose reduction is the counted "for each" form, fall through to
// the plain lowerEquipAbility path unchanged.
func lowerEquipCostReductionAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	reduction := ability.SourceAbilityCostReduction
	if reduction == nil {
		return abilityLowering{}, false, nil
	}
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordEquip ||
		ability.Kind != compiler.AbilityStatic {
		return abilityLowering{}, false, nil
	}
	// The parser recognizes only the flat "This ability costs {N} less to
	// activate" reduction on a keyword (static) ability, so a reduction reaching
	// this Equip path never carries the counted "for each" selection.
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		len(keyword.ManaCost) == 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return abilityLowering{}, false, nil
	}
	var body game.ActivatedAbility
	if keyword.EquipRestriction != nil {
		body = game.EquipRestrictedActivatedAbility(
			slices.Clone(keyword.ManaCost),
			slices.Clone(keyword.EquipRestriction.Supertypes),
			slices.Clone(keyword.EquipRestriction.Subtypes),
		)
	} else {
		body = game.EquipActivatedAbility(slices.Clone(keyword.ManaCost))
	}
	modifier := game.CostModifier{
		Kind:             game.CostModifierAbility,
		GenericReduction: reduction.Amount,
	}
	spans := keywordSpans(ability, syntax)
	spans = append(spans, reduction.Span)
	switch len(ability.Content.Conditions) {
	case 0:
		// Unconditional flat reduction ("This ability costs {N} less to activate.").
	case 1:
		cond, ok := lowerCondition(ability.Content.Conditions[0], conditionContextSpellCostReduction)
		if !ok {
			return abilityLowering{}, true, executableDiagnostic(
				ability,
				"unsupported Equip ability",
				"the equip cost-reduction condition is not representable by the runtime condition vocabulary",
			)
		}
		modifier.ReductionCondition = opt.Val(cond)
		spans = append(spans, ability.Content.Conditions[0].Span)
	default:
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported Equip ability",
			"the executable source backend supports only a single equip cost-reduction condition",
		)
	}
	body.CostModifiers = append(body.CostModifiers, modifier)
	return abilityLowering{
		activatedAbility: opt.Val(body),
		consumed: semanticConsumption{
			keywords:   1,
			conditions: len(ability.Content.Conditions),
		},
		sourceSpans: spans,
	}, true, nil
}
