package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// recognizeAnyOpponentDamageThisTurnCondition matches "an opponent was dealt N
// or more damage this turn" (Spinerock Knoll's Hideaway play gate). It reads the
// per-turn damage dealt to any single opponent, so Threshold carries the N of
// the "N or more" inclusive threshold. It fails closed on any other wording.
func recognizeAnyOpponentDamageThisTurnCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "an", "opponent", "was", "dealt")
	if !ok || len(rest) < 1 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(rest[1:], "or", "more", "damage", "this", "turn") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast,
		Threshold: value,
	}, true
}

// recognizeAnyOpponentLostLifeThisTurnCondition matches "an opponent lost N or
// more life this turn" (Bloodchief Ascension's each-end-step quest-counter
// intervening-if). It reads the per-turn life lost by any single opponent, so
// Threshold carries the N of the "N or more" inclusive threshold. Every kind of
// life loss counts — combat and noncombat damage (CR 120.3, the printed "(Damage
// causes loss of life.)" reminder), life paid as a cost, and direct loss — which
// the runtime aggregate reflects. It is the life-loss counterpart of
// recognizeAnyOpponentDamageThisTurnCondition and fails closed on any other
// wording.
func recognizeAnyOpponentLostLifeThisTurnCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "an", "opponent", "lost")
	if !ok || len(rest) < 1 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(rest[1:], "or", "more", "life", "this", "turn") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast,
		Threshold: value,
	}, true
}

// recognizeAnyLibrarySizeCondition matches "a library has N or fewer cards in
// it" (Shelldock Isle's Hideaway play gate). "A library" is any player's
// library, so the predicate is an existential over all players' library sizes;
// Threshold carries the N of the "N or fewer" inclusive ceiling. It fails closed
// on any other wording.
func recognizeAnyLibrarySizeCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "a", "library", "has")
	if !ok || len(rest) < 1 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(rest[1:], "or", "fewer", "cards", "in", "it") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateAnyLibrarySizeAtMost,
		Threshold: value,
	}, true
}

// recognizeAllPlayersHandEmptyCondition matches "each player has no cards in
// hand" (Howltooth Hollow's Hideaway play gate), a universal hand-empty gate
// over every player. It fails closed on any other wording.
func recognizeAllPlayersHandEmptyCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "each", "player", "has", "no", "cards", "in", "hand") {
		return ConditionClause{Predicate: ConditionPredicateAllPlayersHandEmpty}, true
	}
	return ConditionClause{}, false
}
