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
// sentence before its effect verb: a "for each X" iteration prefix or a
// reflexive/delayed trigger preamble ("When you do," / "Whenever ... this turn,").
// The trailing effect rounds-trips, but effectCreditSpan clips backward to the
// clause boundary before the verb, leaving the recognized leading clause
// uncovered. Each leading clause is credited up to its terminating comma.
func appendLeadingClauseSpans(spans []shared.Span, sentences []Sentence) []shared.Span {
	for i := range sentences {
		sentence := &sentences[i]
		if !sentenceHasRepresentedEffect(sentence) {
			continue
		}
		tokens := semanticEffectTokens(sentence.Tokens)
		if span, ok := leadingClauseSpan(tokens); ok {
			spans = append(spans, span)
		}
	}
	return spans
}

// leadingClauseSpan returns the span of a recognized leading clause — a "for each"
// iteration prefix or a "when"/"whenever" trigger preamble — that precedes the
// sentence's effect verb, bounded by the sentence's first top-level comma.
func leadingClauseSpan(tokens []shared.Token) (shared.Span, bool) {
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return shared.Span{}, false
	}
	lead := tokens[:comma]
	if !isLeadingForEach(lead) && !isLeadingTriggerPreamble(lead) {
		return shared.Span{}, false
	}
	return shared.SpanOf(lead), true
}

func isLeadingForEach(lead []shared.Token) bool {
	return len(lead) >= 2 && equalWord(lead[0], "for") && equalWord(lead[1], "each")
}

func isLeadingTriggerPreamble(lead []shared.Token) bool {
	return len(lead) >= 1 && (equalWord(lead[0], "when") || equalWord(lead[0], "whenever"))
}

func sentenceHasRepresentedEffect(sentence *Sentence) bool {
	for i := range sentence.Effects {
		if effectRepresented(&sentence.Effects[i]) {
			return true
		}
	}
	return false
}
