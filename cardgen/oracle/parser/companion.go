package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// CompanionClause is the recognized companion keyword ability (CR 702.139). The
// deckbuilding condition that follows "Companion —" (and the from-outside-the-game
// reminder) is a sideboard and deck-construction restriction the deterministic
// playtester does not enforce, so the parser captures only the keyword identity
// and the condition's source span; downstream stages model companion as an inert
// recognized static keyword.
type CompanionClause struct {
	// Span is the source span of the whole companion ability paragraph.
	Span shared.Span `json:"-"`
	// ConditionSpan is the source span of the deckbuilding condition: the text
	// after "Companion —" for the standard form, or the zero span for the
	// "<X>'s companion" partner variant, whose condition lives entirely in its
	// reminder.
	ConditionSpan shared.Span `json:"-"`
}

// emitCompanionAbility recognizes the companion keyword ability on each static
// paragraph and, when found, records the typed companion clause and clears the
// paragraph's competing effect, declaration, and condition semantics so the
// compiler sees an empty static shell. The companion deckbuilding condition is
// free-form rules-irrelevant text for the goldfish playtester, so capturing the
// keyword identity and covering the span is sufficient.
func emitCompanionAbility(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		clause, ok := recognizeCompanion(ability)
		if !ok {
			continue
		}
		ability.Companion = &clause
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

// recognizeCompanion reports whether a static paragraph is a companion keyword
// ability and returns its typed clause. It recognizes the standard
// "Companion — <deckbuilding condition>" form by its "Companion" ability-word
// label and the "<X>'s companion" partner variant by the companion keyword word
// in the paragraph body outside its reminder.
func recognizeCompanion(ability *Ability) (CompanionClause, bool) {
	if ability.Kind != AbilityStatic || ability.Modal != nil {
		return CompanionClause{}, false
	}
	if ability.AbilityWord != nil && ability.AbilityWord.Label == "Companion" {
		return CompanionClause{
			Span:          ability.Span,
			ConditionSpan: shared.SpanOf(ability.bodyTokens()),
		}, true
	}
	if ability.AbilityWord != nil {
		return CompanionClause{}, false
	}
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	for _, keyword := range ability.Atoms.KeywordsWithin(tokens) {
		if keyword.Kind == KeywordCompanion {
			return CompanionClause{Span: ability.Span}, true
		}
	}
	return CompanionClause{}, false
}
