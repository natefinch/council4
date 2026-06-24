package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// PartnerWithClause is the recognized "Partner with <name>" keyword ability (CR
// 702.124e). It names a specific partner card and grants an enters trigger that
// lets the chosen player tutor the named partner into hand, then shuffle. The
// "partner commander" deck-construction permission and the pair-fetch ETB are
// mechanics the deterministic playtester does not simulate, so the parser
// captures only the keyword identity and the named-partner source span;
// downstream stages model partner-with as an inert recognized static keyword.
type PartnerWithClause struct {
	// Span is the source span of the whole partner-with ability paragraph.
	Span shared.Span `json:"-"`
	// NameSpan is the source span of the named partner card following
	// "Partner with", up to (but excluding) the reminder text.
	NameSpan shared.Span `json:"-"`
}

// emitPartnerWithAbility recognizes the "Partner with <name>" keyword ability on
// each static paragraph and, when found, records the typed partner-with clause
// and clears the paragraph's competing effect, declaration, and condition
// semantics so the compiler sees an empty static shell. The named partner and
// its pair-fetch reminder are rules-irrelevant for the goldfish playtester, so
// capturing the keyword identity and covering the span is sufficient.
func emitPartnerWithAbility(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		clause, ok := recognizePartnerWith(ability)
		if !ok {
			continue
		}
		ability.PartnerWith = &clause
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

// recognizePartnerWith reports whether a static paragraph is a "Partner with
// <name>" keyword ability and returns its typed clause. It recognizes the
// keyword by the Partner-with keyword word at the head of the paragraph body
// (outside its pair-fetch reminder) and covers the named partner span up to the
// end of the paragraph's body.
func recognizePartnerWith(ability *Ability) (PartnerWithClause, bool) {
	if ability.Kind != AbilityStatic || ability.Modal != nil {
		return PartnerWithClause{}, false
	}
	if ability.AbilityWord != nil {
		return PartnerWithClause{}, false
	}
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	for _, keyword := range ability.Atoms.KeywordsWithin(tokens) {
		if keyword.Kind != KeywordPartnerWith {
			continue
		}
		return PartnerWithClause{
			Span:     ability.Span,
			NameSpan: shared.SpanOf(tokens),
		}, true
	}
	return PartnerWithClause{}, false
}
