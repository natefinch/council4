package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// computeSemanticReferences returns the explicit references recognized in the
// ability's semantic tokens: reminder and quoted spans removed, and the
// activation-timing clause removed for activated abilities.
func (a *Ability) computeSemanticReferences() []Reference {
	tokens := eventHistorySemanticTokens(a.Tokens, a.Reminders, a.Quoted)
	if span, ok := a.activationTimingSpan(); ok {
		tokens = tokensOutsideParserSpan(tokens, span)
	}
	return a.Atoms.ReferencesWithin(tokens)
}

// computeSemanticKeywords returns the keywords recognized in the ability's
// semantic body tokens: the resolving body after any ability-word, cost colon, or
// trigger comma prefix, with the activation-timing clause and reminder/quoted
// spans removed.
func (a *Ability) computeSemanticKeywords() []Keyword {
	body := a.bodyTokens()
	if span, ok := a.activationTimingSpan(); ok {
		body = tokensOutsideParserSpan(body, span)
	}
	tokens := eventHistorySemanticTokens(body, a.Reminders, a.Quoted)
	return a.Atoms.KeywordsWithin(tokens)
}

// computeSemanticReferences returns the explicit references recognized in a modal
// option's semantic tokens, with rendered Text.
func (m *Mode) computeSemanticReferences() []Reference {
	tokens := eventHistorySemanticTokens(m.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.ReferencesWithin(tokens)
}

// computeSemanticKeywords returns the keywords recognized in a modal option's
// semantic tokens.
func (m *Mode) computeSemanticKeywords() []Keyword {
	tokens := eventHistorySemanticTokens(m.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.KeywordsWithin(tokens)
}

// emitSemanticAccessors materializes the parser's on-demand semantic views as
// plain fields so the parse result is a serializable data structure. Each value
// is computed once, from the same fully populated ability and mode state the
// compiler historically read lazily, so the stored values are identical.
func emitSemanticAccessors(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		ability.SemanticReferences = ability.computeSemanticReferences()
		ability.SemanticKeywords = ability.computeSemanticKeywords()
		ability.ContentSpan = ability.computeContentSpan()
		ability.ConditionSegments = ability.computeConditionSegments()
		ability.TriggerConditionSegments = ability.computeTriggerConditionSegments()
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			mode.SemanticReferences = mode.computeSemanticReferences()
			mode.SemanticKeywords = mode.computeSemanticKeywords()
			mode.ConditionSegments = mode.computeConditionSegments()
		}
	}
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
		return tokensWithinParserSpan(a.Tokens, a.BodySpan)
	default:
	}
	return tokens
}

// computeContentSpan returns the source span of the ability's resolving content:
// the body tokens after the shell cost/timing extraction. For an optional
// triggered ability ("you may ...") the span begins after the "you may" choice.
func (a *Ability) computeContentSpan() shared.Span {
	body := a.bodyTokens()
	if span, ok := a.activationTimingSpan(); ok {
		body = tokensOutsideParserSpan(body, span)
	}
	span := shared.SpanOf(body)
	if a.Optional {
		semantic := eventHistorySemanticTokens(body, a.Reminders, a.Quoted)
		if len(semantic) >= 3 {
			span.Start = semantic[2].Span.Start
		}
	}
	return span
}
