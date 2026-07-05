package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// appendConstructRecognizedSpans credits three recognized constructs whose typed
// output stops short of the source tokens it accounts for, so the coverage harness
// can assert generated cards are parser-complete without the recognized-span union
// loosening to swallow an adjacent unrepresented clause. Each construct emits a
// span tightly bounded to grammar the parser already typed: a coordinated
// card-type/subtype list, a "for each" iteration prefix, or a reflexive/delayed
// trigger preamble. The compiler never consumes these spans, so lowering is
// unaffected.
func appendConstructRecognizedSpans(spans []shared.Span, a *Ability) []shared.Span {
	spans = appendCoordinatedTypeListSpans(spans, a.Tokens, a.Atoms)
	spans = appendLeadingClauseSpans(spans, a.Sentences)
	spans = appendCoinFlipSpans(spans, a)
	spans = appendVoteSpans(spans, a)
	spans = appendEachPlayerChooseDestroySpans(spans, a)
	spans = appendPileSplitSpans(spans, a)
	return spans
}

// appendPileSplitSpans credits the zero-effect middle sentence of a recognized
// pile-split sequence ("An opponent separates those cards into two piles." / "An
// opponent chooses one of those piles."). The recognizer typed the reveal and
// put effects and recorded the middle sentence's span on the put effect; that
// sentence produces no effect, so without crediting it the coverage union would
// leave its tokens uncovered.
func appendPileSplitSpans(spans []shared.Span, a *Ability) []shared.Span {
	for i := range a.Sentences {
		for j := range a.Sentences[i].Effects {
			effect := &a.Sentences[i].Effects[j]
			if effect.PileSplitSequence && effect.PileSplitMiddleSpan != (shared.Span{}) {
				spans = append(spans, effect.PileSplitMiddleSpan)
			}
		}
	}
	return spans
}

// appendCoinFlipSpans credits the source spans of every sentence a recognized
// coin flip consumed (the "Flip a coin." line and each win/lose branch). The
// recognizer re-parsed each branch clause into typed effects and shed the
// consumed sentences' effects and condition wording, so the construct fully
// accounts for its tokens; crediting the whole sentence spans keeps the coverage
// union from leaving the condition prefixes or the flip line uncovered.
func appendCoinFlipSpans(spans []shared.Span, a *Ability) []shared.Span {
	if a.CoinFlip == nil {
		return spans
	}
	for _, span := range a.CoinFlip.Spans {
		if span != (shared.Span{}) {
			spans = append(spans, span)
		}
	}
	return spans
}

// appendEachPlayerChooseDestroySpans credits the source spans of both sentences a
// recognized "Starting with you, each player may choose <permanent>. Destroy each
// permanent chosen this way." construct consumed. The recognizer re-parsed the
// candidate filter into a typed pool and shed both sentences' effects, so
// crediting the whole sentence spans keeps the coverage union from leaving the
// choose sentence (which carries no typed effect of its own) uncovered.
func appendEachPlayerChooseDestroySpans(spans []shared.Span, a *Ability) []shared.Span {
	if a.EachPlayerChooseDestroy == nil {
		return spans
	}
	for _, span := range a.EachPlayerChooseDestroy.Spans {
		if span != (shared.Span{}) {
			spans = append(spans, span)
		}
	}
	return spans
}

// appendVoteSpans credits the source spans of every sentence a recognized vote
// consumed (the "Starting with you, each player votes ..." line and each arm).
// The recognizer re-parsed each arm clause into typed effects and shed the
// consumed sentences' effects and condition wording, so the construct fully
// accounts for its tokens; crediting the whole sentence spans keeps the coverage
// union from leaving the condition prefixes or the voting line uncovered.
func appendVoteSpans(spans []shared.Span, a *Ability) []shared.Span {
	if a.Vote == nil {
		return spans
	}
	for _, span := range a.Vote.Spans {
		if span != (shared.Span{}) {
			spans = append(spans, span)
		}
	}
	return spans
}

// appendCoordinatedTypeListSpans credits a coordinated list of two or more
// card-type or subtype atoms joined by commas and "or"/"and" (optionally closed by
// a "spell" noun). The trigger spell-selection and condition-subject grammars type
// only the first list item ("an instant, sorcery, or Wizard spell" types just
// instant; "a Fish, Octopus, ... or Whale" types just Fish), leaving the remaining
// recognized type atoms and their list glue uncovered. The whole run is recognized
// grammar — every list item is a typed atom — so its span is credited as a unit.
func appendCoordinatedTypeListSpans(spans []shared.Span, tokens []shared.Token, atoms Atoms) []shared.Span {
	atomSpans := typeAtomSpans(atoms)
	if len(atomSpans) < 2 {
		return spans
	}
	for i := 0; i < len(tokens); {
		if !spanInUnion(tokens[i].Span, atomSpans) {
			i++
			continue
		}
		end, count := coordinatedTypeListRun(tokens, i, atomSpans)
		if count < 2 {
			i++
			continue
		}
		spans = append(spans, shared.SpanOf(tokens[i:end+1]))
		i = end + 1
	}
	return spans
}

// coordinatedTypeListRun scans forward from a type atom at start, consuming a
// coordinated list of distinct type atoms joined by commas and "or"/"and". It
// returns the index of the last token in the run (a type atom or a trailing
// "spell" noun) and the number of distinct type atoms consumed.
func coordinatedTypeListRun(tokens []shared.Token, start int, atomSpans []shared.Span) (end, count int) {
	end = start
	count = 1
	prevAtom, _ := coveringSpan(tokens[start].Span, atomSpans)
	for j := start + 1; j < len(tokens); j++ {
		token := tokens[j]
		if atom, ok := coveringSpan(token.Span, atomSpans); ok {
			if atom != prevAtom {
				count++
				prevAtom = atom
			}
			end = j
			continue
		}
		if token.Kind == shared.Comma || isListConjunction(token) {
			continue
		}
		break
	}
	if count >= 2 && end+1 < len(tokens) && isSpellNoun(tokens[end+1]) {
		end++
	}
	return end, count
}

// typeAtomSpans returns the source spans of every card-type and subtype atom in
// an ability, the recognized list items a coordinated type list is built from.
func typeAtomSpans(atoms Atoms) []shared.Span {
	cardTypes := atoms.CardTypes()
	subtypes := atoms.Subtypes()
	spans := make([]shared.Span, 0, len(cardTypes)+len(subtypes))
	for i := range cardTypes {
		spans = append(spans, cardTypes[i].Span)
	}
	for i := range subtypes {
		spans = append(spans, subtypes[i].Span)
	}
	return spans
}

// coveringSpan returns the first span in the union that covers the given span.
func coveringSpan(span shared.Span, union []shared.Span) (shared.Span, bool) {
	for _, candidate := range union {
		if spanCovers(candidate, span) {
			return candidate, true
		}
	}
	return shared.Span{}, false
}

func isListConjunction(token shared.Token) bool {
	return equalWord(token, "or") || equalWord(token, "and")
}

func isSpellNoun(token shared.Token) bool {
	return equalWord(token, "spell") || equalWord(token, "spells")
}

// appendLeadingClauseSpans credits a recognized clause that leads a resolving
// sentence before its effect verb: a "for each X" iteration prefix backed by a
// typed for-each amount, or a closed-form reflexive/delayed trigger preamble
// ("When you do," / "Whenever ... this turn,"). The trailing effect rounds-trips,
// but effectCreditSpan clips backward to the clause boundary before the verb,
// leaving the recognized leading clause uncovered. Each credit is gated on the
// parser actually typing the construct (or, for trigger preambles the parser
// types no node for, on the clause matching a closed surface form), so an
// unrecognized clause interior fails closed and keeps its tokens uncovered.
func appendLeadingClauseSpans(spans []shared.Span, sentences []Sentence) []shared.Span {
	for i := range sentences {
		sentence := &sentences[i]
		tokens := semanticEffectTokens(sentence.Tokens)
		if span, ok := leadingClauseSpan(tokens, sentence); ok {
			spans = append(spans, span)
		}
	}
	return spans
}

// leadingClauseSpan returns the span of a recognized leading clause that precedes
// the sentence's effect verb, bounded by the sentence's first top-level comma. A
// "for each X" prefix is credited only when the sentence carries a typed for-each
// amount; a reflexive/delayed trigger preamble is credited only when its surface
// form is one the parser leaves in-sentence and the sentence has a represented
// effect.
func leadingClauseSpan(tokens []shared.Token, sentence *Sentence) (shared.Span, bool) {
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return shared.Span{}, false
	}
	lead := tokens[:comma]
	if isLeadingForEach(lead) && sentenceForEachClauseExact(sentence, shared.SpanOf(lead)) {
		return shared.SpanOf(lead), true
	}
	if isRecognizedTriggerPreamble(lead) && sentenceHasRepresentedEffect(sentence) {
		return shared.SpanOf(lead), true
	}
	return shared.Span{}, false
}

func isLeadingForEach(lead []shared.Token) bool {
	return len(lead) >= 2 && equalWord(lead[0], "for") && equalWord(lead[1], "each")
}

// sentenceForEachClauseExact reports whether the sentence carries a represented,
// exactly round-tripped effect whose ownership clause absorbs the leading "for
// each X" prefix. The parser folds the iteration prefix into the create effect's
// clause and marks it exact, but effectCreditSpan clips the credit back to the
// verb, leaving the prefix uncovered. Gating on the owning effect's exactness ties
// the credit to the parser's own round-trip claim for these tokens, so a "for
// each" prefix the parser did not fold into an exact clause fails closed.
func sentenceForEachClauseExact(sentence *Sentence, leadSpan shared.Span) bool {
	for i := range sentence.Effects {
		effect := &sentence.Effects[i]
		if effectRepresented(effect) && effectExact(effect) && spanCovers(effect.ClauseSpan, leadSpan) {
			return true
		}
	}
	return false
}

// isRecognizedTriggerPreamble reports whether a leading clause is one of the two
// closed reflexive/delayed trigger preambles the parser leaves in-sentence: a
// reflexive "when you do" backreference, or a delayed "whenever ... this turn"
// clause. The parser types no reflexive/delayed-trigger node, so the preamble is
// matched by its closed surface form and fails closed on any other wording rather
// than crediting an arbitrary "when"/"whenever" clause interior.
func isRecognizedTriggerPreamble(lead []shared.Token) bool {
	return isReflexivePreamble(lead) || isDelayedThisTurnPreamble(lead)
}

func isReflexivePreamble(lead []shared.Token) bool {
	return len(lead) == 3 && equalWord(lead[0], "when") && equalWord(lead[1], "you") && equalWord(lead[2], "do")
}

func isDelayedThisTurnPreamble(lead []shared.Token) bool {
	if len(lead) < 2 || (!equalWord(lead[0], "when") && !equalWord(lead[0], "whenever")) {
		return false
	}
	for i := 0; i+1 < len(lead); i++ {
		if equalWord(lead[i], "this") && equalWord(lead[i+1], "turn") {
			return true
		}
	}
	return false
}

func sentenceHasRepresentedEffect(sentence *Sentence) bool {
	for i := range sentence.Effects {
		if effectRepresented(&sentence.Effects[i]) {
			return true
		}
	}
	return false
}
