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
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerReplacementAbility(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if hasOptionalResolvingEffect(ability.Content.Effects) {
		if replacementAbility, ok := lowerOptionalEntryPayment(ability); ok {
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
	replacementAbility, diagnostic := lowerEntersTappedReplacement(ability)
	return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
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
		if !plusOneCounterDoublingEffects(ability.Content.Effects) {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling replacement amounts")
		}
		return game.CounterPlacementReplacement(ability.Text, 2, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
	case compiler.ConditionPredicateControllerCounterPlacement:
		if !anyCounterDoublingEffects(ability.Content.Effects) {
			return unsupported("the executable source backend supports only all-counter-doubling replacement amounts")
		}
		return game.AnyCounterPlacementReplacement(ability.Text, 2, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-creature +1/+1 or broad permanent/player counter-doubling replacements")
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
		condition.Predicate == compiler.ConditionPredicateCounterPlacementOnControlledCreature &&
			condition.Counter == compiler.ConditionCounterPlusOnePlusOne
}

func plusOneCounterDoublingEffects(effects []compiler.CompiledEffect) bool {
	second := effects[1]
	return second.Replacement.Kind == parser.EffectReplacementTwiceThatMany &&
		!second.Replacement.EachCounterKind &&
		!replacementSelectorHasUnsupportedQualifier(second.Selector)
}

func anyCounterDoublingEffects(effects []compiler.CompiledEffect) bool {
	return effects[1].Replacement.Kind == parser.EffectReplacementTwiceThatMany &&
		effects[1].Replacement.EachCounterKind &&
		!replacementSelectorHasUnsupportedQualifier(effects[1].Selector)
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
	if !effect.Amount.Known ||
		effect.Amount.Value <= 0 {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	if !effect.CounterKindKnown {
		return unsupported("the executable source backend does not support this enters-with-counters counter kind")
	}
	if !effect.Exact {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	placement := game.CounterPlacement{
		Kind:   effect.CounterKind,
		Amount: effect.Amount.Value,
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
