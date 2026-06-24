package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// ChooseABackgroundClause is the recognized "Choose a Background" keyword ability
// (CR 702.124f): "You can have a Background as a second commander." The
// permission is a deck-construction mechanic the deterministic playtester does
// not simulate, so the parser captures only the keyword identity and the
// paragraph span; downstream stages model choose-a-background as an inert
// recognized static keyword.
type ChooseABackgroundClause struct {
	// Span is the source span of the whole choose-a-background ability paragraph.
	Span shared.Span `json:"-"`
}

// emitChooseABackgroundAbility recognizes the "Choose a Background" keyword
// ability on each static paragraph and, when found, records the typed clause and
// clears the paragraph's competing effect, declaration, and condition semantics
// so the compiler sees an empty static shell. The second-commander permission and
// its reminder are rules-irrelevant for the goldfish playtester, so capturing the
// keyword identity and covering the span is sufficient.
func emitChooseABackgroundAbility(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		clause, ok := recognizeChooseABackground(ability)
		if !ok {
			continue
		}
		ability.ChooseABackground = &clause
		ability.AbilityWord = nil
		ability.Sentences = nil
		ability.StaticDeclarations = nil
		ability.ConditionBoundaries = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.EventHistoryConditions = nil
	}
}

// recognizeChooseABackground reports whether a static paragraph is a "Choose a
// Background" keyword ability and returns its typed clause. It recognizes the
// keyword by the choose-a-background keyword word in the paragraph body (outside
// its second-commander reminder).
func recognizeChooseABackground(ability *Ability) (ChooseABackgroundClause, bool) {
	if ability.Kind != AbilityStatic || ability.Modal != nil {
		return ChooseABackgroundClause{}, false
	}
	if ability.AbilityWord != nil {
		return ChooseABackgroundClause{}, false
	}
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	for _, keyword := range ability.Atoms.KeywordsWithin(tokens) {
		if keyword.Kind == KeywordChooseABackground {
			return ChooseABackgroundClause{Span: ability.Span}, true
		}
	}
	return ChooseABackgroundClause{}, false
}
