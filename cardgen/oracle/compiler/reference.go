package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// bindReferences assigns each recognized reference phrase one conservative
// referent. It never guesses between multiple target occurrences or an
// unsupported antecedent.
func bindReferences(
	references []CompiledReference,
	targets []CompiledTarget,
	effects []CompiledEffect,
	trigger *CompiledTrigger,
) []CompiledReference {
	bound := append([]CompiledReference(nil), references...)
	for i := range bound {
		reference := &bound[i]
		switch reference.Kind {
		case ReferenceSelfName, ReferenceThisObject:
			reference.Binding = ReferenceBindingSource
			continue
		case ReferenceChosenCards:
			occurrence, ok := chosenCardsTargetAntecedent(*reference, targets)
			if ok {
				reference.Binding = ReferenceBindingTarget
				reference.Occurrence = occurrence
			} else {
				reference.Binding = ReferenceBindingUnsupported
			}
			continue
		case ReferencePronoun, ReferenceThatObject, ReferenceThatPlayer:
		default:
			reference.Binding = ReferenceBindingUnsupported
			continue
		}

		// "that creature" in a combat block trigger body names the other creature
		// in the combat, the event's related permanent ("Whenever this creature
		// blocks or becomes blocked by a creature, ~ deals N damage to that
		// creature.", Inferno Elemental). It binds ahead of the event-permanent
		// fallback so the demonstrative points at the opposing combatant the
		// runtime resolves through EventRelatedPermanentReference, not the source
		// permanent the event names as its primary subject.
		if trigger != nil &&
			reference.Kind == ReferenceThatObject &&
			reference.Order.Start >= trigger.Order.Start &&
			triggerPatternBindsThatCreature(&trigger.Pattern) {
			reference.Binding = ReferenceBindingEventRelatedPermanent
			continue
		}

		if trigger != nil &&
			(reference.Kind != ReferenceThatPlayer || !triggerPatternBindsThatPlayer(&trigger.Pattern)) &&
			!thatObjectPrefersEventPermanent(*reference, trigger) &&
			precedingSourceReferenceAfter(bound[:i], reference.Order, trigger.Order.End) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		// In a self-source cycle trigger ("When you cycle this card, it deals 1
		// damage to each opponent.", CR 702.29e) the cycled card is the ability's
		// own source, so the object pronoun "it"/"that card" in the body names
		// the source. This is scoped to the self-source cycle event, which has no
		// event permanent of its own, so it never reinterprets other triggers.
		if trigger != nil &&
			trigger.Pattern.Source == TriggerSourceSelf &&
			trigger.Pattern.Event == TriggerEventCycled &&
			reference.Order.Start >= trigger.Order.Start &&
			(reference.Kind == ReferenceThatObject ||
				(reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounIt)) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if prior, ok := priorInstructionAntecedent(*reference, effects); ok {
			if precedingSourceReferenceAfter(
				bound[:i],
				reference.Order,
				effects[prior].VerbOrder.Start,
			) {
				reference.Binding = ReferenceBindingSource
			} else {
				reference.Binding = ReferenceBindingPriorInstructionResult
				reference.PriorInstruction = prior
			}
			continue
		}
		if trigger != nil &&
			reference.Kind == ReferenceThatPlayer &&
			reference.Order.Start >= trigger.Order.Start &&
			triggerPatternBindsThatPlayer(&trigger.Pattern) {
			reference.Binding = ReferenceBindingEventPlayer
			continue
		}
		if occurrence, ok, ambiguous := targetAntecedent(*reference, targets); ok || ambiguous {
			if ambiguous {
				reference.Binding = ReferenceBindingAmbiguous
			} else {
				reference.Binding = ReferenceBindingTarget
				reference.Occurrence = occurrence
			}
			continue
		}
		if trigger == nil && precedingSourceReference(bound[:i], reference.Order) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if delayedEffectBindsSource(*reference, effects) &&
			(trigger == nil || trigger.Pattern.Source == TriggerSourceSelf) {
			reference.Binding = ReferenceBindingSource
			continue
		}
		if trigger != nil &&
			trigger.Condition != nil &&
			trigger.Condition.Intervening &&
			trigger.Condition.Order.Contains(reference.Order) &&
			!trigger.Pattern.OneOrMore &&
			triggerEventBindsPermanent(trigger.Pattern.Event) {
			reference.Binding = ReferenceBindingEventPermanent
			continue
		}
		if trigger != nil && triggerReferenceBindsEventCard(&trigger.Pattern, *reference, effects) {
			reference.Binding = ReferenceBindingEventCard
			continue
		}
		if trigger != nil &&
			reference.Order.Start >= trigger.Order.Start &&
			!trigger.Pattern.OneOrMore &&
			triggerEventBindsPermanent(trigger.Pattern.Event) {
			reference.Binding = ReferenceBindingEventPermanent
			continue
		}
		if trigger != nil &&
			reference.Kind == ReferencePronoun &&
			reference.Order.Start >= trigger.Order.Start &&
			triggerEventBindsPlayer(trigger.Pattern.Event) &&
			(reference.Pronoun == ReferencePronounThey ||
				reference.Pronoun == ReferencePronounTheir ||
				reference.Pronoun == ReferencePronounThem) {
			reference.Binding = ReferenceBindingEventPlayer
			continue
		}
		if trigger != nil &&
			reference.Order.Start >= trigger.Order.Start &&
			!trigger.Pattern.OneOrMore &&
			triggerEventBindsStackObject(trigger.Pattern.Event) &&
			(reference.Kind == ReferenceThatObject ||
				(reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounIt)) {
			reference.Binding = ReferenceBindingEventStackObject
			continue
		}
		if reference.Kind == ReferencePronoun {
			reference.Binding = ReferenceBindingAmbiguous
		} else {
			reference.Binding = ReferenceBindingUnsupported
		}

	}
	return bound
}

func bindActivationCostReferences(kind AbilityKind, cost *CompiledCost, references []CompiledReference) []CompiledReference {
	if kind != AbilityActivated || cost == nil {
		return references
	}
	bound := append([]CompiledReference(nil), references...)
	for i := range bound {
		reference := &bound[i]
		if reference.Kind == ReferencePronoun &&
			reference.Pronoun == ReferencePronounIt &&
			cost.Order.Contains(reference.Order) {
			if activationCostPronounBindsSource(cost, reference.Order) {
				reference.Binding = ReferenceBindingSource
			} else {
				reference.Binding = ReferenceBindingAmbiguous
			}
		}
	}
	return bound
}

func activationCostPronounBindsSource(cost *CompiledCost, reference shared.SourceOrder) bool {
	componentIndex := -1
	for i, component := range cost.Components {
		if component.Order.Contains(reference) {
			componentIndex = i
			break
		}
	}
	if componentIndex < 0 {
		return false
	}
	switch cost.Components[componentIndex].Kind {
	case CostRemoveCounter, CostPutCounter, CostExert:
	default:
		return false
	}
	for _, component := range cost.Components[:componentIndex] {
		switch component.Kind {
		case CostMana, CostTap, CostUntap, CostPayLife, CostEnergy, CostLoyalty:
		default:
			return false
		}
	}
	return true
}

func delayedEffectBindsSource(reference CompiledReference, effects []CompiledEffect) bool {
	if len(effects) != 1 ||
		effects[0].DelayedTiming == 0 ||
		reference.Order.Start < effects[0].VerbOrder.End {
		return false
	}
	switch effects[0].Kind {
	case EffectExile, EffectReturn, EffectSacrifice:
		return true
	default:
		return false
	}
}

func priorInstructionAntecedent(reference CompiledReference, effects []CompiledEffect) (int, bool) {
	current := -1
	for i := range effects {
		effect := &effects[i]
		if effect.VerbOrder.Start >= reference.Order.Start {
			continue
		}
		if current < 0 || effect.VerbOrder.Start > effects[current].VerbOrder.Start {
			current = i
		}
	}
	if current < 0 {
		return 0, false
	}
	// "That token" / "those tokens" reads as the subject of a following clause
	// ("That token gains ...") rather than the object of the current one, so the
	// nearest preceding verb belongs to the antecedent effect itself. When that
	// effect creates a token, bind the reference straight to it. A "that
	// <permanent>" reference contained within the create effect's own clause is
	// instead that effect's copy source ("create a token that's a copy of that
	// creature"), not a following subject, so leave it for later binding.
	if effects[current].Kind == EffectCreate && reference.Kind == ReferenceThatObject &&
		!effects[current].Order.Contains(reference.Order) {
		return current, true
	}
	prior := -1
	for i := range effects {
		effect := &effects[i]
		if effect.VerbOrder.Start >= effects[current].VerbOrder.Start {
			continue
		}
		if prior < 0 || effect.VerbOrder.Start > effects[prior].VerbOrder.Start {
			prior = i
		}
	}
	if prior < 0 {
		return 0, false
	}
	if search, ok := priorSearchMoveAntecedent(current, effects); ok {
		return search, true
	}
	switch effects[prior].Kind {
	case EffectDig, EffectExile, EffectManifestDread, EffectReveal, EffectSearch:
		return prior, true
	case EffectPut, EffectReturn:
		effect := effects[prior]
		if reference.Kind == ReferenceThatObject &&
			effect.FromZone == zone.Graveyard &&
			effect.ToZone == zone.Battlefield &&
			effect.UnderYourControl {
			return prior, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func priorSearchMoveAntecedent(current int, effects []CompiledEffect) (int, bool) {
	if effects[current].Kind == EffectPut && current >= 2 &&
		effects[current-1].Kind == EffectShuffle {
		search := current - 2
		if effects[search].Kind == EffectReveal {
			search--
		}
		if search >= 0 && effects[search].Kind == EffectSearch &&
			effects[search].SearchDestination == parser.EffectDestinationTop {
			return search, true
		}
	}
	if current < 3 ||
		effects[current-1].Kind != EffectShuffle ||
		effects[current-2].Kind != EffectPut ||
		effects[current-3].Kind != EffectSearch {
		return 0, false
	}
	return current - 3, true
}

func targetAntecedent(reference CompiledReference, targets []CompiledTarget) (occurrence int, ok, ambiguous bool) {
	closest := -1
	for i, target := range targets {
		if target.Order.Start >= reference.Order.Start {
			continue
		}
		if closest < 0 || target.Order.Start > targets[closest].Order.Start {
			closest = i
		}
	}
	if closest < 0 {
		return 0, false, false
	}
	target := targets[closest]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return 0, false, true
	}
	for i := closest + 1; i < len(targets); i++ {
		if targets[i].Order.Start < reference.Order.Start &&
			targets[i].Order.Start == target.Order.Start {
			return 0, false, true
		}
	}
	return closest, true, false
}

func chosenCardsTargetAntecedent(reference CompiledReference, targets []CompiledTarget) (int, bool) {
	closest := -1
	for i, target := range targets {
		if target.Order.Start >= reference.Order.Start {
			continue
		}
		if closest < 0 || target.Order.Start > targets[closest].Order.Start {
			closest = i
		}
	}
	if closest < 0 {
		return 0, false
	}
	target := targets[closest]
	if target.Cardinality.Min < 0 ||
		target.Cardinality.Max < target.Cardinality.Min ||
		target.Cardinality.Max < 2 {
		return 0, false
	}
	for i := closest + 1; i < len(targets); i++ {
		if targets[i].Order.Start < reference.Order.Start &&
			targets[i].Order.Start == target.Order.Start {
			return 0, false
		}
	}
	return closest, true
}

func precedingSourceReference(references []CompiledReference, order shared.SourceOrder) bool {
	return precedingSourceReferenceAfter(references, order, 0)
}

func precedingSourceReferenceAfter(references []CompiledReference, order shared.SourceOrder, after int) bool {
	for _, reference := range references {
		if reference.Order.Start >= after &&
			reference.Order.Start < order.Start &&
			reference.Binding == ReferenceBindingSource {
			return true
		}
	}
	return false
}

// thatObjectPrefersEventPermanent reports whether a demonstrative "that
// <object>" reference should bind to the triggering permanent rather than be
// captured by the preceding-source heuristic. After a permanent-binding trigger
// (e.g. "another creature you control enters"), "that creature" names the
// triggering object, not the source permanent the effect clause also mentions
// ("this creature deals damage equal to that creature's power"). Letting it fall
// through to the target and event-permanent bindings keeps the demonstrative
// pointed at the entering permanent.
func thatObjectPrefersEventPermanent(reference CompiledReference, trigger *CompiledTrigger) bool {
	return reference.Kind == ReferenceThatObject &&
		!trigger.Pattern.OneOrMore &&
		triggerEventBindsPermanent(trigger.Pattern.Event)
}

func triggerEventBindsPermanent(event TriggerEvent) bool {
	switch event {
	case TriggerEventPermanentEnteredBattlefield,
		TriggerEventPermanentDied,
		TriggerEventZoneChanged,
		TriggerEventCountersAdded,
		TriggerEventDamageDealt,
		TriggerEventAttackerDeclared,
		TriggerEventBlockerDeclared,
		TriggerEventPermanentTapped,
		TriggerEventPermanentUntapped,
		TriggerEventPermanentTurnedFaceUp,
		TriggerEventPermanentSacrificed,
		TriggerEventObjectBecameTarget,
		TriggerEventPermanentMutated,
		TriggerEventAttackerBecameBlocked,
		TriggerEventAttackerBecameUnblocked,
		TriggerEventTokenCreated:
		return true
	default:
		return false
	}
}

func triggerReferenceBindsEventCard(
	trigger *TriggerPattern,
	reference CompiledReference,
	effects []CompiledEffect,
) bool {
	if trigger.Event != TriggerEventPermanentDied &&
		(trigger.Event != TriggerEventZoneChanged ||
			!trigger.MatchToZone ||
			(trigger.ToZone != TriggerZoneGraveyard && trigger.ToZone != TriggerZoneExile)) {
		return false
	}
	for i := range effects {
		effect := &effects[i]
		if !effect.Order.Contains(reference.Order) {
			continue
		}
		switch effect.Kind {
		case EffectPut:
			// "Put it/them onto the battlefield" reanimates the triggering
			// card(s); a coalesced one-or-more trigger binds the whole batch.
			return effect.ToZone == zone.Battlefield
		case EffectReturn:
			// "Return them to the battlefield" likewise reanimates the batch;
			// any other singular return binds the lone event card.
			if trigger.OneOrMore {
				return effect.ToZone == zone.Battlefield
			}
			return true
		case EffectExile, EffectCast:
			return !trigger.OneOrMore
		default:
			return false
		}
	}
	return false
}

// triggerEventBindsStackObject reports whether the trigger event has an
// authoritative stack object subject. When true, "that spell"/"it" in the
// trigger body binds to the triggering event's stack object so effects like
// "copy that spell" copy the spell that was cast.
func triggerEventBindsStackObject(event TriggerEvent) bool {
	return event == TriggerEventSpellCast
}

// triggerEventBindsPlayer reports whether the trigger event has an authoritative
// player subject. When true, player pronouns in the trigger body (they/their/them)
// are conservatively bound to ReferenceBindingEventPlayer.
func triggerEventBindsPlayer(event TriggerEvent) bool {
	switch event {
	case TriggerEventSpellCast,
		TriggerEventCardDrawn,
		TriggerEventCardDiscarded,
		TriggerEventCycled,
		TriggerEventScry,
		TriggerEventSurveil,
		TriggerEventLifeGained,
		TriggerEventLifeLost,
		TriggerEventLibrarySearched:
		return true
	default:
		return false
	}
}

// triggerPatternBindsThatPlayer reports whether a trigger pattern has an
// authoritative player subject that the explicit "that player" reference binds
// to. It extends triggerEventBindsPlayer with the combat/noncombat
// damage-to-a-player event ("deals combat damage to a player, that player ..."),
// whose damaged player the runtime resolves through EventPlayerReference, the
// beginning-of-step event ("at the beginning of each player's draw step, that
// player ..."), whose active player the runtime resolves the same way, and the
// permanent-tapped event ("whenever a player taps a land for mana, that player
// adds ..."), whose "that player" is the controller of the tapped permanent.
func triggerPatternBindsThatPlayer(pattern *TriggerPattern) bool {
	if triggerEventBindsPlayer(pattern.Event) {
		return true
	}
	if pattern.Event == TriggerEventBeginningOfStep {
		return true
	}
	if pattern.Event == TriggerEventPermanentTapped {
		return true
	}
	return pattern.Event == TriggerEventDamageDealt &&
		pattern.DamageRecipient == TriggerDamageRecipientPlayer
}

// triggerPatternBindsThatCreature reports whether a trigger pattern has an
// authoritative related-creature subject that the object demonstrative "that
// creature" binds to. It requires a self-source combat block trigger: "this
// creature blocks a creature" (blocker-declared), "this creature becomes blocked
// by a creature" (attacker-became-blocked), or the union of the two. For these
// the source permanent is the event's primary subject and the opposing
// combatant is the event's related permanent, so "that creature" in the body
// names that other creature. It fails closed for a non-self source ("Whenever a
// creature blocks, ... that creature" names the blocking creature itself, the
// event permanent) and for every other event, leaving "that creature" to the
// existing event-permanent and target bindings.
func triggerPatternBindsThatCreature(pattern *TriggerPattern) bool {
	if pattern.Source != TriggerSourceSelf {
		return false
	}
	return triggerEventIsCombatBlock(pattern.Event) ||
		triggerEventIsCombatBlock(pattern.UnionEvent)
}

// triggerEventIsCombatBlock reports whether a trigger event is one of the two
// combat block events whose related permanent is the opposing combatant: the
// blocker-declared event ("blocks a creature") and the attacker-became-blocked
// event ("becomes blocked by a creature").
func triggerEventIsCombatBlock(event TriggerEvent) bool {
	return event == TriggerEventBlockerDeclared ||
		event == TriggerEventAttackerBecameBlocked
}
