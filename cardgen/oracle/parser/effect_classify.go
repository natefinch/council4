package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func durationScopesAcrossAnd(current, next EffectKind) bool {
	return temporaryModifierEffect(current) && temporaryModifierEffect(next)
}

func temporaryModifierEffect(kind EffectKind) bool {
	switch kind {
	case EffectModifyPT, EffectGain, EffectGrantKeyword:
		return true
	default:
		return false
	}
}

func targetsInSpan(targets []TargetSyntax, span shared.Span) []TargetSyntax {
	var result []TargetSyntax
	for i := range targets {
		target := &targets[i]
		if target.Span.Start.Offset >= span.Start.Offset && target.Span.End.Offset <= span.End.Offset {
			result = append(result, *target)
		}
	}
	return result
}

func resolvingClauseStart(tokens []shared.Token, indices []int, effectIndex int) int {
	if effectIndex == 0 {
		return 0
	}
	for i := indices[effectIndex] - 1; i > indices[effectIndex-1]; i-- {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Semicolon ||
			equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
			return i + 1
		}
	}
	return 0
}

func parseEffectReplacement(tokens []shared.Token, atoms Atoms) EffectReplacementSyntax {
	if replacement, ok := parseInsteadOneOfEachReplacement(tokens); ok {
		return replacement
	}
	if len(tokens) < 2 ||
		!equalWord(tokens[len(tokens)-2], "instead") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return EffectReplacementSyntax{}
	}
	replacement := EffectReplacementSyntax{
		Kind: EffectReplacementInstead,
		Span: tokens[len(tokens)-2].Span,
	}
	if replacementHasUnsupportedSelectionModifier(tokens, atoms) {
		return replacement
	}
	twiceMany := effectHasTokenWords(tokens, "twice", "that", "many")
	thatMuchPlus := effectHasTokenWords(tokens, "that", "much", "damage", "plus")
	thatManyPlus := effectHasTokenWords(tokens, "that", "many", "plus")
	doubleThat := effectHasTokenWords(tokens, "double", "that", "damage") ||
		effectHasTokenWords(tokens, "twice", "that", "damage")
	if boolCount(twiceMany, thatMuchPlus, thatManyPlus, doubleThat) != 1 {
		return replacement
	}
	switch {
	case twiceMany:
		replacement.Kind = EffectReplacementTwiceThatMany
	case thatMuchPlus:
		for i := range tokens {
			if !equalWord(tokens[i], "plus") || i+1 >= len(tokens) {
				continue
			}
			if amount, ok := effectNumber(tokens[i+1], atoms); ok {
				replacement.Kind = EffectReplacementThatMuchPlus
				replacement.Amount = amount
			}
			break
		}
	case thatManyPlus:
		for i := range tokens {
			if !equalWord(tokens[i], "plus") || i+1 >= len(tokens) {
				continue
			}
			if amount, ok := effectNumber(tokens[i+1], atoms); ok {
				replacement.Kind = EffectReplacementThatManyPlus
				replacement.Amount = amount
			}
			break
		}
	case doubleThat:
		replacement.Kind = EffectReplacementDoubleThat
	default:
	}
	replacement.EachCounterKind = effectHasTokenWords(tokens, "each", "of", "those", "kinds", "of", "counters")
	return replacement
}

// parseInsteadOneOfEachReplacement recognizes the "instead create one of each"
// output of a token-type replacement (Academy Manufactor: "If you would create a
// Clue, Food, or Treasure token, instead create one of each."). The replaced set
// of token types is carried by the create effect that owns this clause.
func parseInsteadOneOfEachReplacement(tokens []shared.Token) (EffectReplacementSyntax, bool) {
	words := normalizedWords(tokens)
	if !slices.Contains(words, "instead") {
		return EffectReplacementSyntax{}, false
	}
	if len(words) < 3 ||
		words[len(words)-3] != "one" ||
		words[len(words)-2] != "of" ||
		words[len(words)-1] != "each" {
		return EffectReplacementSyntax{}, false
	}
	return EffectReplacementSyntax{
		Kind: EffectReplacementOneOfEach,
		Span: shared.SpanOf(tokens),
	}, true
}

func replacementHasUnsupportedSelectionModifier(tokens []shared.Token, atoms Atoms) bool {
	selection := parseSelection(tokens, atoms)
	return selection.Controller != SelectionControllerAny ||
		selection.Another || selection.Other || selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped || selection.Keyword != KeywordUnknown ||
		selection.Zone != zone.None ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		len(selection.ExcludedTypes) != 0 || len(selection.Supertypes) != 0 ||
		len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
		len(selection.SubtypesAny) != 0
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func effectHasTokenWords(tokens []shared.Token, words ...string) bool {
	for i := range tokens {
		if effectWordsAt(tokens, i, words...) {
			return true
		}
	}
	return false
}

// stripLeadingAdditionalMana drops a leading "additional" qualifier, with its
// optional preceding article, from an add-mana body so "adds an additional {G}"
// parses to the same typed mana as "{G}" (Wild Growth and the mana-additional
// aura family). It is a no-op for bodies that do not begin with "additional".
func stripLeadingAdditionalMana(body []shared.Token) []shared.Token {
	rest := body
	if len(rest) >= 1 && (equalWord(rest[0], "a") || equalWord(rest[0], "an")) {
		if len(rest) >= 2 && equalWord(rest[1], "additional") {
			return rest[2:]
		}
		return body
	}
	if len(rest) >= 1 && equalWord(rest[0], "additional") {
		return rest[1:]
	}
	return body
}

func parseEffectMana(kind EffectKind, tokens []shared.Token, connected bool) EffectManaSyntax {
	if kind != EffectAddMana || len(tokens) == 0 {
		return EffectManaSyntax{}
	}
	body := tokens
	if tokens[len(tokens)-1].Kind == shared.Period {
		body = tokens[:len(tokens)-1]
	} else if !connected && !equalWord(tokens[len(tokens)-1], "instead") {
		return EffectManaSyntax{}
	}
	body = stripLeadingAdditionalMana(body)
	if len(body) == 5 && effectWordsAt(body, 0, "one", "mana", "of", "any", "color") {
		return EffectManaSyntax{Span: shared.SpanOf(body), AnyColor: true}
	}
	if len(body) == 6 && effectWordsAt(body, 1, "mana", "of", "any", "one", "color") {
		if count, ok := manaAnyOneColorCount(body[0]); ok {
			return EffectManaSyntax{Span: shared.SpanOf(body), AnyColor: true, AnyColorCount: count}
		}
	}
	if len(body) == 10 &&
		effectWordsAt(body, 0, "an", "amount", "of") &&
		body[3].Kind == shared.Symbol &&
		strings.EqualFold(body[3].Text, "{C}") &&
		effectWordsAt(body, 4, "equal", "to", "that") &&
		referencePossessiveObjectNoun(body[7]) &&
		effectWordsAt(body, 8, "mana", "value") {
		return EffectManaSyntax{Span: shared.SpanOf(body), DynamicColorless: true}
	}
	if len(body) == 10 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "color", "in", "your", "commander's", "color", "identity") {
		return EffectManaSyntax{Span: shared.SpanOf(body), CommanderIdentity: true}
	}
	if len(body) == 12 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "color", "that", "a", "land", "you", "control", "could", "produce") {
		return EffectManaSyntax{Span: shared.SpanOf(body), LandsProduce: true, LandsProduceScope: ManaLandsProduceYou}
	}
	if len(body) == 13 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "color", "that", "a", "land", "an", "opponent", "controls", "could", "produce") {
		return EffectManaSyntax{Span: shared.SpanOf(body), LandsProduce: true, LandsProduceScope: ManaLandsProduceOpponent}
	}
	if len(body) == 12 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "type", "that", "a", "land", "you", "control", "could", "produce") {
		return EffectManaSyntax{Span: shared.SpanOf(body), LandsProduce: true, LandsProduceScope: ManaLandsProduceYou, LandsProduceAnyType: true}
	}
	if len(body) == 13 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "type", "that", "a", "land", "an", "opponent", "controls", "could", "produce") {
		return EffectManaSyntax{Span: shared.SpanOf(body), LandsProduce: true, LandsProduceScope: ManaLandsProduceOpponent, LandsProduceAnyType: true}
	}
	if len(body) == 9 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "of", "the", "exiled", "card's", "colors") {
		return EffectManaSyntax{Span: shared.SpanOf(body), LinkedExileColors: true}
	}
	if len(body) == 6 && effectWordsAt(body, 0, "one", "mana", "of", "the", "chosen", "color") {
		return EffectManaSyntax{Span: shared.SpanOf(body), ChosenColor: true}
	}
	if len(body) == 14 &&
		effectWordsAt(body, 0, "an", "amount", "of", "mana", "of", "that", "color", "equal", "to", "your", "devotion", "to", "that", "color") {
		return EffectManaSyntax{Span: shared.SpanOf(body), ChosenColorDevotion: true}
	}
	if len(body) == 8 && body[0].Kind == shared.Symbol && equalWord(body[1], "or") &&
		effectWordsAt(body, 2, "one", "mana", "of", "the", "chosen", "color") {
		if fixed, ok := effectManaColor(body[0].Text); ok {
			return EffectManaSyntax{
				Span:                  shared.SpanOf(body),
				ChosenColor:           true,
				ChosenColorFixed:      fixed,
				ChosenColorFixedKnown: true,
			}
		}
	}
	if first, second, ok := filterPairManaBody(body); ok {
		return EffectManaSyntax{
			Span:         shared.SpanOf(body),
			FilterPair:   true,
			FilterColors: []mana.Color{first, second},
		}
	}
	// A trailing "instead" marks a conditional alternative mana production
	// ("Add {B}{B}{B}{B}{B} instead if ...", the Threshold cycle). The word
	// itself adds no mana, so strip it from the symbol body while recording the
	// flag and keeping it in the consumed span.
	instead := false
	loopBody := body
	if n := len(loopBody); n > 0 && equalWord(loopBody[n-1], "instead") {
		instead = true
		loopBody = loopBody[:n-1]
	}
	symbols, choice, ok := parseManaSymbolBody(loopBody)
	if !ok {
		return EffectManaSyntax{}
	}
	colors, colorsKnown := effectManaColors(symbols)
	return EffectManaSyntax{
		Span:        shared.SpanOf(body),
		Symbols:     symbols,
		Colors:      colors,
		ColorsKnown: colorsKnown,
		Choice:      choice,
		Instead:     instead,
	}
}

// parseManaSymbolBody reads a fixed mana-symbol sequence ("{B}{B}{B}") or a
// single-symbol choice list ("{W}, {U}, or {B}") from a body of tokens. It
// reports the recognized symbols, whether the body is a choice among them, and
// whether the body is a well-formed mana sequence at all.
func parseManaSymbolBody(body []shared.Token) (symbols []string, choice, ok bool) {
	expectSymbol := true
	for i := 0; i < len(body); i++ {
		token := body[i]
		if expectSymbol {
			if token.Kind != shared.Symbol {
				return nil, false, false
			}
			symbols = append(symbols, token.Text)
			expectSymbol = false
			continue
		}
		switch {
		case token.Kind == shared.Symbol:
			if choice {
				return nil, false, false
			}
			symbols = append(symbols, token.Text)
		case token.Kind == shared.Comma:
			if len(symbols) != 1 && !choice {
				return nil, false, false
			}
			choice = true
			expectSymbol = true
			if i+1 < len(body) && equalWord(body[i+1], "or") {
				i++
			}
		case equalWord(token, "or"):
			if len(symbols) != 1 && !choice {
				return nil, false, false
			}
			choice = true
			expectSymbol = true
		default:
			return nil, false, false
		}
	}
	if len(symbols) == 0 || expectSymbol || choice && len(symbols) < 2 {
		return nil, false, false
	}
	return symbols, choice, true
}

// manaAnyOneColorCount resolves the leading count of the body "<N> mana of any
// one color" (Gilded Lotus). It accepts an integer or cardinal-word count and
// requires N >= 2 so the single "one mana of any color" body keeps its own
// exact branch and "any combination of colors" wordings fail closed.
func manaAnyOneColorCount(token shared.Token) (int, bool) {
	count, ok := additionalLandCountWord(token)
	if !ok || count < 2 {
		return 0, false
	}
	return count, true
}

// effectManaColors maps every add-mana symbol to its typed basic mana color. It
// reports false (and discards the partial result) when any symbol is not one of
// the basic color tokens {W}{U}{B}{R}{G}{C}, so a consumer fails closed instead
// of re-reading the rendered symbol strings.
func effectManaColors(symbols []string) ([]mana.Color, bool) {
	colors := make([]mana.Color, 0, len(symbols))
	for _, symbol := range symbols {
		color, ok := effectManaColor(symbol)
		if !ok {
			return nil, false
		}
		colors = append(colors, color)
	}
	return colors, true
}

func effectManaColor(symbol string) (mana.Color, bool) {
	inner, ok := strings.CutPrefix(strings.ToUpper(symbol), "{")
	if !ok {
		return "", false
	}
	inner, ok = strings.CutSuffix(inner, "}")
	if !ok {
		return "", false
	}
	switch inner {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	case "C":
		return mana.C, true
	default:
		return "", false
	}
}

// filterPairManaBody recognizes the "filter land" add-mana output body
// "{X}{X}, {X}{Y}, or {Y}{Y}." (after the leading "Add" verb, period removed),
// returning the pair's two distinct basic colors X and Y. The body must be
// exactly the nine tokens {X}{X}, {X}{Y}, or {Y}{Y} in that fixed order; every
// symbol must be one of the five basic colors W, U, B, R, or G; and the two
// colors must differ. Any deviation fails closed with ok=false so callers fall
// through to the generic add-mana parse.
func filterPairManaBody(body []shared.Token) (first, second mana.Color, ok bool) {
	if len(body) != 9 {
		return "", "", false
	}
	if body[2].Kind != shared.Comma || body[5].Kind != shared.Comma || !equalWord(body[6], "or") {
		return "", "", false
	}
	colors := make([]mana.Color, 0, 6)
	for _, index := range []int{0, 1, 3, 4, 7, 8} {
		if body[index].Kind != shared.Symbol {
			return "", "", false
		}
		manaColor, valid := effectManaColor(body[index].Text)
		if !valid || manaColor == mana.C {
			return "", "", false
		}
		colors = append(colors, manaColor)
	}
	first, second = colors[0], colors[3]
	if first == second {
		return "", "", false
	}
	// The three printed groups must be exactly {X}{X}, {X}{Y}, and {Y}{Y}, with
	// first=X (colors[0]) and second=Y (colors[3]).
	if colors[1] != first || colors[2] != first || colors[4] != second || colors[5] != second {
		return "", "", false
	}
	return first, second, true
}

func effectConnection(tokens []shared.Token, indices []int, effectIndex int) (EffectConnectionKind, shared.Span) {
	if effectIndex == 0 {
		if indices[effectIndex] > 0 && equalWord(tokens[0], "then") {
			return EffectConnectionThen, tokens[0].Span
		}
		return EffectConnectionNone, shared.Span{}
	}
	for i := indices[effectIndex] - 1; i > indices[effectIndex-1]; i-- {
		switch {
		case equalWord(tokens[i], "then"):
			return EffectConnectionThen, tokens[i].Span
		case equalWord(tokens[i], "and"):
			return EffectConnectionAnd, tokens[i].Span
		}
	}
	return EffectConnectionNone, shared.Span{}
}

func effectOptional(tokens []shared.Token, index int) (bool, shared.Span) {
	start := max(0, index-3)
	for i, token := range tokens[start:index] {
		if equalWord(token, "may") {
			span := token.Span
			tokenIndex := start + i
			if tokenIndex > 0 && equalWord(tokens[tokenIndex-1], "you") {
				span.Start = tokens[tokenIndex-1].Span.Start
			}
			return true, span
		}
	}
	return false, shared.Span{}
}

func parseEffectDestination(tokens []shared.Token) EffectDestinationPosition {
	words := normalizedWords(tokens)
	switch {
	case effectContainsWords(words, "on", "top", "of", "your", "library") ||
		effectContainsWords(words, "on", "the", "top", "of", "your", "library") ||
		effectContainsWords(words, "on", "top", "of", "its", "owner's", "library") ||
		effectContainsWords(words, "on", "the", "top", "of", "its", "owner's", "library") ||
		effectContainsWords(words, "on", "top", "of", "their", "owner's", "library") ||
		effectContainsWords(words, "on", "the", "top", "of", "their", "owner's", "library"):
		return EffectDestinationTop
	case effectContainsWords(words, "on", "bottom", "of", "your", "library") ||
		effectContainsWords(words, "on", "the", "bottom", "of", "your", "library") ||
		effectContainsWords(words, "on", "bottom", "of", "its", "owner's", "library") ||
		effectContainsWords(words, "on", "the", "bottom", "of", "its", "owner's", "library") ||
		effectContainsWords(words, "on", "bottom", "of", "their", "owner's", "library") ||
		effectContainsWords(words, "on", "the", "bottom", "of", "their", "owner's", "library"):
		return EffectDestinationBottom
	default:
		return EffectDestinationUnspecified
	}
}

func effectWordsAtAny(tokens []shared.Token, first, second string) bool {
	for i := range tokens {
		if equalWord(tokens[i], first) {
			for _, token := range tokens[i+1:] {
				if equalWord(token, second) {
					return true
				}
			}
		}
	}
	return false
}

func effectContextAt(tokens []shared.Token, index int, atoms Atoms) EffectContextKind {
	start := effectSubjectStart(tokens, index, atoms.SelfNameSpans())
	subject := tokens[start:index]
	// "You and target <player> each <verb>" splits on its "and" so the retained
	// subject is "target <player> each". Recognize the dropped "you and" prefix
	// from the raw tokens before the split point to classify the compound
	// controller-and-target recipient.
	youAndPrefix := start >= 2 && equalWord(tokens[start-1], "and") && equalWord(tokens[start-2], "you")
	for len(subject) > 0 && equalWord(subject[0], "then") {
		subject = subject[1:]
	}
	// A "random" or "named" word in the subject marks a shape this resolver does
	// not classify (e.g. "a creature named X"); the subject portion is scanned
	// rather than the whole sentence so an object-position token name ("... token
	// named X") does not suppress an otherwise-recognized controller subject.
	for _, token := range subject {
		if equalWord(token, "random") || equalWord(token, "named") {
			return EffectContextUnknown
		}
	}
	if len(subject) == 0 {
		return EffectContextController
	}
	words := normalizedWords(subject)
	if len(words) == 0 {
		return EffectContextUnknown
	}
	if words[len(words)-1] == "may" {
		words = words[:len(words)-1]
	}
	if len(words) == 0 {
		return EffectContextUnknown
	}
	switch {
	case effectContainsWords(words, "each", "other", "player") ||
		effectContainsWords(words, "each", "other", "players"):
		return EffectContextEachOtherPlayer
	case effectContainsWords(words, "each", "opponent") || effectContainsWords(words, "each", "opponents"):
		return EffectContextEachOpponent
	case effectContainsWords(words, "each", "player"):
		return EffectContextEachPlayer
	case len(words) >= 2 && youAndPrefix && words[0] == "target" &&
		words[len(words)-1] == "each":
		// "You and target <player> each <verb>": the controller and a single
		// player target both receive the effect.
		return EffectContextControllerAndTarget
	case effectContainsWords(words, "target"):
		return EffectContextTarget
	case len(words) >= 2 && words[len(words)-2] == "that" && words[len(words)-1] == "player":
		return EffectContextReferencedPlayer
	case words[len(words)-1] == "controller" && subjectReferencesObject(subject, atoms):
		return EffectContextReferencedObjectController
	case words[len(words)-1] == "they":
		return EffectContextEventPlayer
	case words[len(words)-1] == "you" || len(words) >= 2 && words[len(words)-2] == "you" && words[len(words)-1] == "may":
		return EffectContextController
	}
	span := shared.SpanOf(subject)
	for _, reference := range atoms.References() {
		if !spanCovers(span, reference.Span) {
			continue
		}
		switch {
		case reference.Kind == ReferenceSelfName || reference.Kind == ReferenceThisObject:
			return EffectContextSource
		case reference.Kind == ReferencePronoun && reference.Pronoun == PronounThey:
			return EffectContextEventPlayer
		case reference.Kind == ReferenceThatObject:
			return EffectContextReferencedObject
		case reference.Kind == ReferenceThatPlayer:
			return EffectContextReferencedPlayer
		case reference.Kind == ReferencePronoun && reference.Pronoun == PronounIt:
			return EffectContextReferencedObject
		}
	}
	return EffectContextUnknown
}

// subjectReferencesObject reports whether the subject tokens contain a
// referenced-object pronoun ("it"/"its") or a "that <object>" reference,
// identifying a "<referenced object>'s controller" recipient.
func subjectReferencesObject(subject []shared.Token, atoms Atoms) bool {
	span := shared.SpanOf(subject)
	for _, reference := range atoms.References() {
		if !spanCovers(span, reference.Span) {
			continue
		}
		switch {
		case reference.Kind == ReferenceThatObject:
			return true
		case reference.Kind == ReferencePronoun &&
			(reference.Pronoun == PronounIt || reference.Pronoun == PronounIts):
			return true
		}
	}
	return false
}

func effectHasExplicitSubject(tokens []shared.Token, index int, selfNames []shared.Span) bool {
	return effectSubjectStart(tokens, index, selfNames) < index
}

func effectSubjectStart(tokens []shared.Token, index int, selfNames []shared.Span) int {
	start := 0
	for i := range index {
		if spanWithinAny(tokens[i].Span, selfNames) {
			continue
		}
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period || tokens[i].Kind == shared.Semicolon ||
			equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
			start = i + 1
		}
	}
	return start
}

// spanWithinAny reports whether span is covered by any of the given spans. It
// lets subject-boundary detection ignore commas and conjunctions that fall
// inside the card's own printed name (e.g. "Syr Konrad, the Grim"), which would
// otherwise truncate the subject at the name's internal comma.
func spanWithinAny(span shared.Span, spans []shared.Span) bool {
	for _, outer := range spans {
		if spanCovers(outer, span) {
			return true
		}
	}
	return false
}

func parseEffectPayment(tokens []shared.Token, atoms Atoms) EffectPaymentSyntax {
	for i := range tokens {
		var payer EffectPaymentPayerKind
		switch {
		case effectWordsAt(tokens, i, "unless", "its", "controller", "pays"):
			payer = EffectPaymentPayerTargetController
		case effectWordsAt(tokens, i, "unless", "that", "player", "pays"):
			payer = EffectPaymentPayerEventPlayer
		default:
			continue
		}
		manaCost, end, ok := parseKeywordManaCost(tokens, i+4)
		if !ok || end >= len(tokens) {
			return EffectPaymentSyntax{}
		}
		var genericAmount EffectAmountSyntax
		paymentEnd := end
		switch {
		case tokens[end].Kind == shared.Period && end == len(tokens)-1:
		case len(manaCost) == 1 && manaCost[0] == cost.X &&
			tokens[end].Kind == shared.Comma && end+1 < len(tokens):
			amount, attempted, amountOK := parseDynamicEffectAmount(tokens[end+1:], atoms)
			if !attempted || !amountOK ||
				amount.DynamicForm != EffectDynamicAmountFormWhereX ||
				amount.DynamicKind != EffectDynamicAmountSourcePower ||
				amount.Multiplier != 1 ||
				amount.Span.End != tokens[len(tokens)-1].Span.Start {
				return EffectPaymentSyntax{}
			}
			genericAmount = amount
			paymentEnd = len(tokens) - 1
		default:
			return EffectPaymentSyntax{}
		}
		return EffectPaymentSyntax{
			Span:              shared.SpanOf(tokens[i:paymentEnd]),
			Form:              EffectPaymentFormUnless,
			Payer:             payer,
			ManaCost:          manaCost,
			GenericManaAmount: genericAmount,
		}
	}
	return EffectPaymentSyntax{}
}

func effectIndices(tokens []shared.Token, atoms Atoms) []int {
	var result []int
	for i := range tokens {
		if effectKindAt(tokens, i) != EffectUnknown &&
			!atoms.SelfNameAt(tokens[i].Span) &&
			!effectNounAt(tokens, i) {
			result = append(result, i)
		}
	}
	return result
}

func effectNounAt(tokens []shared.Token, index int) bool {
	return index > 0 && index+1 < len(tokens) &&
		equalWord(tokens[index], "untap") &&
		equalWord(tokens[index-1], "next") &&
		equalWord(tokens[index+1], "step")
}

// cantBeBlockedThisTurnVerbAt reports whether the temporary prohibition "can't
// be blocked this turn" / "cannot be blocked this turn" begins at index. It
// anchors the temporary can't-be-blocked resolving effect ("Target creature
// can't be blocked this turn.") on the negated "can't"/"cannot" so the subject
// is the targeted creature. The "this turn" tail distinguishes this resolving,
// until-end-of-turn effect from the continuous static prohibitions ("Enchanted
// creature can't be blocked.", "... can't be blocked by ...") that carry no
// turn duration, so those keep flowing through the static-declaration path. The
// exactness recognizer reconstructs the full clause, so any other wording (a
// different operation, an "except by" qualifier) still fails closed.
func cantBeBlockedThisTurnVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		index+4 < len(tokens) &&
		equalWord(tokens[index+1], "be") &&
		equalWord(tokens[index+2], "blocked") &&
		equalWord(tokens[index+3], "this") &&
		equalWord(tokens[index+4], "turn")
}

// pastCastCountPhraseAt reports whether the "cast" verb at index is the past
// participle inside a "spell[s] you've cast this turn" / "...you have cast this
// turn" count phrase rather than a casting effect. The storm-counter dynamic
// amount ("you gain 1 life for each spell you've cast this turn") consumes that
// span as a count, so the bare "cast" must not also seed a separate cast effect.
func pastCastCountPhraseAt(tokens []shared.Token, index int) bool {
	if index == 0 {
		return false
	}
	contracted := equalWord(tokens[index-1], "you've")
	expanded := index >= 2 && equalWord(tokens[index-1], "have") && equalWord(tokens[index-2], "you")
	if !contracted && !expanded {
		return false
	}
	return effectWordsAt(tokens, index+1, "this", "turn")
}

func resolvingClauseEnd(tokens []shared.Token, indices []int, effectIndex int) int {
	start := indices[effectIndex] + 1
	end := len(tokens)
	for _, next := range indices[effectIndex+1:] {
		for i := next - 1; i >= start; i-- {
			if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Semicolon {
				end = i
				break
			}
			if equalWord(tokens[i], "then") || equalWord(tokens[i], "and") {
				end = i
				if i > start && tokens[i-1].Kind == shared.Comma {
					end--
				}
				break
			}
		}
		if end != len(tokens) {
			break
		}
	}
	for i := start; i < end; i++ {
		if equalWord(tokens[i], "if") || equalWord(tokens[i], "unless") ||
			(i+1 < end && equalWord(tokens[i], "only") && equalWord(tokens[i+1], "if")) {
			return i
		}
	}
	return end
}

// gainLoseLifeObject reports whether a gain/lose effect's grammatical object is
// the player's life rather than a keyword or quoted ability. It scans the
// post-verb clause for a top-level "life" word, ignoring tokens inside quoted
// granted abilities so that "gains \"... gain that much life\"" is treated as an
// ability grant, not a life change.
func gainLoseLifeObject(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectGain && kind != EffectLose {
		return false
	}
	quoted := false
	for _, token := range clause {
		switch token.Kind {
		case shared.Quote:
			quoted = !quoted
		case shared.Word:
			if !quoted && equalWord(token, "life") {
				return true
			}
		default:
		}
	}
	return false
}

// loseGameObject reports whether a lose effect's grammatical object is "the
// game" rather than life or a keyword. It scans the post-verb clause for a
// top-level "game" word outside any quoted granted ability.
func loseGameObject(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectLose {
		return false
	}
	quoted := false
	for _, token := range clause {
		switch token.Kind {
		case shared.Quote:
			quoted = !quoted
		case shared.Word:
			if !quoted && equalWord(token, "game") {
				return true
			}
		default:
		}
	}
	return false
}

// winGameVerbAt reports whether the "win"/"wins" verb at index governs the
// object "the game" ("you win the game"). The verb is generic, so the win
// classification is confirmed by scanning forward to the clause's terminating
// period for a top-level "game" word outside any quoted granted ability. This
// anchors EffectWinGame so effectIndices treats the verb as an effect start,
// mirroring how loseGameObject promotes the lose verb.
func winGameVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "win") && !equalWord(tokens[index], "wins") {
		return false
	}
	quoted := false
	for i := index + 1; i < len(tokens); i++ {
		switch tokens[i].Kind {
		case shared.Period, shared.Semicolon:
			return false
		case shared.Quote:
			quoted = !quoted
		case shared.Word:
			if !quoted && equalWord(tokens[i], "game") {
				return true
			}
		default:
		}
	}
	return false
}

func effectKindAt(tokens []shared.Token, index int) EffectKind {
	kind := effectWordKind(tokens[index])
	switch {
	case equalWord(tokens[index], "manifest"):
		switch {
		case effectWordsAt(tokens, index+1, "dread") && len(tokens) == index+3 && tokens[index+2].Kind == shared.Period:
			return EffectManifestDread
		case effectWordsAt(tokens, index+1, "the", "top", "card", "of", "your", "library") &&
			len(tokens) == index+8 && tokens[index+7].Kind == shared.Period:
			return EffectManifest
		default:
			return EffectManifest
		}
	case equalWord(tokens[index], "look"):
		if digLookInstruction(tokens[index:]) {
			return EffectDig
		}
		return EffectManifestDread
	case equalWord(tokens[index], "win") || equalWord(tokens[index], "wins"):
		if winGameVerbAt(tokens, index) {
			return EffectWinGame
		}
		return EffectUnknown
	case cantBeBlockedThisTurnVerbAt(tokens, index):
		return EffectCantBeBlocked
	case kind == EffectGrantKeyword && index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you"):
		return EffectUnknown
	case kind == EffectGrantKeyword && playerPossessionVerb(tokens, index):
		return EffectUnknown
	case kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared"):
		return EffectEnterPrepared
	case kind == EffectCast && index > 0 && (equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")):
		return EffectUnknown
	case kind == EffectCast && pastCastCountPhraseAt(tokens, index):
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case chooseNewTargetsVerbAt(tokens, index):
		return EffectChooseNewTargets
	case kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control"):
		return EffectGainControl
	case kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike"):
		return EffectUnknown
	case kind == EffectGrantKeyword && priorPTChange(tokens, index):
		return EffectUnknown
	case kind == EffectGrantKeyword && effectWordsAt(tokens, index+1, "the", "same", "name"):
		return EffectUnknown
	default:
		return kind
	}
}

func effectWordKind(token shared.Token) EffectKind {
	if token.Kind != shared.Word {
		return EffectUnknown
	}
	switch strings.ToLower(token.Text) {
	case "add", "adds":
		return EffectAddMana
	case "attach", "attaches":
		return EffectAttach
	case "cast", "casts":
		return EffectCast
	case "counter", "counters":
		return EffectCounter
	case "create", "creates":
		return EffectCreate
	case "deal", "deals":
		return EffectDealDamage
	case "destroy", "destroys":
		return EffectDestroy
	case "discard", "discards":
		return EffectDiscard
	case "discover", "discovers":
		return EffectDiscover
	case "double", "doubles":
		return EffectDouble
	case "draw", "draws":
		return EffectDraw
	case "enters":
		return EffectEnterTapped
	case "exile", "exiles":
		return EffectExile
	case "fight", "fights":
		return EffectFight
	case "gain", "gains":
		return EffectGain
	case "has", "have":
		return EffectGrantKeyword
	case "investigate", "investigates":
		return EffectInvestigate
	case "explore", "explores":
		return EffectExplore
	case "lose", "loses":
		return EffectLose
	case "manifest":
		return EffectManifest
	case "mill", "mills":
		return EffectMill
	case "get", "gets":
		return EffectModifyPT
	case "put", "puts":
		return EffectPut
	case "proliferate", "proliferates":
		return EffectProliferate
	case "regenerate", "regenerates":
		return EffectRegenerate
	case "return", "returns":
		return EffectReturn
	case "reveal", "reveals":
		return EffectReveal
	case "sacrifice", "sacrifices":
		return EffectSacrifice
	case "scry", "scries":
		return EffectScry
	case "surveil", "surveils":
		return EffectSurveil
	case "search", "searches":
		return EffectSearch
	case "shuffle", "shuffles":
		return EffectShuffle
	case "tap", "taps":
		return EffectTap
	case "untap", "untaps":
		return EffectUntap
	case "transform", "transforms":
		return EffectTransform
	default:
		return EffectUnknown
	}
}

// digLookInstruction reports whether the sentence is the impulse look clause
// "Look at the top <number> cards of your library." that introduces a dig: the
// player looks at a fixed number of top cards before a following "Put ..."
// sentence sorts them. The looked-at count is any number word; the exactness
// recognizer rejects a variable ("X") or non-numeric word so only fixed digs
// reach the combined lowerer.
func digLookInstruction(tokens []shared.Token) bool {
	return len(tokens) == 10 &&
		effectWordsAt(tokens, 0, "look", "at", "the", "top") &&
		tokens[4].Kind == shared.Word &&
		effectWordsAt(tokens, 5, "cards", "of", "your", "library") &&
		tokens[9].Kind == shared.Period
}

// chooseNewTargetsVerbAt reports whether a retarget effect ("[You may] choose
// new targets for <target spell or ability>.") begins at index. The parser owns
// this wording: the verb "choose" is generic, so the retarget classification is
// anchored on the exact "choose new targets for" lead-in. The copy-spell rider
// "choose new targets for the copy" carries a non-stack ("the copy") object and
// fails the exactness check, so it stays unsupported rather than misclassifying.
func chooseNewTargetsVerbAt(tokens []shared.Token, index int) bool {
	return equalWord(tokens[index], "choose") &&
		index+3 < len(tokens) &&
		equalWord(tokens[index+1], "new") &&
		equalWord(tokens[index+2], "targets") &&
		equalWord(tokens[index+3], "for")
}

func counterVerbAt(tokens []shared.Token, index int) bool {
	if index == 0 {
		return true
	}
	previous := tokens[index-1]
	if previous.Kind == shared.Comma || previous.Kind == shared.Period || previous.Kind == shared.Semicolon ||
		equalWord(previous, "then") || equalWord(previous, "may") || equalWord(previous, "can") {
		return true
	}
	return index+1 < len(tokens) &&
		(equalWord(tokens[index+1], "target") || equalWord(tokens[index+1], "it") || equalWord(tokens[index+1], "that"))
}

// playerPossessionVerb reports whether the "has"/"have" verb at index expresses
// player possession ("you have", "a player has", "an opponent has") rather than
// an object keyword grant. A player never has keyword abilities, so this verb
// never introduces a keyword-grant effect; it typically belongs to a condition
// clause such as "As long as you have seven or more cards in hand".
func playerPossessionVerb(tokens []shared.Token, index int) bool {
	if index < 1 {
		return false
	}
	previous := tokens[index-1]
	return equalWord(previous, "you") || equalWord(previous, "player") ||
		equalWord(previous, "players") || equalWord(previous, "opponent") ||
		equalWord(previous, "opponents")
}

func priorPTChange(tokens []shared.Token, index int) bool {
	for i := range index {
		if equalWord(tokens[i], "get") || equalWord(tokens[i], "gets") {
			power, toughness := parsePTChange(tokens[i+1 : index])
			return power.Known && toughness.Known
		}
	}
	return false
}

func effectIsNegated(tokens []shared.Token, index int) bool {
	start := max(0, index-3)
	for i, token := range tokens[start:index] {
		if equalWord(token, "can't") || equalWord(token, "cannot") ||
			equalWord(token, "doesn't") || equalWord(token, "don't") || equalWord(token, "not") {
			for _, following := range tokens[start+i+1 : index] {
				if equalWord(following, "control") {
					return false
				}
			}
			return true
		}
	}
	return false
}
