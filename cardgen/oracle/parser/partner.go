package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// PartnerClause is the recognized "Partner" keyword ability (CR 702.124a) and
// its "Partner—<quality>" restricted variants (CR 702.124f, e.g.
// "Partner—Survivors", "Partner—Character select"). Partner is a
// deck-construction permission letting two such commanders share a deck; the
// restricted variants only narrow which other partner cards a card may pair
// with. The permission and pairing restrictions are mechanics the deterministic
// playtester does not simulate, so the parser captures only the keyword identity
// and the paragraph span; downstream stages model partner as an inert recognized
// static keyword.
type PartnerClause struct {
	// Span is the source span of the whole partner ability paragraph.
	Span shared.Span `json:"-"`
}

// emitPartnerAbility recognizes the "Partner" keyword ability (and its
// "Partner—<quality>" restricted variants) on each static paragraph and, when
// found, records the typed partner clause and clears the paragraph's competing
// effect, declaration, condition, and ability-word semantics so the compiler
// sees an empty static shell. The pairing permission and its reminder are
// rules-irrelevant for the goldfish playtester, so capturing the keyword
// identity and covering the span is sufficient.
//
// It runs after emitPartnerWithAbility so the longer "Partner with <name>"
// keyword is recognized first and never reclassified as plain partner.
func emitPartnerAbility(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.PartnerWith != nil {
			continue
		}
		clause, ok := recognizePartner(ability)
		if !ok {
			continue
		}
		ability.Partner = &clause
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

// recognizePartner reports whether a static paragraph is a "Partner" keyword
// ability and returns its typed clause. It recognizes the "Partner—<quality>"
// restricted variants by their "Partner" ability-word label (the em dash that
// introduces the quality parses as an ability-word separator) and the plain
// "Partner" form by a sole "Partner" word in the paragraph body outside its
// reminder.
func recognizePartner(ability *Ability) (PartnerClause, bool) {
	if ability.Kind != AbilityStatic || ability.Modal != nil {
		return PartnerClause{}, false
	}
	if ability.AbilityWord != nil {
		if ability.AbilityWord.Label == "Partner" {
			return PartnerClause{Span: ability.Span}, true
		}
		return PartnerClause{}, false
	}
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	if len(tokens) == 1 && equalWord(tokens[0], "Partner") {
		return PartnerClause{Span: ability.Span}, true
	}
	return PartnerClause{}, false
}
