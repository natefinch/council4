package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitObservedCardResultOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		bindObservedCardResultOwnership(ability.Sentences, ability.ConditionSegments, ability.ConditionClauses, ability.SemanticReferences)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				bindObservedCardResultOwnership(mode.Sentences, mode.ConditionSegments, mode.ConditionClauses, mode.SemanticReferences)
			}
		}
	}
}

func bindObservedCardResultOwnership(sentences []Sentence, segments []ConditionSegment, clauses []ConditionClause, references []Reference) {
	effects := conditionEffects(sentences)
	for i := range segments {
		segment := &segments[i]
		if segment.ClauseIndex < 0 || segment.ClauseIndex >= len(clauses) ||
			clauses[segment.ClauseIndex].Predicate != ConditionPredicatePriorInstructionAccepted ||
			len(segment.Ownership.ReferenceNodeIDs) != 0 || segment.Ownership.DelayedBody ||
			(segment.Ownership.Scope != ConditionScopeClause && segment.Ownership.Scope != ConditionScopeGroup) {
			continue
		}
		if producer, handled := observedCardResultProducer(effects, *segment, references); handled {
			segment.Ownership.ResultProducerClauseID = producer
		}
	}
}

func observedCardResultProducer(effects []*EffectSyntax, segment ConditionSegment, references []Reference) (int, bool) {
	producer, handled := 0, false
	for _, consumer := range effects {
		if !slices.Contains(segment.Ownership.ClauseIDs, consumer.ClauseID) ||
			consumer.Kind != EffectPut || consumer.CardSource != EffectCardSourcePriorInstructionResult ||
			consumer.FromZone != zone.Library || consumer.Payment.Form != EffectPaymentFormUnknown {
			continue
		}
		for _, reference := range consumer.References {
			if !reference.LibraryCardObservation || reference.ProducerClauseID <= 0 {
				continue
			}
			var observation *EffectSyntax
			matches := 0
			for _, candidate := range effects {
				if candidate.ClauseID == reference.ProducerClauseID {
					if candidate.Kind == EffectLookAtLibraryTop || candidate.Kind == EffectReveal {
						observation = candidate
					}
					matches++
				}
			}
			if observation == nil {
				continue
			}
			if observation.Kind == EffectLookAtLibraryTop && !observation.Optional {
				continue
			}
			if slices.ContainsFunc(effects, func(candidate *EffectSyntax) bool {
				return candidate.ClauseID == segment.Ownership.ResultProducerClauseID &&
					candidate.ClauseID != observation.ClauseID &&
					(candidate.Optional || candidate.Payment.Form != EffectPaymentFormUnknown)
			}) {
				continue
			}
			if observation.Kind == EffectReveal {
				if segment.Ownership.ResultProducerClauseID != observation.ClauseID {
					return 0, true
				}
				continue
			}
			handled = true
			if (reference.Kind != ReferenceThatObject &&
				(reference.Kind != ReferencePronoun || reference.Pronoun != PronounIt && reference.Pronoun != PronounIts)) ||
				(reference.SubjectNoun != ObjectNounUnknown && reference.SubjectNoun != ObjectNounCard) {
				return 0, true
			}
			var canonical Reference
			nodes := 0
			for _, candidate := range references {
				if candidate.NodeID != reference.NodeID {
					continue
				}
				if candidate.Kind != reference.Kind || candidate.Pronoun != reference.Pronoun ||
					candidate.SubjectNoun != reference.SubjectNoun ||
					(reference.Order != (shared.SourceOrder{}) && candidate.Order != reference.Order) ||
					candidate.CardIdentity != reference.CardIdentity ||
					candidate.ProducerClauseID != reference.ProducerClauseID ||
					candidate.ReachedCardProducerClauseID != 0 || !candidate.LibraryCardObservation {
					return 0, true
				}
				canonical = candidate
				nodes++
			}
			if matches != 1 || nodes != 1 || !exactOptionalTargetObservation(*observation) ||
				!consumer.Exact || consumer.Negated || consumer.ClauseID <= 0 ||
				segment.Order.End <= segment.Order.Start ||
				consumer.Order.End <= consumer.Order.Start ||
				observation.Order.End > segment.Order.Start ||
				reference.ReachedCardProducerClauseID != 0 ||
				canonical.Order.End <= canonical.Order.Start ||
				segment.Order.End > canonical.Order.Start || !consumer.Order.Contains(canonical.Order) ||
				producer != 0 && producer != observation.ClauseID {
				return 0, true
			}
			producer = observation.ClauseID
		}
	}
	return producer, handled
}

func exactOptionalTargetObservation(effect EffectSyntax) bool {
	amount := effect.Amount
	if effect.Kind != EffectLookAtLibraryTop || !effect.Exact || !effect.Optional || effect.Negated ||
		effect.Context != EffectContextTarget || effect.Player != EffectPlayerNone ||
		effect.CardSource != EffectCardSourceTopOfPlayerLibrary ||
		effect.ClauseID <= 0 || effect.Order.End <= effect.Order.Start ||
		len(effect.Targets) != 1 || len(effect.SubjectReferences) != 0 || len(effect.References) != 0 ||
		!amount.Known || amount.Value != 1 || amount.RangeKnown || amount.VariableX || amount.AnyNumber ||
		amount.DynamicKind != EffectDynamicAmountNone {
		return false
	}
	target := effect.Targets[0]
	return target.Exact && target.Cardinality.Min == 1 && target.Cardinality.Max == 1 &&
		(target.Selection.Kind == SelectionPlayer || target.Selection.Kind == SelectionOpponent) &&
		target.Order.End > target.Order.Start && effect.Order.Contains(target.Order)
}
