package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// Leading object predicates and explicit player-result gates own their
// references separately from the gated action's subject.
func separateConditionReferences(effect *EffectSyntax, atoms Atoms) {
	intro, width := conditionIntroAt(effect.Tokens, 0)
	if intro != ConditionIntroIf {
		return
	}
	end := conditionClauseEnd(effect.Tokens, 0)
	if end >= len(effect.Tokens) || effect.Tokens[end].Kind != shared.Comma {
		return
	}
	clause, ok := recognizeConditionPredicate(effect.Tokens[width:end], atoms)
	object := ok && clause.Predicate == ConditionPredicateObjectMatches && clause.HasSubjectSpan
	playerResult := ok &&
		(clause.Predicate == ConditionPredicatePriorInstructionAccepted ||
			clause.Predicate == ConditionPredicatePriorInstructionNotAccepted) &&
		effectWordsAt(effect.Tokens, width, "the", "player")
	if !object && !playerResult {
		return
	}
	span := shared.SpanOf(effect.Tokens[:end])
	effect.References = referencesOutsideSpan(effect.References, span)
	effect.SubjectReferences = referencesOutsideSpan(effect.SubjectReferences, span)
}
