package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/lexer"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// Parse builds a lossless syntax tree for source. It returns a partial tree
// alongside localized diagnostics when the input is malformed.
func Parse(source string, context Context) (Document, []shared.Diagnostic) {
	source = expandBushidoKeyword(source)
	source = expandExtortKeyword(source)
	source = expandDevourKeyword(source)
	source = expandAnnihilatorKeyword(source)
	source = expandAfflictKeyword(source)
	source = expandFrenzyKeyword(source)
	source = expandAfterlifeKeyword(source)
	source = expandRenownKeyword(source)
	source = expandModularKeyword(source)
	source = expandAffinityKeyword(source)
	source = expandBattleCryKeyword(source)
	source = expandTributeKeyword(source)
	source = expandMentorKeyword(source)
	source = expandFusedTrigger(source)
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
			if dash >= 0 && dash+1 < len(modalTokens) {
				headerTokens = modalTokens[:dash+1]
			}
			modal := &Modal{header: phraseFromTokens(source, headerTokens)}
			j := i + 1
			if dash >= 0 && dash+1 < len(modalTokens) {
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
		} else if table, next, ok := parseDiceTable(source, lines, i); ok {
			ability.DiceTable = table
			lastRow := table.Rows[len(table.Rows)-1]
			ability.Span.End = lastRow.Span.End
			ability.Text = shared.SliceSpan(source, ability.Span)
			i = next
		} else if isSpreeHeader(lines[i]) {
			modal := &Modal{header: phraseFromTokens(source, lines[i]), Spree: true}
			j := i + 1
			for j < len(lines) && startsWith(lines[j], shared.Plus) {
				mode, modeDiagnostics := parseSpreeMode(source, lines[j][1:])
				modal.Options = append(modal.Options, mode)
				diagnostics = append(diagnostics, modeDiagnostics...)
				j++
			}
			if len(modal.Options) == 0 {
				i++
			} else {
				ability.Span.End = modal.Options[len(modal.Options)-1].Span.End
				ability.Text = shared.SliceSpan(source, ability.Span)
				ability.Modal = modal
				i = j
			}
		} else {
			i++
		}
		document.Abilities = append(document.Abilities, ability)
	}
	emitAtoms(document.Abilities, context.CardName)
	emitSelfNameStaticRules(document.Abilities)
	emitCost(document.Abilities)
	emitOptional(document.Abilities)
	emitConditionBoundaries(document.Abilities, context.CardName)
	emitTriggerEventClauses(document.Abilities, context.CardName)
	emitEventHistoryConditions(document.Abilities)
	emitConditionClauses(document.Abilities)
	emitSourceAbilityCostReduction(document.Abilities)
	emitResolvingSyntax(document.Abilities)
	emitSourceSpellCostReduction(document.Abilities)
	emitSourceSpellCostReductionDynamic(document.Abilities)
	emitStaticDeclarations(document.Abilities)
	stripCastThisFromExileEffectSemantics(document.Abilities)
	emitSemanticAccessors(document.Abilities)
	stripImpulseExileSemantics(document.Abilities)
	emitCoinFlipSequences(document.Abilities)
	emitReminderInner(document.Abilities)
	emitSourceOrder(document.Abilities)
	stripConditionalModalHeaderSemantics(document.Abilities)
	return document, diagnostics
}

// stripCastThisFromExileEffectSemantics suppresses the resolving-effect reading
// of "You may cast this card from exile." (Misthollow Griffin, Eternal Scourge).
// Because "exile" is also an effect verb, the resolving-syntax scanner mis-reads
// the zone noun in "from exile" as an exile effect, leaving the ability with
// spurious effects that block the text-blind compiler's empty-content
// recognition of the cast-from-exile player rule. When the static idiom is
// recognized, the static declaration owns the whole sentence, so its competing
// effect and target syntax is cleared.
func stripCastThisFromExileEffectSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasCastThisFromExileDeclaration(ability) {
			continue
		}
		for j := range ability.Sentences {
			sentence := &ability.Sentences[j]
			sentence.Effects = nil
			sentence.Targets = nil
			sentence.LegacyEffects = false
		}
	}
}

func abilityHasCastThisFromExileDeclaration(ability *Ability) bool {
	for i := range ability.StaticDeclarations {
		if ability.StaticDeclarations[i].PlayerRule == StaticDeclarationPlayerRuleCastThisFromExile {
			return true
		}
	}
	return false
}

func stripImpulseExileSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if sentencesHaveImpulseExile(ability.Sentences) {
			ability.SemanticReferences = nil
			ability.SemanticKeywords = nil
			ability.ConditionBoundaries = nil
			ability.EventHistoryConditions = nil
			ability.ConditionClauses = nil
			ability.ConditionSegments = nil
		}
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			if sentencesHaveImpulseExile(mode.Sentences) {
				mode.SemanticReferences = nil
				mode.SemanticKeywords = nil
				mode.ConditionBoundaries = nil
				mode.EventHistoryConditions = nil
				mode.ConditionClauses = nil
				mode.ConditionSegments = nil
			}
		}
	}
}

func stripConditionalModalHeaderSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal == nil ||
			ability.Modal.ChoiceBonus.Condition != ModalChoiceBonusConditionControlsCommander {
			continue
		}
		ability.Sentences = nil
		ability.ConditionBoundaries = nil
		ability.EventHistoryConditions = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.SemanticReferences = nil
		ability.SemanticKeywords = nil
	}
}

// emitReminderInner parses the inner content of each fully-parenthesized reminder
// ability once, so a consumer lowering a reminder mana ability ("({T}: Add {G}.)")
// reads typed inner syntax instead of re-parsing the reminder wording. The inner
// parse runs through the same pipeline with an empty context, reproducing exactly
// what a consumer's own re-parse would have produced.
func emitReminderInner(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Kind != AbilityReminder {
			continue
		}
		if len(ability.Text) < 2 || ability.Text[0] != '(' || ability.Text[len(ability.Text)-1] != ')' {
			continue
		}
		inner := strings.TrimSpace(ability.Text[1 : len(ability.Text)-1])
		document, diagnostics := Parse(inner, Context{})
		ability.reminderInner = &reminderInner{document: document, diagnostics: diagnostics}
	}
}

// emitOptional records the leading optional "you may" choice on each triggered
// ability whose resolving body begins with those two words. The compiler reads
// the typed flag and span instead of inspecting "you"/"may" tokens.
func emitOptional(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Kind != AbilityTriggered {
			continue
		}
		body := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
		semantic := eventHistorySemanticTokens(body, ability.Reminders, ability.Quoted)
		if len(semantic) >= 2 && equalWord(semantic[0], "you") && equalWord(semantic[1], "may") {
			ability.Optional = true
			ability.OptionalSpan = shared.Span{Start: semantic[0].Span.Start, End: semantic[1].Span.End}
		}
	}
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
		if abilities[i].DiceTable != nil {
			for k := range abilities[i].DiceTable.Rows {
				row := &abilities[i].DiceTable.Rows[k]
				row.Atoms = collectAtoms(row.Tokens, nil, nil, cardName)
			}
		}
		if abilities[i].Modal == nil {
			continue
		}
		abilities[i].Modal.Atoms = collectAtoms(abilities[i].Modal.header.Tokens, nil, nil, cardName)
		if abilities[i].Modal.Spree {
			abilities[i].Modal.MinModes = 1
			abilities[i].Modal.MaxModes = len(abilities[i].Modal.Options)
			abilities[i].Modal.ChoiceKind = ModalChoiceKindOneOrMore
			abilities[i].Modal.ChoiceKnown = true
		} else {
			choice := recognizeModalChoice(abilities[i].Modal.header, abilities[i].Modal.Atoms)
			abilities[i].Modal.MinModes = choice.minModes
			abilities[i].Modal.MaxModes = choice.maxModes
			if choice.maxModes < 0 {
				abilities[i].Modal.MaxModes = len(abilities[i].Modal.Options)
			}
			abilities[i].Modal.ChoiceKind = choice.kind
			abilities[i].Modal.ChoiceBonus = choice.bonus
			abilities[i].Modal.ChoiceKnown = choice.ok
		}
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			mode.Atoms = collectAtoms(mode.Body.Tokens, mode.Reminders, mode.Quoted, cardName)
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
		} else if flashback, costPhrase, ok := flashbackAlternativeCostClause(source, tokens, dash); ok {
			ability.AlternativeCost = flashback
			ability.costPhrase = &costPhrase
		} else if escape, costPhrase, ok := escapeAlternativeCostClause(source, tokens, dash); ok {
			ability.AlternativeCost = escape
			ability.costPhrase = &costPhrase
		} else {
			phrase := phraseFromTokens(source, tokens[:dash])
			ability.AbilityWord = &AbilityWordClause{
				Label:         phrase.Text,
				Span:          phrase.Span,
				SeparatorSpan: tokens[dash].Span,
			}
		}
		body = tokens[dash+1:]
	}
	if colon := shared.TopLevelIndex(body, shared.Colon); colon >= 0 {
		phrase := phraseFromTokens(source, body[:colon])
		ability.costPhrase = &phrase
	}
	switch {
	case len(ability.Chapters) > 0:
		ability.Kind = AbilityChapter
	case ability.AlternativeCost != nil:
		ability.Kind = AbilitySpellAlternativeCost
	default:
		ability.Kind = classifyAbility(body, context)
	}
	// The "As an additional cost to cast this spell," prefix applies to any card
	// cast as a spell, including permanent spells (creatures, artifacts), so it
	// is recognized regardless of how the paragraph was otherwise classified.
	// It is excluded when the paragraph already carries a colon activation cost.
	if ability.Kind != AbilityChapter && ability.costPhrase == nil {
		if phrase, ok := spellAdditionalCostClause(source, body); ok {
			ability.Kind = AbilitySpellAdditionalCost
			ability.costPhrase = &phrase
		}
	}
	if ability.Kind != AbilityChapter && ability.costPhrase == nil {
		if alternative, ok := spellAlternativeCostClause(body); ok {
			ability.Kind = AbilitySpellAlternativeCost
			ability.AlternativeCost = alternative
		}
	}
	if ability.Kind == AbilityTriggered {
		ability.Trigger = parseTriggerClause(source, body, context.CardName)
	}
	resolvingBody := resolvingBodyTokens(body, ability.Kind, context.CardName)
	ability.BodySpan = shared.SpanOf(resolvingBody)
	ability.BodySeparatorSpan = separatorBeforeBody(ability.Tokens, ability.BodySpan)
	ability.Sentences = ParseSentences(source, resolvingBody)
	if ability.Kind == AbilityTriggered {
		ability.TriggerFrequency = parseTrailingTriggerFrequency(source, resolvingBody)
	}
	var diagnostics []shared.Diagnostic
	ability.Reminders, ability.Quoted, diagnostics = parseDelimited(source, body, diagnostics)
	if ability.Kind == AbilityReminder && context.Saga {
		ability.SagaReminder = recognizeSagaLoreReminder(ability.Text)
	}
	if context.Class {
		if ability.Kind == AbilityReminder {
			ability.ClassReminder = recognizeClassLevelReminder(ability.Text)
		}
		if ability.Kind == AbilityActivated {
			ability.ClassLevelGain = recognizeClassLevelGain(resolvingBody)
		}
	}
	ability.ReadAheadSacrificeChapter, ability.ReadAheadRecognized = recognizeReadAheadReminder(ability.Text)
	ability.DevoidRecognized = ability.Text == "Devoid (This card has no color.)"
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

func resolvingBodyTokens(tokens []shared.Token, kind AbilityKind, cardName string) []shared.Token {
	switch kind {
	case AbilityActivated, AbilityLoyalty:
		if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case AbilityTriggered:
		if comma := triggerBodyComma(tokens, cardName); comma >= 0 {
			return tokens[comma+1:]
		}
	case AbilitySpellAdditionalCost, AbilitySpellAlternativeCost:
		return nil
	default:
	}
	return tokens
}

// spellAdditionalCostClause recognizes a spell paragraph of the fixed form
// "As an additional cost to cast this spell, <cost>." and returns the trailing
// cost clause as a Phrase (without the leading prefix or trailing period). The
// prefix is fixed Oracle boilerplate, so it is matched by its exact words; the
// cost clause itself is left for the shared cost machinery to recognize.
func spellAdditionalCostClause(source string, body []shared.Token) (Phrase, bool) {
	comma := shared.TopLevelIndex(body, shared.Comma)
	if comma < 0 {
		return Phrase{}, false
	}
	if !slices.Equal(normalizedWords(body[:comma]),
		[]string{"as", "an", "additional", "cost", "to", "cast", "this", "spell"}) {
		return Phrase{}, false
	}
	clause := body[comma+1:]
	for len(clause) > 0 && clause[len(clause)-1].Kind == shared.Period {
		clause = clause[:len(clause)-1]
	}
	if len(clause) == 0 {
		return Phrase{}, false
	}
	return phraseFromTokens(source, clause), true
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

// recognizeSagaLoreReminder reports whether text is a Saga's intrinsic
// lore-counter reminder. The fixed reminder vocabulary is parser-owned; the
// optional trailing "Sacrifice after <chapter>" clause is validated through the
// roman-numeral chapter grammar so the recognition is composable rather than a
// single frozen phrase.
func recognizeSagaLoreReminder(text string) bool {
	const (
		withComma    = "(As this Saga enters and after your draw step, add a lore counter."
		withoutComma = "(As this Saga enters and after your draw step add a lore counter."
	)
	remainder, ok := strings.CutPrefix(text, withComma)
	if !ok {
		remainder, ok = strings.CutPrefix(text, withoutComma)
	}
	if !ok {
		return false
	}
	if remainder == ")" {
		return true
	}
	chapter, ok := strings.CutPrefix(remainder, " Sacrifice after ")
	if !ok || !strings.HasSuffix(chapter, ".)") {
		return false
	}
	_, ok = romanChapter(strings.TrimSuffix(chapter, ".)"))
	return ok
}

// recognizeReadAheadReminder reports whether text is the canonical "Read ahead"
// keyword line and reminder, returning the final lore chapter named by the
// optional "Sacrifice after <chapter>" clause (0 when omitted). The fixed
// reminder vocabulary is parser-owned; the chapter is recognized through the
// roman-numeral chapter grammar.
func recognizeReadAheadReminder(text string) (chapter int, ok bool) {
	const prefix = "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger."
	remainder, ok := strings.CutPrefix(text, prefix)
	if !ok {
		return 0, false
	}
	if remainder == ")" {
		return 0, true
	}
	suffix, ok := strings.CutPrefix(remainder, " Sacrifice after ")
	if !ok || !strings.HasSuffix(suffix, ".)") {
		return 0, false
	}
	chapter, ok = romanChapter(strings.TrimSuffix(suffix, ".)"))
	if !ok {
		return 0, false
	}
	return chapter, true
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

// lineEndsWithRollDie reports whether a line's final clause is "roll a d<N>."
// (ignoring any trailing period) and returns N. It detects the header line of a
// die-roll outcome table, whose result rows follow on subsequent lines.
func lineEndsWithRollDie(tokens []shared.Token) (int, bool) {
	end := len(tokens)
	for end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	if end < 4 {
		return 0, false
	}
	words := tokens[end-4 : end]
	if !equalWord(words[0], "roll") ||
		!equalWord(words[1], "a") ||
		!equalWord(words[2], "d") ||
		words[3].Kind != shared.Integer {
		return 0, false
	}
	sides, err := strconv.Atoi(words[3].Text)
	if err != nil || sides < 2 {
		return 0, false
	}
	return sides, true
}

// parseDiceTableRow recognizes a single outcome row "<low>[—<high>] | <body>"
// or "<value>+ | <body>" and returns its inclusive interval and resolving
// sentences. The em-dash, en-dash, or ASCII hyphen separate a range; a trailing
// plus opens the interval up to dieSides. It fails closed on any other shape so
// non-table lines flow through the ordinary ability parser.
func parseDiceTableRow(source string, tokens []shared.Token, dieSides int) (DiceTableRow, bool) {
	if len(tokens) == 0 || tokens[0].Kind != shared.Integer {
		return DiceTableRow{}, false
	}
	low, err := strconv.Atoi(tokens[0].Text)
	if err != nil {
		return DiceTableRow{}, false
	}
	high := low
	idx := 1
	switch {
	case idx < len(tokens) &&
		(tokens[idx].Kind == shared.EmDash || tokens[idx].Kind == shared.EnDash || tokens[idx].Kind == shared.Minus):
		if idx+1 >= len(tokens) || tokens[idx+1].Kind != shared.Integer {
			return DiceTableRow{}, false
		}
		high, err = strconv.Atoi(tokens[idx+1].Text)
		if err != nil {
			return DiceTableRow{}, false
		}
		idx += 2
	case idx < len(tokens) && tokens[idx].Kind == shared.Plus:
		high = dieSides
		idx++
	default:
	}
	if idx >= len(tokens) || tokens[idx].Kind != shared.Glyph || tokens[idx].Text != "|" {
		return DiceTableRow{}, false
	}
	idx++
	bodyTokens := tokens[idx:]
	if len(bodyTokens) == 0 || low > high {
		return DiceTableRow{}, false
	}
	span := shared.SpanOf(tokens)
	return DiceTableRow{
		Span:      span,
		Text:      shared.SliceSpan(source, span),
		Tokens:    cloneTokens(tokens),
		Min:       low,
		Max:       high,
		Sentences: ParseSentences(source, bodyTokens),
	}, true
}

// parseDiceTable consumes a die-roll outcome table starting at line i: the
// header line at i must end with "roll a d<N>." and line i+1 must be an outcome
// row. It collects every consecutive outcome row and returns the table together
// with the index of the first unconsumed line.
func parseDiceTable(source string, lines [][]shared.Token, i int) (*DiceTable, int, bool) {
	sides, ok := lineEndsWithRollDie(lines[i])
	if !ok || i+1 >= len(lines) {
		return nil, 0, false
	}
	firstRow, ok := parseDiceTableRow(source, lines[i+1], sides)
	if !ok {
		return nil, 0, false
	}
	table := &DiceTable{DieSides: sides, Rows: []DiceTableRow{firstRow}}
	j := i + 2
	for j < len(lines) {
		row, rowOK := parseDiceTableRow(source, lines[j], sides)
		if !rowOK {
			break
		}
		table.Rows = append(table.Rows, row)
		j++
	}
	return table, j, true
}

func parseMode(source string, tokens []shared.Token) (Mode, []shared.Diagnostic) {
	bodyTokens := tokens
	mode := Mode{
		Span:   shared.SpanOf(tokens),
		Text:   shared.SliceSpan(source, shared.SpanOf(tokens)),
		Tokens: cloneTokens(tokens),
	}
	if label, body, ok := parseModeLabel(source, tokens); ok {
		mode.Label = label
		bodyTokens = body
	}
	mode.Body = phraseFromTokens(source, bodyTokens)
	mode.Sentences = ParseSentences(source, bodyTokens)
	var diagnostics []shared.Diagnostic
	mode.Reminders, mode.Quoted, diagnostics = parseDelimited(source, bodyTokens, diagnostics)
	return mode, diagnostics
}

func parseModeLabel(source string, tokens []shared.Token) (*ModeLabelClause, []shared.Token, bool) {
	dash := shared.TopLevelIndex(tokens, shared.EmDash)
	if dash <= 0 || dash+1 >= len(tokens) {
		return nil, nil, false
	}
	labelTokens := tokens[:dash]
	var kind ModeLabelKind
	switch strings.ToLower(strings.TrimSpace(joinedTokenText(labelTokens))) {
	case "sell contraband":
		kind = ModeLabelSellContraband
	case "buy information":
		kind = ModeLabelBuyInformation
	case "hire a mercenary":
		kind = ModeLabelHireMercenary
	default:
		return nil, nil, false
	}
	span := shared.SpanOf(labelTokens)
	return &ModeLabelClause{
		Kind:          kind,
		Text:          shared.SliceSpan(source, span),
		Span:          span,
		SeparatorSpan: tokens[dash].Span,
	}, tokens[dash+1:], true
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
	words := normalizedWords(tokensOutsideParens(tokens))
	if len(words) >= 2 && words[0] == "as" && slices.Contains(words, "enters") {
		return true
	}
	if slices.Contains(words, "enters") &&
		(slices.Contains(words, "tapped") || slices.Contains(words, "with") || slices.Contains(words, "as")) {
		return true
	}
	if slices.Contains(words, "copy") && slices.Contains(words, "as") &&
		(slices.Contains(words, "enter") || slices.Contains(words, "enters")) {
		return true
	}
	if groupEntersTappedWording(words) {
		return true
	}
	return slices.Contains(words, "would") && slices.Contains(words, "instead")
}

// groupEntersTappedWording reports whether the tokens read as a static group
// enters-tapped replacement ("Creatures your opponents control enter tapped.").
// The plural "enter" distinguishes the group form from the self "enters tapped"
// already handled above; the leading permanent-type plural keeps unrelated
// "... enter ... tapped" spell text from being misclassified as a replacement.
func groupEntersTappedWording(words []string) bool {
	if len(words) < 3 || !slices.Contains(words, "enter") || !slices.Contains(words, "tapped") {
		return false
	}
	switch words[0] {
	case "creatures", "lands", "artifacts", "enchantments", "planeswalkers", "permanents":
		return true
	default:
		return false
	}
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

// recognizeModalChoice reads the typed cardinal grammar of a choose header and
// returns its (minModes, maxModes) choice range. It accepts "Choose <word> —"
// where <word> is a cardinal number spelled as a single word ("one", "two",
// etc.), plus the exact "Choose one or both —" header. The boolean result is
// false when the header is not one of those recognized shapes. Downstream
// lowering consumes the typed range instead of re-reading these tokens.
type modalChoiceRecognition struct {
	minModes int
	maxModes int
	kind     ModalChoiceKind
	bonus    ModalChoiceBonusSyntax
	ok       bool
}

func recognizeModalChoice(header Phrase, atoms Atoms) modalChoiceRecognition {
	tokens := header.Tokens
	if strings.EqualFold(
		strings.TrimSpace(header.Text),
		"Choose one. If you control a commander as you cast this spell, you may choose both instead.",
	) {
		return modalChoiceRecognition{
			minModes: 1,
			maxModes: 1,
			bonus: ModalChoiceBonusSyntax{
				Condition:          ModalChoiceBonusConditionControlsCommander,
				AdditionalMaxModes: 1,
			},
			ok: true,
		}
	}
	if len(tokens) == 5 &&
		tokens[0].Kind == shared.Word && strings.EqualFold(tokens[0].Text, "choose") &&
		tokens[1].Kind == shared.Word && strings.EqualFold(tokens[1].Text, "one") &&
		tokens[2].Kind == shared.Word && strings.EqualFold(tokens[2].Text, "or") &&
		tokens[3].Kind == shared.Word && strings.EqualFold(tokens[3].Text, "both") &&
		tokens[4].Kind == shared.EmDash {
		return modalChoiceRecognition{minModes: 1, maxModes: 2, ok: true}
	}
	if len(tokens) == 5 &&
		tokens[0].Kind == shared.Word && strings.EqualFold(tokens[0].Text, "choose") &&
		tokens[1].Kind == shared.Word && strings.EqualFold(tokens[1].Text, "one") &&
		tokens[2].Kind == shared.Word && strings.EqualFold(tokens[2].Text, "or") &&
		tokens[3].Kind == shared.Word && strings.EqualFold(tokens[3].Text, "more") &&
		tokens[4].Kind == shared.EmDash {
		return modalChoiceRecognition{minModes: 1, maxModes: -1, kind: ModalChoiceKindOneOrMore, ok: true}
	}
	if len(tokens) == 5 &&
		tokens[0].Kind == shared.Word && strings.EqualFold(tokens[0].Text, "choose") &&
		tokens[1].Kind == shared.Word && strings.EqualFold(tokens[1].Text, "up") &&
		tokens[2].Kind == shared.Word && strings.EqualFold(tokens[2].Text, "to") &&
		tokens[3].Kind == shared.Word &&
		tokens[4].Kind == shared.EmDash {
		// "Choose up to <number> —" is an optional modal choice: the controller
		// may pick between zero and <number> distinct modes.
		if n, numOK := atoms.CardinalAt(tokens[3].Span); numOK {
			return modalChoiceRecognition{minModes: 0, maxModes: n, ok: true}
		}
		return modalChoiceRecognition{}
	}
	// Expected: [Word("Choose"), Word(<number>), EmDash]
	if len(tokens) != 3 ||
		tokens[0].Kind != shared.Word || !strings.EqualFold(tokens[0].Text, "choose") ||
		tokens[1].Kind != shared.Word ||
		tokens[2].Kind != shared.EmDash {
		return modalChoiceRecognition{}
	}
	n, numOK := atoms.CardinalAt(tokens[1].Span)
	if !numOK {
		return modalChoiceRecognition{}
	}
	return modalChoiceRecognition{minModes: n, maxModes: n, ok: true}
}

func isModalHeader(tokens []shared.Token) bool {
	if strings.EqualFold(
		strings.TrimSpace(joinedTokenText(tokens)),
		"Choose one. If you control a commander as you cast this spell, you may choose both instead.",
	) {
		return true
	}
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

func joinedTokenText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && token.Span.Start.Offset > tokens[i-1].Span.End.Offset {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func modalHeaderStart(tokens []shared.Token) int {
	if isModalHeader(tokens) {
		return 0
	}
	for _, separator := range []shared.Kind{shared.Colon, shared.Comma} {
		index := shared.TopLevelIndex(tokens, separator)
		if index >= 0 && index+1 < len(tokens) && isModalHeader(tokens[index+1:]) {
			return index + 1
		}
	}
	return -1
}

// tokensOutsideParens returns the tokens that lie outside any parenthesized
// group, dropping the parentheses and their contents. Parenthesized spans are
// reminder text, which carries no game meaning, so ability classification must
// not read its wording.
func tokensOutsideParens(tokens []shared.Token) []shared.Token {
	outside := make([]shared.Token, 0, len(tokens))
	depth := 0
	for _, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			depth++
		case shared.RightParen:
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				outside = append(outside, token)
			}
		}
	}
	return outside
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

// tokensWithinParserSpan returns the tokens that lie within span. Because a
// token stream is contiguous and span is the span of a contiguous sub-slice,
// this selects exactly that sub-slice. An empty span selects no tokens.
func tokensWithinParserSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	var result []shared.Token
	for _, token := range tokens {
		if token.Span.Start.Offset >= span.Start.Offset && token.Span.End.Offset <= span.End.Offset {
			result = append(result, token)
		}
	}
	return result
}

// separatorBeforeBody returns the span of the single token that immediately
// precedes body in the ability's token stream — the cost colon, the triggered
// event comma, or a Saga chapter heading's em dash. It is the zero span when the
// body starts at the first token (no separator) or body is empty.
func separatorBeforeBody(tokens []shared.Token, body shared.Span) shared.Span {
	if body == (shared.Span{}) {
		return shared.Span{}
	}
	var separator shared.Span
	for _, token := range tokens {
		if token.Span.End.Offset <= body.Start.Offset {
			separator = token.Span
		}
	}
	return separator
}

// TokensInSpan returns the contiguous sub-slice of stream that lies within span.
// It lets a consumer slice an ability's token stream at a parser-emitted
// boundary (such as BodySpan) without scanning for separator token kinds.
func TokensInSpan(stream []shared.Token, span shared.Span) []shared.Token {
	return tokensWithinParserSpan(stream, span)
}

// TokensFrom returns the suffix of stream whose tokens start at or after offset.
// It lets a consumer slice an ability's token stream at a parser-emitted source
// offset without scanning for separator token kinds.
func TokensFrom(stream []shared.Token, offset int) []shared.Token {
	for i, token := range stream {
		if token.Span.Start.Offset >= offset {
			return stream[i:]
		}
	}
	return nil
}
