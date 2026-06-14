package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// ConditionSegment is a pre-segmented condition clause the parser emits for the
// compiler. The parser owns introducer recognition, clause segmentation, and
// rendering; the compiler consumes the typed kind, span, and rendered text
// mechanically instead of scanning Oracle tokens for clause boundaries or
// rebuilding display text.
type ConditionSegment struct {
	// Kind is the grammatical introducer that opens the clause.
	Kind ConditionIntroKind
	// Span is the source span of the segmented clause.
	Span shared.Span
	// Text is the parser-rendered display spelling of the clause.
	Text string
	// Intervening reports that this clause is a triggered ability's
	// intervening-if. It is only ever set when the segments were emitted for a
	// triggered ability.
	Intervening bool
	// ActivationKeyword is the source span of an "Activate" keyword that
	// immediately precedes an "only if" introducer, or the zero span when absent.
	ActivationKeyword shared.Span
}

// ConditionSegments returns the ability's condition clauses, pre-segmented over
// the same semantic token stream the compiler historically scanned: a triggered
// ability's raw tokens (filtered to semantic tokens), or a non-triggered
// ability's resolving body tokens with the activation-timing clause removed.
func (a *Ability) ConditionSegments() []ConditionSegment {
	triggered := a.Kind == AbilityTriggered
	var tokens []shared.Token
	if triggered {
		tokens = eventHistorySemanticTokens(a.Tokens, a.Reminders, a.Quoted)
	} else {
		body := a.bodyTokens()
		if span, ok := a.activationTimingSpan(); ok {
			body = tokensOutsideParserSpan(body, span)
		}
		tokens = eventHistorySemanticTokens(body, a.Reminders, a.Quoted)
	}
	return conditionSegments(tokens, a.ConditionBoundaries(), triggered)
}

// TriggerConditionSegments returns the ability's condition clauses segmented over
// its raw tokens, used to locate a triggered ability's intervening-if condition.
func (a *Ability) TriggerConditionSegments() []ConditionSegment {
	return conditionSegments(a.Tokens, a.ConditionBoundaries(), true)
}

// ConditionSegments returns a modal option's condition clauses, pre-segmented
// over its semantic tokens.
func (m *Mode) ConditionSegments() []ConditionSegment {
	tokens := eventHistorySemanticTokens(m.Tokens, m.Reminders, m.Quoted)
	return conditionSegments(tokens, m.ConditionBoundaries, false)
}

// conditionSegments walks the supplied token stream, matching each typed
// boundary to a token by source position and segmenting its clause by token
// kind. Duration-skip boundaries are consumed without emitting a segment,
// matching the compiler's historical scan.
func conditionSegments(tokens []shared.Token, boundaries []ConditionBoundary, triggered bool) []ConditionSegment {
	var segments []ConditionSegment
	for i := 0; i < len(tokens); i++ {
		boundary, ok := conditionBoundaryAt(boundaries, tokens[i].Span.Start)
		if !ok {
			continue
		}
		end := conditionClauseEnd(tokens, i)
		if boundary.DurationSkip {
			i = end - 1
			continue
		}
		phrase := tokens[i:end]
		segments = append(segments, ConditionSegment{
			Kind:              boundary.Kind,
			Span:              shared.SpanOf(phrase),
			Text:              joinTokens(phrase),
			Intervening:       triggered && boundary.Intervening,
			ActivationKeyword: boundary.ActivationKeyword,
		})
		i = end - 1
	}
	return segments
}

func conditionBoundaryAt(boundaries []ConditionBoundary, position shared.Position) (ConditionBoundary, bool) {
	for _, boundary := range boundaries {
		if boundary.Start.Offset == position.Offset {
			return boundary, true
		}
	}
	return ConditionBoundary{}, false
}
