package parser

import "slices"

func resultConditionPredicate(predicate ConditionPredicateKind) bool {
	return predicate == ConditionPredicateResultThisWay ||
		predicate == ConditionPredicateCounterSucceeded ||
		predicate == ConditionPredicatePriorInstructionAccepted ||
		predicate == ConditionPredicatePriorInstructionNotAccepted
}

// Resolve result grammar once, while the parser still owns sentence structure.
func emitResultConditionOwnership(sentences []Sentence, segments []ConditionSegment, clauses []ConditionClause) {
	effects := conditionEffects(sentences)
	emitOptionalActionOwnership(sentences, segments)
	emitOptionalPostfixElseOwnership(effects, segments)
	resultOwner := make(map[int]int)
	for ci := range segments {
		segment := &segments[ci]
		if segment.ClauseIndex < 0 || segment.ClauseIndex >= len(clauses) ||
			!resultConditionPredicate(clauses[segment.ClauseIndex].Predicate) ||
			len(segment.Ownership.ClauseIDs) == 0 {
			continue
		}

		clause := clauses[segment.ClauseIndex]
		first := slices.IndexFunc(effects, func(effect *EffectSyntax) bool {
			return effect.ClauseID == segment.Ownership.ClauseIDs[0]
		})
		if first <= 0 {
			continue
		}
		producer := first - 1
		switch {
		case clause.Predicate == ConditionPredicateCounterSucceeded:
			producer = slices.IndexFunc(effects, func(effect *EffectSyntax) bool {
				return effect.ClauseID == segment.Ownership.ResultProducerClauseID
			})
		case clause.Predicate == ConditionPredicateResultThisWay:
			for producer >= 0 && effects[producer].Kind != clause.ThisWayOutcome {
				producer--
			}
		case clause.Predicate == ConditionPredicatePriorInstructionNotAccepted && !effects[producer].Optional:
			// "If you don't" complements the action, not an intervening
			// filtered predicate or its consequence.
			if previous, exists := resultOwner[effects[producer].ClauseID]; exists {
				producer = slices.IndexFunc(effects, func(effect *EffectSyntax) bool {
					return effect.ClauseID == segments[previous].Ownership.ResultProducerClauseID
				})
			}
		default:
		}
		if producer < 0 {
			continue
		}
		if clause.Predicate != ConditionPredicateResultThisWay && producer > 0 &&
			effects[producer].Span == effects[producer-1].Span &&
			effects[producer-1].Kind != EffectGrantKeyword {
			continue
		}
		segment.Ownership.ResultProducerClauseID = effects[producer].ClauseID
		for _, id := range segment.Ownership.ClauseIDs {
			resultOwner[id] = ci
		}
		// A new sentence naming the same subject continues the consequence.
		// A fresh independent subject or condition ends this scope.
		last := slices.IndexFunc(effects, func(effect *EffectSyntax) bool {
			return effect.ClauseID == segment.Ownership.ClauseIDs[len(segment.Ownership.ClauseIDs)-1]
		})
		for next := last + 1; next < len(effects); next++ {
			effect := effects[next]
			if !resultSubjectContinuation(effect, segments) {
				break
			}
			segment.Ownership.ClauseIDs = append(segment.Ownership.ClauseIDs, effect.ClauseID)
			segment.Ownership.Scope = ConditionScopeGroup
			resultOwner[effect.ClauseID] = ci
		}
	}
	for ei, effect := range effects {
		if effect.Connection != EffectConnectionOtherwise || ei == 0 {
			continue
		}
		ci, exists := resultOwner[effects[ei-1].ClauseID]
		if !exists {
			continue
		}
		segment := segments[ci]
		if segment.Ownership.ResultProducerClauseID == 0 ||
			clauses[segment.ClauseIndex].Predicate == ConditionPredicatePriorInstructionNotAccepted {
			continue
		}
		j := ei
		for ; j < len(effects) && effects[j].Span == effect.Span; j++ {
			effects[j].ResultElseOfClauseID = segment.Ownership.ClauseIDs[0]
		}
		for ; j < len(effects) && resultSubjectContinuation(effects[j], segments); j++ {
			effects[j].ResultElseOfClauseID = segment.Ownership.ClauseIDs[0]
		}
	}
}

func resultSubjectContinuation(effect *EffectSyntax, segments []ConditionSegment) bool {
	return effect.Connection != EffectConnectionOtherwise &&
		(effect.Context == EffectContextReferencedObject || effect.Context == EffectContextPriorSubject) &&
		!slices.ContainsFunc(segments, func(segment ConditionSegment) bool {
			return slices.Contains(segment.Ownership.ClauseIDs, effect.ClauseID)
		})
}

func emitOptionalActionOwnership(sentences []Sentence, segments []ConditionSegment) {
	for si := range sentences {
		emitOptionalEffectGroups(sentences[si].Effects, segments)
	}
}

func emitOptionalEffectGroups(effects []EffectSyntax, segments []ConditionSegment) {
	for ei := range effects {
		effect := &effects[ei]
		emitOptionalEffectGroups(effect.RepeatBody, segments)
		if !effect.Optional {
			continue
		}
		effect.OptionalActionClauseIDs = []int{effect.ClauseID}
		for next := ei + 1; next < len(effects); next++ {
			consequence := &effects[next]
			if consequence.Span != effect.Span || consequence.Optional ||
				slices.ContainsFunc(segments, func(segment ConditionSegment) bool {
					return slices.Contains(segment.Ownership.ClauseIDs, consequence.ClauseID) &&
						!slices.Contains(segment.Ownership.ClauseIDs, effect.ClauseID)
				}) {
				break
			}
			effect.OptionalActionClauseIDs = append(effect.OptionalActionClauseIDs, consequence.ClauseID)
		}
	}
}
