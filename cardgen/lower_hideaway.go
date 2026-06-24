package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerHideawayAbility lowers the Hideaway N land keyword (CR 702.75) to its
// canonical enters-the-battlefield triggered ability: look at the top N cards of
// your library, exile one face down, then put the rest on the bottom in a random
// order. Only the exact keyword with a fixed positive integer and no other rules
// text is supported.
func lowerHideawayAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordHideaway {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterInteger ||
		keyword.Integer < 1 ||
		ability.Kind != compiler.AbilityStatic ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Hideaway ability",
			"the executable source backend supports only exact Hideaway with one integer parameter",
		)
	}
	return game.HideawayTriggeredAbility(keyword.Integer), true, nil
}

// lowerHideawayPlayAbility lowers the activated half of a Hideaway land: the
// "{cost}, {T}: You may play the exiled card without paying its mana cost if
// <condition>" ability that reads the face-down card linked by the Hideaway
// enters trigger. It emits a single PlayHideawayCard primitive gated, as an
// effect gate, on the activation's "if" condition and made optional by the "may"
// permission. Anything outside this exact shape fails closed.
func lowerHideawayPlayAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if ability.Kind != compiler.AbilityActivated ||
		len(ability.Content.Effects) != 1 ||
		!ability.Content.Effects[0].PlayHideawayExiledCard {
		return abilityLowering{}, false, nil
	}
	unsupported := func(detail string) (abilityLowering, bool, *shared.Diagnostic) {
		return abilityLowering{}, true, executableDiagnostic(
			ability,
			"unsupported Hideaway play ability",
			detail,
		)
	}
	effect := ability.Content.Effects[0]
	if !effect.Optional ||
		!effect.CastWithoutPayingManaCost ||
		ability.Cost == nil ||
		len(ability.Cost.Components) == 0 ||
		ability.Trigger != nil ||
		ability.AbilityWord != "" ||
		ability.ActivationTiming != compiler.ActivationTimingNone ||
		ability.SourceAbilityCostReduction != nil ||
		len(ability.Content.Conditions) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Modes) != 0 {
		return unsupported(
			"the executable source backend supports only an exact \"{cost}, {T}: You may play the exiled card without paying its mana cost if <condition>\" ability",
		)
	}
	// The only bound reference this ability may carry is the "exiled card"
	// antecedent inside the play effect itself. The PlayHideawayCard primitive
	// resolves that card through the persistent Hideaway link, not through
	// reference binding, so the reference is internal to the effect. Reject any
	// reference that escapes the effect span so unexpected antecedents fail
	// closed.
	for _, reference := range ability.Content.References {
		if !spanCovered(reference.Span, []shared.Span{effect.Span}) {
			return unsupported("the executable source backend cannot lower a Hideaway play reference outside the play effect")
		}
	}
	gate, ok := lowerCondition(ability.Content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return unsupported("the executable source backend cannot lower this Hideaway activation condition")
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return unsupported("the executable source backend cannot lower every typed Hideaway activation cost component")
	}
	zoneOfFunction, ok := lowerActivationZone(ability.ActivationZone)
	if !ok || zoneOfFunction != zone.Battlefield {
		return unsupported("the executable source backend supports only a battlefield Hideaway play ability")
	}
	instruction := game.Instruction{
		Primitive: game.PlayHideawayCard{},
		Optional:  true,
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(gate)}),
	}
	result := game.ActivatedAbility{
		Text:            ability.Text,
		AdditionalCosts: additionalCosts,
		ZoneOfFunction:  zoneOfFunction,
		Content:         game.Mode{Sequence: []game.Instruction{instruction}}.Ability(),
	}
	if manaCost != nil {
		result.ManaCost = opt.Val(manaCost)
	}
	spans := []shared.Span{ability.Cost.Span, effect.Span, ability.Content.Conditions[0].Span}
	if ability.Content.Conditions[0].ActivationKeywordSpan != (shared.Span{}) {
		spans = append(spans, ability.Content.Conditions[0].ActivationKeywordSpan)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		activatedAbility: opt.Val(result),
		consumed: semanticConsumption{
			cost:       true,
			conditions: 1,
			effects:    1,
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, true, nil
}
