package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// A singular card reference cannot choose one member of an explicitly plural
// top-of-library product.
func singularReferenceToPluralTopCards(reference CompiledReference, producer CompiledEffect) bool {
	singular := reference.Kind == ReferenceThatObject && reference.Pronoun == ReferencePronounUnknown ||
		reference.Kind == ReferencePronoun &&
			(reference.Pronoun == ReferencePronounIt || reference.Pronoun == ReferencePronounIts)
	return singular && producer.CardSource == parser.EffectCardSourceTopOfPlayerLibrary &&
		producer.Amount.Known && producer.Amount.Value != 1
}

func referenceIsOwnedCastCostPolicy(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIts ||
		reference.CardIdentity || reference.ProducerClauseID != 0 {
		return false
	}
	matches := 0
	for _, effect := range effects {
		if effectOwnsReference(effect, reference) && effect.Kind == EffectCast &&
			effect.CastWithoutPayingManaCost && effect.Selector.Kind == SelectorSpell &&
			!effect.Negated && len(effect.Targets) == 0 {
			matches++
		}
	}
	return matches == 1
}

func referenceIsCorrelatedPlayerPolicy(reference CompiledReference, effects []CompiledEffect) bool {
	for _, effect := range effects {
		if !effect.Exact || effect.Negated || effect.Optional || !effectOwnsReference(effect, reference) {
			continue
		}
		if effect.Kind == EffectExile && effect.ExileEachOpponentChoosesGreatestPower &&
			effect.Context == parser.EffectContextEachOpponent && reference.Kind == ReferenceThatPlayer {
			return true
		}
		if effect.Kind == EffectDealDamage && effect.DamageEachOpponentCorrelatedExiledPower &&
			reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounThey {
			return true
		}
	}
	return false
}

func referenceIsOwnedPaymentDecision(reference CompiledReference, content AbilityContent, source ReferenceSourceContext) bool {
	if reference.Kind != ReferenceThatPlayer || source.kind != AbilityTriggered ||
		source.event != TriggerEventAttackerDeclared {
		return false
	}
	matches := 0
	for _, condition := range content.Conditions {
		if condition.Predicate != ConditionPredicateDefendingPlayerDoesNotPay ||
			!slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID) ||
			len(condition.Ownership.ClauseIDs) != 1 {
			continue
		}
		for _, effect := range content.Effects {
			if effect.ClauseID == condition.Ownership.ClauseIDs[0] && effect.Exact &&
				!effect.Negated && effect.Payment.Payer == parser.EffectPaymentPayerDefendingPlayer &&
				effect.Payment.Form == parser.EffectPaymentFormMayPayThenIfDoesNot &&
				len(effect.Payment.ManaCost) > 0 && effect.Payment.AdditionalCost == nil {
				matches++
			}
		}
	}
	return matches == 1
}

func referenceIsCastEntryPolicy(reference CompiledReference, domain ReferenceSubjectDomain,
	source ReferenceSourceContext, effects []CompiledEffect) bool {
	if domain != ReferenceSubjectStackObject || source.event != TriggerEventSpellCast ||
		!slices.Contains(source.eventTypes, types.Creature) ||
		reference.SubjectNoun != parser.ObjectNounCreature {
		return false
	}
	for _, effect := range effects {
		if effectOwnsReference(effect, reference) && effect.Kind == EffectEnterTapped &&
			effect.EntersWithCounters && effect.Exact && !effect.Negated &&
			!effect.Optional && len(effect.Targets) == 0 {
			return true
		}
	}
	return false
}

func effectOwnsReference(effect CompiledEffect, reference CompiledReference) bool {
	for _, candidate := range effect.References {
		if candidate.NodeID == reference.NodeID {
			return true
		}
	}
	return false
}

func spellParagraphReferenceTargets(content AbilityContent, context Context, kind AbilityKind) []CompiledTarget {
	if kind != AbilitySpell || len(content.Targets) != 0 || len(content.Modes) != 0 ||
		context.spellParagraphModal || len(context.spellParagraphTargets) != 1 {
		return content.Targets
	}
	target := context.spellParagraphTargets[0]
	if !target.Exact || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
		targetSubjectDomain(target) != ReferenceSubjectPermanent {
		return content.Targets
	}
	return []CompiledTarget{target}
}

func referenceIsOwnedSpellDamageSubject(reference CompiledReference, content AbilityContent) bool {
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIt {
		return false
	}
	for _, effect := range content.Effects {
		if effect.Kind != EffectDealDamage || !effect.Exact || effect.Negated ||
			len(effect.SubjectReferences) != 1 || effect.SubjectReferences[0].NodeID != reference.NodeID {
			continue
		}
		sources := 0
		for _, condition := range content.Conditions {
			if !slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) {
				continue
			}
			for _, candidate := range content.References {
				if slices.Contains(condition.Ownership.ReferenceNodeIDs, candidate.NodeID) &&
					candidate.Kind == ReferenceThisObject && candidate.SubjectNoun == parser.ObjectNounSpell &&
					candidate.Binding == ReferenceBindingSource {
					sources++
				}
			}
		}
		return sources == 1
	}
	return false
}

func referenceIsOwnedCountPolicy(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounThose {
		return false
	}
	for i, effect := range effects {
		if i == 0 || !effectOwnsReference(effect, reference) || effect.Kind != EffectCreate ||
			effect.Replacement.Kind != parser.EffectReplacementInstead || !effect.Amount.Known ||
			effects[i-1].Kind != EffectCreate || !effects[i-1].Exact ||
			!effects[i-1].Amount.Known || effects[i-1].Context != effect.Context {
			continue
		}
		return true
	}
	return false
}

func referenceIsOwnedGroupPlayer(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun ||
		(reference.Pronoun != ReferencePronounThey && reference.Pronoun != ReferencePronounTheir &&
			reference.Pronoun != ReferencePronounThem) {
		return false
	}
	for _, effect := range effects {
		if !effectOwnsReference(effect, reference) {
			continue
		}
		if effect.Context == parser.EffectContextEachPlayer || effect.Context == parser.EffectContextEachOpponent ||
			effect.Context == parser.EffectContextEachOtherPlayer {
			return true
		}
		if effect.Kind == EffectDealDamage {
			for _, selector := range effect.DamageRecipient.GroupSelectors {
				if selector.Kind == SelectorOpponent {
					return true
				}
			}
		}
	}
	return false
}

func referenceIsEventPlayerProjection(reference CompiledReference, source ReferenceSourceContext, effects []CompiledEffect) bool {
	if source.event == TriggerEventUnknown {
		return false
	}
	for _, effect := range effects {
		if effectOwnsReference(effect, reference) &&
			(effect.Context == parser.EffectContextReferencedPlayer ||
				effect.Context == parser.EffectContextControllerAndReferencedPlayer ||
				effect.DamageRecipient.Reference == parser.DamageRecipientReferenceThatPlayer) {
			return true
		}
	}
	return false
}

func referenceIsTargetPlayerProjection(reference CompiledReference, target CompiledTarget, effects []CompiledEffect) bool {
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return false
	}
	for i, effect := range effects {
		if !effectOwnsReference(effect, reference) ||
			(effect.Context != parser.EffectContextReferencedPlayer && effect.Context != parser.EffectContextEventPlayer) || i == 0 {
			continue
		}
		for _, antecedent := range effects[:i] {
			if (antecedent.Kind != EffectDestroy && antecedent.Kind != EffectReturn &&
				antecedent.Kind != EffectExile && antecedent.Kind != EffectCounter) || len(antecedent.Targets) != 1 ||
				antecedent.Targets[0].Order != target.Order {
				continue
			}
			return true
		}
	}
	return false
}

func referenceIsCardCharacteristicUse(reference CompiledReference, domain ReferenceSubjectDomain, effects []CompiledEffect) bool {
	if domain != ReferenceSubjectCard || reference.CardIdentity {
		return false
	}
	for _, effect := range effects {
		if !effectOwnsReference(effect, reference) {
			continue
		}
		if effect.Amount.ReferenceNodeID == reference.NodeID && damageAmountObservesCharacteristic(effect.Amount.DynamicKind) ||
			effect.Kind == EffectCreate && effect.TokenCopyOfReference ||
			effect.Kind == EffectReturn || effect.Kind == EffectPut || effect.Kind == EffectCast || effect.Kind == EffectExile {
			return true
		}
	}
	return false
}

func referenceIsResolvingDamageSubject(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounIt {
		return false
	}
	for _, effect := range effects {
		if effect.Kind != EffectDealDamage {
			continue
		}
		for _, subject := range effect.SubjectReferences {
			if subject.NodeID == reference.NodeID {
				return true
			}
		}
	}
	return false
}

func ownedTargetSubjectDomain(target CompiledTarget, effects []CompiledEffect) ReferenceSubjectDomain {
	for _, effect := range effects {
		for _, owned := range effect.Targets {
			if owned.Order != target.Order || owned.Cardinality != target.Cardinality {
				continue
			}
			if effect.Kind == EffectCast && effect.FromZone != zone.None && effect.FromZone != zone.Battlefield {
				return ReferenceSubjectCard
			}
			if target.Selector.Zone == zone.None &&
				(effect.Kind == EffectPut && effect.ToZone == zone.None && effect.CounterKindKnown ||
					effect.Kind == EffectModifyPT || effect.Kind == EffectTap || effect.Kind == EffectUntap ||
					effect.Kind == EffectFight || effect.Kind == EffectGainControl ||
					effect.Kind == EffectGain && !effect.LifeObject) {
				return ReferenceSubjectPermanent
			}
		}
	}
	return ReferenceSubjectUnknown
}
