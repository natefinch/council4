package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// SemanticReferences returns the explicit references recognized in the ability's
// semantic tokens: reminder and quoted spans removed, and the activation-timing
// clause removed for activated abilities. The parser owns recognition and
// rendering, so the compiler consumes these typed references (including their
// rendered Text) instead of filtering raw token streams.
func (a *Ability) SemanticReferences() []Reference {
	tokens := eventHistorySemanticTokens(a.Tokens, a.Reminders, a.Quoted)
	if span, ok := a.activationTimingSpan(); ok {
		tokens = tokensOutsideParserSpan(tokens, span)
	}
	return a.Atoms.ReferencesWithin(tokens)
}

// SemanticKeywords returns the keywords recognized in the ability's semantic body
// tokens: the resolving body after any ability-word, cost colon, or trigger
// comma prefix, with the activation-timing clause and reminder/quoted spans
// removed. The compiler consumes these typed keywords instead of filtering raw
// token streams.
func (a *Ability) SemanticKeywords() []Keyword {
	body := a.bodyTokens()
	if span, ok := a.activationTimingSpan(); ok {
		body = tokensOutsideParserSpan(body, span)
	}
	tokens := eventHistorySemanticTokens(body, a.Reminders, a.Quoted)
	return a.Atoms.KeywordsWithin(tokens)
}

// SemanticReferences returns the explicit references recognized in a modal
// option's semantic tokens, with rendered Text.
func (m *Mode) SemanticReferences() []Reference {
	tokens := eventHistorySemanticTokens(m.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.ReferencesWithin(tokens)
}

// SemanticKeywords returns the keywords recognized in a modal option's semantic
// tokens.
func (m *Mode) SemanticKeywords() []Keyword {
	tokens := eventHistorySemanticTokens(m.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.KeywordsWithin(tokens)
}

// activationTimingSpan returns the source span of an activated ability's trailing
// "Activate only ..." restriction clause, when present. The compiler removes this
// span from the reference and keyword token sets; the parser owns the equivalent
// structural span so neither stage inspects token spelling.
func (a *Ability) activationTimingSpan() (shared.Span, bool) {
	if a.Kind != AbilityActivated || len(a.ActivationRestrictions) == 0 {
		return shared.Span{}, false
	}
	return shared.Span{
		Start: a.ActivationRestrictions[0].Span.Start,
		End:   a.ActivationRestrictions[len(a.ActivationRestrictions)-1].Span.End,
	}, true
}

// bodyTokens returns the ability's resolving body tokens: the tokens after an
// ability-word em dash, an activated/loyalty cost colon, or a triggered event
// comma. It selects the body structurally so the compiler need not slice tokens
// by Oracle punctuation.
func (a *Ability) bodyTokens() []shared.Token {
	tokens := a.Tokens
	if a.AbilityWord != nil {
		if dash := shared.TopLevelIndex(tokens, shared.EmDash); dash >= 0 {
			tokens = tokens[dash+1:]
		}
	}
	switch a.Kind {
	case AbilityActivated, AbilityLoyalty:
		if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case AbilityTriggered:
		return tokensWithinParserSpan(a.Tokens, a.BodySpan())
	default:
	}
	return tokens
}

// ContentSpan returns the source span of the ability's resolving content: the
// body tokens after the shell cost/timing extraction. For an optional triggered
// ability ("you may ...") the span begins after the "you may" choice. The parser
// owns this structural span so the compiler need not slice body tokens.
func (a *Ability) ContentSpan() shared.Span {
	body := a.bodyTokens()
	if span, ok := a.activationTimingSpan(); ok {
		body = tokensOutsideParserSpan(body, span)
	}
	span := shared.SpanOf(body)
	if a.Optional() {
		semantic := eventHistorySemanticTokens(body, a.Reminders, a.Quoted)
		if len(semantic) >= 3 {
			span.Start = semantic[2].Span.Start
		}
	}
	return span
}
