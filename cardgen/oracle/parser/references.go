package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// ReferenceKind identifies the exact explicit-reference wording recognized in
// Oracle text before any semantic antecedent binding. The parser owns this
// vocabulary, including the card's own name supplied through Context.CardName;
// downstream stages bind referents from the typed kind and span without
// reinspecting Oracle spelling.
type ReferenceKind uint8

// Explicit reference kinds recognized by the parser.
const (
	ReferenceUnknown ReferenceKind = iota
	ReferenceSelfName
	ReferenceThisObject
	ReferenceThatObject
	ReferenceThatPlayer
	ReferencePronoun
)

// PronounKind identifies the exact grammatical pronoun carried by a reference.
type PronounKind uint8

// Explicit pronouns recognized by the parser.
const (
	PronounUnknown PronounKind = iota
	PronounIt
	PronounIts
	PronounThey
	PronounTheir
	PronounThem
	PronounThose
)

// Reference is a source-spanned explicit reference atom. Tokens retains the
// matched source slice so downstream stages can render exact text without
// re-recognizing meaning.
type Reference struct {
	Kind    ReferenceKind
	Pronoun PronounKind
	Span    shared.Span
	Tokens  []shared.Token
}

// collectReferences recognizes explicit self-name, this-object, that-object,
// and pronoun references in a token slice. The card's own name arrives as
// cardName so the parser, not the compiler, owns recognition of the source
// name's spelling. Source-tied duration subjects ("for as long as you control
// [CardName]"/"this [type]") are intentionally not reported as references.
func collectReferences(tokens []shared.Token, cardName string) []Reference {
	var references []Reference
	for _, nameWords := range selfNameReferenceAliases(cardName) {
		for i := 0; i+len(nameWords) <= len(tokens); i++ {
			if i >= 6 {
				pre := shared.NormalizedWords(tokens[i-6 : i])
				if referenceContainsSequence(pre, "for", "as", "long", "as", "you", "control") {
					i += len(nameWords) - 1
					continue
				}
			}
			if referencePossessiveNameAt(tokens, i, nameWords) {
				phrase := tokens[i : i+len(nameWords)]
				if referenceSpanOverlaps(references, shared.SpanOf(phrase)) {
					continue
				}
				references = append(references, Reference{
					Kind:   ReferenceSelfName,
					Span:   shared.SpanOf(phrase),
					Tokens: phrase,
				})
				i += len(nameWords) - 1
				continue
			}
			if referenceTokenWordsEqual(tokens[i:i+len(nameWords)], nameWords) {
				phrase := tokens[i : i+len(nameWords)]
				if referenceSpanOverlaps(references, shared.SpanOf(phrase)) {
					continue
				}
				references = append(references, Reference{
					Kind:   ReferenceSelfName,
					Span:   shared.SpanOf(phrase),
					Tokens: phrase,
				})
				i += len(nameWords) - 1
			}
		}
	}
	for i := 0; i < len(tokens); i++ {
		switch {
		case i+1 < len(tokens) &&
			equalWord(tokens[i], "this") &&
			strings.EqualFold(tokens[i+1].Text, "creature's"):
			phrase := tokens[i : i+2]
			references = append(references, Reference{
				Kind:   ReferenceThisObject,
				Span:   shared.SpanOf(phrase),
				Tokens: phrase,
			})
			i++
		case i+1 < len(tokens) && equalWord(tokens[i], "this") && referenceSelfMarkerNoun(tokens[i+1]):
			if i >= 6 {
				pre := shared.NormalizedWords(tokens[i-6 : i])
				if referenceContainsSequence(pre, "for", "as", "long", "as", "you", "control") {
					i++
					break
				}
			}
			phrase := tokens[i : i+2]
			references = append(references, Reference{
				Kind:   ReferenceThisObject,
				Span:   shared.SpanOf(phrase),
				Tokens: phrase,
			})
			i++
		case i+1 < len(tokens) && equalWord(tokens[i], "that") && referenceObjectNoun(tokens[i+1]):
			phrase := tokens[i : i+2]
			kind := ReferenceThatObject
			if equalWord(tokens[i+1], "player") {
				kind = ReferenceThatPlayer
			}
			references = append(references, Reference{
				Kind:   kind,
				Span:   shared.SpanOf(phrase),
				Tokens: phrase,
			})
			i++
		case pronounKind(tokens[i]) != PronounUnknown:
			references = append(references, Reference{
				Kind:    ReferencePronoun,
				Pronoun: pronounKind(tokens[i]),
				Span:    tokens[i].Span,
				Tokens:  tokens[i : i+1],
			})
		default:
		}
	}
	return references
}

func pronounKind(token shared.Token) PronounKind {
	switch {
	case equalWord(token, "it"):
		return PronounIt
	case equalWord(token, "its"):
		return PronounIts
	case equalWord(token, "they"):
		return PronounThey
	case equalWord(token, "their"):
		return PronounTheir
	case equalWord(token, "them"):
		return PronounThem
	case equalWord(token, "those"):
		return PronounThose
	default:
		return PronounUnknown
	}
}

// referenceObjectNoun reports whether token is one of the object nouns that can
// follow "this"/"that" in an explicit object reference.
func referenceSelfMarkerNoun(token shared.Token) bool {
	return referenceObjectNoun(token)
}

func referenceObjectNoun(token shared.Token) bool {
	noun, ok := recognizeObjectNoun(token)
	if !ok {
		return false
	}
	switch noun {
	case ObjectNounArtifact, ObjectNounCard, ObjectNounCreature, ObjectNounEnchantment,
		ObjectNounEquipment, ObjectNounLand, ObjectNounPermanent, ObjectNounPlayer, ObjectNounSpell, ObjectNounToken:
		return true
	default:
		return false
	}
}

func referencePossessiveNameAt(tokens []shared.Token, start int, nameWords []string) bool {
	if len(nameWords) == 0 || start < 0 || start+len(nameWords) > len(tokens) {
		return false
	}
	last := len(nameWords) - 1
	for i := range last {
		if !equalWord(tokens[start+i], nameWords[i]) {
			return false
		}
	}
	return strings.EqualFold(tokens[start+last].Text, nameWords[last]+"'s")
}

func referenceTokenWordsEqual(tokens []shared.Token, words []string) bool {
	if len(tokens) != len(words) {
		return false
	}
	for i := range words {
		normalized := strings.ToLower(strings.Trim(tokens[i].Text, ",.'\u2019"))
		if tokens[i].Kind != shared.Word || normalized != words[i] {
			return false
		}
	}
	return true
}

func referenceContainsSequence(words []string, expected ...string) bool {
	for i := 0; i+len(expected) <= len(words); i++ {
		if referenceStartsWords(words[i:], expected...) {
			return true
		}
	}
	return false
}

// collectSelfNameSpans returns the span of every occurrence of the card's own
// name, including possessive forms, using the same case-insensitive word
// matching the compiler historically applied. Unlike collectReferences it does
// not skip duration-context occurrences: callers that filter card-name tokens
// out of effect, duration, and amount grammar need every occurrence.
func collectSelfNameSpans(tokens []shared.Token, cardName string) []shared.Span {
	var spans []shared.Span
	for _, nameWords := range selfNameSpanAliases(cardName) {
		for start := 0; start+len(nameWords) <= len(tokens); start++ {
			phrase := tokens[start : start+len(nameWords)]
			if referenceWordsAt(phrase, nameWords) || referencePossessiveNameAt(tokens, start, nameWords) {
				span := shared.SpanOf(phrase)
				if !selfSpanOverlaps(spans, span) {
					spans = append(spans, span)
				}
			}
		}
	}
	return spans
}

func selfNameSpanAliases(cardName string) [][]string {
	cardName = strings.TrimSpace(cardName)
	if cardName == "" {
		return nil
	}
	var aliases [][]string
	appendAlias := func(name string) {
		words := strings.Fields(strings.ToLower(strings.TrimSpace(name)))
		if len(words) == 0 {
			return
		}
		for _, alias := range aliases {
			if strings.Join(alias, " ") == strings.Join(words, " ") {
				return
			}
		}
		aliases = append(aliases, words)
	}
	appendAlias(cardName)
	if shortName, _, ok := strings.Cut(cardName, ","); ok {
		appendAlias(shortName)
	}
	if frontName, _, ok := strings.Cut(cardName, " // "); ok {
		appendAlias(frontName)
	}
	return aliases
}

func collectSourceNameSpans(tokens []shared.Token, cardName string) []shared.Span {
	var spans []shared.Span
	for _, nameWords := range selfNameSubjectAliases(cardName) {
		for start := 0; start+len(nameWords) <= len(tokens); start++ {
			phrase := tokens[start : start+len(nameWords)]
			if referenceWordsAt(phrase, nameWords) || referencePossessiveNameAt(tokens, start, nameWords) {
				span := shared.SpanOf(phrase)
				if !selfSpanOverlaps(spans, span) {
					spans = append(spans, span)
				}
			}
		}
	}
	return spans
}

func collectSourceMarkerSpans(tokens []shared.Token) []shared.Span {
	var spans []shared.Span
	for i := 0; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "this") || !sourceSubjectMarkerNoun(tokens[i+1]) {
			continue
		}
		spans = append(spans, shared.SpanOf(tokens[i:i+2]))
		i++
	}
	return spans
}

func sourceSubjectMarkerNoun(token shared.Token) bool {
	if referenceObjectNoun(token) {
		return true
	}
	for _, word := range []string{"aura", "battle", "vehicle", "siege", "case", "class", "planeswalker", "spacecraft"} {
		if equalWord(token, word) {
			return true
		}
	}
	return false
}

func referenceSpanOverlaps(references []Reference, span shared.Span) bool {
	for _, reference := range references {
		if spansOverlap(reference.Span, span) {
			return true
		}
	}
	return false
}

func selfSpanOverlaps(spans []shared.Span, span shared.Span) bool {
	for _, existing := range spans {
		if spansOverlap(existing, span) {
			return true
		}
	}
	return false
}

func spansOverlap(left, right shared.Span) bool {
	return left.Start.Offset < right.End.Offset && right.Start.Offset < left.End.Offset
}

func selfNameReferenceAliases(cardName string) [][]string {
	cardName = strings.TrimSpace(cardName)
	if cardName == "" {
		return nil
	}
	words := referenceNameWords(cardName)
	if len(words) == 0 {
		return nil
	}
	return [][]string{words}
}

func selfNameSubjectAliases(cardName string) [][]string {
	cardName = strings.TrimSpace(cardName)
	if cardName == "" {
		return nil
	}
	var aliases [][]string
	appendAlias := func(name string) {
		words := referenceNameWords(name)
		if len(words) == 0 {
			return
		}
		for _, alias := range aliases {
			if strings.Join(alias, " ") == strings.Join(words, " ") {
				return
			}
		}
		aliases = append(aliases, words)
	}
	appendAlias(cardName)
	if shortName, _, ok := strings.Cut(cardName, ","); ok {
		appendAlias(shortName)
	}
	if frontName, _, ok := strings.Cut(cardName, " // "); ok {
		appendAlias(frontName)
	}
	firstWord, _, hasMore := strings.Cut(cardName, " ")
	if hasMore && !referenceNameArticle(firstWord) {
		appendAlias(firstWord)
	}
	return aliases
}

func referenceNameWords(name string) []string {
	tokens, _ := lexAll(strings.TrimSpace(name))
	words := make([]string, 0, len(tokens))
	for _, token := range tokens {
		switch token.Kind {
		case shared.Word, shared.Integer, shared.Ampersand, shared.Slash, shared.Period, shared.Comma:
			words = append(words, strings.ToLower(token.Text))
		default:
		}
	}
	return words
}

func referenceNameArticle(word string) bool {
	switch strings.ToLower(word) {
	case "a", "an", "the":
		return true
	default:
		return false
	}
}

func referenceWordsAt(tokens []shared.Token, words []string) bool {
	if len(tokens) != len(words) {
		return false
	}
	for i, word := range words {
		if word == "&" || word == "/" || word == "." || word == "," {
			if !strings.EqualFold(tokens[i].Text, word) {
				return false
			}
			continue
		}
		if !equalWord(tokens[i], word) && !strings.EqualFold(tokens[i].Text, word) {
			return false
		}
	}
	return true
}

func referenceStartsWords(words []string, expected ...string) bool {
	if len(words) < len(expected) {
		return false
	}
	for i := range expected {
		if words[i] != expected[i] {
			return false
		}
	}
	return true
}
