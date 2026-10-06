package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func recognizeCounterSucceededCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "that", "spell", "is", "countered", "this", "way") ||
		tokenWordsEqual(body, "that", "spell", "was", "countered", "this", "way") {
		return ConditionClause{Predicate: ConditionPredicateCounterSucceeded}, true
	}
	return ConditionClause{}, false
}

func emitCounterResultOwnership(sentences []Sentence, segments []ConditionSegment, clauses []ConditionClause, references []Reference) {
	effects := conditionEffects(sentences)
	var targets []TargetSyntax
	for _, sentence := range sentences {
		targets = append(targets, sentence.Targets...)
	}
	for ci := range segments {
		segment := &segments[ci]
		if segment.ClauseIndex < 0 || segment.ClauseIndex >= len(clauses) ||
			clauses[segment.ClauseIndex].Predicate != ConditionPredicateCounterSucceeded {
			continue
		}
		segment.Ownership.ResultSubjectReferenceNodeID = -1
		segment.Ownership.ResultSubjectTargetOccurrence = -1
		var subject *Reference
		for ri := range references {
			reference := &references[ri]
			if reference.Kind == ReferenceThatObject && slices.Contains(segment.Ownership.ReferenceNodeIDs, reference.NodeID) {
				if subject != nil {
					subject = nil
					break
				}
				subject = reference
			}
		}
		if subject == nil {
			continue
		}
		segment.Ownership.ResultSubjectReferenceNodeID = subject.NodeID
		// A specific demonstrative must name the nearest target occurrence,
		// not merely the nearest effect with a matching counter verb.
		occurrence := -1
		for ti, target := range targets {
			if target.Span.End.Offset <= subject.Span.Start.Offset &&
				(occurrence < 0 || target.Span.Start.Offset > targets[occurrence].Span.Start.Offset) {
				occurrence = ti
			}
		}
		if occurrence < 0 {
			continue
		}
		nearest := targets[occurrence]
		var owner *EffectSyntax
		for _, effect := range effects {
			if slices.ContainsFunc(effect.Targets, func(target TargetSyntax) bool { return target.Span == nearest.Span }) {
				if owner != nil {
					owner = nil
					break
				}
				owner = effect
			}
		}
		if owner == nil || owner.Kind != EffectCounter || len(owner.Targets) != 1 ||
			nearest.Selection.Kind != SelectionSpell ||
			nearest.Cardinality.Min != 1 || nearest.Cardinality.Max != 1 {
			continue
		}
		segment.Ownership.ResultProducerClauseID = owner.ClauseID
		segment.Ownership.ResultSubjectTargetOccurrence = occurrence
	}
}
