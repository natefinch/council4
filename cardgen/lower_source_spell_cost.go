package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerSourceSpellCostReduction lowers the exact source-scoped dynamic cast cost
// reduction "This spell costs {N} less to cast for each <countable object>." into
// a static ability that carries an AffectedSource spell cost modifier. The
// modifier holds the per-object generic reduction N and the typed count
// selection; the rules layer counts the matching battlefield permanents, or the
// matching cards in the caster's own graveyard or hand, at cost time and applies
// the reduction only while this exact spell is being cast.
//
// It runs for both instant/sorcery (AbilitySpell) and permanent (AbilityStatic)
// abilities, since the same wording reduces a permanent's own cast cost. The
// count selection is derived from the effect's typed Amount through the shared
// dynamic-count machinery, which fails closed on unrepresentable zones and
// controllers.
func lowerSourceSpellCostReduction(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Effects) != 1 {
		return abilityLowering{}, false, nil
	}
	effect := ability.Content.Effects[0]
	if !effect.SourceSpellCostReduction &&
		!effect.SourceSpellCostReductionDynamic &&
		!effect.SourceSpellCostReductionConditional {
		return abilityLowering{}, false, nil
	}
	conditional := effect.SourceSpellCostReductionConditional
	allowedConditions := 0
	if conditional {
		allowedConditions = 1
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != allowedConditions ||
		len(ability.Content.Keywords) != 0 ||
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction shell",
			"a source-spell cast cost reduction requires an otherwise empty ability shell",
		)
	}
	modifier, diagnostic := sourceSpellCostModifier(ability, &effect)
	if diagnostic != nil {
		return abilityLowering{}, true, diagnostic
	}
	body := game.StaticAbility{
		Text: ability.Text,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier:   modifier,
		}},
	}
	spans := make([]shared.Span, 0, 1+len(syntax.Reminders))
	spans = append(spans, ability.Span)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{Body: body}},
		consumed: semanticConsumption{
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, true, nil
}

// sourceSpellCostModifier builds the AffectedSource spell cost modifier for a
// source-spell cast cost reduction. The per-object form ("costs {N} less to cast
// for each <object>") yields a PerObjectReduction with a battlefield count
// selection, or, when the counted objects are cards in the caster's own
// graveyard or hand, a PerObjectReduction with that card-zone count selection and
// CountZone; the dynamic form ("costs {X} less to cast, where X is <dynamic
// amount>") yields a DynamicReduction carrying the typed dynamic amount the
// runtime evaluates at cost time. All fail closed when the counted/measured
// objects are not battlefield permanents the runtime represents.
func sourceSpellCostModifier(ability compiler.CompiledAbility, effect *compiler.CompiledEffect) (game.CostModifier, *shared.Diagnostic) {
	if effect.SourceSpellCostReductionConditional {
		if effect.SourceSpellCostReductionAmount <= 0 {
			return game.CostModifier{}, executableDiagnostic(
				ability,
				"unsupported source-spell cost reduction",
				"the flat generic reduction must be positive",
			)
		}
		if len(ability.Content.Conditions) != 1 {
			return game.CostModifier{}, executableDiagnostic(
				ability,
				"unsupported source-spell cost reduction",
				"a conditional cast cost reduction requires exactly one condition",
			)
		}
		condition, ok := lowerCondition(ability.Content.Conditions[0], conditionContextSpellCostReduction)
		if !ok {
			return game.CostModifier{}, executableDiagnostic(
				ability,
				"unsupported source-spell cost reduction",
				"the reduction condition is not representable by the runtime",
			)
		}
		return game.CostModifier{
			Kind:               game.CostModifierSpell,
			GenericReduction:   effect.SourceSpellCostReductionAmount,
			ReductionCondition: opt.Val(condition),
		}, nil
	}
	if effect.SourceSpellCostReductionDynamic {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok {
			return game.CostModifier{}, executableDiagnostic(
				ability,
				"unsupported source-spell cost reduction",
				"the dynamic reduction amount is not representable by the runtime",
			)
		}
		return game.CostModifier{
			Kind:             game.CostModifierSpell,
			DynamicReduction: &dynamic,
		}, nil
	}
	if effect.SourceSpellCostReductionAmount <= 0 {
		return game.CostModifier{}, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction",
			"the per-object generic reduction must be positive",
		)
	}
	if selector := effect.Amount.Selector(); selector.Zone != zone.None {
		zoneAmount, ok := dynamicCardZoneAmount(selector, effect.Amount.Multiplier)
		if !ok || zoneAmount.Selection == nil || zoneAmount.Selection.Empty() {
			return game.CostModifier{}, executableDiagnostic(
				ability,
				"unsupported source-spell cost reduction",
				"the counted cards in that zone are not representable by the runtime selection vocabulary",
			)
		}
		return game.CostModifier{
			Kind:               game.CostModifierSpell,
			PerObjectReduction: effect.SourceSpellCostReductionAmount,
			CountSelection:     zoneAmount.Selection,
			CountZone:          opt.Val(zoneAmount.CardZone),
		}, nil
	}
	selection, ok := dynamicAmountSelection(effect.Amount.Selector())
	if !ok {
		return game.CostModifier{}, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction",
			"the counted battlefield objects are not representable by the runtime selection vocabulary",
		)
	}
	return game.CostModifier{
		Kind:               game.CostModifierSpell,
		PerObjectReduction: effect.SourceSpellCostReductionAmount,
		CountSelection:     &selection,
	}, nil
}
