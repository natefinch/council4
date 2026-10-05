package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// Preserve the existing resolving-sacrifice grammar when there is no sacrifice
// cost antecedent. A paid subject must not steal an independent action's result.
func preserveExistingSacrificeResult(abilities []Ability, consumer int, conditions []ConditionClause,
	references *[]Reference, sentences []Sentence, tokens []shared.Token, atoms Atoms,
) {
	candidates, _ := paidCostComponents(abilities, consumer, PaidCostDomainSacrificedPermanent)
	if len(candidates) != 0 {
		return
	}
	for ci := range conditions {
		condition := &conditions[ci]
		if condition.Predicate != ConditionPredicateObjectMatches || !condition.HasSubjectSpan ||
			!paidCostHasCompetingEffect(sentences, PaidCostDomainSacrificedPermanent, condition.Span) {
			continue
		}
		clause := tokensWithinParserSpan(tokens, condition.Span)
		if len(clause) < 2 || !equalWord(clause[0], "if") {
			continue
		}
		legacy, ok := recognizePriorInstructionSubjectMatchCondition(clause[1:], atoms)
		if !ok {
			continue
		}
		subjectID := condition.SubjectRefID
		legacy.Span, legacy.Intro = condition.Span, condition.Intro
		*condition = legacy
		*references = slices.DeleteFunc(*references, func(reference Reference) bool {
			return reference.Kind == ReferencePaidCostSubject && reference.NodeID == subjectID
		})
	}
}
