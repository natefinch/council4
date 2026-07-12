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
			equalWord(tokens[i], "then") || equalWord(tokens[i], "and") || equalWord(tokens[i], "or") {
			return i + 1
		}
	}
	return 0
}

// leadingInsteadReplacement recognizes the clause-initial "instead" replacement
// marker ("If <condition>, instead <effect>"), where the alternative effect
// replaces the immediately preceding effect when the condition holds. It is
// distinguished from the trailing "... instead." form by requiring the word to
// sit immediately after a comma, or to open the effect's ownership tokens when
// the preceding clause boundary (the condition's comma) has already been
// stripped ("... , instead exile it ..." whose effect clause begins at
// "instead"). It never treats a final trailing "instead" as a leading marker, so
// an ordinary trailing replacement is never matched here.
func leadingInsteadReplacement(tokens []shared.Token) (EffectReplacementSyntax, bool) {
	if len(tokens) > 1 && equalWord(tokens[0], "instead") {
		return EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[0].Span,
		}, true
	}
	for i := 1; i < len(tokens)-1; i++ {
		if tokens[i-1].Kind == shared.Comma && equalWord(tokens[i], "instead") {
			return EffectReplacementSyntax{
				Kind: EffectReplacementInstead,
				Span: tokens[i].Span,
			}, true
		}
	}
	return EffectReplacementSyntax{}, false
}

// parsePlusAdditionalReplacement recognizes the token-creation addend rider
// "... plus an additional <Type> token" (Xorn: "instead create those tokens plus
// an additional Treasure token"), under which the effect creates extra tokens of
// the same kind in addition to those it would create. The amount defaults to one
// additional token, or the explicit number in "plus N additional ... tokens".
func parsePlusAdditionalReplacement(tokens []shared.Token, atoms Atoms) (EffectReplacementSyntax, bool) {
	for i := range tokens {
		if !equalWord(tokens[i], "plus") || i+2 >= len(tokens) {
			continue
		}
		if equalWord(tokens[i+1], "an") && equalWord(tokens[i+2], "additional") {
			return EffectReplacementSyntax{
				Kind:   EffectReplacementPlusAdditional,
				Amount: 1,
				Span:   tokens[i].Span,
			}, true
		}
		if amount, ok := effectNumber(tokens[i+1], atoms); ok && equalWord(tokens[i+2], "additional") {
			return EffectReplacementSyntax{
				Kind:   EffectReplacementPlusAdditional,
				Amount: amount,
				Span:   tokens[i].Span,
			}, true
		}
	}
	return EffectReplacementSyntax{}, false
}

// trailingInsteadBeforeConditionReplacement recognizes the plain "instead"
// replacement marker that ends an effect clause whose trailing "if" condition
// has already been stripped from the ownership tokens ("That creature gets
// -13/-13 until end of turn instead if a creature died this turn.", Tragic
// Slip, whose ownership tokens end at "instead"). The "instead" word marks this
// effect as replacing the preceding effect; the stripped trailing condition
// gates when the replacement applies. It is distinguished from the final
// "... instead." form the caller handles next by requiring the clause to end at
// the bare "instead" word with no closing period.
func trailingInsteadBeforeConditionReplacement(tokens []shared.Token) (EffectReplacementSyntax, bool) {
	if len(tokens) == 0 || !equalWord(tokens[len(tokens)-1], "instead") {
		return EffectReplacementSyntax{}, false
	}
	return EffectReplacementSyntax{
		Kind: EffectReplacementInstead,
		Span: tokens[len(tokens)-1].Span,
	}, true
}

func parseEffectReplacement(tokens []shared.Token, atoms Atoms) EffectReplacementSyntax {
	if replacement, ok := parseInsteadOneOfEachReplacement(tokens); ok {
		return replacement
	}
	if replacement, ok := parsePlusAdditionalReplacement(tokens, atoms); ok {
		return replacement
	}
	if replacement, ok := leadingInsteadReplacement(tokens); ok {
		return replacement
	}
	if replacement, ok := trailingInsteadBeforeConditionReplacement(tokens); ok {
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
	twiceMuch := effectHasTokenWords(tokens, "twice", "that", "much")
	thatMuchPlus := effectHasTokenWords(tokens, "that", "much", "damage", "plus") ||
		effectHasTokenWords(tokens, "that", "much", "life", "plus")
	thatManyPlus := effectHasTokenWords(tokens, "that", "many", "plus")
	doubleThat := effectHasTokenWords(tokens, "double", "that", "damage") ||
		effectHasTokenWords(tokens, "twice", "that", "damage")
	tripleThat := effectHasTokenWords(tokens, "triple", "that", "damage")
	if boolCount(twiceMany, twiceMuch, thatMuchPlus, thatManyPlus, doubleThat, tripleThat) != 1 {
		return replacement
	}
	switch {
	case twiceMany:
		replacement.Kind = EffectReplacementTwiceThatMany
	case twiceMuch:
		replacement.Kind = EffectReplacementTwiceThatMuch
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
	case tripleThat:
		replacement.Kind = EffectReplacementTripleThat
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
	if len(body) == 8 &&
		effectWordsAt(body, 0, "one", "mana", "of", "any", "type", "that", "land", "produced") {
		return EffectManaSyntax{Span: shared.SpanOf(body), TriggerLandProducedType: true}
	}
	if len(body) == 5 && effectWordsAt(body, 0, "one", "mana", "of", "any", "color") {
		return EffectManaSyntax{Span: shared.SpanOf(body), AnyColor: true}
	}
	if len(body) == 6 && effectWordsAt(body, 1, "mana", "of", "any", "one", "color") {
		if count, ok := manaAnyOneColorCount(body[0]); ok {
			return EffectManaSyntax{Span: shared.SpanOf(body), AnyColor: true, AnyColorCount: count}
		}
	}
	if len(body) >= 7 && effectWordsAt(body, 1, "mana", "in", "any", "combination", "of") {
		if count, ok := manaAnyOneColorCount(body[0]); ok {
			if colors, ok := combinationManaColorList(body[6:]); ok {
				return EffectManaSyntax{
					Span:              shared.SpanOf(body),
					Combination:       true,
					CombinationColors: colors,
					CombinationCount:  count,
				}
			}
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
		counted, cok := countedSingleManaSymbols(loopBody)
		if !cok {
			return EffectManaSyntax{}
		}
		symbols, choice = counted, false
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

// countedSingleManaSymbols expands a counted single-symbol add-mana body
// ("Add six {R}", The Flux; "Add seven {R}", Irencrag Feat) into N copies of the
// lone mana symbol. It requires a leading cardinal count of two or more followed
// by exactly one mana symbol, so ordinary single- and multi-symbol bodies keep
// their own branch and any other shape fails closed.
func countedSingleManaSymbols(body []shared.Token) ([]string, bool) {
	if len(body) != 2 || body[1].Kind != shared.Symbol {
		return nil, false
	}
	count, ok := additionalLandCountWord(body[0])
	if !ok || count < 2 {
		return nil, false
	}
	symbols := make([]string, count)
	for i := range symbols {
		symbols[i] = body[1].Text
	}
	return symbols, true
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

// combinationManaColorList parses the color list of an "in any combination of
// <colors>" body. The single word "colors" denotes all five basic colors in
// WUBRG order (Manamorphose, Cascading Cataracts); otherwise the tokens are a
// list of basic color symbols joined by commas and/or "and"/"or"/"/"
// connectives ("{R} and/or {G}", Goblin Clearcutter). It requires two or more
// distinct basic colors and rejects colorless ({C}) and any adjacent symbols,
// so an unmodeled wording fails closed.
func combinationManaColorList(tokens []shared.Token) ([]mana.Color, bool) {
	if len(tokens) == 1 && equalWord(tokens[0], "colors") {
		return []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G}, true
	}
	var colors []mana.Color
	seen := make(map[mana.Color]bool, len(tokens))
	prevSymbol := false
	started := false
	for _, token := range tokens {
		switch {
		case token.Kind == shared.Symbol:
			if prevSymbol {
				return nil, false
			}
			color, ok := effectManaColor(token.Text)
			if !ok || color == mana.C || seen[color] {
				return nil, false
			}
			colors = append(colors, color)
			seen[color] = true
			prevSymbol = true
			started = true
		case token.Kind == shared.Comma, token.Kind == shared.Slash,
			equalWord(token, "and"), equalWord(token, "or"):
			if !started {
				return nil, false
			}
			prevSymbol = false
		default:
			return nil, false
		}
	}
	if !prevSymbol || len(colors) < 2 {
		return nil, false
	}
	return colors, true
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
		// A sentence-initial "Otherwise," introduces the else branch of the
		// preceding sentence's conditional effect.
		if len(tokens) > 1 && equalWord(tokens[0], "otherwise") && tokens[1].Kind == shared.Comma {
			return EffectConnectionOtherwise, tokens[0].Span
		}
		return EffectConnectionNone, shared.Span{}
	}
	for i := indices[effectIndex] - 1; i > indices[effectIndex-1]; i-- {
		switch {
		case equalWord(tokens[i], "then"):
			return EffectConnectionThen, tokens[i].Span
		case equalWord(tokens[i], "and"):
			return EffectConnectionAnd, tokens[i].Span
		case equalWord(tokens[i], "or"):
			return EffectConnectionOr, tokens[i].Span
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
	case len(words) >= 3 && youAndPrefix && words[0] == "that" && words[1] == "player" &&
		words[len(words)-1] == "each":
		// "You and that player each <verb>": the controller and the triggering
		// event's player ("that player") both receive the effect. It mirrors the
		// controller-and-target case but for an event-bound recipient.
		return EffectContextControllerAndReferencedPlayer
	case effectContainsWords(words, "target"):
		return EffectContextTarget
	case len(words) >= 2 && words[len(words)-2] == "that" && words[len(words)-1] == "player":
		return EffectContextReferencedPlayer
	case len(words) >= 2 && words[len(words)-2] == "defending" && words[len(words)-1] == "player":
		return EffectContextDefendingPlayer
	case words[len(words)-1] == "controller" && subjectReferencesObject(subject, atoms):
		return EffectContextReferencedObjectController
	case words[len(words)-1] == "owner" && subjectReferencesObject(subject, atoms):
		return EffectContextReferencedObjectOwner
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
		boundary := tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period ||
			tokens[i].Kind == shared.Semicolon || equalWord(tokens[i], "then") || equalWord(tokens[i], "and")
		// A clause-leading "also" ("..., also create a token") is an additive
		// adverb that carries no subject; skip it so the controller subject and
		// exact verb coverage are recognized. A non-leading "also" (e.g.
		// "creatures you control also gain first strike") follows a real subject
		// and must be retained. Similarly, a clause-leading "instead" ("...,
		// instead it gets +3/+3") is a replacement marker that precedes the
		// subject and must be skipped for exact subject recognition.
		leadingAlso := (equalWord(tokens[i], "also") || equalWord(tokens[i], "instead")) && i == start
		if boundary || leadingAlso {
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
		var costStart int
		switch {
		case effectWordsAt(tokens, i, "unless", "its", "controller", "pays"):
			payer = EffectPaymentPayerTargetController
			costStart = i + 4
		case effectWordsAt(tokens, i, "unless", "that", "player", "pays"):
			payer = EffectPaymentPayerEventPlayer
			costStart = i + 4
		case effectWordsAt(tokens, i, "unless", "you", "pay"):
			payer = EffectPaymentPayerController
			costStart = i + 3
		default:
			continue
		}
		manaCost, end, ok := parseKeywordManaCost(tokens, costStart)
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
		// "{N} for each <count subject>" repeats the fixed generic payment per
		// counted object ("pays {1} for each card in your graveyard.", Circular
		// Logic). The fixed mana cost stays in ManaCost and the trailing for-each
		// count rides in GenericManaAmount; lowering repeats the cost by the count.
		case fixedGenericManaCost(manaCost) &&
			effectWordsAt(tokens, end, "for", "each") &&
			tokens[len(tokens)-1].Kind == shared.Period:
			amount, attempted, amountOK := parseDynamicEffectAmount(tokens[end:len(tokens)-1], atoms)
			if !attempted || !amountOK ||
				amount.DynamicForm != EffectDynamicAmountFormForEach ||
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

// fixedGenericManaCost reports whether a parsed payment mana cost is a single
// fixed generic symbol such as {1}. It backs the "{N} for each <subject>"
// resolution-payment form, where the generic cost is repeated per counted object.
func fixedGenericManaCost(manaCost cost.Mana) bool {
	return len(manaCost) == 1 && manaCost[0].Kind == cost.GenericSymbol
}

// classifiedVerb is one candidate effect verb produced by the single
// effect-classification pass: its token index, the authoritative effect kind
// from effectKindAt, and whether it sits inside a leading if/unless condition
// clause. Both the real effect segmentation (effectIndices) and the
// ordered-lowering count (orderedEffectCount) derive from the same records so a
// sentence's parsed effect list and its ordered-lowering metadata cannot
// classify a verb differently.
type classifiedVerb struct {
	Index           int
	Kind            EffectKind
	WithinCondition bool
}

// classifyEffectVerbs returns one classifiedVerb per token that the authoritative
// effectKindAt classifier recognizes as an effect verb, after applying the
// exclusions shared by every consumer: a self-name reference, the inner "untap"
// of a "tap or untap" choice, and a copy-token "except" rider boundary. The
// per-consumer exclusions (the noun-form "next untap step" for segmentation, the
// leading-condition membership for the ordered count) are left to the callers so
// this single pass owns all verb-kind overrides in one place.
func classifyEffectVerbs(tokens []shared.Token, atoms Atoms) []classifiedVerb {
	var result []classifiedVerb
	for i := range tokens {
		kind := effectKindAt(tokens, i)
		if kind == EffectUnknown ||
			atoms.SelfNameAt(tokens[i].Span) ||
			tapOrUntapInnerUntapAt(tokens, i) ||
			withinAsThoughDidntHaveDefenderTail(tokens, i) ||
			untilBecomesMonarchBoundaryAt(tokens, i) ||
			copyTokenExceptRiderBoundaryAt(tokens, i) {
			continue
		}
		result = append(result, classifiedVerb{
			Index:           i,
			Kind:            kind,
			WithinCondition: effectWithinCondition(tokens, i),
		})
	}
	return result
}

func effectIndices(tokens []shared.Token, atoms Atoms) []int {
	var result []int
	for _, verb := range classifyEffectVerbs(tokens, atoms) {
		if !effectNounAt(tokens, verb.Index) {
			result = append(result, verb.Index)
		}
	}
	return result
}

// orderedEffectCount returns the number of effect verbs that make a sentence
// drive the ordered-lowering path. It derives from the same classifyEffectVerbs
// pass as the real effect segmentation, excluding verbs inside a leading
// condition clause (an "if"/"unless" guard is not a sequenced effect) and
// collapsing a mass reanimation/exchange to a single effect.
func orderedEffectCount(tokens []shared.Token, atoms Atoms) int {
	if _, ok := massReanimationExchangeWords(tokens); ok {
		return 1
	}
	count := 0
	for _, verb := range classifyEffectVerbs(tokens, atoms) {
		if !verb.WithinCondition {
			count++
		}
	}
	return count
}

// untilBecomesMonarchBoundaryAt reports whether the effect-boundary verb at index
// is the "becomes"/"become" of an "until <player> becomes the monarch" duration
// clause ("exile <target> until an opponent becomes the monarch.", Palace
// Jailer). That trailing clause names the exile's return condition, not a second
// effect, so it must not split the sentence into an exile and a stranded
// become-monarch sibling.
func untilBecomesMonarchBoundaryAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "becomes") && !equalWord(tokens[index], "become") {
		return false
	}
	if index+2 >= len(tokens) ||
		!equalWord(tokens[index+1], "the") ||
		!equalWord(tokens[index+2], "monarch") {
		return false
	}
	for i := range index {
		if equalWord(tokens[i], "until") {
			return true
		}
	}
	return false
}

// copyTokenExceptRiderBoundaryAt reports whether the effect-boundary verb at
// index lies inside a copy-token "Create ... a copy of <source>, except <rider>"
// clause, where the rider modifies the created copy rather than starting a new
// effect. A keyword-grant rider verb ("the token has flying", Irenicus's Vile
// Duplication) would otherwise split the rider into a stranded EffectGrantKeyword
// sibling; keeping it inside the create clause lets the copy-token recognizer
// fold the copiable rider into the create (or fail closed for an unrecognized
// rider). The guard requires a copy-token create head ("Create ... a copy of")
// before a ", except" that precedes the verb, so only copy-token create riders
// are affected; every such card with a verb rider is multi-effect unsupported
// today, so folding it strands no existing output.
func copyTokenExceptRiderBoundaryAt(tokens []shared.Token, index int) bool {
	if index == 0 {
		return false
	}
	exceptIndex := -1
	for i := 1; i < index; i++ {
		if equalWord(tokens[i], "except") && tokens[i-1].Kind == shared.Comma {
			exceptIndex = i
		}
	}
	if exceptIndex < 0 {
		return false
	}
	return createCopyTokenHead(tokens[:exceptIndex])
}

// createCopyTokenHead reports whether the head tokens begin a copy-token creation
// clause: a leading "Create" verb followed by an "a copy of" phrase. It anchors
// the copy-token "except" rider guard so unrelated "Create" clauses without a
// copy-of phrase, and non-creation clauses, are never affected.
func createCopyTokenHead(head []shared.Token) bool {
	if len(head) == 0 || !equalWord(head[0], "create") {
		return false
	}
	for i := 1; i+1 < len(head); i++ {
		if equalWord(head[i], "copy") && equalWord(head[i-1], "a") && equalWord(head[i+1], "of") {
			return true
		}
	}
	return false
}

// tapOrUntapInnerUntapAt reports whether the "untap" at index is the second verb
// of a "tap or untap" choice ("Tap or untap target creature."), so it is not a
// separate untap effect. The "tap or untap" phrase lowers to one TapOrUntap
// instruction anchored on the leading "tap" verb.
func tapOrUntapInnerUntapAt(tokens []shared.Token, index int) bool {
	return index >= 2 &&
		equalWord(tokens[index], "untap") &&
		equalWord(tokens[index-1], "or") &&
		(equalWord(tokens[index-2], "tap") || equalWord(tokens[index-2], "taps"))
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
		(equalWord(tokens[index+4], "turn") || equalWord(tokens[index+4], "combat"))
}

// canAttackAsThoughDefenderVerbAt reports whether the temporary combat
// permission "can attack this turn as though it didn't have defender" begins at
// index. It anchors the resolving EffectCanAttackAsThoughDefender effect ("This
// creature can attack this turn as though it didn't have defender.", Krotiq
// Nestguard, Skyclave Squid) on the "can" so the subject is the source creature
// scanned before it. The exactness recognizer reconstructs the full clause, so
// any other wording (a different subject, a missing "this turn" duration, an
// added rider) still fails closed.
func canAttackAsThoughDefenderVerbAt(tokens []shared.Token, index int) bool {
	return index+9 < len(tokens) &&
		equalWord(tokens[index], "can") &&
		equalWord(tokens[index+1], "attack") &&
		equalWord(tokens[index+2], "this") &&
		equalWord(tokens[index+3], "turn") &&
		equalWord(tokens[index+4], "as") &&
		equalWord(tokens[index+5], "though") &&
		equalWord(tokens[index+6], "it") &&
		equalWord(tokens[index+7], "didn't") &&
		equalWord(tokens[index+8], "have") &&
		equalWord(tokens[index+9], "defender")
}

// asThoughDidntHaveDefenderTailAt reports whether the "as though it didn't have
// defender" reminder tail of the attack permission begins at index. It anchors
// the span removed from semantic reference and keyword scanning and the
// suppression of the "have" keyword-grant verb inside that tail.
func asThoughDidntHaveDefenderTailAt(tokens []shared.Token, index int) bool {
	return index+5 < len(tokens) &&
		equalWord(tokens[index], "as") &&
		equalWord(tokens[index+1], "though") &&
		equalWord(tokens[index+2], "it") &&
		equalWord(tokens[index+3], "didn't") &&
		equalWord(tokens[index+4], "have") &&
		equalWord(tokens[index+5], "defender")
}

// withinAsThoughDidntHaveDefenderTail reports whether the token at index lies
// inside an "as though it didn't have defender" tail, so the "have" within it is
// not segmented as a separate keyword-grant verb of the attack permission.
func withinAsThoughDidntHaveDefenderTail(tokens []shared.Token, index int) bool {
	for start := 0; start <= index; start++ {
		if asThoughDidntHaveDefenderTailAt(tokens, start) && index <= start+5 {
			return true
		}
	}
	return false
}

// cantBeSacrificedThisTurnVerbAt reports whether the temporary prohibition
// "can't be sacrificed this turn" / "cannot be sacrificed this turn" begins at
// index. It anchors the temporary can't-be-sacrificed resolving effect ("... it
// can't be sacrificed this turn.", Slicer, Hired Muscle) on the negated
// "can't"/"cannot" so the subject is the permanent scanned before it. The "this
// turn" tail distinguishes this resolving, until-end-of-turn effect from the
// continuous static prohibition ("Creatures you control but don't own ... can't
// be sacrificed.", Garland, Royal Kidnapper) that carries no turn duration, so
// that keeps flowing through the static-declaration path. The exactness
// recognizer reconstructs the full clause, so any other wording still fails
// closed.
func cantBeSacrificedThisTurnVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		index+4 < len(tokens) &&
		equalWord(tokens[index+1], "be") &&
		equalWord(tokens[index+2], "sacrificed") &&
		equalWord(tokens[index+3], "this") &&
		equalWord(tokens[index+4], "turn")
}

// cantBlockThisTurnVerbAt reports whether the temporary prohibition "can't block
// this turn" / "cannot block this turn" begins at index. It anchors the
// temporary can't-block resolving effect ("Target creature can't block this
// turn.", "Up to three target creatures can't block this turn.") on the negated
// "can't"/"cannot" so the subject is the targeted creature(s). The "this turn"
// tail distinguishes this resolving, until-end-of-turn effect from the
// continuous static prohibitions ("Creatures can't block.", "Creatures with
// power less than this creature's power can't block it.") that carry no turn
// duration, so those keep flowing through the static-declaration path. The
// exactness recognizer reconstructs the full clause, so any other wording (a
// "can't block creatures you control" qualifier, an "except" rider) still fails
// closed.
func cantBlockThisTurnVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		index+3 < len(tokens) &&
		equalWord(tokens[index+1], "block") &&
		equalWord(tokens[index+2], "this") &&
		equalWord(tokens[index+3], "turn")
}

// cantAttackThisTurnVerbAt reports whether the temporary prohibition "can't
// attack this turn" / "cannot attack this turn" begins at index. It anchors the
// temporary can't-attack resolving effect ("Target creature can't attack this
// turn.") on the negated "can't"/"cannot" so the subject is the targeted
// creature. Like the sibling can't-block recognizer, the trailing "this turn"
// distinguishes this resolving, until-end-of-turn effect from the continuous
// static "can't attack" prohibitions, which keep flowing through the
// static-declaration path. The combined "can't attack or block this turn" form
// is recognized separately by cantAttackOrBlockThisTurnVerbAt, so this matcher
// excludes it.
func cantAttackThisTurnVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		index+3 < len(tokens) &&
		equalWord(tokens[index+1], "attack") &&
		equalWord(tokens[index+2], "this") &&
		equalWord(tokens[index+3], "turn")
}

// cantAttackOrBlockThisTurnVerbAt reports whether the combined temporary
// prohibition "can't attack or block this turn" / "cannot attack or block this
// turn" begins at index. It anchors the temporary can't-attack-or-block
// resolving effect ("Target creature can't attack or block this turn.",
// Thundersong Trumpeter) on the negated "can't"/"cannot". The fixed "attack or
// block this turn" tail distinguishes this resolving effect from the continuous
// static prohibition that carries no turn duration.
func cantAttackOrBlockThisTurnVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		index+5 < len(tokens) &&
		equalWord(tokens[index+1], "attack") &&
		equalWord(tokens[index+2], "or") &&
		equalWord(tokens[index+3], "block") &&
		equalWord(tokens[index+4], "this") &&
		equalWord(tokens[index+5], "turn")
}

// mustAttackTargetVerbAt reports whether the single-target forced-attack verb
// "attacks this turn if able" begins at index. It anchors the temporary
// single-target must-attack resolving effect ("Target creature attacks this
// turn if able.", Kookus, Norritt) on the "attacks" verb so the subject is the
// targeted creature. The fixed "this turn if able" tail distinguishes the
// one-shot, until-end-of-turn requirement from the continuous static "attacks
// each combat if able" rule and the directed "attacks <player> this turn if
// able" form, both of which fail closed here.
func mustAttackTargetVerbAt(tokens []shared.Token, index int) bool {
	return equalWord(tokens[index], "attacks") &&
		index+4 < len(tokens) &&
		equalWord(tokens[index+1], "this") &&
		equalWord(tokens[index+2], "turn") &&
		equalWord(tokens[index+3], "if") &&
		equalWord(tokens[index+4], "able")
}

// negatedNextUntapStepVerbAt reports whether the token at index begins the
// standalone stun predicate "doesn't/don't untap during ... next untap step"
// that follows a leading target subject ("Target creature doesn't untap during
// its controller's next untap step."). Unlike a forward effect verb, the negated
// "doesn't" contraction is not itself an effect word, so target scanning would
// otherwise absorb it into the target noun phrase. Breaking here keeps the target
// ("Target creature") clean so it reconstructs exactly.
func negatedNextUntapStepVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "doesn't") || equalWord(tokens[index], "don't")) &&
		index+1 < len(tokens) &&
		equalWord(tokens[index+1], "untap")
}

// phasesOutVerbAt reports whether the token at index begins the subject-final
// phase-out verb "phases out" (singular subject) or "phase out" (plural subject)
// that closes a leading target subject ("Target creature phases out.", "Any
// number of target nonland permanents you control phase out."). Neither "phase"
// nor "phases" is a forward effect word, so target scanning would otherwise
// absorb the verb into the target noun phrase and corrupt the selection. Breaking
// here keeps the target phrase clean so it reconstructs exactly. A preceding
// "can't"/"cannot" marks the restriction "... can't phase out" rather than the
// phase-out action, so it is left for that clause to consume.
func phasesOutVerbAt(tokens []shared.Token, index int) bool {
	if index > 0 && (equalWord(tokens[index-1], "can't") || equalWord(tokens[index-1], "cannot")) {
		return false
	}
	return (equalWord(tokens[index], "phases") || equalWord(tokens[index], "phase")) &&
		index+1 < len(tokens) &&
		equalWord(tokens[index+1], "out")
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
	if effectWordsAt(tokens, index+1, "this", "turn") {
		return true
	}
	return effectWordsAt(tokens, index+1, "your", "commander", "from", "the", "command", "zone", "this", "game")
}

// castDuringMainPhaseConditionAt reports whether the "cast" verb at index begins
// the Addendum cast-timing condition "you cast this spell during your main
// phase". That phrase is a condition predicate, not a cast effect, so the
// effect classifier must not treat it as an EffectCast.
func castDuringMainPhaseConditionAt(tokens []shared.Token, index int) bool {
	if index == 0 || !equalWord(tokens[index-1], "you") {
		return false
	}
	return effectWordsAt(tokens, index, "cast", "this", "spell", "during", "your", "main", "phase")
}

// castSpellsFromLibraryTopAt reports whether the "cast" verb at index begins the
// cast-from-library-top static permission "cast [<types>] spells [of the chosen
// type] from the top of your library" (Future Sight, Bolas's Citadel,
// Realmwalker). That phrase is a continuous player-rule static, not a cast
// effect, so the effect classifier must not treat it as an EffectCast and let the
// static-declaration path recognize it.
func castSpellsFromLibraryTopAt(tokens []shared.Token, index int) bool {
	i := index + 1
	matchedSpells := false
	for i < len(tokens) {
		if equalWord(tokens[i], "from") || equalWord(tokens[i], "of") {
			break
		}
		switch {
		case equalWord(tokens[i], "spells"):
			matchedSpells = true
		case equalWord(tokens[i], "and") || equalWord(tokens[i], "or"):
		case tokens[i].Kind == shared.Comma:
		case equalWord(tokens[i], "colorless"):
		default:
			if _, ok := recognizeCardTypeWord(tokens[i].Text); !ok {
				return false
			}
		}
		i++
	}
	if !matchedSpells {
		return false
	}
	if effectWordsAt(tokens, i, "of", "the", "chosen", "type") {
		i += 4
	}
	return effectWordsAt(tokens, i, "from", "the", "top", "of", "your", "library")
}

// castThisFromGraveyardAt reports whether the "cast" verb at index begins the
// self-scoped cast-from-graveyard static permission "cast this card from your
// graveyard" (Gravecrawler, Hogaak). That phrase is a continuous player-rule
// static, not a cast effect, so the effect classifier must not treat it as an
// EffectCast and let the static-declaration path recognize it.
func castThisFromGraveyardAt(tokens []shared.Token, index int) bool {
	return effectWordsAt(tokens, index, "cast", "this", "card", "from", "your", "graveyard")
}

// manaSpentToCastPhraseAt reports whether the "cast" verb at index is the
// infinitive inside the Converge count phrase "mana spent to cast it" rather
// than a casting effect. The Converge dynamic amount ("for each color of mana
// spent to cast it") consumes that span as a count, so the bare "cast" must not
// also seed a separate cast effect that would split the enters-with-counters
// sentence.
func manaSpentToCastPhraseAt(tokens []shared.Token, index int) bool {
	return index >= 3 &&
		equalWord(tokens[index-3], "mana") &&
		equalWord(tokens[index-2], "spent") &&
		equalWord(tokens[index-1], "to") &&
		effectWordsAt(tokens, index, "cast", "it")
}

// manaWasSpentToCastSpellPhraseAt reports whether the "cast" verb at index is
// the infinitive inside the Adamant condition phrase "mana was spent to cast
// this spell" rather than a casting effect. The Adamant condition ("if at least
// three <color> mana was spent to cast this spell") leads an
// enters-with-counters sentence, so the bare "cast" must not seed a separate
// cast effect that would split the sentence into two effects and defeat the
// enters-with-counters recognizer.
func manaWasSpentToCastSpellPhraseAt(tokens []shared.Token, index int) bool {
	return index >= 3 &&
		equalWord(tokens[index-3], "was") &&
		equalWord(tokens[index-2], "spent") &&
		equalWord(tokens[index-1], "to") &&
		effectWordsAt(tokens, index, "cast", "this", "spell")
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
			if equalWord(tokens[i], "then") || equalWord(tokens[i], "and") || equalWord(tokens[i], "or") {
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

// resolvingAttachedPossessiveSubject reports whether a resolving effect's
// possessive subject is the source's attached permanent named by "equipped
// creature's" or "enchanted creature's" ("Double equipped creature's power…",
// Junk Jet). The leading non-possessive "equipped/enchanted creature" subject is
// parsed as a StaticSubject and handled by the static-group path, so this is
// gated on there being no static subject: only the possessive object form, which
// leaves no static subject, routes to the source-attached permanent.
func resolvingAttachedPossessiveSubject(ownership []shared.Token, staticSubject EffectStaticSubjectSyntax) bool {
	if staticSubject.Kind != EffectStaticSubjectNone {
		return false
	}
	for i := 0; i+1 < len(ownership); i++ {
		if (equalWord(ownership[i], "equipped") || equalWord(ownership[i], "enchanted")) &&
			strings.EqualFold(ownership[i+1].Text, "creature's") {
			return true
		}
	}
	return false
}

// loseAllAbilitiesObject reports whether a lose effect's grammatical object is
// the total "all abilities" removal ("<subject> loses all abilities until end of
// turn"), rather than life, the game, or a named keyword. The parser strips a
// quoted ability class ("loses all \"bands with other\" abilities", Shelkin
// Brownie) out of the clause tokens, leaving them identical to the total form, so
// the raw sentence text is checked for the contiguous "lose(s) all abilities"
// phrase: a specific-ability-class removal keeps its quoted class between "all"
// and "abilities" and so is not mistaken for total ability removal.
func loseAllAbilitiesObject(kind EffectKind, sentenceText string) bool {
	if kind != EffectLose {
		return false
	}
	lower := strings.ToLower(sentenceText)
	return strings.Contains(lower, "loses all abilities") ||
		strings.Contains(lower, "lose all abilities")
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

// payLifeVerbAt reports whether the "pay"/"pays" verb at index governs a bare
// "pay N life" life payment ("Pay 2 life."). Paying life is losing that much
// life (CR 119.1b), so this anchors the generic "pay" verb to EffectLose. The
// classification is confirmed by scanning forward to the clause terminator for a
// top-level "life" word outside any quoted granted ability, with no mana Symbol
// in the clause: a combined "pay {mana} and N life" cost carries a mana symbol
// and is folded by the optional-payment recognizers, not treated as a resolving
// life-loss effect.
func payLifeVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "pay") && !equalWord(tokens[index], "pays") {
		return false
	}
	// A preceding top-level "enters"/"enter" marks the "As this <permanent>
	// enters, you may pay N life. If you don't, it enters tapped." entry-payment
	// replacement (the dual-land cycle), where the life amount is folded onto the
	// leading enters effect rather than parsed as a resolving life loss. Leave
	// that shape to the optional-entry-payment recognizer.
	for i := range index {
		if tokens[i].Kind == shared.Word &&
			(equalWord(tokens[i], "enters") || equalWord(tokens[i], "enter")) {
			return false
		}
	}
	quoted := false
	for i := index + 1; i < len(tokens); i++ {
		switch tokens[i].Kind {
		case shared.Period, shared.Semicolon, shared.Comma, shared.Symbol:
			return false
		case shared.Quote:
			quoted = !quoted
		case shared.Word:
			if !quoted && equalWord(tokens[i], "life") {
				return true
			}
		default:
		}
	}
	return false
}

// manifestDreadClauseBoundary reports whether the token following "manifest
// dread" ends that keyword-action clause: end of the token run, a sentence
// terminator, or a clause separator such as the comma before "then put ...
// counters on that creature" (Weight Room). "manifest dread" is a fixed
// keyword action with no "manifest dread <noun>" phrasing, so recognizing it
// before a continuation lets a following clause reference the manifested
// creature instead of misreading the verb as a plain "manifest the top card".
func manifestDreadClauseBoundary(tokens []shared.Token, index int) bool {
	if index >= len(tokens) {
		return true
	}
	switch tokens[index].Kind {
	case shared.Period, shared.Comma, shared.Semicolon:
		return true
	default:
		return equalWord(tokens[index], "then")
	}
}

func effectKindAt(tokens []shared.Token, index int) EffectKind {
	kind := effectWordKind(tokens[index])
	switch {
	case equalWord(tokens[index], "manifest") || equalWord(tokens[index], "manifests"):
		switch {
		case effectWordsAt(tokens, index+1, "dread") && manifestDreadClauseBoundary(tokens, index+2):
			return EffectManifestDread
		case effectWordsAt(tokens, index+1, "the", "top", "card", "of", "your", "library") &&
			len(tokens) == index+8 && tokens[index+7].Kind == shared.Period:
			return EffectManifest
		default:
			return EffectManifest
		}
	case equalWord(tokens[index], "cloak") || equalWord(tokens[index], "cloaks"):
		return EffectCloak
	case equalWord(tokens[index], "look"):
		if digLookInstruction(tokens[index:]) {
			return EffectDig
		}
		if lookAtTopCardAnyTimeInstruction(tokens[index:]) {
			return EffectUnknown
		}
		if lookAtLibraryTopInstruction(tokens[index:]) {
			return EffectLookAtLibraryTop
		}
		if lookAtHandInstruction(tokens[index:]) {
			return EffectLookAtHand
		}
		return EffectManifestDread
	case equalWord(tokens[index], "win") || equalWord(tokens[index], "wins"):
		if winGameVerbAt(tokens, index) {
			return EffectWinGame
		}
		return EffectUnknown
	case payLifeVerbAt(tokens, index):
		return EffectLose
	case canAttackAsThoughDefenderVerbAt(tokens, index):
		return EffectCanAttackAsThoughDefender
	case cantBeBlockedThisTurnVerbAt(tokens, index):
		return EffectCantBeBlocked
	case cantBeSacrificedThisTurnVerbAt(tokens, index):
		return EffectCantBeSacrificed
	case cantBlockThisTurnVerbAt(tokens, index):
		return EffectCantBlock
	case cantAttackOrBlockThisTurnVerbAt(tokens, index):
		return EffectCantAttackOrBlock
	case cantAttackThisTurnVerbAt(tokens, index):
		return EffectCantAttack
	case mustAttackTargetVerbAt(tokens, index):
		return EffectMustAttack
	case kind == EffectGrantKeyword && index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you"):
		return EffectUnknown
	case kind == EffectGrantKeyword && playerPossessionVerb(tokens, index):
		return EffectUnknown
	case kind == EffectGrantKeyword && counterPossessionVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectGrantKeyword && totalPowerPossessionVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared"):
		return EffectEnterPrepared
	case kind == EffectCast && index > 0 && (equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")):
		return EffectUnknown
	case kind == EffectCast && pastCastCountPhraseAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && castDuringMainPhaseConditionAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && castSpellsFromLibraryTopAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && castThisFromGraveyardAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && manaSpentToCastPhraseAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && manaWasSpentToCastSpellPhraseAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && spellCostModifierCastAt(tokens, index):
		// A resolving spell-cost-modifier sentence carries two "cast" tokens
		// ("spells you cast ... cost {N} less to cast"); its dedicated recognizer
		// produces a single effect, so neither the effect segmentation nor the
		// ordered-lowering count may treat the casts as separate effects.
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectCopyStackObject && !copyVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectTransform && index > 0 &&
		(equalWord(tokens[index-1], "can't") || equalWord(tokens[index-1], "cannot")):
		// "<subject> can't transform." is a continuous transform prohibition, not
		// a resolving transform effect; it is owned by the static-rule
		// declaration path, so it carries no resolving effect.
		return EffectUnknown
	case chooseNewTargetsVerbAt(tokens, index):
		return EffectChooseNewTargets
	case chooseCreatureTypeVerbAt(tokens, index):
		return EffectChooseCreatureType
	case kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control"):
		return EffectGainControl
	case kind == EffectGain && everyCreatureTypeGainRiderAt(tokens, index) && priorBasePowerToughnessSet(tokens, index):
		// "gain all/every creature type(s)" folded onto a base power/toughness set
		// (Mirror Entity) is a rider on that set, not a standalone effect, so it
		// is suppressed from both segmentation and the ordered-lowering count.
		return EffectUnknown
	case kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike"):
		return EffectUnknown
	case kind == EffectGrantKeyword && priorPTChange(tokens, index):
		return EffectUnknown
	case kind == EffectGrantKeyword && effectWordsAt(tokens, index+1, "the", "same", "name"):
		return EffectUnknown
	case kind == EffectModifyPT && playerCounterGainVerbAt(tokens, index):
		return EffectGainPlayerCounter
	case becomeMonarchVerbAt(tokens, index):
		return EffectBecomeMonarch
	case cantBecomeMonarchVerbAt(tokens, index):
		return EffectCantBecomeMonarch
	case kind == EffectTap && index+2 < len(tokens) &&
		equalWord(tokens[index+1], "or") && equalWord(tokens[index+2], "untap"):
		return EffectTapOrUntap
	case removeFromCombatVerbAt(tokens, index):
		return EffectRemoveFromCombat
	case removeCounterVerbAt(tokens, index):
		return EffectRemoveCounter
	case ellipticalOrRemoveCounterAt(tokens, index):
		return EffectRemoveCounter
	default:
		return kind
	}
}

// removeCounterVerbAt reports whether the verb at index begins the resolving
// effect "Remove <amount> [<kind> ]counter(s) from <object>." (Ferropede,
// "Whenever this creature deals combat damage to a player, you may remove a
// counter from target permanent."). The verb "remove" is otherwise used only in
// counter-removal costs (parsed before the colon) and the "Remove ... from
// combat" effect, so the classification is anchored on the "remove"/"removes"
// verb followed by a "counter"/"counters" word and a "from" clause that is not
// "from combat". The exact-syntax matcher reconstructs the supported single
// recognized-target forms; richer shapes (mass "all counters", dynamic counts)
// stay non-exact and fail closed.
func removeCounterVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "remove") && !equalWord(tokens[index], "removes") {
		return false
	}
	if removeFromCombatClauseStartsAt(tokens, index+1) >= 0 {
		return false
	}
	sawCounter := false
	for i := index + 1; i < len(tokens); i++ {
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			sawCounter = true
		}
	}
	if !sawCounter {
		return false
	}
	for i := index + 1; i+1 < len(tokens); i++ {
		if equalWord(tokens[i], "from") {
			return true
		}
	}
	return false
}

// ellipticalOrRemoveCounterAt reports whether the "remove" verb at index begins
// the kind-elided counter removal alternative "...or remove <amount> from
// <it/them>." that follows a counter-placement alternative in the same sentence
// ("Put a lore counter on target Saga you control or remove one from it.",
// Sigurd, Jarl of Ravensthorpe; "...put a charge counter on it or remove one
// from it.", Immard). The placed counter's noun is elided after "remove", so the
// ordinary removeCounterVerbAt — which anchors on an explicit "counter" noun —
// does not classify it. The verb is recognized only as the second arm of an
// "or" choice whose first arm already named a "counter", and only when a "from"
// clause (not "from combat") follows without its own "counter" noun, so a
// removal that does spell out its counter noun keeps flowing through
// removeCounterVerbAt and no unrelated "remove" wording is captured.
func ellipticalOrRemoveCounterAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "remove") && !equalWord(tokens[index], "removes") {
		return false
	}
	if index == 0 || !equalWord(tokens[index-1], "or") {
		return false
	}
	priorCounter := false
	for i := range index {
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			priorCounter = true
			break
		}
	}
	if !priorCounter {
		return false
	}
	if removeFromCombatClauseStartsAt(tokens, index+1) >= 0 {
		return false
	}
	fromIndex := -1
	for i := index + 1; i+1 < len(tokens); i++ {
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			return false
		}
		if equalWord(tokens[i], "from") {
			fromIndex = i
			break
		}
	}
	return fromIndex >= 0
}

// removeFromCombatVerbAt reports whether the verb at index begins the resolving
// effect "Remove <object> from combat." (Reconnaissance, "Remove target
// attacking creature you control from combat."). The verb "remove" is otherwise
// used only in counter-removal costs, so the classification is anchored on the
// "remove"/"removes" verb followed later by the "from combat" clause.
func removeFromCombatVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "remove") && !equalWord(tokens[index], "removes") {
		return false
	}
	return removeFromCombatClauseStartsAt(tokens, index+1) >= 0
}

// removeFromCombatClauseStartsAt returns the index of the "from" token of a
// "from combat" clause at or after start, or -1 if none precedes the sentence
// end. It anchors both the "Remove ... from combat" verb classification and the
// target-boundary scan that keeps "from combat" out of the target noun phrase.
func removeFromCombatClauseStartsAt(tokens []shared.Token, start int) int {
	for i := start; i+1 < len(tokens); i++ {
		if equalWord(tokens[i], "from") && equalWord(tokens[i+1], "combat") {
			return i
		}
	}
	return -1
}

// playerCounterGainVerbAt reports whether the "get"/"gets" verb at index is
// followed by a player-counter object — energy symbols ("You get {E}{E}.") or a
// named player counter ("You get an experience counter.") — rather than a
// power/toughness modification ("gets +1/+1"). The recipient and exact count are
// resolved separately; this only distinguishes the verb's object.
func playerCounterGainVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "get") && !equalWord(tokens[index], "gets") {
		return false
	}
	return len(energySymbolsAfter(tokens, index+1)) > 0 ||
		playerCounterWordAfter(tokens, index+1)
}

// becomeMonarchVerbAt reports whether the "become"/"becomes" verb at index heads
// a "<subject> become(s) the monarch" designation effect (CR 720). The object is
// the fixed "the monarch" noun phrase that ends the sentence; any other object
// leaves the verb unclassified so unrelated "becomes" wordings ("becomes a
// copy", "becomes an artifact") keep their own whole-sentence recognizers.
func becomeMonarchVerbAt(tokens []shared.Token, index int) bool {
	if !equalWord(tokens[index], "become") && !equalWord(tokens[index], "becomes") {
		return false
	}
	return index+3 < len(tokens) &&
		effectWordsAt(tokens, index+1, "the", "monarch") &&
		tokens[index+3].Kind == shared.Period
}

// cantBecomeMonarchVerbAt reports whether the temporary prohibition "can't
// become the monarch this turn." begins at index (Jared Carthalion). It anchors
// on the negated "can't"/"cannot" so lowering blocks the controller from
// becoming the monarch for the rest of the turn.
func cantBecomeMonarchVerbAt(tokens []shared.Token, index int) bool {
	return (equalWord(tokens[index], "can't") || equalWord(tokens[index], "cannot")) &&
		effectWordsAt(tokens, index+1, "become", "the", "monarch", "this", "turn") &&
		index+6 < len(tokens) &&
		tokens[index+6].Kind == shared.Period
}

// playerCounterWordAfter reports whether the tokens beginning at start name a
// player-only counter kind immediately followed by the "counter"/"counters"
// noun ("an experience counter", "two poison counters"). The kind word and count
// are recognized later from counter atoms and the effect amount; this only gates
// classification so a "gets +1/+1" P/T change never matches.
func playerCounterWordAfter(tokens []shared.Token, start int) bool {
	for i := start; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "experience") && !equalWord(tokens[i], "poison") {
			continue
		}
		if equalWord(tokens[i+1], "counter") || equalWord(tokens[i+1], "counters") {
			return true
		}
	}
	return false
}

// energySymbolsAfter returns the run of consecutive energy ({E}) symbol tokens
// beginning at start, stopping at the first non-energy token. It is empty when
// the run is interrupted before any energy symbol, so a "gets +1/+1" power
// modification never matches.
func energySymbolsAfter(tokens []shared.Token, start int) []shared.Token {
	end := start
	for end < len(tokens) && tokens[end].Kind == shared.Symbol && strings.EqualFold(tokens[end].Text, "{E}") {
		end++
	}
	if end == start {
		return nil
	}
	return tokens[start:end]
}

func effectWordKind(token shared.Token) EffectKind {
	if token.Kind != shared.Word {
		return EffectUnknown
	}
	switch strings.ToLower(token.Text) {
	case "add", "adds":
		return EffectAddMana
	case "amass":
		return EffectAmass
	case "bolster", "bolsters":
		return EffectBolster
	case "renown":
		return EffectRenown
	case "monstrosity":
		return EffectMonstrosity
	case "adapt", "adapts":
		return EffectAdapt
	case "attach", "attaches":
		return EffectAttach
	case "cast", "casts":
		return EffectCast
	case "counter", "counters":
		return EffectCounter
	case "copy", "copies":
		return EffectCopyStackObject
	case "connive", "connives":
		return EffectConnive
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
	case "exchange", "exchanges":
		return EffectExchange
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
	case "manifest", "manifests":
		return EffectManifest
	case "cloak", "cloaks":
		return EffectCloak
	case "mill", "mills":
		return EffectMill
	case "move", "moves":
		return EffectMoveCounters
	case "get", "gets":
		return EffectModifyPT
	case "put", "puts", "distribute", "distributes":
		// "Distribute N <kind> counters among ... target creatures" places
		// counters split among the chosen targets; it is a counter placement
		// whose distribution the DistributeCounters flag and lowerer model.
		return EffectPut
	case "populate", "populates":
		return EffectPopulate
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
	case "goad", "goads":
		return EffectGoad
	case "untap", "untaps":
		return EffectUntap
	// "convert"/"converts" is the Transformers-flavored name for the transform
	// keyword action (CR 701.30): it flips a transforming double-faced permanent
	// to its other face. The adjective "converted" (as in "converted mana cost"
	// or "cast this card converted") is a different word and does not match here.
	case "transform", "transforms", "convert", "converts":
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
// lookAtHandInstruction reports whether the sentence is the private
// hand-inspection effect "Look at <player>'s hand." (Gitaxian Probe, Peek). The
// verb "look" is generic, so the classification is anchored on the "look at"
// lead-in, a possessive player reference (a token ending in "'s"), and a
// trailing "hand." clause boundary. This distinguishes it from the library
// "look at the top ..." dig/visibility wordings handled before it.
func lookAtHandInstruction(tokens []shared.Token) bool {
	if len(tokens) < 5 || !effectWordsAt(tokens, 0, "look", "at") {
		return false
	}
	last := len(tokens) - 1
	if tokens[last].Kind != shared.Period || !equalWord(tokens[last-1], "hand") {
		return false
	}
	for _, token := range tokens[2 : last-1] {
		if strings.HasSuffix(token.Text, "'s") {
			return true
		}
	}
	return false
}

func digLookInstruction(tokens []shared.Token) bool {
	return len(tokens) == 10 &&
		effectWordsAt(tokens, 0, "look", "at", "the", "top") &&
		tokens[4].Kind == shared.Word &&
		effectWordsAt(tokens, 5, "cards", "of", "your", "library") &&
		tokens[9].Kind == shared.Period
}

// lookAtTopCardAnyTimeInstruction reports whether the sentence is the
// continuous-visibility static "look at the top card of your library any time."
// (Bolas's Citadel, Vizier of the Menagerie, Sphinx of Jwar Isle). It is a
// player-rule permission rather than a resolving effect, so the effect
// classifier leaves the "look" verb for the static-declaration recognizer.
func lookAtTopCardAnyTimeInstruction(tokens []shared.Token) bool {
	return len(tokens) == 11 &&
		effectWordsAt(tokens, 0, "look", "at", "the", "top", "card", "of", "your", "library", "any", "time") &&
		tokens[10].Kind == shared.Period
}

// lookAtLibraryTopInstruction reports whether the sentence is the one-shot peek
// "look at the top card of your library." (the Kinship ability word's leading
// instruction). It is the resolving-effect counterpart of
// lookAtTopCardAnyTimeInstruction's continuous "any time" permission: the player
// privately sees the top card once as the ability resolves, conveying hidden
// information without moving the card.
func lookAtLibraryTopInstruction(tokens []shared.Token) bool {
	n := len(tokens)
	if n < 9 || tokens[n-1].Kind != shared.Period || !equalWord(tokens[n-2], "library") {
		return false
	}
	if !effectWordsAt(tokens, 0, "look", "at", "the", "top", "card", "of") {
		return false
	}
	// The library owner sits between "of" and "library": "your" (the controller,
	// Kinship's leading peek) or the possessive "target player's"/"target
	// opponent's"/"that player's" (Merfolk Observer, Dewdrop Spy, Saheeli's
	// Silverwing), where the controller peeks another player's library.
	owner := tokens[6 : n-2]
	if len(owner) == 1 && equalWord(owner[0], "your") {
		return true
	}
	return len(owner) == 2 &&
		(strings.EqualFold(owner[1].Text, "player's") || strings.EqualFold(owner[1].Text, "opponent's")) &&
		(equalWord(owner[0], "target") || equalWord(owner[0], "that"))
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

// chooseCreatureTypeVerbAt reports whether a resolution-time creature-type choice
// ("Choose a creature type.") begins at index. The parser owns this wording: the
// verb "choose" is generic, so the choice classification is anchored on the exact
// "choose a creature type" lead-in followed by a clause boundary. Other "choose"
// wordings (modal "Choose one", "choose a color", "choose target ...") fail this
// check and stay classified elsewhere.
// chooseCreatureTypeVerbAt reports whether a standalone "Choose a creature type."
// effect clause begins at index. It fires only at a sentence boundary (clause
// start or just after a period) so the shared "choose" verb in an entry
// replacement ("As this creature enters, choose a creature type.") is left to the
// entry-choice path and not classified as a separate top-level effect.
func chooseCreatureTypeVerbAt(tokens []shared.Token, index int) bool {
	if index != 0 && tokens[index-1].Kind != shared.Period {
		return false
	}
	return (equalWord(tokens[index], "choose") || equalWord(tokens[index], "chooses")) &&
		effectWordsAt(tokens, index+1, "a", "creature", "type") &&
		(index+4 >= len(tokens) || tokens[index+4].Kind == shared.Period)
}
func copyVerbAt(tokens []shared.Token, index int) bool {
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

// counterPossessionVerbAt reports whether the possession verb "has"/"have" at
// index introduces a counter-possession state ("has a +1/+1 counter on it",
// "has ten or more +1/+1 counters on it") rather than a keyword grant. A keyword
// grant's object ("has trample", "has protection from red") never opens with the
// determiner "a"/"an" or a count word, so requiring that opener and a
// "counter"/"counters" head before the next clause boundary reclassifies the
// counter-possession gate that appears inside an "as long as it has ... counter
// on it" condition as a non-effect, while leaving real keyword grants intact.
func counterPossessionVerbAt(tokens []shared.Token, index int) bool {
	if index+1 >= len(tokens) || !counterPossessionOpener(tokens[index+1]) {
		return false
	}
	for i := index + 1; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period || tokens[i].Kind == shared.Semicolon {
			return false
		}
		if equalWord(tokens[i], "counter") || equalWord(tokens[i], "counters") {
			return true
		}
	}
	return false
}

// totalPowerPossessionVerbAt reports whether the "have"/"has" verb at index opens
// a collective-power qualifier "have total power ...". Such a clause is a
// condition predicate ("If those creatures have total power 8 or greater,
// convert Ultra Magnus.", Ultra Magnus, Armored Carrier; "If creatures you
// control have total power 8 or greater, ..."), not a keyword-grant effect, so
// the effect classifier must not treat it as an effect verb and strand a phantom
// EffectGrantKeyword when the clause opens a standalone leading condition.
func totalPowerPossessionVerbAt(tokens []shared.Token, index int) bool {
	return effectWordsAt(tokens, index+1, "total", "power")
}

// counterPossessionOpener reports whether token opens a counter-possession object
// phrase: the determiner "a"/"an" of the single-counter form or a leading count
// word of the threshold form ("ten or more ... counters").
func counterPossessionOpener(token shared.Token) bool {
	if equalWord(token, "a") || equalWord(token, "an") {
		return true
	}
	_, ok := CardinalWordValue(token.Text)
	return ok
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

// everyCreatureTypeGainRiderAt reports whether the tokens at index begin a "gain
// all creature types" / "gain every creature type" rider, the all-creature-type
// grant folded onto a base power/toughness set (Mirror Entity).
func everyCreatureTypeGainRiderAt(tokens []shared.Token, index int) bool {
	return staticWordsAt(tokens, index+1, "all", "creature", "types") ||
		staticWordsAt(tokens, index+1, "every", "creature", "type")
}

// priorBasePowerToughnessSet reports whether a "base power and toughness" set
// phrase precedes index in the same sentence, marking the gain-every-creature
// rider as folded onto that set rather than a standalone effect.
func priorBasePowerToughnessSet(tokens []shared.Token, index int) bool {
	for i := 0; i+3 < index; i++ {
		if staticWordsAt(tokens, i, "base", "power", "and", "toughness") {
			return true
		}
	}
	return false
}

// effectFallbackOnInability reports whether an effect's subject is a "who can't"
// relative clause naming the players who couldn't satisfy a preceding required
// action ("Each player who can't discards a card."). It detects "who"
// immediately followed by "can't"/"cannot" within the subject token range
// [start,index), the verb-leading subject of the effect.
func effectFallbackOnInability(tokens []shared.Token, start, index int) bool {
	if start < 0 || index > len(tokens) {
		return false
	}
	for i := start; i+1 < index; i++ {
		if equalWord(tokens[i], "who") &&
			(equalWord(tokens[i+1], "can't") || equalWord(tokens[i+1], "cannot")) {
			return true
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
