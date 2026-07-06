package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerPreventDamageToCountersReplacement lowers the continuous static "If
// <permanent> would be dealt damage, prevent that damage and put that many
// +1/+1 counters on it." (Panther Habit, Jared Carthalion, Anti-Venom) into a
// DamagePreventionToPlusOneCountersReplacement. The damaged permanent is the
// ability's own source or the permanent it is attached to; the runtime prevents
// the whole event and adds that many +1/+1 counters to the recipient.
func lowerPreventDamageToCountersReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if ability.Kind != compiler.AbilityReplacement ||
		len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateDamageWouldBeDealtToPermanent {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported damage prevention replacement",
			detail,
		)
	}
	if ability.Cost != nil || ability.Trigger != nil || ability.Optional ||
		len(ability.Content.Targets) != 0 || len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 {
		return unsupported("the executable source backend supports only exact prevent-damage-to-counters replacements")
	}
	if len(ability.Content.Effects) != 1 {
		return unsupported("the executable source backend supports only a single prevent-and-add-counters effect")
	}
	effect := ability.Content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		!effect.CounterKindKnown || effect.CounterKind != counter.PlusOnePlusOne ||
		effect.Amount.DynamicKind == compiler.DynamicAmountNone {
		return unsupported("the executable source backend supports only the 'put that many +1/+1 counters' prevention effect")
	}
	selection := ability.Content.Conditions[0].Selection
	if selection.DamageRecipientAttached == selection.DamageRecipientSelf {
		return unsupported("the executable source backend supports only a self or attached prevent-damage recipient")
	}
	return game.DamagePreventionToPlusOneCountersReplacement(
		ability.Text,
		selection.DamageRecipientAttached,
		opt.V[game.Condition]{},
	), true, nil
}

func lowerReplacementAbility(ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	if hasOptionalResolvingEffect(ability.Content.Effects) {
		if replacementAbility, ok := lowerOptionalEntryPayment(ability); ok {
			return replacementAbilityLowering(ability, &replacementAbility, nil)
		}
		if replacementAbility, ok := lowerOptionalEntryZoneReplacement(ability); ok {
			return replacementAbilityLowering(ability, &replacementAbility, nil)
		}
		// The self "You may have this creature enter as a copy of <filter>"
		// replacement (Clone family) carries its optionality on the enters-as-copy
		// effect itself; the enters-as-copy lowerer already honors that optional
		// flag, so route the optional copy replacement to it rather than failing
		// closed as an unlowered optional replacement.
		if replacementAbility, handled, diagnostic := lowerEntersAsCopyReplacement(ability); handled || diagnostic != nil {
			return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
		}
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported optional replacement effect",
			"the executable source backend does not yet lower optional replacement effects",
		)
	}
	if replacementAbility, handled, diagnostic := lowerPreventDamageToCountersReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDamagePreventionReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
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
	if replacementAbility, handled, diagnostic := lowerGraveyardRedirectReplacement(ability); handled || diagnostic != nil {
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
	if replacementAbility, handled, diagnostic := lowerDevourReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerTributeReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerGroupEntersTappedReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDrawEmptyLibraryWinReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDrawReplacementDig(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerDrawDoublingReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerLifeGainReplacement(ability); handled || diagnostic != nil {
		return replacementAbilityLowering(ability, &replacementAbility, diagnostic)
	}
	if replacementAbility, handled, diagnostic := lowerLifeLossReplacement(ability); handled || diagnostic != nil {
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
	if len(ability.Content.Effects) != 1 || !ability.Content.Effects[0].EntersTappedGroup() {
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
	controller, ok := groupEntersTappedController(effect.GroupEntryModification.ControllerScope)
	if !ok {
		return unsupported("the executable source backend does not lower this enters-tapped controller scope")
	}
	return game.EntersTappedGroupReplacement(ability.Text, controller, effect.GroupEntryModification.Types...), true, nil
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

// lowerDrawReplacementDig lowers the draw-replacement dig ("If you would draw a
// card, instead look at the top N cards of your library, then put one into your
// hand and the rest into your graveyard.", Underrealm Lich) to a persistent
// replacement that, each time the controller would draw a card, instead looks at
// the top N cards, puts the take count into hand, and routes the rest to the
// recorded remainder. It reports handled=false unless a recognized
// would-draw-card condition gates a single instead-dig effect so unrelated
// replacements (including the plain draw-doubling form) keep flowing down the
// chain.
func lowerDrawReplacementDig(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 {
		return game.ReplacementAbility{}, false, nil
	}
	predicate := ability.Content.Conditions[0].Predicate
	if predicate != compiler.ConditionPredicateWouldDrawCard {
		return game.ReplacementAbility{}, false, nil
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectDig ||
		!ability.Content.Effects[0].Dig.Put ||
		ability.Content.Effects[0].Replacement.Kind != parser.EffectReplacementInstead {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported draw-replacement dig",
			detail,
		)
	}
	effect := ability.Content.Effects[0]
	look := effect.Amount.Value
	take := effect.Dig.Take
	if !effect.Exact || effect.Optional || effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.Known || take < 1 || look <= take ||
		len(effect.Targets) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only the exact draw-replacement dig")
	}
	return game.DrawCardDigReplacement(ability.Text, look, take, digRemainder(effect.Dig.Remainder)), true, nil
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
	if isMaxSpeedAbilityWord(ability.AbilityWord) {
		// "Max speed — If you would draw a card, draw N cards instead." (Vnwxt,
		// Verbose Host) gates the draw-doubling replacement on the controller
		// having maximum speed (CR 702.179). The runtime evaluates the condition
		// against the in-flight draw event, so the multiplier applies only while
		// the controller is at max speed.
		return game.MaxSpeedDrawCardMultiplierReplacement(ability.Text, multiplier, exceptFirstInDrawStep), true, nil
	}
	if ability.AbilityWord != "" {
		return unsupported("the executable source backend supports only the exact draw-doubling replacement")
	}
	return game.DrawCardMultiplierReplacement(ability.Text, multiplier, exceptFirstInDrawStep), true, nil
}

// lowerLifeLossReplacement lowers the life-loss replacement "If an opponent
// would lose life during your turn, they lose twice that much life instead."
// (Bloodletter of Aclazotz) and its untimed/any-player generalizations to a
// persistent replacement that scales the matched player's life loss.
func lowerLifeLossReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 {
		return game.ReplacementAbility{}, false, nil
	}
	var recipientOpponent, duringControllerTurn bool
	switch ability.Content.Conditions[0].Predicate {
	case compiler.ConditionPredicateOpponentLifeLossDuringControllerTurn:
		recipientOpponent = true
		duringControllerTurn = true
	case compiler.ConditionPredicateOpponentLifeLoss:
		recipientOpponent = true
	case compiler.ConditionPredicateAnyPlayerLifeLoss:
	default:
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported life-loss replacement",
			detail,
		)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectLose ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact additive or multiplicative life-loss replacements")
	}
	switch ability.Content.Effects[0].Replacement.Kind {
	case parser.EffectReplacementTwiceThatMuch:
		return game.LifeLossReplacement(ability.Text, 2, 0, recipientOpponent, duringControllerTurn), true, nil
	case parser.EffectReplacementThatMuchPlus:
		addend := ability.Content.Effects[0].Replacement.Amount
		if addend <= 0 {
			return unsupported("the executable source backend supports only positive additive life-loss replacements")
		}
		return game.LifeLossReplacement(ability.Text, 1, addend, recipientOpponent, duringControllerTurn), true, nil
	default:
		return unsupported("the executable source backend supports only double or additive life-loss replacements")
	}
}

// lowerLifeGainReplacement lowers the life-gain replacement "If you would gain
// life, you gain twice that much life instead." (multiplier two) and "you gain
// that much life plus N instead." (additive bonus) to a persistent replacement
// that scales the controller's life gain (Boon Reflection, Rhox Faithmender,
// Angel of Vitality, Alhammarret's Archive's life clause).
func lowerLifeGainReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateControllerLifeGain {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported life-gain replacement",
			detail,
		)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectGain ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact additive or multiplicative life-gain replacements")
	}
	switch ability.Content.Effects[0].Replacement.Kind {
	case parser.EffectReplacementTwiceThatMuch:
		return game.LifeGainReplacement(ability.Text, 2, 0), true, nil
	case parser.EffectReplacementThatMuchPlus:
		addend := ability.Content.Effects[0].Replacement.Amount
		if addend <= 0 {
			return unsupported("the executable source backend supports only positive additive life-gain replacements")
		}
		return game.LifeGainReplacement(ability.Text, 1, addend), true, nil
	default:
		return unsupported("the executable source backend supports only double or additive life-gain replacements")
	}
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
			keywords:   entersAsCopyConsumedKeywords(ability),
			references: len(ability.Content.References),
		},
		sourceSpans: replacementSourceSpans(ability, replacementAbility),
	}, nil
}

// entersAsCopyConsumedKeywords counts the ability's content keywords that an
// enters-as-copy replacement consumes as copiable "except it has <keyword>"
// riders. Replacements without an enters-as-copy effect carry no rider keywords,
// so the count is zero and their content keywords remain unconsumed.
func entersAsCopyConsumedKeywords(ability compiler.CompiledAbility) int {
	var riders []parser.KeywordKind
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersAsCopy {
			riders = append(riders, ability.Content.Effects[i].EntersAsCopyAddKeywords...)
		}
	}
	count := 0
	for i := range ability.Content.Keywords {
		if slices.Contains(riders, ability.Content.Keywords[i].Kind) {
			count++
		}
	}
	return count
}

func appendKeywordSpans(spans []shared.Span, keywords []compiler.CompiledKeyword) []shared.Span {
	for _, keyword := range keywords {
		spans = append(spans, keyword.Span)
	}
	return spans
}

// replacementAbilityWordConsumed reports whether a lowered replacement ability
// absorbed its leading ability word into a rules-bearing gate that the lowerer
// recognized. Two cases qualify: the "Max speed —" draw-doubling replacement
// (Vnwxt, Verbose Host), which gates on the controller having maximum speed, and
// the "Adamant —" enters-with-counters replacement (the Throne of Eldraine
// Paladin cycle), which gates on the per-color mana spent to cast the spell. In
// both cases the ability word span is therefore covered.
func replacementAbilityWordConsumed(lowered abilityLowering) bool {
	if !lowered.replacementAbility.Exists {
		return false
	}
	if replacementManaSpentCondition(&lowered.replacementAbility.Val) {
		return true
	}
	return lowered.replacementAbility.Val.Replacement.Condition.Exists &&
		lowered.replacementAbility.Val.Replacement.Condition.Val.ControllerHasMaxSpeed
}

func replacementSourceSpans(ability compiler.CompiledAbility, replacementAbility *game.ReplacementAbility) []shared.Span {
	spans := make([]shared.Span, 0, len(ability.Content.Effects))
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	// An Adamant enters-with-counters replacement ("Adamant — If at least three
	// white mana was spent to cast this spell, this creature enters with a +1/+1
	// counter on it.") carries its gate in a leading condition clause whose
	// source span and "this spell" reference span fall outside the
	// enters-with-counters effect span, so cover them here. The mana-spent
	// condition fields mark this case; every other replacement keeps the prior
	// effect-only coverage so its source accounting is unchanged.
	if replacementManaSpentCondition(replacementAbility) {
		for i := range ability.Content.Conditions {
			spans = append(spans, ability.Content.Conditions[i].Span)
		}
		for i := range ability.Content.References {
			spans = append(spans, ability.Content.References[i].Span)
		}
	}
	return spans
}

// replacementManaSpentCondition reports whether a lowered replacement gates on an
// Adamant "mana was spent to cast this spell" condition, whose leading clause and
// "this spell" reference need explicit source-span coverage.
func replacementManaSpentCondition(replacementAbility *game.ReplacementAbility) bool {
	if replacementAbility == nil || !replacementAbility.Replacement.Condition.Exists {
		return false
	}
	condition := replacementAbility.Replacement.Condition.Val
	return condition.SpellColorManaSpent.Count > 0 || condition.SpellSameColorManaSpentAtLeast > 0
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

func lowerGraveyardRedirectReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != compiler.ConditionIf ||
		ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateCardWouldGoToGraveyard {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported graveyard-redirect replacement",
			detail,
		)
	}
	if len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact graveyard-redirect replacements")
	}
	if len(ability.Content.Effects) == 0 {
		return unsupported("the executable source backend supports only exile graveyard-redirect replacements")
	}
	destination, ok := selfZoneReplacementDestination(ability.Content.Effects)
	if !ok || destination != zone.Exile ||
		replacementSelectorHasUnsupportedQualifier(ability.Content.Effects[len(ability.Content.Effects)-1].Selector) {
		return unsupported("the executable source backend supports only exile graveyard-redirect replacements")
	}
	condition := ability.Content.Conditions[0]
	if !triggerCardTypesValid(condition.GraveyardSubjectTypesAny) {
		return unsupported("the executable source backend does not support this graveyard-redirect card-type filter")
	}
	ownerFilter, ok := graveyardRedirectOwnerFilter(condition.GraveyardRedirectScope)
	if !ok {
		return unsupported("the executable source backend does not support this graveyard-redirect scope")
	}
	controlFilter, ok := graveyardRedirectControlFilter(condition.GraveyardRedirectControlScope)
	if !ok {
		return unsupported("the executable source backend does not support this graveyard-redirect control scope")
	}
	return game.GraveyardRedirectReplacement(
		ability.Text,
		ownerFilter,
		controlFilter,
		condition.GraveyardFromBattlefieldOnly,
		condition.GraveyardSubjectTypesAny...,
	), true, nil
}

// triggerCardTypesValid reports whether every entry is a recognized card type.
// It preserves the fail-closed guard the deleted lowerTriggerCardTypes helper
// applied: an unset (empty) card type is rejected.
func triggerCardTypesValid(cardTypes []types.Card) bool {
	return !slices.Contains(cardTypes, "")
}

func graveyardRedirectOwnerFilter(scope compiler.GraveyardRedirectScope) (game.TriggerControllerFilter, bool) {
	switch scope {
	case compiler.GraveyardRedirectScopeAny:
		return game.TriggerControllerAny, true
	case compiler.GraveyardRedirectScopeYou:
		return game.TriggerControllerYou, true
	case compiler.GraveyardRedirectScopeOpponent:
		return game.TriggerControllerOpponent, true
	default:
		return game.TriggerControllerAny, false
	}
}

func graveyardRedirectControlFilter(scope compiler.GraveyardRedirectControlScope) (game.TriggerControllerFilter, bool) {
	switch scope {
	case compiler.GraveyardRedirectControlScopeAny:
		return game.TriggerControllerAny, true
	case compiler.GraveyardRedirectControlScopeYou:
		return game.TriggerControllerYou, true
	case compiler.GraveyardRedirectControlScopeOpponent:
		return game.TriggerControllerOpponent, true
	default:
		return game.TriggerControllerAny, false
	}
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
	case compiler.ConditionPredicateCounterPlacementOnSelf:
		multiplier, addend, ok := controlledCreatureCounterReplacementAmount(ability.Content.Effects)
		if !ok {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling or additive replacement amounts")
		}
		return game.SelfCounterPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne), true, nil
	case compiler.ConditionPredicateCounterPlacementOnAnyCreature:
		multiplier, addend, ok := controlledCreatureCounterReplacementAmount(ability.Content.Effects)
		if !ok {
			return unsupported("the executable source backend supports only +1/+1 counter-doubling or additive replacement amounts")
		}
		return game.CounterPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, game.TriggerControllerAny), true, nil
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
		condition := ability.Content.Conditions[0]
		// "of each of those kinds of counters" modifies every counter kind being
		// placed, so it always maps to the kind-agnostic all-counter constructors
		// regardless of any specific counter named in the condition.
		eachKind := ability.Content.Effects[1].Replacement.EachCounterKind
		plusOnePlusOne := condition.Counter == compiler.ConditionCounterPlusOnePlusOne && !eachKind
		if condition.CounterRecipientExcludesSource {
			recipient, ok := controlledPermanentRecipientSelection(condition)
			if !ok {
				return unsupported("the executable source backend does not support this counter-recipient filter")
			}
			if plusOnePlusOne {
				return game.ControlledPermanentSelectionCounterKindPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, recipient, game.TriggerControllerYou), true, nil
			}
			return game.ControlledPermanentSelectionCounterPlacementReplacement(ability.Text, multiplier, addend, recipient, game.TriggerControllerYou), true, nil
		}
		if len(condition.CounterRecipientTypesAny) > 0 {
			if !triggerCardTypesValid(condition.CounterRecipientTypesAny) {
				return unsupported("the executable source backend does not support this counter-recipient card-type filter")
			}
			recipientTypes := condition.CounterRecipientTypesAny
			if plusOnePlusOne {
				return game.ControlledPermanentTypesCounterKindPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, recipientTypes, game.TriggerControllerYou), true, nil
			}
			return game.ControlledPermanentTypesCounterPlacementReplacement(ability.Text, multiplier, addend, recipientTypes, game.TriggerControllerYou), true, nil
		}
		if plusOnePlusOne {
			return game.ControlledPermanentCounterKindPlacementReplacement(ability.Text, multiplier, addend, counter.PlusOnePlusOne, game.TriggerControllerYou), true, nil
		}
		return game.ControlledPermanentCounterPlacementReplacement(ability.Text, multiplier, addend, game.TriggerControllerYou), true, nil
	default:
		return unsupported("the executable source backend supports only controlled-creature +1/+1, controlled-permanent, or broad permanent/player counter-doubling or additive replacements")
	}
}

// lowerDamagePreventionReplacement lowers a continuous static "If a [qualifier]
// source would deal damage to you, prevent N of that damage." replacement
// (the Sphere of Law/Duty/Reason/Grace/Truth, Sphere of Purity, Urza's Armor,
// Protection of the Hekma, and Guardian Seraph family) into a filtered damage
// prevention replacement. It reports handled=false when the ability is not a
// prevention candidate so the additive/multiplicative path keeps flowing to
// lowerDamageReplacement.
func lowerDamagePreventionReplacement(
	ability compiler.CompiledAbility,
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if !damagePreventionReplacementCandidate(ability) {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(
			ability,
			"unsupported damage prevention replacement",
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
		return unsupported("the executable source backend supports only exact static damage prevention replacements")
	}
	effect := damagePreventionReplacementEffect(ability.Content.Effects)
	if effect.PreventDamageThatAmount <= 0 {
		return unsupported("the executable source backend supports only fixed-amount damage prevention replacements")
	}
	condition := ability.Content.Conditions[0]
	if !condition.Selection.DamageRecipientController {
		return unsupported("the executable source backend supports only damage prevention that protects you")
	}
	if condition.Selection.ExcludeSource ||
		condition.Selection.DamageRecipientOpponent ||
		condition.Selection.DamageNoncombatOnly {
		return unsupported("the executable source backend supports only unqualified static damage prevention replacements")
	}
	sourceColors, ok := conditionColors(condition.Selection.ColorsAny)
	if !ok {
		return unsupported("the executable source backend supports only known source colors in damage prevention replacements")
	}
	sourceTypes, ok := conditionCardTypes(condition.Selection.RequiredTypes)
	if !ok {
		return unsupported("the executable source backend supports only known source card types in damage prevention replacements")
	}
	return game.DamagePreventionReplacement(ability.Text, &game.DamagePreventionSpec{
		Amount:                   effect.PreventDamageThatAmount,
		SourceColors:             sourceColors,
		SourceTypes:              sourceTypes,
		SourceControllerOpponent: condition.Selection.DamageSourceControllerOpponent,
	}), true, nil
}

// damagePreventionReplacementCandidate reports whether the ability is a
// controlled-source damage replacement whose sole damage effect is a fixed-amount
// "prevent N of that damage" static.
func damagePreventionReplacementCandidate(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityReplacement || len(ability.Content.Conditions) == 0 {
		return false
	}
	if ability.Content.Conditions[0].Predicate != compiler.ConditionPredicateDamageByControlledSource {
		return false
	}
	return damagePreventionReplacementEffect(ability.Content.Effects).PreventDamageThatAmount > 0
}

// damagePreventionReplacementEffect returns the fixed-amount prevention effect
// from a static damage-prevention replacement, or the zero value when none is
// present.
func damagePreventionReplacementEffect(effects []compiler.CompiledEffect) compiler.CompiledEffect {
	for i := range effects {
		if effects[i].Kind == compiler.EffectPreventDamage &&
			effects[i].PreventDamageThatAmount > 0 {
			return effects[i]
		}
	}
	return compiler.CompiledEffect{}
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
		return unsupported("the executable source backend supports only controlled-source damage replacements")
	}
	multiplier, addend, ok := damageReplacementAmount(replacement.Replacement)
	if !ok {
		return unsupported("the executable source backend supports only double, triple, or additive damage replacements")
	}
	sourceColors, ok := conditionColors(condition.Selection.ColorsAny)
	if !ok {
		return unsupported("the executable source backend supports only known source colors in damage replacements")
	}
	sourceTypes, ok := conditionCardTypes(condition.Selection.RequiredTypes)
	if !ok {
		return unsupported("the executable source backend supports only known source card types in damage replacements")
	}
	controller := game.TriggerControllerYou
	if condition.Selection.DamageSourceAnyController {
		controller = game.TriggerControllerAny
	}
	return game.DamageReplacementFiltered(ability.Text, &game.DamageReplacementSpec{
		Multiplier:        multiplier,
		Addend:            addend,
		SourceColors:      sourceColors,
		SourceTypes:       sourceTypes,
		ExcludeSource:     condition.Selection.ExcludeSource,
		RecipientOpponent: condition.Selection.DamageRecipientOpponent,
		NoncombatOnly:     condition.Selection.DamageNoncombatOnly,
		Controller:        controller,
	}), true, nil
}

// damageReplacementAmount maps the parsed "double/triple that damage" and "that
// much damage plus N" wordings onto the runtime multiplier and additive bonus.
// The additive form uses a zero multiplier so only the bonus is applied.
func damageReplacementAmount(replacement parser.EffectReplacementSyntax) (multiplier, addend int, ok bool) {
	switch replacement.Kind {
	case parser.EffectReplacementDoubleThat:
		return 2, 0, true
	case parser.EffectReplacementTripleThat:
		return 3, 0, true
	case parser.EffectReplacementThatMuchPlus:
		if replacement.Amount <= 0 {
			return 0, 0, false
		}
		return 0, replacement.Amount, true
	default:
		return 0, 0, false
	}
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
		(condition.Predicate == compiler.ConditionPredicateCounterPlacementOnSelf &&
			condition.Counter == compiler.ConditionCounterPlusOnePlusOne) ||
		(condition.Predicate == compiler.ConditionPredicateCounterPlacementOnControlledCreature ||
			condition.Predicate == compiler.ConditionPredicateCounterPlacementOnAnyCreature) &&
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
	if replacementSelectorHasUnsupportedQualifier(second.Selector) {
		return 0, 0, false
	}
	return counterReplacementAmount(second.Replacement)
}

// controlledPermanentRecipientSelection builds the recipient characteristic
// filter of a controlled-permanent counter-placement replacement that excludes
// the source permanent ("another creature you control", Benevolent Hydra). The
// controller scope ("you control") lives outside the selection on the runtime
// constructor, so it is not encoded here. It fails closed on any card type
// outside the permanent-selection vocabulary.
func controlledPermanentRecipientSelection(condition compiler.CompiledCondition) (game.Selection, bool) {
	if len(condition.CounterRecipientTypesAny) > 0 && !triggerCardTypesValid(condition.CounterRecipientTypesAny) {
		return game.Selection{}, false
	}
	selection := game.Selection{
		RequiredTypesAny: append([]types.Card(nil), condition.CounterRecipientTypesAny...),
		ExcludeSource:    condition.CounterRecipientExcludesSource,
	}
	if len(selection.RequiredTypesAny) == 0 && !selection.ExcludeSource {
		return game.Selection{}, false
	}
	return selection, true
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
		len(ability.Content.Effects) != 2 ||
		ability.Content.Effects[0].Kind != compiler.EffectCreate ||
		ability.Content.Effects[1].Kind != compiler.EffectCreate ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact token-creation replacements")
	}
	filter := game.TriggerControllerYou
	if ability.Content.Conditions[0].Predicate == compiler.ConditionPredicateTokenCreationAnyController {
		filter = game.TriggerControllerAny
	}
	requiredTypes, ok := tokenCreationReplacementRequiredTypes(ability.Content.Effects[0].Selector)
	if !ok {
		return unsupported("the executable source backend supports only an artifact, creature, or untyped would-create filter")
	}
	output := ability.Content.Effects[1]
	switch output.Replacement.Kind {
	case parser.EffectReplacementTwiceThatMany:
		if replacementSelectorHasUnsupportedQualifier(output.Selector) {
			return unsupported("the executable source backend supports only token-doubling replacement amounts")
		}
		// The doubling spec carries a card-type filter but no subtype filter, so
		// a subtype-restricted would-create group ("one or more Treasure tokens
		// ... twice that many") must fail closed rather than lower to a
		// subtype-blind doubler. The additive branch below threads the
		// would-create subtypes through explicitly, so it has no such gap.
		if len(ability.Content.Effects[0].Selector.SubtypesAny()) != 0 ||
			len(ability.Content.Effects[0].Selector.ExcludedSubtypes()) != 0 {
			return unsupported("the executable source backend supports only type-filtered token-doubling replacements")
		}
		if filter == game.TriggerControllerYou && len(requiredTypes) == 0 {
			return game.TokenCreationReplacement(ability.Text, 2, filter), true, nil
		}
		return game.TokenCreationReplacementFiltered(ability.Text, &game.TokenCreationReplacementSpec{
			Multiplier: 2,
			Types:      requiredTypes,
			Filter:     filter,
		}), true, nil
	case parser.EffectReplacementPlusAdditional:
		addendSubtypes := output.Selector.SubtypesAny()
		if len(addendSubtypes) == 0 {
			return unsupported("the executable source backend supports only named additive token-creation replacements")
		}
		if slices.Equal(addendSubtypes, ability.Content.Effects[0].Selector.SubtypesAny()) {
			return game.TokenCreationReplacementFiltered(ability.Text, &game.TokenCreationReplacementSpec{
				Multiplier: 1,
				Addend:     output.Replacement.Amount,
				Subtypes:   addendSubtypes,
				Types:      requiredTypes,
				Filter:     filter,
			}), true, nil
		}
		addendDef, ok := synthesizeNamedArtifactTokenDef(&output)
		if !ok {
			addendDef, ok = synthesizeCreatureTokenDef(&output, nil)
		}
		if !ok {
			return unsupported("the executable source backend supports only same-token, predefined-token, or fixed creature additive token-creation replacements")
		}
		return game.TokenCreationReplacementFiltered(ability.Text, &game.TokenCreationReplacementSpec{
			Multiplier: 1,
			Addend:     output.Replacement.Amount,
			Subtypes:   ability.Content.Effects[0].Selector.SubtypesAny(),
			Types:      requiredTypes,
			Filter:     filter,
			AddendDef:  addendDef,
		}), true, nil
	case parser.EffectReplacementThatManyIdentity:
		// The identity substitution replaces each would-be-created token with one
		// copy of a fully spelled-out substitute token ("... that many 4/4 white
		// Angel creature tokens with flying and vigilance are created instead.").
		// The substitute is synthesized from the output effect exactly as the
		// active create-verb path builds a creature token, with the output
		// clause's keywords ("flying and vigilance") threaded in. The would-create
		// group's subtypes restrict which token-creation events are replaced.
		replaceDef, ok := synthesizeCreatureTokenDef(&output, output.TokenKeywords)
		if !ok {
			return unsupported("the executable source backend supports only fixed creature substitute token-creation replacements")
		}
		return game.TokenCreationReplacementFiltered(ability.Text, &game.TokenCreationReplacementSpec{
			Multiplier: 1,
			Subtypes:   ability.Content.Effects[0].Selector.SubtypesAny(),
			Types:      requiredTypes,
			Filter:     filter,
			ReplaceDef: replaceDef,
		}), true, nil
	default:
		return unsupported("the executable source backend supports only token-doubling, additive, or identity replacement amounts")
	}
}

// tokenCreationReplacementRequiredTypes derives the card-type filter restricting
// which token-creation events a replacement matches from the would-create
// group's selector ("one or more artifact tokens", "one or more creature
// tokens"). The artifact/creature/enchantment/land typed selections carry their
// type in the selector kind; an untyped, "permanent", or unparsed-noun selection
// imposes no type filter ("one or more tokens"). It fails closed for a
// would-create group carrying any other qualifier so an unmodeled restriction
// never lowers to a broader replacement.
func tokenCreationReplacementRequiredTypes(selector compiler.CompiledSelector) ([]types.Card, bool) {
	if selector.Controller != compiler.ControllerAny ||
		selector.Another || selector.Other || selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped || selector.Keyword != parser.KeywordUnknown ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		len(selector.ExcludedTypes()) != 0 || len(selector.Supertypes()) != 0 ||
		len(selector.ColorsAny()) != 0 || len(selector.ExcludedColors()) != 0 {
		return nil, false
	}
	var cardTypes []types.Card
	switch selector.Kind {
	case compiler.SelectorUnknown, compiler.SelectorAny, compiler.SelectorPermanent:
	case compiler.SelectorArtifact:
		cardTypes = append(cardTypes, types.Artifact)
	case compiler.SelectorCreature:
		cardTypes = append(cardTypes, types.Creature)
	case compiler.SelectorEnchantment:
		cardTypes = append(cardTypes, types.Enchantment)
	case compiler.SelectorLand:
		cardTypes = append(cardTypes, types.Land)
	default:
		return nil, false
	}
	cardTypes = append(cardTypes, selector.RequiredTypesAny()...)
	return cardTypes, true
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
	predicate := ability.Content.Conditions[0].Predicate
	return predicate == compiler.ConditionPredicateTokenCreationUnderController ||
		predicate == compiler.ConditionPredicateTokenCreationAnyController
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
	if ability.Content.Effects[0].EntersWithCountersGroup() {
		return lowerGroupEntersWithCountersReplacement(ability, unsupported)
	}
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		!selfEntersWithCountersReferences(ability.Content.References, ability.Content.Effects[0], ability.Content.Conditions) {
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
	var dynamic opt.V[*game.DynamicAmount]
	if !amountFromX &&
		(!effect.Amount.Known || effect.Amount.Value <= 0) {
		// "This creature enters with a +1/+1 counter on it for each <X>."
		// (Golgari Grave-Troll) places a rules-derived number of counters; reuse
		// the shared dynamic-amount lowering so every supported count form works.
		lowered, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok {
			return unsupported("the executable source backend does not support this dynamic enters-with-counters quantity")
		}
		dynamic = opt.Val(&lowered)
	}
	if !effect.CounterKindKnown {
		return unsupported("the executable source backend does not support this enters-with-counters counter kind")
	}
	// A concretely resolved count (a known positive amount, the spell's X, or a
	// lowered dynamic amount) places an exact number of counters even when the
	// parser flagged the sentence non-exact only because the word numeral above
	// four ("five", "six", "seven") is not an integer token. Such fixed self
	// counts (Pentavus, Ghave) still place a definite number, so only a count
	// the runtime cannot resolve stays unsupported.
	concreteAmount := amountFromX || dynamic.Exists || (effect.Amount.Known && effect.Amount.Value > 0)
	if !effect.Exact && !concreteAmount {
		return unsupported("the executable source backend does not yet support dynamic enters-with-counters quantities")
	}
	placement := game.CounterPlacement{
		Kind:        effect.CounterKind,
		Amount:      effect.Amount.Value,
		AmountFromX: amountFromX,
		Dynamic:     dynamic,
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

func selfEntersWithCountersReferences(references []compiler.CompiledReference, effect compiler.CompiledEffect, conditions []compiler.CompiledCondition) bool {
	if !referencesBindTo(references, compiler.ReferenceBindingSource, 0) {
		return false
	}
	if len(references) == 2 {
		return true
	}
	// The Converge count "for each color of mana spent to cast it" (Crystalline
	// Crawler) and the Multikicker count "for each time it was kicked"
	// (Everflowing Chalice) each add a third self reference for the "it" inside
	// their count phrase, so a three-reference self replacement is accepted only
	// for those amounts.
	if len(references) == 3 &&
		(effect.Amount.DynamicKind == compiler.DynamicAmountColorsOfManaSpent ||
			effect.Amount.DynamicKind == compiler.DynamicAmountTimesKicked) {
		return true
	}
	// An Adamant condition ("if at least three white mana was spent to cast this
	// spell", the Throne of Eldraine Paladin cycle) adds a third self reference
	// for the "this spell" inside its gate, so a three-reference self replacement
	// is accepted when that condition is present.
	if len(references) == 3 && hasManaSpentToCastCondition(conditions) {
		return true
	}
	// A kicker gate that names the source by "this <type>" ("If this creature was
	// kicked, it enters with N +1/+1 counters on it." — the Invasion kicker
	// cycle) adds a third self reference for that named subject inside the gate.
	// Only the bare "enters with N counters on it" form (effect.EntersWithCounters)
	// is accepted; a combined "... and with <keyword>" clause is not represented
	// here and must stay unsupported.
	return len(references) == 3 && effect.EntersWithCounters && hasEventSubjectKickedCondition(conditions)
}

// hasEventSubjectKickedCondition reports whether any condition is the
// event-subject "was kicked" gate, whose source-naming subject ("this creature")
// binds the source and so adds a self reference to an enters-with-counters
// replacement.
func hasEventSubjectKickedCondition(conditions []compiler.CompiledCondition) bool {
	for i := range conditions {
		if conditions[i].Predicate == compiler.ConditionPredicateEventSubjectWasKicked {
			return true
		}
	}
	return false
}

// hasManaSpentToCastCondition reports whether any condition is an Adamant
// per-color or same-color "mana was spent to cast this spell" gate, whose "this
// spell" reference binds the source and so adds a self reference to an
// enters-with-counters replacement.
func hasManaSpentToCastCondition(conditions []compiler.CompiledCondition) bool {
	for i := range conditions {
		switch conditions[i].Predicate {
		case compiler.ConditionPredicateColoredManaSpentToCastAtLeast,
			compiler.ConditionPredicateSameColorManaSpentToCastAtLeast:
			return true
		default:
			// Other predicates do not match; continue scanning.
		}
	}
	return false
}

// lowerGroupEntersWithCountersReplacement lowers a static enters-with-counters
// replacement that adds a single counter to a group of the controller's
// permanents as they enter ("Each other creature you control enters with an
// additional vigilance counter on it." — Tayam, Luminous Enigma). The recipient
// group is read from the parser-recognized Selector; dynamic quantities and
// recipient shapes the runtime Selection cannot represent fail closed.
func lowerGroupEntersWithCountersReplacement(
	ability compiler.CompiledAbility,
	unsupported func(string) (game.ReplacementAbility, bool, *shared.Diagnostic),
) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional {
		return unsupported("the executable source backend supports only exact group enters-with-counters replacements")
	}
	effect := ability.Content.Effects[0]
	if effect.Duration != compiler.DurationNone || effect.Negated ||
		effect.EntersColorChoice || effect.EntersTypeChoice || effect.Selector.Tapped {
		return unsupported("the executable source backend supports only exact group enters-with-counters replacements")
	}
	if !effect.CounterKindKnown {
		return unsupported("the executable source backend does not support this enters-with-counters counter kind")
	}
	if !effect.Exact {
		return unsupported("the executable source backend does not yet support dynamic group enters-with-counters quantities")
	}
	if effect.Amount.VariableX {
		return unsupported("the executable source backend does not yet support dynamic group enters-with-counters quantities")
	}
	recipient, ok := lowerGroupEntersWithCountersRecipient(effect.Selector)
	if !ok {
		return unsupported("the executable source backend does not support this group enters-with-counters recipient")
	}
	placement := game.CounterPlacement{Kind: effect.CounterKind, Amount: 1}
	if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		// "Each other creature you control enters with a number of additional
		// +1/+1 counters on it equal to <amount>." (Arwen, Weaver of Hope) scales
		// the placement by a rules-derived amount read from the replacement's
		// source permanent at resolution; reuse the shared dynamic-amount
		// lowering so every supported amount form works.
		lowered, dynamicOK := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !dynamicOK {
			return unsupported("the executable source backend does not support this dynamic group enters-with-counters quantity")
		}
		placement.Amount = 0
		placement.Dynamic = opt.Val(&lowered)
	} else if effect.Amount.Value > 0 {
		placement.Amount = effect.Amount.Value
	}
	return game.EntersWithCountersGroupReplacement(ability.Text, recipient, placement), true, nil
}

// lowerGroupEntersWithCountersRecipient maps the recipient selector of a group
// enters-with-counters replacement to a runtime Selection scoped to the
// controller's permanents. Only the controller scope, card-type, subtype,
// excluded-subtype, color, keyword, token-status, and "other" filters are
// supported; any other selector shape fails closed.
func lowerGroupEntersWithCountersRecipient(selector compiler.CompiledSelector) (*game.Selection, bool) {
	if selector.Controller != compiler.ControllerYou {
		return nil, false
	}
	if selector.All || selector.Another || selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped || selector.MatchCounter ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.BasicLandType || selector.PlayerOrPlaneswalker ||
		selector.SubtypeFromChosenType ||
		len(selector.Alternatives) != 0 || selector.Zone != zone.None {
		return nil, false
	}
	if _, ok := groupEntersWithCountersRequiredType(selector); !ok {
		return nil, false
	}
	selection, ok := SelectionForSelectorMasked(selector, groupEntersWithCountersRecipientMask)
	if !ok {
		return nil, false
	}
	return &selection, true
}

// groupEntersWithCountersRecipientMask drops the canonical dimensions a group
// enters-with-counters recipient never carries: the excluded supertype,
// kind-agnostic counter, "aren't of the chosen type" exclusion, conjunctive type
// set, and historic disjunction. It fails closed on a source-relative power
// comparison: an enters-with-counters group has no source permanent to compare
// against, so the predecessor projector rejected that filter rather than dropping
// it.
var groupEntersWithCountersRecipientMask = SelectionMask{}.Ignoring(
	DimExcludedSupertype,
	DimMatchAnyCounter,
	DimSubtypeChoiceExcluded,
	DimConjunctiveTypes,
	DimHistoric,
).Rejecting(
	DimPowerVsSource,
	DimRequiredName,
)

// groupEntersWithCountersRequiredType maps a group enters-with-counters
// recipient selector kind to the runtime card type the entering permanent must
// have. SelectorPermanent imposes no type ("" with ok=true); the bare
// subtype-named group form ("Each Dragon you control", SelectorUnknown) is
// accepted only when a subtype constraint scopes it. Any other kind fails closed.
func groupEntersWithCountersRequiredType(selector compiler.CompiledSelector) (types.Card, bool) {
	switch selector.Kind {
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPermanent:
		return "", true
	case compiler.SelectorUnknown:
		if len(selector.SubtypesAny()) > 0 {
			return "", true
		}
		return "", false
	default:
		return "", false
	}
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
		hasForeignReference(ability.Content.References) {
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
		hasForeignReference(ability.Content.References) {
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

// hasForeignReference reports whether any reference binds to an object other than
// the source. Unlike allReferencesBindToSource it accepts an empty reference set,
// so a self entry-choice replacement whose only subject is named by its own
// subtype ("As this Aura enters, ...", which surfaces no object reference) is
// still recognized as referring solely to its source.
func hasForeignReference(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Binding != compiler.ReferenceBindingSource {
			return true
		}
	}
	return false
}

func contentKeywordsAreCopyRiders(keywords []compiler.CompiledKeyword, riders []parser.KeywordKind) bool {
	for i := range keywords {
		if !slices.Contains(riders, keywords[i].Kind) {
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
// lowerDevourReplacement lowers the Devour as-enters replacement (CR 702.81)
// produced by the keyword expansion. It accepts only the exact unconditional
// self replacement (a single EntersDevour effect with a positive multiplier and
// no targets, conditions, cost, or trigger) and builds the matching
// game.Devour*Replacement: the typed variants ("Devour artifact N", "Devour land
// N", "Devour Food N") carry their sacrifice filter and the creature form uses
// the plain constructor; anything else keeps the card unsupported.
func lowerDevourReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	devourIndex := -1
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersDevour {
			devourIndex = i
			break
		}
	}
	if devourIndex < 0 {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(ability, "unsupported devour replacement", detail)
	}
	effect := ability.Content.Effects[devourIndex]
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		effect.Negated ||
		effect.EntersDevourMultiplier <= 0 ||
		!allReferencesBindToSource(ability.Content.References) {
		return unsupported("the executable source backend supports only the exact unconditional self devour replacement")
	}
	if effect.EntersDevourSubtype != "" {
		return game.DevourSubtypeReplacement(ability.Text, effect.EntersDevourMultiplier, effect.EntersDevourSubtype), true, nil
	}
	if effect.EntersDevourType != "" {
		return game.DevourTypeReplacement(ability.Text, effect.EntersDevourMultiplier, effect.EntersDevourType), true, nil
	}
	return game.DevourReplacement(ability.Text, effect.EntersDevourMultiplier), true, nil
}

// lowerTributeReplacement lowers the Tribute as-enters replacement (CR 702.110)
// produced by the keyword expansion. It accepts only the exact unconditional
// self replacement (a single EntersTribute effect with a positive count and no
// targets, conditions, cost, or trigger) and builds a game.TributeReplacement;
// anything else keeps the card unsupported.
func lowerTributeReplacement(ability compiler.CompiledAbility) (game.ReplacementAbility, bool, *shared.Diagnostic) {
	tributeIndex := -1
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].EntersTribute {
			tributeIndex = i
			break
		}
	}
	if tributeIndex < 0 {
		return game.ReplacementAbility{}, false, nil
	}
	unsupported := func(detail string) (game.ReplacementAbility, bool, *shared.Diagnostic) {
		return game.ReplacementAbility{}, true, executableDiagnostic(ability, "unsupported tribute replacement", detail)
	}
	effect := ability.Content.Effects[tributeIndex]
	if len(ability.Content.Effects) != 1 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		effect.Negated ||
		effect.EntersTributeCount <= 0 ||
		!allReferencesBindToSource(ability.Content.References) {
		return unsupported("the executable source backend supports only the exact unconditional self tribute replacement")
	}
	return game.TributeReplacement(ability.Text, effect.EntersTributeCount), true, nil
}

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
		!contentKeywordsAreCopyRiders(ability.Content.Keywords, ability.Content.Effects[copyIndex].EntersAsCopyAddKeywords) ||
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
	var conditionalCounters []game.ConditionalCounterPlacement
	if len(effect.EntersAsCopyConditionalCounters) > 0 {
		placements, diagnostic := entersAsCopyConditionalCounterPlacements(ability, effect.EntersAsCopyConditionalCounters)
		if diagnostic != nil {
			return game.ReplacementAbility{}, true, diagnostic
		}
		conditionalCounters = placements
	}
	var addKeywords []game.Keyword
	for _, keyword := range effect.EntersAsCopyAddKeywords {
		runtime, ok := runtimeKeyword(keyword)
		if !ok {
			return unsupported("the executable source backend does not support this enters-as-copy keyword rider")
		}
		if _, ok := game.KeywordStaticBody(runtime); !ok {
			return unsupported("the executable source backend does not support this enters-as-copy keyword rider")
		}
		addKeywords = append(addKeywords, runtime)
	}
	replacement := game.EntersAsCopyReplacement(
		ability.Text,
		&selection,
		effect.EntersAsCopyOptional,
		effect.EntersAsCopyNotLegendary,
		conditionalCounters,
		effect.EntersAsCopyUntilEndOfTurn,
		addKeywords,
		effect.EntersAsCopyAddSubtypes,
		effect.EntersAsCopyAddTypes...,
	)
	if effect.EntersAsCopyTapped {
		replacement = game.EntersTappedAsCopy(replacement)
	}
	return replacement, true, nil
}

// entersAsCopyConditionalCounterPlacements lowers the parsed conditional copiable
// counter riders into runtime placements, failing closed on a non-positive
// amount or an unset counter kind so a malformed rider keeps the card
// unsupported rather than emitting an inert placement.
func entersAsCopyConditionalCounterPlacements(ability compiler.CompiledAbility, riders []parser.EntersAsCopyConditionalCounter) ([]game.ConditionalCounterPlacement, *shared.Diagnostic) {
	placements := make([]game.ConditionalCounterPlacement, 0, len(riders))
	for _, rider := range riders {
		if rider.Amount <= 0 || rider.IfType == "" {
			return nil, executableDiagnostic(ability, "unsupported enters-as-copy replacement", "the executable source backend does not support this conditional copiable counter rider")
		}
		placements = append(placements, game.ConditionalCounterPlacement{
			Kind:   rider.Kind,
			Amount: rider.Amount,
			IfType: rider.IfType,
		})
	}
	return placements, nil
}
