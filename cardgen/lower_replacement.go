package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerReplacementAbility(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if hasOptionalResolvingEffect(ability.Content.Effects) {
		if replacementAbility, ok := lowerOptionalEntryPayment(ability); ok {
			return replacementAbilityLowering(ability, &replacementAbility, nil)
		}
		if replacementAbility, ok := lowerOptionalEntryZoneReplacement(ability); ok {
			return replacementAbilityLowering(ability, &replacementAbility, nil)
		}
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported optional replacement effect",
			"the executable source backend does not yet lower optional replacement effects",
		)
	}
	if replacementAbility, handled, diagnostic := lowerDamageReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerCounterPlacementReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerTokenCreationReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerNamedTokenSetReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerSelfZoneDestinationReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerEntersWithCountersReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerEntryColorChoiceReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerEntryTypeChoiceReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerEntersAsCopyReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerGroupEntersTappedReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDrawEmptyLibraryWinReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDrawDoublingReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	replacementAbility, diagnostic := lowerEntersTappedReplacement(ability)
	return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
}

// lowerGroupEntersTappedReplacement lowers a static "<permanents> [your
// opponents/you] control enter [the battlefield] tapped." replacement to a
// continuous controller- and type-scoped enters-tapped replacement (Authority of
// the Consuls and the Kismet/Frozen Aether family). It reports handled=false for
// the self enters-tapped form so that path keeps flowing to
// lowerEntersTappedReplacement.
func lowerGroupEntersTappedReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Effects) != 1 || !ability.Content.Effects[0].EntersTappedGroup {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			detail,
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.References) != 0 {
		return unsupported("the executable source backend supports only unconditional group enters-tapped replacements")
	}
	effect := ability.Content.Effects[0]
	controller, ok := groupEntersTappedController(effect.EntersTappedGroupScope)
	if !ok {
		return unsupported("the executable source backend does not lower this enters-tapped controller scope")
	}
	return game.EntersTappedGroupReplacement(ability.Text, controller, effect.EntersTappedGroupTypes...), true, nil
}

// lowerDrawEmptyLibraryWinReplacement lowers the draw-from-empty-library win
// replacement ("If you would draw a card while your library has no cards in it,
// you win the game instead.") to a persistent replacement that wins the game for
// the controller. It reports handled=false unless the recognized would-draw
// condition is present so unrelated replacements keep flowing down the chain.
func lowerDrawEmptyLibraryWinReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateWouldDrawFromEmptyLibrary {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported draw-from-empty-library win replacement",
			detail,
		)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectWinGame ||
		ability.Content.Effects[0].Replacement.Kind != parser.EffectReplacementInstead ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only the exact draw-from-empty-library win replacement")
	}
	return game.DrawFromEmptyLibraryWinReplacement(ability.Text), true, nil
}

// lowerDrawDoublingReplacement lowers the draw-doubling replacement ("If you
// would draw a card[ except the first one you draw in each of your draw steps],
// draw two cards instead.", Thought Reflection, Teferi's Ageless Insight) to a
// persistent replacement that multiplies the controller's card draws. It reports
// handled=false unless a recognized would-draw-card condition is present so
// unrelated replacements keep flowing down the chain.
func lowerDrawDoublingReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 {
		return game.ReplacementAbility{}, false, nil
	}
	predicate := ability.Content.Conditions[0].Predicate
	exceptFirstInDrawStep := predicate == compiler.ConditionPredicateWouldDrawCardExceptFirstInDrawStep
	if predicate != compiler.ConditionPredicateWouldDrawCard && !exceptFirstInDrawStep {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported draw-doubling replacement",
			detail,
		)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectDraw ||
		ability.Content.Effects[0].Replacement.Kind != parser.EffectReplacementInstead ||
		!ability.Content.Effects[0].Amount.Known ||
		ability.Content.Effects[0].Amount.Value < 2 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only the exact draw-doubling replacement")
	}
	multiplier := ability.Content.Effects[0].Amount.Value
	return game.DrawCardMultiplierReplacement(ability.Text, multiplier, exceptFirstInDrawStep), true, nil
}

// groupEntersTappedController maps the parsed controller scope of a group
// enters-tapped replacement to the runtime trigger-controller filter.
func groupEntersTappedController(scope parser.EntersTappedGroupControllerScope) (game.TriggerControllerFilter, bool) {
	switch scope {
	case parser.EntersTappedGroupControllerOpponents:
		return game.TriggerControllerOpponent, true
	case parser.EntersTappedGroupControllerYou:
		return game.TriggerControllerYou, true
	case parser.EntersTappedGroupControllerEach:
		return game.TriggerControllerAny, true
	default:
		return game.TriggerControllerAny, false
	}
}

func replacementAbilityLowering(ability compiler.CompiledAbility, replacementAbility *game.ReplacementAbility, diagnostic *shared.Diagnostic) (abilityLowering, *shared.Diagnostic) {
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	return abilityLowering{
		replacementAbility: opt.Val(*replacementAbility),
		consumed: semanticConsumption{
			effects:    len(ability.Content.Effects),
			conditions: len(ability.Content.Conditions),
			references: len(ability.Content.References),
		},
		sourceSpans: replacementSourceSpans(ability),
	}, nil
}

func appendKeywordSpans(spans []shared.Span, keywords []compiler.CompiledKeyword) []shared.Span {
	for _, keyword := range keywords {
		spans = append(spans, keyword.Span)
	}
	return spans
}

func replacementSourceSpans(ability compiler.CompiledAbility) []shared.Span {
	spans := make([]shared.Span, 0, len(ability.Content.Effects))
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	return spans
}

func lowerEntersTappedReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, *shared.Diagnostic) {
	if replacement, ok := lowerOptionalEntryPayment(ability); ok {
		return replacement, nil
	}
	if !entersTappedReplacementEffectsSupported(ability) ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != compiler.ReferenceBindingSource {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	if len(ability.Content.Conditions) == 1 {
		return lowerConditionalEntersTappedReplacement(ability)
	}
	if len(ability.Content.Conditions) != 0 {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only zero or one condition for self enters-tapped replacements",
		)
	}
	effect := ability.Content.Effects[0]
	if !effect.EntersTappedSelf {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported enters-tapped replacement",
			"the executable source backend supports only exact unconditional self enters-tapped replacements",
		)
	}
	return game.EntersTappedReplacement(ability.Text), nil
}

func lowerSelfZoneDestinationReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	event, eventOK := selfZoneDestinationReplacedEvent(ability)
	if !eventOK {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported self zone-destination replacement",
			detail,
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfZoneDestinationReferencesSupported(ability) {
		return unsupported("the executable source backend supports only exact self graveyard-destination replacements")
	}
	destination, ok := selfZoneReplacementDestination(ability.Content.Effects)
	if !ok || replacementSelectorHasUnsupportedQualifier(ability.Content.Effects[len(ability.Content.Effects)-1].Selector) {
		return unsupported("the executable source backend supports only exile or shuffle-into-library self zone-destination replacements")
	}
	return game.ReplacementAbility{
		Text: ability.Text,
		Replacement: game.ReplacementEffect{
			MatchEvent:         game.EventZoneChanged,
			MatchFromZone:      event.matchFromZone,
			FromZone:           event.fromZone,
			MatchToZone:        true,
			ToZone:             zone.Graveyard,
			ReplaceToZone:      destination,
			ShuffleIntoLibrary: destination == zone.Library,
			RevealSource:       destination == zone.Library,
			Duration:           game.DurationPermanent,
		},
	}, true, nil
}

func lowerCounterPlacementReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !counterPlacementReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported counter-placement replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf ||
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectPut ||
		ability.Content.Effects[1].Kind != compiler.EffectPut ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact counter-doubling replacements")
	}
	switch ability.Content.Conditions[0].Predicate {
	case compiler.ConditionPredicateCounterPlacementOnControlledCreature:
		multiplier, addend, ok := controlledCreatureCounterReplacementAmount(ability.Content.Effects)
		if !ok {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling or additive replacement amounts")
		}
		return game.CounterPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
	case compiler.ConditionPredicateControllerCounterPlacement:
		multiplier, addend, ok := anyCounterReplacementAmount(ability.Content.Effects)
		if !ok {
			return unsupported("the executable source backend supports only all-counter-doubling or additive replacement amounts")
		}
		return game.AnyCounterPlacementReplacement(ability.Text, multiplier, addend, game.TriggerControllerYou), true, nil
	case compiler.ConditionPredicateCounterPlacementOnControlledPermanent:
		multiplier, addend, ok := controlledPermanentCounterReplacementAmount(ability.Content.Effects)
		if !ok {
			return unsupported("the executable source backend supports only controlled-permanent counter-doubling or additive replacement amounts")
		}
		if ability.Content.Conditions[0].Counter == compiler.ConditionCounterPlusOnePlusOne {
			return game.ControlledPermanentCounterKindPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
		}
		return game.ControlledPermanentCounterPlacementReplacement(ability.Text, multiplier, addend, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-creature +1/+1, controlled-permanent, or broad permanent/player counter-doubling or additive replacements")
	}
}

func lowerDamageReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !damageReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported damage replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact additive or multiplicative damage replacements")
	}
	replacement := damageReplacementEffect(ability.Content.Effects)
	if replacementSelectorHasUnsupportedQualifier(replacement.Selector) {
		return unsupported("the executable source backend supports only exact additive or multiplicative damage replacements")
	}
	condition := ability.Content.Conditions[0]
	if condition.Predicate != compiler.ConditionPredicateDamageByControlledSource {
		return unsupported("the executable source backend supports only controlled-source red +1 damage or controlled-source double-damage replacements")
	}
	if len(condition.Selection.ColorsAny) == 1 &&
		condition.Selection.ColorsAny[0] == compiler.ConditionColorRed {
		if replacement.Replacement.Kind != parser.EffectReplacementThatMuchPlus ||
			replacement.Replacement.Amount != 1 {
			return unsupported("the executable source backend supports only +1 red-source damage replacements")
		}
		if condition.Selection.ExcludeSource {
			return game.DamageReplacementExcludingSource(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
		}
		return game.DamageReplacement(ability.Text, 0, 1, []color.Color{color.Red}, game.TriggerControllerYou), true, nil
	}
	if len(condition.Selection.ColorsAny) == 0 && !condition.Selection.ExcludeSource {
		if replacement.Replacement.Kind != parser.EffectReplacementDoubleThat {
			return unsupported("the executable source backend supports only double-damage replacements")
		}
		return game.DamageReplacement(ability.Text, 2, 0, nil, game.TriggerControllerYou), true, nil
	}
	return unsupported("the executable source backend supports only controlled-source red +1 damage or controlled-source double-damage replacements")
}

func damageReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	return ability.Content.Conditions[0].Predicate == compiler.ConditionPredicateDamageByControlledSource
}

func damageReplacementEffect(effects []compiler.CompiledEffect) compiler.CompiledEffect {
	for i := range effects {
		if effects[i].Kind == compiler.EffectDealDamage &&
			effects[i].Replacement.Kind != parser.EffectReplacementNone {
			return effects[i]
		}
	}
	return compiler.CompiledEffect{}
}

func counterPlacementReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	condition := ability.Content.Conditions[0]
	return condition.Predicate == compiler.ConditionPredicateControllerCounterPlacement ||
		condition.Predicate == compiler.ConditionPredicateCounterPlacementOnControlledPermanent ||
		condition.Predicate == compiler.ConditionPredicateCounterPlacementOnControlledCreature &&
			condition.Counter == compiler.ConditionCounterPlusOnePlusOne
}

func controlledCreatureCounterReplacementAmount(effects []compiler.CompiledEffect) (multiplier, addend int, ok bool) {
	second := effects[1]
	if second.Replacement.EachCounterKind ||
		replacementSelectorHasUnsupportedQualifier(second.Selector) {
		return 0, 0, false
	}
	return counterReplacementAmount(second.Replacement)
}

func anyCounterReplacementAmount(effects []compiler.CompiledEffect) (multiplier, addend int, ok bool) {
	second := effects[1]
	if !second.Replacement.EachCounterKind ||
		replacementSelectorHasUnsupportedQualifier(second.Selector) {
		return 0, 0, false
	}
	return counterReplacementAmount(second.Replacement)
}

func controlledPermanentCounterReplacementAmount(effects []compiler.CompiledEffect) (multiplier, addend int, ok bool) {
	second := effects[1]
	if second.Replacement.EachCounterKind ||
		replacementSelectorHasUnsupportedQualifier(second.Selector) {
		return 0, 0, false
	}
	return counterReplacementAmount(second.Replacement)
}

// counterReplacementAmount derives the multiplier and additive bonus a
// counter-placement replacement applies from the parsed "twice that many"
// (doubling) and "that many plus N" (additive) wordings.
func counterReplacementAmount(replacement parser.EffectReplacementSyntax) (multiplier, addend int, ok bool) {
	switch replacement.Kind {
	case parser.EffectReplacementTwiceThatMany:
		return 2, 0, true
	case parser.EffectReplacementThatManyPlus:
		if replacement.Amount <= 0 {
			return 0, 0, false
		}
		return 0, replacement.Amount, true
	default:
		return 0, 0, false
	}
}

func lowerTokenCreationReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !tokenCreationReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported token-creation replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateTokenCreationUnderController ||
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectCreate ||
		ability.Content.Effects[1].Kind != compiler.EffectCreate ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact token-doubling replacements under your control")
	}
	if ability.Content.Effects[1].Replacement.Kind != parser.EffectReplacementTwiceThatMany ||
		replacementSelectorHasUnsupportedQualifier(ability.Content.Effects[1].Selector) {
		return unsupported("the executable source backend supports only token-doubling replacement amounts")
	}
	return game.TokenCreationReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
}

func replacementSelectorHasUnsupportedQualifier(selector compiler.CompiledSelector) bool {
	return selector.Controller != compiler.ControllerAny ||
		selector.Another || selector.Other || selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped || selector.Keyword != parser.KeywordUnknown ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		len(selector.ExcludedTypes()) != 0 || len(selector.Supertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 || len(selector.ExcludedColors()) != 0 ||
		len(selector.SubtypesAny()) != 0
}

func tokenCreationReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	return ability.Content.Conditions[0].Predicate == compiler.ConditionPredicateTokenCreationUnderController
}

// lowerNamedTokenSetReplacement lowers Academy Manufactor's token-type
// replacement ("If you would create a Clue, Food, or Treasure token, instead
// create one of each.") to a persistent replacement that creates one of each
// named token. The replaced set comes from the would-create effect's selector
// subtypes; the trailing create effect carries the one-of-each output marker.
func lowerNamedTokenSetReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !namedTokenSetReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported token-creation replacement",
			detail,
		)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateControllerWouldCreateNamedToken ||
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectCreate ||
		ability.Content.Effects[1].Kind != compiler.EffectCreate ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact one-of-each token-type replacements under your control")
	}
	if ability.Content.Effects[1].Replacement.Kind != parser.EffectReplacementOneOfEach {
		return unsupported("the executable source backend supports only one-of-each token-type replacement amounts")
	}
	selector := ability.Content.Effects[0].Selector
	subtypes := selector.SubtypesAny()
	if len(subtypes) < 2 || namedTokenSelectorHasUnsupportedQualifier(selector) {
		return unsupported("the executable source backend supports only one-of-each replacements over a fixed set of named tokens")
	}
	defs := make([]*game.CardDef, 0, len(subtypes))
	for _, sub := range subtypes {
		def, ok := namedArtifactTokenDef(sub)
		if !ok {
			return unsupported("the executable source backend does not model one of the named tokens in this replacement")
		}
		defs = append(defs, def)
	}
	return game.NamedTokenSetReplacement(ability.Text, defs, game.TriggerControllerYou), true, nil
}

func namedTokenSetReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Effects) == 0 {
		return false
	}
	last := ability.Content.Effects[len(ability.Content.Effects)-1]
	return last.Replacement.Kind == parser.EffectReplacementOneOfEach
}

// namedTokenSelectorHasUnsupportedQualifier rejects a one-of-each replacement
// whose token selector carries any modifier beyond the named subtypes that
// identify the predefined artifact tokens.
func namedTokenSelectorHasUnsupportedQualifier(selector compiler.CompiledSelector) bool {
	return selector.Controller != compiler.ControllerAny ||
		selector.Another || selector.Other || selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped || selector.Keyword != parser.KeywordUnknown ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		len(selector.ExcludedTypes()) != 0 || len(selector.Supertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 || len(selector.ExcludedColors()) != 0
}

type selfZoneDestinationEvent struct {
	fromZone      zone.Type
	matchFromZone bool
}

func selfZoneDestinationReplacedEvent(ability compiler.CompiledAbility) (selfZoneDestinationEvent, bool) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf {
		return selfZoneDestinationEvent{}, false
	}
	switch ability.Content.Conditions[0].Predicate {
	case compiler.ConditionPredicateSourceWouldGoToGraveyard:
		if !referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
			return selfZoneDestinationEvent{}, false
		}
		return selfZoneDestinationEvent{}, true
	case compiler.ConditionPredicateSourceWouldDie:
		if !referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
			return selfZoneDestinationEvent{}, false
		}
		return selfZoneDestinationEvent{fromZone: zone.Battlefield, matchFromZone: true}, true
	default:
		return selfZoneDestinationEvent{}, false
	}
}

func selfZoneDestinationReferencesSupported(ability compiler.CompiledAbility) bool {
	return referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0)
}

func selfZoneReplacementDestination(effects []compiler.CompiledEffect) (zone.Type, bool) {
	for i := range effects {
		effect := &effects[i]
		if effect.Replacement.Kind == parser.EffectReplacementNone {
			continue
		}
		switch effect.Kind {
		case compiler.EffectExile:
			return zone.Exile, true
		case compiler.EffectShuffle:
			return zone.Library, true
		default:
		}
	}
	return zone.None, false
}

func lowerEntersWithCountersReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !isEntersWithCountersReplacement(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported enters-with-counters replacement",
			detail,
		)
	}
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfEntersWithCountersReferences(ability.Content.References) {
		return unsupported("the executable source backend supports only exact self enters-with-counters replacements")
	}
	effect := ability.Content.Effects[0]
	if effect.Duration != compiler.DurationNone || effect.Negated ||
		effect.EntersColorChoice || effect.EntersTypeChoice {
		return unsupported("the executable source backend supports only exact self enters-with-counters replacements")
	}
	// "This creature enters with X +1/+1 counters on it." (Walking Ballista,
	// Hangarback Walker, Endless One) places counters equal to the spell's
	// chosen X, resolved by the runtime from the entering permanent.
	amountFromX := effect.Amount.VariableX
	if !amountFromX &&
		(!effect.Amount.Known || effect.Amount.Value <= 0) {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	if !effect.CounterKindKnown {
		return unsupported("the executable source backend does not support this enters-with-counters counter kind")
	}
	if !effect.Exact {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	placement := game.CounterPlacement{
		Kind:        effect.CounterKind,
		Amount:      effect.Amount.Value,
		AmountFromX: amountFromX,
	}
	// "... enters with N counters on it if <condition>" (Raid, Morbid, Ferocious).
	if len(ability.Content.Conditions) == 1 {
		if effect.Selector.Tapped {
			return unsupported("the executable source backend does not yet support conditional enters-tapped-with-counters replacements")
		}
		condition, ok := lowerCondition(ability.Content.Conditions[0], conditionContextEntryCounters)
		if !ok {
			return unsupported("the executable source backend does not support this enters-with-counters condition")
		}
		return game.EntersWithCountersIfReplacement(ability.Text, &condition, placement), true, nil
	}
	if len(ability.Content.Conditions) != 0 {
		return unsupported("the executable source backend supports only zero or one condition for self enters-with-counters replacements")
	}
	// "This permanent enters tapped with N counters on it." (the Vivid land cycle).
	if effect.Selector.Tapped {
		return game.EntersTappedWithCountersReplacement(ability.Text, placement), true, nil
	}
	return game.EntersWithCountersReplacement(ability.Text, placement), true, nil
}

// isEntersWithCountersReplacement recognizes a self enters-with-counters
// replacement. The parser's EntersWithCounters flag covers the bare "enters with
// N counters" phrasing, while the conditional ("... if a creature died this
// turn") and combined enters-tapped ("enters tapped with N counters") phrasings
// instead surface a known counter kind on the enters effect, so both signals
// route here and lowering decides the exact supported subset.
func isEntersWithCountersReplacement(ability compiler.CompiledAbility) bool {
	if len(ability.Content.Effects) == 0 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped {
		return false
	}
	effect := ability.Content.Effects[0]
	return effect.EntersWithCounters || effect.CounterKindKnown
}

func selfEntersWithCountersReferences(references []compiler.CompiledReference) bool {
	return len(references) == 2 &&
		referencesBindTo(references, compiler.ReferenceBindingSource, 0)
}

func lowerOptionalEntryPayment(ability compiler.CompiledAbility) (game.ReplacementAbility, bool) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return game.ReplacementAbility{}, false
	}
	// "As this land enters, you may pay N life. If you don't, it enters tapped."
	// The parser encodes the optional life payment as the leading enters effect's
	// known amount, so the dual-land cycle (pay 1, 2, or 3 life) is read from that
	// amount rather than fixed at a single value.
	if len(ability.Content.Effects) == 2 &&
		ability.Content.Effects[0].Kind == compiler.EffectEnterTapped &&
		ability.Content.Effects[0].Amount.Known &&
		ability.Content.Effects[0].Amount.Value >= 1 &&
		!ability.Content.Effects[0].Selector.Tapped &&
		ability.Content.Effects[1].Kind == compiler.EffectEnterTapped &&
		ability.Content.Effects[1].Selector.Tapped &&
		len(ability.Content.References) == 2 &&
		referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
		life := ability.Content.Effects[0].Amount.Value
		return game.EntersTappedUnlessPaidReplacement(ability.Text, game.ResolutionPayment{
			Prompt: fmt.Sprintf("Pay %d life?", life),
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalPayLife,
				Amount: life,
			}},
		}), true
	}
	if len(ability.Content.Effects) != 3 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterTapped ||
		ability.Content.Effects[0].Selector.Tapped ||
		ability.Content.Effects[1].Kind != compiler.EffectReveal ||
		ability.Content.Effects[1].Amount.Value != 1 ||
		!ability.Content.Effects[1].Amount.Known ||
		len(ability.Content.Effects[1].Selector.SubtypesAny()) == 0 ||
		len(ability.Content.Effects[1].Selector.SubtypesAny()) > 2 ||
		ability.Content.Effects[2].Kind != compiler.EffectEnterTapped ||
		!ability.Content.Effects[2].Selector.Tapped ||
		len(ability.Content.References) != 2 ||
		!referencesBindTo(ability.Content.References, compiler.ReferenceBindingSource, 0) {
		return game.ReplacementAbility{}, false
	}
	var subtypeSet cost.SubtypeSet
	copy(subtypeSet[:], ability.Content.Effects[1].Selector.SubtypesAny())
	return game.EntersTappedUnlessPaidReplacement(ability.Text, game.ResolutionPayment{
		Prompt: "Reveal a matching card?",
		AdditionalCosts: []cost.Additional{{
			Kind:        cost.AdditionalReveal,
			Amount:      1,
			SubtypesAny: subtypeSet,
			Source:      zone.Hand,
		}},
	}), true
}

// lowerOptionalEntryZoneReplacement lowers the optional self enters-the-
// battlefield replacement "If this permanent would enter, you may <pay an
// alternative cost> instead. If you do, put it onto the battlefield. If you
// don't, put it into its owner's graveyard." (Mox Diamond). The controller may
// pay the alternative cost (discard a card, sacrifice a permanent, pay life) to
// keep the permanent on the battlefield; if the cost is not paid the permanent
// is put into the destination zone instead. It fails closed on any other shape
// so other optional replacements continue to route elsewhere.
func lowerOptionalEntryZoneReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Effects) != 3 ||
		len(ability.Content.Conditions) != 3 ||
		!allReferencesBindToSource(ability.Content.References) {
		return game.ReplacementAbility{}, false
	}
	if !optionalEntryConditionsMatch(ability.Content.Conditions) {
		return game.ReplacementAbility{}, false
	}
	pay := ability.Content.Effects[0]
	keep := ability.Content.Effects[1]
	miss := ability.Content.Effects[2]
	if !pay.Optional ||
		pay.Replacement.Kind != parser.EffectReplacementInstead ||
		keep.Kind != compiler.EffectPut ||
		keep.Negated ||
		keep.ToZone != zone.Battlefield ||
		miss.Kind != compiler.EffectPut ||
		!miss.Negated ||
		miss.ToZone == zone.None ||
		miss.ToZone == zone.Battlefield {
		return game.ReplacementAbility{}, false
	}
	payment, ok := optionalEntryAlternativeCost(&pay)
	if !ok {
		return game.ReplacementAbility{}, false
	}
	return game.EntersUnlessPaidElseZoneReplacement(ability.Text, payment, miss.ToZone), true
}

// optionalEntryConditionsMatch verifies the three conditions guarding an
// optional self-entry replacement: the would-enter trigger, the "If you do"
// branch (prior instruction accepted) and the "If you don't" branch (prior
// instruction not accepted), in source order.
func optionalEntryConditionsMatch(conditions []compiler.CompiledCondition) bool {
	return conditions[0].Predicate == compiler.ConditionPredicateUnsupported &&
		conditions[1].Predicate == compiler.ConditionPredicatePriorInstructionAccepted &&
		conditions[2].Predicate == compiler.ConditionPredicatePriorInstructionNotAccepted
}

// optionalEntryAlternativeCost builds the resolution payment from the optional
// "you may <cost> instead" effect. It supports discarding a card (optionally
// constrained by card type), sacrificing a permanent (optionally constrained by
// type) and paying life, covering the optional-ETB-cost family.
func optionalEntryAlternativeCost(effect *compiler.CompiledEffect) (game.ResolutionPayment, bool) {
	switch effect.Kind {
	case compiler.EffectDiscard:
		additional := cost.Additional{
			Kind:   cost.AdditionalDiscard,
			Amount: 1,
			Source: zone.Hand,
		}
		if cardType, ok := selectorCardType(effect.Selector.Kind); ok {
			additional.MatchCardType = true
			additional.CardType = cardType
		}
		return game.ResolutionPayment{
			Prompt:          "Pay the alternative cost?",
			AdditionalCosts: []cost.Additional{additional},
		}, true
	case compiler.EffectSacrifice:
		additional := cost.Additional{
			Kind:   cost.AdditionalSacrifice,
			Amount: 1,
		}
		if cardType, ok := selectorCardType(effect.Selector.Kind); ok {
			additional.MatchPermanentType = true
			additional.PermanentType = cardType
		}
		return game.ResolutionPayment{
			Prompt:          "Pay the alternative cost?",
			AdditionalCosts: []cost.Additional{additional},
		}, true
	default:
		return game.ResolutionPayment{}, false
	}
}

// selectorCardType maps a card/permanent selector kind to its card type for an
// optional-entry alternative cost. Generic card/permanent selectors carry no
// type constraint and return false.
func selectorCardType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// lowerEntryColorChoiceReplacement lowers the exact self entry color-choice
// replacement "As this <permanent> enters, choose a color." into an entry-time
// color choice that stores the chosen color on the permanent (CR 614.12). It
// fails closed on any other shape (conditions, targets, additional effects), so
// the enters verb's other constructs continue to route elsewhere.
func lowerEntryColorChoiceReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	choiceIndex := -1
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersColorChoice {
			choiceIndex = i
			break
		}
	}
	if choiceIndex < 0 {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func() (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported entry-choice replacement",
			"the executable source backend supports only exact unconditional self \"choose a color\" entry replacements, optionally combined with self enters-tapped",
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!allReferencesBindToSource(ability.Content.References) {
		return unsupported()
	}
	for i := range ability.Content.Effects {
		effect := ability.Content.Effects[i]
		if effect.Kind != compiler.EffectEnterTapped || effect.Negated {
			return unsupported()
		}
	}
	switch len(ability.Content.Effects) {
	case 1:
		exclude := ability.Content.Effects[choiceIndex].EntersColorChoiceExclude
		if exclude != "" {
			return game.EntryColorChoiceExcludingReplacement(ability.Text, exclude), true, nil
		}
		return game.EntryColorChoiceReplacement(ability.Text), true, nil
	case 2:
		other := ability.Content.Effects[1-choiceIndex]
		if !other.EntersTappedSelf {
			return unsupported()
		}
		exclude := ability.Content.Effects[choiceIndex].EntersColorChoiceExclude
		if exclude != "" {
			return game.EntersTappedColorChoiceExcludingReplacement(ability.Text, exclude), true, nil
		}
		return game.EntersTappedColorChoiceReplacement(ability.Text), true, nil
	default:
		return unsupported()
	}
}

// lowerEntryTypeChoiceReplacement lowers the exact self entry creature-type
// choice replacement "As this <permanent> enters, choose a creature type." into
// an entry-time type choice that stores the chosen creature type on the
// permanent (CR 614.12). It fails closed on any other shape (conditions,
// targets, additional effects, combined enters-tapped), so the enters verb's
// other constructs continue to route elsewhere.
func lowerEntryTypeChoiceReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	choiceIndex := -1
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersTypeChoice {
			choiceIndex = i
			break
		}
	}
	if choiceIndex < 0 {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func() (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported entry-choice replacement",
			"the executable source backend supports only the exact unconditional self \"choose a creature type\" entry replacement",
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		len(ability.Content.Effects) != 1 ||
		!allReferencesBindToSource(ability.Content.References) {
		return unsupported()
	}
	effect := ability.Content.Effects[choiceIndex]
	if effect.Kind != compiler.EffectEnterTapped || effect.Negated {
		return unsupported()
	}
	return game.EntryTypeChoiceReplacement(ability.Text), true, nil
}

func allReferencesBindToSource(references []compiler.CompiledReference) bool {
	if len(references) == 0 {
		return false
	}
	for i := range references {
		if references[i].Binding != compiler.ReferenceBindingSource {
			return false
		}
	}
	return true
}

func entersTappedReplacementEffectsSupported(ability compiler.CompiledAbility) bool {
	if len(ability.Content.Effects) == 0 {
		return false
	}
	if len(ability.Content.Effects) == 1 {
		return true
	}
	if len(ability.Content.Conditions) != 1 {
		return false
	}
	conditionSpans := []shared.Span{ability.Content.Conditions[0].Span}
	for i := 1; i < len(ability.Content.Effects); i++ {
		if !spanCovered(ability.Content.Effects[i].VerbSpan, conditionSpans) {
			return false
		}
	}
	return true
}

func lowerConditionalEntersTappedReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, *shared.Diagnostic) {
	condition := ability.Content.Conditions[0]
	replacementCondition, ok := lowerCondition(condition, conditionContextReplacement)
	if !ok {
		return game.ReplacementAbility{}, executableDiagnostic(
			ability,
			"unsupported conditional enters-tapped replacement",
			"the executable source backend does not support this enters-tapped condition",
		)
	}
	return game.EntersTappedIfReplacement(ability.Text, &replacementCondition), nil
}

// lowerEntersAsCopyReplacement lowers the self "You may have this creature enter
// the battlefield as a copy of <filter>[, except <rider>]." replacement (Clone,
// Clever Impersonator, Phyrexian Metamorph) into an enters-as-copy replacement
// whose copied-permanent filter is the effect's selector (CR 706). It fails
// closed on any other ability shape (conditions, targets, costs, triggers,
// additional effects), so unrelated wordings keep their existing handling.
func lowerEntersAsCopyReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	copyIndex := -1
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersAsCopy {
			copyIndex = i
			break
		}
	}
	if copyIndex < 0 {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(ability, "unsupported enters-as-copy replacement", detail)
	}
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		!allReferencesBindToSource(ability.Content.References) {
		return unsupported("the executable source backend supports only the exact unconditional self enters-as-copy replacement")
	}
	effect := ability.Content.Effects[copyIndex]
	if effect.Negated {
		return unsupported("the executable source backend does not support a negated enters-as-copy replacement")
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return unsupported("the executable source backend does not support this enters-as-copy filter")
	}
	return game.EntersAsCopyReplacement(
		ability.Text,
		&selection,
		effect.EntersAsCopyOptional,
		effect.EntersAsCopyNotLegendary,
		effect.EntersAsCopyAddTypes...,
	), true, nil
}
