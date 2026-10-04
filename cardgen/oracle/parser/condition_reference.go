package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// A leading object's condition owns its reference, not the gated effect's
// subject. Separate the two before exact effect reconstruction.
func separateObjectConditionReferences(effect *EffectSyntax, atoms Atoms) {
	intro, width := conditionIntroAt(effect.Tokens, 0)
	if intro != ConditionIntroIf {
		return
	}
	end := conditionClauseEnd(effect.Tokens, 0)
	if end >= len(effect.Tokens) || effect.Tokens[end].Kind != shared.Comma {
		return
	}
	clause, ok := recognizeConditionPredicate(effect.Tokens[width:end], atoms)
	if !ok || clause.Predicate != ConditionPredicateObjectMatches || !clause.HasSubjectSpan {
		return
	}
	span := shared.SpanOf(effect.Tokens[:end])
	effect.References = referencesOutsideSpan(effect.References, span)
	effect.SubjectReferences = referencesOutsideSpan(effect.SubjectReferences, span)
}
