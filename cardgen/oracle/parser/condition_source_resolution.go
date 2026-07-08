package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// recognizeSourceAbilityResolutionOrdinalCondition matches "this is the Nth time
// this ability has resolved this turn" (Prowl, Pursuit Vehicle's back-face
// enters trigger: "If this is the second time this ability has resolved this
// turn, convert Prowl."). It reads the ordinal word (second -> 2) into Threshold
// and gates its consequence on the resolving triggered ability's per-turn
// resolution count reaching exactly N. It fails closed on any other wording.
func recognizeSourceAbilityResolutionOrdinalCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "this", "is", "the")
	if !ok || len(rest) == 0 {
		return ConditionClause{}, false
	}
	ordinal, ok := OrdinalWordValue(rest[0].Text)
	if !ok || ordinal <= 0 {
		return ConditionClause{}, false
	}
	rest = rest[1:]
	if !tokenWordsEqual(rest, "time", "this", "ability", "has", "resolved", "this", "turn") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateSourceAbilityResolutionOrdinalThisTurn,
		Threshold: ordinal,
	}, true
}
