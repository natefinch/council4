package oracle

import (
	"slices"
	"strings"
)

// Parse builds a lossless syntax tree for source. It returns a partial tree
// alongside localized diagnostics when the input is malformed.
func Parse(source string, context ParseContext) (Document, []Diagnostic) {
	tokens, diagnostics := lexAll(source)
	lines := splitLines(tokens)
	document := Document{
		Source: source,
		Span: Span{
			Start: Position{Line: 1, Column: 1},
			End:   eofPosition(tokens),
		},
	}

	for i := 0; i < len(lines); {
		if len(lines[i]) == 0 {
			i++
			continue
		}
		ability, abilityDiagnostics := parseAbility(source, lines[i], context)
		diagnostics = append(diagnostics, abilityDiagnostics...)
		if modalStart := modalHeaderStart(lines[i]); modalStart >= 0 {
			modalTokens := lines[i][modalStart:]
			dash := topLevelIndex(modalTokens, EmDash)
			headerTokens := modalTokens
			if dash+1 < len(modalTokens) {
				headerTokens = modalTokens[:dash+1]
			}
			modal := &Modal{Header: phraseFromTokens(source, headerTokens)}
			j := i + 1
			if dash+1 < len(modalTokens) {
				for _, modeTokens := range inlineModeTokens(modalTokens[dash+1:]) {
					mode, modeDiagnostics := parseMode(source, modeTokens)
					modal.Options = append(modal.Options, mode)
					diagnostics = append(diagnostics, modeDiagnostics...)
				}
			} else {
				for j < len(lines) && startsWith(lines[j], Bullet) {
					mode, modeDiagnostics := parseMode(source, lines[j][1:])
					modal.Options = append(modal.Options, mode)
					diagnostics = append(diagnostics, modeDiagnostics...)
					j++
				}
			}
			if len(modal.Options) == 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityError,
					Summary:  "modal ability has no options",
					Detail:   "a choose header must be followed by one or more bullet lines",
					Span:     ability.Span,
				})
			} else {
				ability.Span.End = modal.Options[len(modal.Options)-1].Span.End
				ability.Text = sliceSpan(source, ability.Span)
				ability.Modal = modal
			}
			i = j
		} else {
			i++
		}
		document.Abilities = append(document.Abilities, ability)
	}
	return document, diagnostics
}

func parseAbility(
	source string,
	tokens []Token,
	context ParseContext,
) (Ability, []Diagnostic) {
	ability := Ability{
		Span:   spanOf(tokens),
		Text:   sliceSpan(source, spanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	body := tokens
	if dash, modalStart := topLevelIndex(tokens, EmDash), modalHeaderStart(tokens); dash > 0 && (modalStart < 0 || dash < modalStart) {
		if chapters, ok := parseChapterHeading(tokens[:dash]); context.Saga && ok {
			ability.Chapters = chapters
			ability.ChapterSpan = spanOf(tokens[:dash])
		} else {
			phrase := phraseFromTokens(source, tokens[:dash])
			ability.AbilityWord = &phrase
		}
		body = tokens[dash+1:]
	}
	if colon := topLevelIndex(body, Colon); colon >= 0 {
		phrase := phraseFromTokens(source, body[:colon])
		ability.Cost = &phrase
	}
	if len(ability.Chapters) > 0 {
		ability.Kind = AbilityChapter
	} else {
		ability.Kind = classifyAbility(body, context)
	}
	ability.Sentences = parseSentences(source, body)
	var diagnostics []Diagnostic
	ability.Reminders, ability.Quoted, diagnostics = parseDelimited(source, body, diagnostics)
	return ability, diagnostics
}

func parseChapterHeading(tokens []Token) ([]int, bool) {
	parts := splitTopLevel(tokens, Comma)
	chapters := make([]int, 0, len(parts))
	for _, part := range parts {
		if len(part) != 1 || part[0].Kind != Word {
			return nil, false
		}
		chapter, ok := romanChapter(part[0].Text)
		if !ok {
			return nil, false
		}
		chapters = append(chapters, chapter)
	}
	return chapters, len(chapters) > 0
}

func romanChapter(text string) (int, bool) {
	switch strings.ToUpper(text) {
	case "I":
		return 1, true
	case "II":
		return 2, true
	case "III":
		return 3, true
	case "IV":
		return 4, true
	case "V":
		return 5, true
	case "VI":
		return 6, true
	default:
		return 0, false
	}
}

func parseMode(source string, tokens []Token) (Mode, []Diagnostic) {
	mode := Mode{
		Span:   spanOf(tokens),
		Text:   sliceSpan(source, spanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	mode.Sentences = parseSentences(source, tokens)
	var diagnostics []Diagnostic
	mode.Reminders, mode.Quoted, diagnostics = parseDelimited(source, tokens, diagnostics)
	return mode, diagnostics
}

func inlineModeTokens(tokens []Token) [][]Token {
	parts := splitTopLevelTokens(tokens, Semicolon)
	if len(parts) < 2 {
		return nil
	}
	for i := 1; i < len(parts); i++ {
		if startsWithWord(parts[i], "or") {
			parts[i] = parts[i][1:]
		}
	}
	return parts
}

func splitTopLevelTokens(tokens []Token, separator Kind) [][]Token {
	var parts [][]Token
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if !quoted {
				depth++
			}
		case RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case Quote:
			quoted = !quoted
		default:
			if token.Kind == separator && depth == 0 && !quoted {
				parts = append(parts, cloneTokens(tokens[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, cloneTokens(tokens[start:]))
}

func classifyAbility(tokens []Token, context ParseContext) AbilityKind {
	if len(tokens) == 0 {
		return AbilityUnknown
	}
	if tokens[0].Kind == LeftParen && matchingOuter(tokens, LeftParen, RightParen) {
		return AbilityReminder
	}
	if colon := topLevelIndex(tokens, Colon); colon >= 0 {
		if context.Planeswalker && loyaltyCost(tokens[:colon]) {
			return AbilityLoyalty
		}
		return AbilityActivated
	}
	if startsWithWord(tokens, "when", "whenever", "at") {
		return AbilityTriggered
	}
	if replacementWording(tokens) {
		return AbilityReplacement
	}
	if joinedSourceText(tokens) == "This spell can't be countered." {
		return AbilityStatic
	}
	if context.InstantOrSorcery {
		return AbilitySpell
	}
	return AbilityStatic
}

func loyaltyCost(tokens []Token) bool {
	if len(tokens) == 1 && loyaltyValue(tokens[0]) {
		return true
	}
	return len(tokens) == 2 &&
		(tokens[0].Kind == Plus || tokens[0].Kind == Minus) &&
		loyaltyValue(tokens[1])
}

func loyaltyValue(token Token) bool {
	return token.Kind == Integer || (token.Kind == Word && strings.EqualFold(token.Text, "x"))
}

func replacementWording(tokens []Token) bool {
	words := normalizedWords(tokens)
	if len(words) >= 2 && words[0] == "as" && containsWord(words, "enters") {
		return true
	}
	if containsWord(words, "enters") &&
		(containsWord(words, "tapped") || containsWord(words, "with") || containsWord(words, "as")) {
		return true
	}
	return containsWord(words, "would") && containsWord(words, "instead")
}

func parseSentences(source string, tokens []Token) []Sentence {
	var sentences []Sentence
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if !quoted {
				depth++
			}
		case RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case Quote:
			quoted = !quoted
		case Period:
			if depth == 0 && !quoted {
				sentences = appendSentence(sentences, source, tokens[start:i+1])
				start = i + 1
			}
		default:
		}
	}
	return appendSentence(sentences, source, tokens[start:])
}

func appendSentence(sentences []Sentence, source string, tokens []Token) []Sentence {
	if len(tokens) == 0 {
		return sentences
	}
	span := spanOf(tokens)
	return append(sentences, Sentence{
		Span:   span,
		Text:   sliceSpan(source, span),
		Tokens: cloneTokens(tokens),
	})
}

func parseDelimited(
	source string,
	tokens []Token,
	diagnostics []Diagnostic,
) (reminders, quoted []Delimited, updatedDiagnostics []Diagnostic) {
	updatedDiagnostics = diagnostics
	var parenStack []int
	quoteStart := -1
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if quoteStart < 0 {
				parenStack = append(parenStack, i)
			}
		case RightParen:
			if quoteStart >= 0 {
				continue
			}
			if len(parenStack) == 0 {
				updatedDiagnostics = append(updatedDiagnostics, unmatchedDiagnostic(token, "parenthesis"))
				continue
			}
			start := parenStack[len(parenStack)-1]
			parenStack = parenStack[:len(parenStack)-1]
			if len(parenStack) == 0 {
				reminders = append(reminders, delimitedFromTokens(source, tokens[start:i+1]))
			}
		case Quote:
			if len(parenStack) > 0 {
				continue
			}
			if quoteStart < 0 {
				quoteStart = i
			} else {
				quoted = append(quoted, delimitedFromTokens(source, tokens[quoteStart:i+1]))
				quoteStart = -1
			}
		default:
		}
	}
	for _, start := range parenStack {
		updatedDiagnostics = append(updatedDiagnostics, Diagnostic{
			Severity: SeverityError,
			Summary:  "unclosed parenthesis",
			Detail:   "the parenthesized text is not closed before the paragraph ends",
			Span:     tokens[start].Span,
		})
	}
	if quoteStart >= 0 {
		updatedDiagnostics = append(updatedDiagnostics, Diagnostic{
			Severity: SeverityError,
			Summary:  "unclosed quote",
			Detail:   "the quoted text is not closed before the paragraph ends",
			Span:     tokens[quoteStart].Span,
		})
	}
	return reminders, quoted, updatedDiagnostics
}

func lexAll(source string) ([]Token, []Diagnostic) {
	lexer := NewLexer(source)
	var tokens []Token
	var diagnostics []Diagnostic
	for {
		token := lexer.Next()
		if token.Kind == EOF {
			tokens = append(tokens, token)
			return tokens, diagnostics
		}
		if token.Kind == Invalid {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityError,
				Summary:  "invalid Oracle text",
				Detail:   "the input contains malformed encoding or an unclosed symbol",
				Span:     token.Span,
			})
		}
		tokens = append(tokens, token)
	}
}

func splitLines(tokens []Token) [][]Token {
	var lines [][]Token
	start := 0
	protected := protectedByMultilineOuterDelimiter(tokens)
	for i, token := range tokens {
		if token.Kind == Newline && protected[i] {
			continue
		}
		if token.Kind == Newline || token.Kind == EOF {
			lines = append(lines, cloneTokens(tokens[start:i]))
			start = i + 1
		}
	}
	return lines
}

func protectedByMultilineOuterDelimiter(tokens []Token) []bool {
	difference := make([]int, len(tokens)+1)
	addPair := func(start, end int) {
		difference[start+1]++
		difference[end]--
	}
	for start := 0; start < len(tokens); {
		end := start
		for end < len(tokens) && tokens[end].Kind != Newline && tokens[end].Kind != EOF {
			end++
		}
		if start < end {
			switch tokens[start].Kind {
			case LeftParen:
				if end := matchingDelimiter(tokens, start, LeftParen, RightParen); end >= 0 {
					addPair(start, end)
				}
			case Quote:
				if end := matchingDelimiter(tokens, start, Quote, Quote); end >= 0 {
					addPair(start, end)
				}
			default:
			}
		}
		start = end + 1
	}
	protected := make([]bool, len(tokens))
	depth := 0
	for i := range tokens {
		depth += difference[i]
		protected[i] = depth > 0
	}
	return protected
}

func matchingDelimiter(tokens []Token, start int, open, closeKind Kind) int {
	depth := 0
	for i := start; i < len(tokens); i++ {
		switch {
		case open == closeKind && tokens[i].Kind == open:
			if depth != 0 {
				return i
			}
			depth = 1
		case open != closeKind && tokens[i].Kind == open:
			depth++
		case open != closeKind && tokens[i].Kind == closeKind:
			depth--
			if depth == 0 {
				return i
			}
		default:
		}
	}
	return -1
}

func isModalHeader(tokens []Token) bool {
	if !startsWithWord(tokens, "choose") {
		return false
	}

	dash := topLevelIndex(tokens, EmDash)
	if dash < 0 {
		return false
	}
	period := topLevelIndex(tokens, Period)
	return period < 0 || dash < period
}

func modalHeaderStart(tokens []Token) int {
	if isModalHeader(tokens) {
		return 0
	}
	colon := topLevelIndex(tokens, Colon)
	if colon >= 0 && colon+1 < len(tokens) && isModalHeader(tokens[colon+1:]) {
		return colon + 1
	}
	return -1
}

func topLevelIndex(tokens []Token, wanted Kind) int {
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case LeftParen:
			if !quoted {
				depth++
			}
		case RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case Quote:
			quoted = !quoted
		default:
			if token.Kind == wanted && depth == 0 && !quoted {
				return i
			}
		}
	}
	return -1
}

func matchingOuter(tokens []Token, open, closeKind Kind) bool {
	depth := 0
	for i, token := range tokens {
		switch token.Kind {
		case open:
			depth++
		case closeKind:
			depth--
			if depth == 0 {
				return i == len(tokens)-1
			}
		default:
		}
	}
	return false
}

func startsWith(tokens []Token, kind Kind) bool {
	return len(tokens) > 0 && tokens[0].Kind == kind
}

func startsWithWord(tokens []Token, words ...string) bool {
	if len(tokens) == 0 || tokens[0].Kind != Word {
		return false
	}
	for _, word := range words {
		if strings.EqualFold(tokens[0].Text, word) {
			return true
		}
	}
	return false
}

func normalizedWords(tokens []Token) []string {
	words := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == Word {
			words = append(words, strings.ToLower(token.Text))
		}
	}
	return words
}

func containsWord(words []string, wanted string) bool {
	return slices.Contains(words, wanted)
}

func phraseFromTokens(source string, tokens []Token) Phrase {
	if len(tokens) == 0 {
		return Phrase{}
	}
	span := spanOf(tokens)
	return Phrase{Span: span, Text: sliceSpan(source, span), Tokens: cloneTokens(tokens)}
}

func delimitedFromTokens(source string, tokens []Token) Delimited {
	span := spanOf(tokens)
	return Delimited{Span: span, Text: sliceSpan(source, span), Tokens: cloneTokens(tokens)}
}

func spanOf(tokens []Token) Span {
	if len(tokens) == 0 {
		return Span{}
	}
	return Span{Start: tokens[0].Span.Start, End: tokens[len(tokens)-1].Span.End}
}

func sliceSpan(source string, span Span) string {
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		return ""
	}
	return source[span.Start.Offset:span.End.Offset]
}

func cloneTokens(tokens []Token) []Token {
	return append([]Token(nil), tokens...)
}

func eofPosition(tokens []Token) Position {
	if len(tokens) == 0 {
		return Position{Line: 1, Column: 1}
	}
	return tokens[len(tokens)-1].Span.End
}

func unmatchedDiagnostic(token Token, delimiter string) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Summary:  "unmatched " + delimiter,
		Detail:   "the closing delimiter has no matching opener in this paragraph",
		Span:     token.Span,
	}
}
