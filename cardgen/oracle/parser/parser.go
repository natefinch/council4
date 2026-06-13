package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/lexer"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// Parse builds a lossless syntax tree for source. It returns a partial tree
// alongside localized diagnostics when the input is malformed.
func Parse(source string, context Context) (Document, []shared.Diagnostic) {
	tokens, diagnostics := lexAll(source)
	lines := splitLines(tokens)
	document := Document{
		Source:   source,
		CardName: context.CardName,
		Span: shared.Span{
			Start: shared.Position{Line: 1, Column: 1},
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
			dash := shared.TopLevelIndex(modalTokens, shared.EmDash)
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
				for j < len(lines) && startsWith(lines[j], shared.Bullet) {
					mode, modeDiagnostics := parseMode(source, lines[j][1:])
					modal.Options = append(modal.Options, mode)
					diagnostics = append(diagnostics, modeDiagnostics...)
					j++
				}
			}
			if len(modal.Options) == 0 {
				diagnostics = append(diagnostics, shared.Diagnostic{
					Severity: shared.SeverityError,
					Summary:  "modal ability has no options",
					Detail:   "a choose header must be followed by one or more bullet lines",
					Span:     ability.Span,
				})
			} else {
				ability.Span.End = modal.Options[len(modal.Options)-1].Span.End
				ability.Text = shared.SliceSpan(source, ability.Span)
				ability.Modal = modal
			}
			i = j
		} else {
			i++
		}
		document.Abilities = append(document.Abilities, ability)
	}
	emitAtoms(document.Abilities, context.CardName)
	emitTriggerEventClauses(document.Abilities, context.CardName)
	emitEventHistoryConditions(document.Abilities)
	emitResolvingSyntax(document.Abilities)
	return document, diagnostics
}

// emitAtoms fills each ability's and modal option's typed atom collection from
// its semantic tokens.
func emitAtoms(abilities []Ability, cardName string) {
	for i := range abilities {
		tokens := abilities[i].Tokens
		if abilities[i].AbilityWord != nil {
			tokens = tokensOutsideParserSpan(tokens, abilities[i].AbilityWord.Span)
		}
		abilities[i].Atoms = collectAtoms(tokens, abilities[i].Reminders, abilities[i].Quoted, cardName)
		if abilities[i].Modal == nil {
			continue
		}
		abilities[i].Modal.Atoms = collectAtoms(abilities[i].Modal.Header.Tokens, nil, nil, cardName)
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			mode.Atoms = collectAtoms(mode.Tokens, mode.Reminders, mode.Quoted, cardName)
		}
	}
}

func parseAbility(
	source string,
	tokens []shared.Token,
	context Context,
) (Ability, []shared.Diagnostic) {
	ability := Ability{
		Span:   shared.SpanOf(tokens),
		Text:   shared.SliceSpan(source, shared.SpanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	body := tokens
	if dash, modalStart := shared.TopLevelIndex(tokens, shared.EmDash), modalHeaderStart(tokens); dash > 0 && (modalStart < 0 || dash < modalStart) {
		if chapters, ok := parseChapterHeading(tokens[:dash]); context.Saga && ok {
			ability.Chapters = chapters
			ability.ChapterSpan = shared.SpanOf(tokens[:dash])
		} else {
			phrase := phraseFromTokens(source, tokens[:dash])
			ability.AbilityWord = &phrase
		}
		body = tokens[dash+1:]
	}
	if colon := shared.TopLevelIndex(body, shared.Colon); colon >= 0 {
		phrase := phraseFromTokens(source, body[:colon])
		ability.Cost = &phrase
	}
	if len(ability.Chapters) > 0 {
		ability.Kind = AbilityChapter
	} else {
		ability.Kind = classifyAbility(body, context)
	}
	if ability.Kind == AbilityTriggered {
		ability.Trigger = parseTriggerClause(source, body)
	}
	ability.Sentences = ParseSentences(source, resolvingBodyTokens(body, ability.Kind))
	var diagnostics []shared.Diagnostic
	ability.Reminders, ability.Quoted, diagnostics = parseDelimited(source, body, diagnostics)
	if ability.Kind == AbilityActivated {
		ability.ActivationRestrictions = parseTrailingActivationRestrictions(
			source,
			body,
			ability.Reminders,
			ability.Quoted,
		)
	}

	return ability, diagnostics
}

func resolvingBodyTokens(tokens []shared.Token, kind AbilityKind) []shared.Token {
	switch kind {
	case AbilityActivated, AbilityLoyalty:
		if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case AbilityTriggered:
		if comma := triggerBodyComma(tokens); comma >= 0 {
			return tokens[comma+1:]
		}
	default:
	}
	return tokens
}

func parseChapterHeading(tokens []shared.Token) ([]int, bool) {
	parts := splitTopLevelTokens(tokens, shared.Comma)
	chapters := make([]int, 0, len(parts))
	for _, part := range parts {
		if len(part) != 1 || part[0].Kind != shared.Word {
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

func parseMode(source string, tokens []shared.Token) (Mode, []shared.Diagnostic) {
	mode := Mode{
		Span:   shared.SpanOf(tokens),
		Text:   shared.SliceSpan(source, shared.SpanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	mode.Sentences = ParseSentences(source, tokens)
	var diagnostics []shared.Diagnostic
	mode.Reminders, mode.Quoted, diagnostics = parseDelimited(source, tokens, diagnostics)
	return mode, diagnostics
}

func inlineModeTokens(tokens []shared.Token) [][]shared.Token {
	parts := splitTopLevelTokens(tokens, shared.Semicolon)
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

func splitTopLevelTokens(tokens []shared.Token, separator shared.Kind) [][]shared.Token {
	var parts [][]shared.Token
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if !quoted {
				depth++
			}
		case shared.RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case shared.Quote:
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

func classifyAbility(tokens []shared.Token, context Context) AbilityKind {
	if len(tokens) == 0 {
		return AbilityUnknown
	}
	if tokens[0].Kind == shared.LeftParen && matchingOuter(tokens, shared.LeftParen, shared.RightParen) {
		return AbilityReminder
	}
	if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
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
	if _, ok := parseStaticRuleSyntax(tokens); ok {
		return AbilityStatic
	}
	if context.InstantOrSorcery {
		return AbilitySpell
	}
	return AbilityStatic
}

func loyaltyCost(tokens []shared.Token) bool {
	if len(tokens) == 1 && loyaltyValue(tokens[0]) {
		return true
	}
	return len(tokens) == 2 &&
		(tokens[0].Kind == shared.Plus || tokens[0].Kind == shared.Minus) &&
		loyaltyValue(tokens[1])
}

func loyaltyValue(token shared.Token) bool {
	return token.Kind == shared.Integer || (token.Kind == shared.Word && strings.EqualFold(token.Text, "x"))
}

func replacementWording(tokens []shared.Token) bool {
	words := shared.NormalizedWords(tokens)
	if len(words) >= 2 && words[0] == "as" && shared.ContainsWord(words, "enters") {
		return true
	}
	if shared.ContainsWord(words, "enters") &&
		(shared.ContainsWord(words, "tapped") || shared.ContainsWord(words, "with") || shared.ContainsWord(words, "as")) {
		return true
	}
	return shared.ContainsWord(words, "would") && shared.ContainsWord(words, "instead")
}

// ParseSentences parses top-level sentences from tokens. It remains available
// for transitional compiler paths that have not yet moved to typed syntax.
func ParseSentences(source string, tokens []shared.Token) []Sentence {
	var sentences []Sentence
	start := 0
	depth := 0
	quoted := false
	for i, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if !quoted {
				depth++
			}
		case shared.RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case shared.Quote:
			quoted = !quoted
		case shared.Period:
			if depth == 0 && !quoted {
				sentences = appendSentence(sentences, source, tokens[start:i+1])
				start = i + 1
			}
		default:
		}
	}
	return appendSentence(sentences, source, tokens[start:])
}

func appendSentence(sentences []Sentence, source string, tokens []shared.Token) []Sentence {
	if len(tokens) == 0 {
		return sentences
	}
	span := shared.SpanOf(tokens)
	sentence := Sentence{
		Span:   span,
		Text:   shared.SliceSpan(source, span),
		Tokens: cloneTokens(tokens),
	}
	if rule, ok := parseStaticRuleSyntax(tokens); ok {
		sentence.StaticRule = rule
	}
	return append(sentences, sentence)
}

func parseDelimited(
	source string,
	tokens []shared.Token,
	diagnostics []shared.Diagnostic,
) (reminders, quoted []Delimited, updatedDiagnostics []shared.Diagnostic) {
	updatedDiagnostics = diagnostics
	var parenStack []int
	quoteStart := -1
	for i, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if quoteStart < 0 {
				parenStack = append(parenStack, i)
			}
		case shared.RightParen:
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
		case shared.Quote:
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
		updatedDiagnostics = append(updatedDiagnostics, shared.Diagnostic{
			Severity: shared.SeverityError,
			Summary:  "unclosed parenthesis",
			Detail:   "the parenthesized text is not closed before the paragraph ends",
			Span:     tokens[start].Span,
		})
	}
	if quoteStart >= 0 {
		updatedDiagnostics = append(updatedDiagnostics, shared.Diagnostic{
			Severity: shared.SeverityError,
			Summary:  "unclosed quote",
			Detail:   "the quoted text is not closed before the paragraph ends",
			Span:     tokens[quoteStart].Span,
		})
	}
	return reminders, quoted, updatedDiagnostics
}

func lexAll(source string) ([]shared.Token, []shared.Diagnostic) {
	scanner := lexer.NewLexer(source)
	var tokens []shared.Token
	var diagnostics []shared.Diagnostic
	for {
		token := scanner.Next()
		if token.Kind == shared.EOF {
			tokens = append(tokens, token)
			return tokens, diagnostics
		}
		if token.Kind == shared.Invalid {
			diagnostics = append(diagnostics, shared.Diagnostic{
				Severity: shared.SeverityError,
				Summary:  "invalid Oracle text",
				Detail:   "the input contains malformed encoding or an unclosed symbol",
				Span:     token.Span,
			})
		}
		tokens = append(tokens, token)
	}
}

func splitLines(tokens []shared.Token) [][]shared.Token {
	var lines [][]shared.Token
	start := 0
	protected := protectedByMultilineOuterDelimiter(tokens)
	for i, token := range tokens {
		if token.Kind == shared.Newline && protected[i] {
			continue
		}
		if token.Kind == shared.Newline || token.Kind == shared.EOF {
			lines = append(lines, cloneTokens(tokens[start:i]))
			start = i + 1
		}
	}
	return lines
}

func protectedByMultilineOuterDelimiter(tokens []shared.Token) []bool {
	difference := make([]int, len(tokens)+1)
	addPair := func(start, end int) {
		difference[start+1]++
		difference[end]--
	}
	for start := 0; start < len(tokens); {
		end := start
		for end < len(tokens) && tokens[end].Kind != shared.Newline && tokens[end].Kind != shared.EOF {
			end++
		}
		if start < end {
			switch tokens[start].Kind {
			case shared.LeftParen:
				if end := matchingDelimiter(tokens, start, shared.LeftParen, shared.RightParen); end >= 0 {
					addPair(start, end)
				}
			case shared.Quote:
				if end := matchingDelimiter(tokens, start, shared.Quote, shared.Quote); end >= 0 {
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

func matchingDelimiter(tokens []shared.Token, start int, open, closeKind shared.Kind) int {
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

func isModalHeader(tokens []shared.Token) bool {
	if !startsWithWord(tokens, "choose") {
		return false
	}

	dash := shared.TopLevelIndex(tokens, shared.EmDash)
	if dash < 0 {
		return false
	}
	period := shared.TopLevelIndex(tokens, shared.Period)
	return period < 0 || dash < period
}

func modalHeaderStart(tokens []shared.Token) int {
	if isModalHeader(tokens) {
		return 0
	}
	colon := shared.TopLevelIndex(tokens, shared.Colon)
	if colon >= 0 && colon+1 < len(tokens) && isModalHeader(tokens[colon+1:]) {
		return colon + 1
	}
	return -1
}

func matchingOuter(tokens []shared.Token, open, closeKind shared.Kind) bool {
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

func startsWith(tokens []shared.Token, kind shared.Kind) bool {
	return len(tokens) > 0 && tokens[0].Kind == kind
}

func startsWithWord(tokens []shared.Token, words ...string) bool {
	if len(tokens) == 0 || tokens[0].Kind != shared.Word {
		return false
	}
	for _, word := range words {
		if strings.EqualFold(tokens[0].Text, word) {
			return true
		}
	}
	return false
}

func phraseFromTokens(source string, tokens []shared.Token) Phrase {
	if len(tokens) == 0 {
		return Phrase{}
	}
	span := shared.SpanOf(tokens)
	return Phrase{Span: span, Text: shared.SliceSpan(source, span), Tokens: cloneTokens(tokens)}
}

func delimitedFromTokens(source string, tokens []shared.Token) Delimited {
	span := shared.SpanOf(tokens)
	return Delimited{Span: span, Text: shared.SliceSpan(source, span), Tokens: cloneTokens(tokens)}
}

func cloneTokens(tokens []shared.Token) []shared.Token {
	return append([]shared.Token(nil), tokens...)
}

func eofPosition(tokens []shared.Token) shared.Position {
	if len(tokens) == 0 {
		return shared.Position{Line: 1, Column: 1}
	}
	return tokens[len(tokens)-1].Span.End
}

func unmatchedDiagnostic(token shared.Token, delimiter string) shared.Diagnostic {
	return shared.Diagnostic{
		Severity: shared.SeverityError,
		Summary:  "unmatched " + delimiter,
		Detail:   "the closing delimiter has no matching opener in this paragraph",
		Span:     token.Span,
	}
}

func tokensOutsideParserSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	var result []shared.Token
	for _, token := range tokens {
		if token.Span.Start.Offset >= span.Start.Offset && token.Span.End.Offset <= span.End.Offset {
			continue
		}
		result = append(result, token)
	}
	return result
}
