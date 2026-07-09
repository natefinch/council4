package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// KeywordShareGrantClause is the recognized team keyword-sharing construct
// (Odric, Lunarch Marshal, CR 702 keyword statics granted en masse). It carries
// the ordered list of shared keyword kinds — the lead keyword named by the first
// resolving sentence followed by every keyword named in the optional "The same
// is true for ..." sentence. Each keyword lowers to one continuous group grant
// gated on the controller controlling a creature that already has that keyword,
// so the whole construct grants each keyword to all the controller's creatures
// until end of turn exactly when at least one of them already has it. Spans
// preserve the consumed sentences' source for coverage.
type KeywordShareGrantClause struct {
	// Keywords are the shared keyword kinds in printed order, beginning with the
	// lead keyword. Each is a simple, unparameterized keyword; the recognizer
	// fails closed on any keyword that carries a parameter.
	Keywords []KeywordKind `json:",omitempty"`
	// Spans are the source spans of the consumed sentences, in order.
	Spans []shared.Span `json:"-"`
	// ConstructSpan covers the whole recognized construct across its sentences.
	ConstructSpan shared.Span `json:"-"`
}

// emitKeywordShareGrants folds the team keyword-sharing family (Odric, Lunarch
// Marshal) onto each phase/step triggered ability whose body matches the exact
// "creatures you control gain <KW> until end of turn if a creature you control
// has <KW>." shape, optionally extended by "The same is true for <KW>, ..., and
// <KW>.". It runs after resolving syntax, atoms, and semantic keywords are
// emitted so the keyword atoms and sentence effects are already classified. When
// the shape matches it records the ordered keyword kinds on
// ability.KeywordShareGrant and strips the consumed sentences' effects and the
// ability's keyword and condition semantics, so the compiler lowers the typed
// construct instead of the per-effect "if a creature you control has <KW>"
// condition it does not model as a state predicate.
func emitKeywordShareGrants(abilities []Ability) {
	for i := range abilities {
		recognizeKeywordShareGrant(&abilities[i])
	}
}

// recognizeKeywordShareGrant matches an ability whose trigger is a phase/step
// trigger ("At the beginning of each combat", "At the beginning of each
// upkeep") and whose resolving body is exactly the keyword-sharing shape. It
// records the ordered keyword kinds on ability.KeywordShareGrant and strips the
// consumed effect, keyword, and condition semantics so coverage credits the
// whole construct and the compiler lowers it as a keyword share. It fails closed
// (leaving the ability untouched) for any other shape: a non-phase trigger, a
// competing recognized construct, a lead sentence that does not match the
// "creatures you control gain <KW> until end of turn if a creature you control
// has <KW>" wording, a mismatched lead/gate keyword, a trailing sentence that is
// not the "The same is true for ..." list, or any parameterized keyword.
func recognizeKeywordShareGrant(ability *Ability) {
	if ability.KeywordShareGrant != nil {
		return
	}
	if ability.Trigger == nil || ability.Trigger.PhaseStep == nil {
		return
	}
	if ability.Modal != nil || ability.DiceTable != nil || ability.CoinFlip != nil ||
		ability.Vote != nil || ability.EachPlayerChooseDestroy != nil || ability.ExactSequence != nil {
		return
	}
	if len(ability.Sentences) < 1 || len(ability.Sentences) > 2 {
		return
	}

	byStart := keywordShareAtomsByStart(ability)
	lead, ok := matchKeywordShareLead(ability.Sentences[0].Tokens, byStart)
	if !ok {
		return
	}
	keywords := []KeywordKind{lead}
	if len(ability.Sentences) == 2 {
		more, ok := matchKeywordShareSame(ability.Sentences[1].Tokens, byStart)
		if !ok {
			return
		}
		keywords = append(keywords, more...)
	}
	for _, keyword := range keywords {
		if !keywordShareGrantable(keyword) {
			return
		}
	}

	spans := make([]shared.Span, len(ability.Sentences))
	construct := ability.Sentences[0].Span
	for i := range ability.Sentences {
		spans[i] = ability.Sentences[i].Span
		ability.Sentences[i].Effects = nil
		ability.Sentences[i].LegacyEffects = false
		if ability.Sentences[i].Span.End.Offset > construct.End.Offset {
			construct.End = ability.Sentences[i].Span.End
		}
	}
	ability.KeywordShareGrant = &KeywordShareGrantClause{
		Keywords:      keywords,
		Spans:         spans,
		ConstructSpan: construct,
	}
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
	ability.TriggerConditionSegments = nil
}

// keywordShareAtomsByStart indexes the ability's recognized keyword atoms by the
// source offset where their printed name begins, so the recognizer can identify
// a keyword that begins at a given token without re-recognizing spelling.
func keywordShareAtomsByStart(ability *Ability) map[int]Keyword {
	keywords := ability.Atoms.KeywordsWithin(ability.Tokens)
	byStart := make(map[int]Keyword, len(keywords))
	for _, keyword := range keywords {
		byStart[keyword.NameSpan.Start.Offset] = keyword
	}
	return byStart
}

// keywordShareSpanAt returns the keyword whose printed name begins at tokens[0]
// and the number of leading tokens that name spans, using the offset index. ok
// is false when no keyword begins exactly at tokens[0].
func keywordShareSpanAt(tokens []shared.Token, byStart map[int]Keyword) (Keyword, int, bool) {
	if len(tokens) == 0 {
		return Keyword{}, 0, false
	}
	keyword, ok := byStart[tokens[0].Span.Start.Offset]
	if !ok {
		return Keyword{}, 0, false
	}
	count := 0
	for count < len(tokens) && tokens[count].Span.Start.Offset < keyword.Span.End.Offset {
		count++
	}
	if count == 0 {
		return Keyword{}, 0, false
	}
	return keyword, count, true
}

// matchKeywordShareLead matches the lead sentence "creatures you control gain
// <KW> until end of turn if a creature you control has <KW>." and returns the
// shared keyword kind. It requires the lead grant keyword and the gate keyword
// to be the same kind and the sentence to end after the gate keyword's period.
func matchKeywordShareLead(tokens []shared.Token, byStart map[int]Keyword) (KeywordKind, bool) {
	rest, ok := cutSyntaxWords(tokens, "creatures", "you", "control", "gain")
	if !ok {
		return "", false
	}
	granted, count, ok := keywordShareSpanAt(rest, byStart)
	if !ok {
		return "", false
	}
	rest = rest[count:]
	rest, ok = cutSyntaxWords(rest, "until", "end", "of", "turn", "if", "a", "creature", "you", "control", "has")
	if !ok {
		return "", false
	}
	gate, gateCount, ok := keywordShareSpanAt(rest, byStart)
	if !ok || gate.Kind != granted.Kind {
		return "", false
	}
	rest = rest[gateCount:]
	if len(rest) != 1 || rest[0].Kind != shared.Period {
		return "", false
	}
	return granted.Kind, true
}

// matchKeywordShareSame matches the extension sentence "The same is true for
// <KW>, <KW>, ..., [and ]<KW>." and returns the listed keyword kinds in order.
// It accepts both an Oxford comma ("..., and X.") and a bare conjunction
// ("... and X.") before the final keyword and requires the sentence to end after
// the final keyword's period.
func matchKeywordShareSame(tokens []shared.Token, byStart map[int]Keyword) ([]KeywordKind, bool) {
	rest, ok := cutSyntaxWords(tokens, "the", "same", "is", "true", "for")
	if !ok {
		return nil, false
	}
	var kinds []KeywordKind
	for {
		keyword, count, ok := keywordShareSpanAt(rest, byStart)
		if !ok {
			return nil, false
		}
		kinds = append(kinds, keyword.Kind)
		rest = rest[count:]
		switch {
		case len(rest) == 1 && rest[0].Kind == shared.Period:
			return kinds, true
		case len(rest) >= 1 && rest[0].Kind == shared.Comma:
			rest = rest[1:]
			if len(rest) > 0 && equalWord(rest[0], "and") {
				rest = rest[1:]
			}
		case len(rest) >= 1 && equalWord(rest[0], "and"):
			rest = rest[1:]
		default:
			return nil, false
		}
		if len(rest) == 0 {
			return nil, false
		}
	}
}

// keywordShareGrantable reports whether a keyword may participate in a keyword
// share: a simple, unparameterized static keyword the lowering can grant to a
// group of creatures. The set is exactly the intersection of the lowering's
// keyword-kind runtime map and the game engine's grantable static keyword
// whitelist; any other keyword fails closed so the whole construct does not
// generate rather than silently dropping the keyword or its gate.
func keywordShareGrantable(kind KeywordKind) bool {
	switch kind {
	case KeywordDeathtouch,
		KeywordDefender,
		KeywordDoubleStrike,
		KeywordFirstStrike,
		KeywordFlying,
		KeywordHaste,
		KeywordHexproof,
		KeywordIndestructible,
		KeywordLifelink,
		KeywordMenace,
		KeywordReach,
		KeywordShroud,
		KeywordSkulk,
		KeywordTrample,
		KeywordVigilance,
		KeywordWither,
		KeywordInfect,
		KeywordExalted,
		KeywordProwess,
		KeywordRiot,
		KeywordEvolve,
		KeywordUnleash,
		KeywordFear,
		KeywordIntimidate:
		return true
	default:
		return false
	}
}
