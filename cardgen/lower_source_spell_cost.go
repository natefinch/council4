package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerSourceSpellCostReduction lowers the exact source-scoped dynamic cast cost
// reduction "This spell costs {N} less to cast for each <countable battlefield
// object>." into a static ability that carries an AffectedSource spell cost
// modifier. The modifier holds the per-object generic reduction N and the typed
// battlefield count selection; the rules layer counts the matching permanents at
// cost time and applies the reduction only while this exact spell is being cast.
//
// It runs for both instant/sorcery (AbilitySpell) and permanent (AbilityStatic)
// abilities, since the same wording reduces a permanent's own cast cost. The
// count selection is derived from the effect's typed Amount through the shared
// dynamic-count machinery, which fails closed on non-battlefield zones and
// unsupported controllers.
func lowerSourceSpellCostReduction(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if len(ability.Content.Effects) != 1 || !ability.Content.Effects[0].SourceSpellCostReduction {
		return abilityLowering{}, false, nil
	}
	effect := ability.Content.Effects[0]
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction shell",
			"a source-spell cast cost reduction requires an otherwise empty ability shell",
		)
	}
	if effect.SourceSpellCostReductionAmount <= 0 {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction",
			"the per-object generic reduction must be positive",
		)
	}
	if _, ok := dynamicCardZoneAmount(effect.Amount.Selector(), effect.Amount.Multiplier); ok {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction",
			"the counted objects must be battlefield permanents",
		)
	}
	selection, ok := dynamicAmountSelection(effect.Amount.Selector())
	if !ok {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported source-spell cost reduction",
			"the counted battlefield objects are not representable by the runtime selection vocabulary",
		)
	}
	body := game.StaticAbility{
		Text: ability.Text,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:               game.CostModifierSpell,
				PerObjectReduction: effect.SourceSpellCostReductionAmount,
				CountSelection:     selection,
			},
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
			effects:    len(ability.Content.Effects),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, true, nil
}
