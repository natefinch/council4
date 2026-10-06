package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// The same passive "this way" grammar can name a cost or a resolving action.
// Only an exact singular cost with no competing body producer is rewritten to
// an object predicate; ordinary action-result predicates are left untouched.
func emitPaidCostThisWaySubjects(abilities []Ability) {
	for ai := range abilities {
		ability := &abilities[ai]
		preserveExistingSacrificeResult(abilities, ai, ability.ConditionClauses,
			&ability.SemanticReferences, ability.Sentences, ability.Tokens, ability.Atoms)
		if ability.Kind == AbilityActivated || ability.Kind == AbilitySpell {
			emitPaidCostThisWayContent(abilities, ai, ability.ConditionClauses,
				&ability.SemanticReferences, ability.Sentences, ability.Tokens, ability.Atoms)
		}
		if ability.Modal != nil {
			for mi := range ability.Modal.Options {
				mode := &ability.Modal.Options[mi]
				preserveExistingSacrificeResult(abilities, ai, mode.ConditionClauses,
					&mode.SemanticReferences, mode.Sentences, mode.Tokens, mode.Atoms)
				if ability.Kind == AbilityActivated || ability.Kind == AbilitySpell {
					emitPaidCostThisWayContent(abilities, ai, mode.ConditionClauses,
						&mode.SemanticReferences, paidCostModeSentences(ability, mi), mode.Tokens, mode.Atoms)
				}
			}
		}
	}
}

func emitPaidCostThisWayContent(abilities []Ability, consumer int, conditions []ConditionClause,
	references *[]Reference, sentences []Sentence, tokens []shared.Token, atoms Atoms,
) {
	sentences = paidCostSpellSentences(abilities, consumer, sentences)
	for ci := range conditions {
		condition := &conditions[ci]
		if condition.Predicate != ConditionPredicateResultThisWay || condition.ThisWaySelection == nil {
			continue
		}
		var domain PaidCostDomain
		switch condition.ThisWayOutcome {
		case EffectSacrifice:
			domain = PaidCostDomainSacrificedPermanent
		case EffectDiscard:
			domain = PaidCostDomainDiscardedCard
		default:
			continue
		}
		if paidCostHasCompetingEffect(sentences, domain, condition.Span) {
			continue
		}
		nodeID := 1
		for _, reference := range *references {
			nodeID = max(nodeID, reference.NodeID+1)
		}
		reference := Reference{
			Kind: ReferencePaidCostSubject, NodeID: nodeID, Span: condition.ThisWaySelection.Span,
			PaidCost: &PaidCostBinding{Domain: domain},
		}
		candidate := []Reference{reference}
		bindPaidCostReferences(abilities, consumer, candidate, sentences)
		if !candidate[0].PaidCost.Known {
			continue
		}
		noun := tokensWithinParserSpan(tokens, condition.ThisWaySelection.Span)
		selection, ok := parseConditionSelection(noun, atoms)
		if !ok {
			continue
		}
		condition.Predicate = ConditionPredicateObjectMatches
		condition.Selection = selection
		condition.SubjectRefID = nodeID
		condition.SubjectSpan = reference.Span
		condition.HasSubjectSpan = true
		condition.SubjectPast = true
		condition.ThisWaySelection = nil
		condition.ThisWayOutcome = EffectUnknown
		condition.ThisWayCardNoun = false
		*references = append(*references, candidate[0])
	}
}
