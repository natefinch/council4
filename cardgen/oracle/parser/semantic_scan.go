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
	for i := range a.Sentences {
		if a.Sentences[i].RegenerationRider {
			tokens = tokensOutsideParserSpan(tokens, a.Sentences[i].Span)
		}
	}
	for i := range a.StaticDeclarations {
		if staticDeclarationConsumesReferences(&a.StaticDeclarations[i]) {
			tokens = tokensOutsideParserSpan(tokens, a.StaticDeclarations[i].Span)
		}
	}
	return a.Atoms.ReferencesWithin(tokens)
}

// staticDeclarationConsumesReferences reports whether a fully recognized static
// declaration absorbs the pronouns within its span as part of a fixed idiom
// rather than as object back-references. The each-player additional-land rule
// ("... on each of their turns.") owns its "their", so that pronoun must not
// surface as a dangling semantic reference that would block the text-blind
// compiler's empty-content recognition of the player rule.
func staticDeclarationConsumesReferences(declaration *StaticDeclarationSyntax) bool {
	return declaration.Kind == StaticDeclarationPlayerRule &&
		declaration.Subject.Kind == StaticDeclarationSubjectEachPlayer
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
	tokens = stripCreatureSpellHasteRiderTokens(tokens)
	return a.Atoms.KeywordsWithin(tokens)
}

// computeSemanticReferences returns the explicit references recognized in a modal
// option's semantic tokens, with rendered Text.
func (m *Mode) computeSemanticReferences() []Reference {
	tokens := eventHistorySemanticTokens(m.Body.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.ReferencesWithin(tokens)
}

// computeSemanticKeywords returns the keywords recognized in a modal option's
// semantic tokens.
func (m *Mode) computeSemanticKeywords() []Keyword {
	tokens := eventHistorySemanticTokens(m.Body.Tokens, m.Reminders, m.Quoted)
	return m.Atoms.KeywordsWithin(tokens)
}

// linkConditionSegments resolves, for each segment, the index of the typed
// condition clause and event-history condition that fill the segment's span, so
// the compiler reads the matching clause by index instead of scanning for an
// equal span. The first clause or event history whose span equals the segment's
// span wins, matching the compiler's historical first-match scan.
func linkConditionSegments(
	segments []ConditionSegment,
	clauses []ConditionClause,
	eventHistories []EventHistoryCondition,
) {
	for s := range segments {
		for i := range clauses {
			if clauses[i].Span == segments[s].Span {
				segments[s].ClauseIndex = i
				break
			}
		}
		for i := range eventHistories {
			if eventHistories[i].Span == segments[s].Span {
				segments[s].EventHistoryIndex = i
				break
			}
		}
	}
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
		linkConditionSegments(ability.ConditionSegments, ability.ConditionClauses, ability.EventHistoryConditions)
		linkConditionSegments(ability.TriggerConditionSegments, ability.ConditionClauses, ability.EventHistoryConditions)
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			mode.SemanticReferences = mode.computeSemanticReferences()
			mode.SemanticKeywords = mode.computeSemanticKeywords()
			mode.ConditionSegments = mode.computeConditionSegments()
			linkConditionSegments(mode.ConditionSegments, mode.ConditionClauses, mode.EventHistoryConditions)
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

// CoverageSpans returns the source spans of the ability's tokens that must be
// accounted for by recognized semantics for the ability to be fully consumed:
// every token except the structural sentence punctuation the parser owns (the
// commas, colons, and periods that separate clauses and costs). A consumer
// asserts each returned span is covered by a recognized semantic span, instead
// of walking the raw token stream and classifying token kinds itself. Reminder,
// quoted, and separator tokens are retained: a consumer accounts for them
// through the matching typed span (a reminder span, an effect span, or the
// ability-word/chapter separator span) so that an ability with un-recognized
// reminder or separator text still fails closed.
func (a *Ability) CoverageSpans() []shared.Span {
	return coverageSpans(a.Tokens)
}

// CoverageSpans returns the modal option's must-cover token spans, with the same
// structural exclusions as Ability.CoverageSpans.
func (m *Mode) CoverageSpans() []shared.Span {
	return coverageSpans(m.Tokens)
}

func coverageSpans(tokens []shared.Token) []shared.Span {
	spans := make([]shared.Span, 0, len(tokens))
	for _, token := range tokens {
		switch token.Kind {
		case shared.Comma, shared.Colon, shared.Period:
			continue
		default:
			spans = append(spans, token.Span)
		}
	}
	return spans
}
