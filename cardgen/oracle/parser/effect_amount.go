package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func parseEffectDuration(tokens []shared.Token, atoms Atoms) EffectDurationKind {
	words := normalizedWords(tokens)
	switch {
	case effectContainsWords(words, "until", "the", "end", "of", "your", "next", "turn"):
		return EffectDurationUntilYourNextTurn
	case effectContainsWords(words, "until", "end", "of", "turn"):
		return EffectDurationUntilEndOfTurn
	case effectContainsWords(words, "until", "your", "next", "turn"):
		return EffectDurationUntilYourNextTurn
	case effectContainsWords(words, "this", "combat"):
		return EffectDurationThisCombat
	case effectContainsWords(words, "this", "turn"):
		return EffectDurationThisTurn
	case effectContainsWords(words, "as", "long", "as", "this") &&
		(effectContainsWords(words, "remains", "on", "the", "battlefield") ||
			effectContainsWords(words, "is", "on", "the", "battlefield")):
		return EffectDurationWhileSourceOnBattlefield
	case effectContainsWords(words, "for", "as", "long", "as", "you", "control", "this"):
		return EffectDurationWhileYouControlSource
	}
	for i := 0; i+6 < len(tokens); i++ {
		if !effectWordsAt(tokens, i, "for", "as", "long", "as", "you", "control") {
			continue
		}
		nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[i+6].Span)
		if !ok {
			continue
		}
		end := i + 6
		for end < len(tokens) && spanCovers(nameSpan, tokens[end].Span) {
			end++
		}
		if end == len(tokens)-1 && tokens[end].Kind == shared.Period {
			return EffectDurationWhileYouControlSource
		}
	}
	return EffectDurationNone
}

func cutDelayedTiming(tokens []shared.Token) ([]shared.Token, DelayedTimingKind) {
	end := len(tokens)
	if end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	for _, suffix := range []struct {
		words  []string
		timing DelayedTimingKind
	}{
		{[]string{"at", "the", "beginning", "of", "the", "next", "end", "step"}, DelayedTimingNextEndStep},
		{[]string{"at", "the", "beginning", "of", "the", "next", "turn's", "upkeep"}, DelayedTimingNextUpkeep},
	} {
		start := end - len(suffix.words)
		if start >= 0 && effectWordsAt(tokens, start, suffix.words...) {
			return append(append([]shared.Token(nil), tokens[:start]...), tokens[end:]...), suffix.timing
		}
	}
	return tokens, DelayedTimingNone
}

// parseTokenPowerToughness finds a created token's fixed unsigned power/toughness
// ("1/1", "2/2") in the create clause: an integer, a slash, and an integer. It
// reports false when no such pattern is present (named tokens with no P/T).
func parseTokenPowerToughness(kind EffectKind, tokens []shared.Token) (power, toughness int, ok bool) {
	if kind != EffectCreate {
		return 0, 0, false
	}
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i].Kind != shared.Integer ||
			tokens[i+1].Kind != shared.Slash ||
			tokens[i+2].Kind != shared.Integer {
			continue
		}
		p, err := strconv.Atoi(tokens[i].Text)
		if err != nil {
			continue
		}
		t, err := strconv.Atoi(tokens[i+2].Text)
		if err != nil {
			continue
		}
		return p, t, true
	}
	return 0, 0, false
}

func parsePTChange(tokens []shared.Token) (power, toughness SignedAmountSyntax) {
	for i := 0; i+4 < len(tokens); i++ {
		power, powerOK := parseSignedAmount(tokens[i], tokens[i+1])
		toughness, toughnessOK := parseSignedAmount(tokens[i+3], tokens[i+4])
		if powerOK && tokens[i+2].Kind == shared.Slash && toughnessOK {
			return power, toughness
		}
	}
	return SignedAmountSyntax{}, SignedAmountSyntax{}
}

func parseSignedAmount(sign, amount shared.Token) (SignedAmountSyntax, bool) {
	if amount.Kind != shared.Integer || sign.Kind != shared.Plus && sign.Kind != shared.Minus {
		return SignedAmountSyntax{}, false
	}
	value, err := strconv.Atoi(amount.Text)
	if err != nil {
		return SignedAmountSyntax{}, false
	}
	return SignedAmountSyntax{
		Span:     shared.Span{Start: sign.Span.Start, End: amount.Span.End},
		Value:    value,
		Known:    true,
		Negative: sign.Kind == shared.Minus,
	}, true
}

func parseEffectAmount(kind EffectKind, tokens []shared.Token, atoms Atoms) EffectAmountSyntax {
	if amount, attempted, ok := parseDynamicEffectAmount(tokens, atoms); attempted {
		if ok {
			return amount
		}
		return EffectAmountSyntax{}
	}
	if kind == EffectEnterTapped {
		for i, token := range tokens {
			if equalWord(token, "with") && i+1 < len(tokens) && equalWord(tokens[i+1], "X") {
				return EffectAmountSyntax{Span: tokens[i+1].Span, VariableX: true}
			}
		}
	}
	for _, token := range tokens {
		if token.Kind != shared.Word {
			continue
		}
		if equalWord(token, "X") {
			return EffectAmountSyntax{Span: token.Span, VariableX: true}
		}
		break
	}
	for i, token := range tokens {
		if value, ok := effectNumber(token, atoms); ok && value > 0 {
			if i > 0 && tokens[i-1].Kind == shared.Minus {
				return EffectAmountSyntax{}
			}
			return EffectAmountSyntax{Span: token.Span, Value: value, Known: true}
		}
		if equalWord(token, "a") || equalWord(token, "an") {
			if i > 0 && equalWord(tokens[i-1], "from") {
				continue
			}
			return EffectAmountSyntax{Span: token.Span, Value: 1, Known: true}
		}
	}
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return EffectAmountSyntax{Span: token.Span, Value: 1, Known: true}
		}
	}
	if (kind == EffectInvestigate || kind == EffectProliferate) &&
		len(tokens) == 1 && tokens[0].Kind == shared.Period {
		return EffectAmountSyntax{Value: 1, Known: true}
	}
	return EffectAmountSyntax{}
}

type dynamicAmountPrefix struct {
	form       EffectDynamicAmountForm
	start      int
	multiplier int
	plural     bool
	count      bool
}

type dynamicAmountSubject struct {
	amount EffectAmountSyntax
	end    int
	plural bool
	count  bool
}

func parseDynamicEffectAmount(tokens []shared.Token, atoms Atoms) (amount EffectAmountSyntax, attempted, ok bool) {
	var matches []EffectAmountSyntax
	for i := range tokens {
		prefix, prefixOK := parseDynamicAmountPrefix(tokens, i, atoms)
		if !prefixOK {
			continue
		}
		attempted = true
		subject, subjectOK := parseDynamicAmountSubject(tokens, prefix.start, atoms)
		if !subjectOK || subject.count != prefix.count || subject.count && subject.plural != prefix.plural {
			continue
		}
		match := subject.amount
		match.DynamicForm = prefix.form
		match.Multiplier = prefix.multiplier
		match.Span = shared.SpanOf(tokens[i:subject.end])
		match.Text = joinedEffectText(tokens[i:subject.end])
		matches = append(matches, match)
	}
	if len(matches) != 1 {
		return EffectAmountSyntax{}, attempted, false
	}
	return matches[0], true, true
}

func parseDynamicAmountPrefix(tokens []shared.Token, index int, atoms Atoms) (dynamicAmountPrefix, bool) {
	switch {
	case effectWordsAt(tokens, index, "equal", "to", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 6, 2, true, true}, true
	case effectWordsAt(tokens, index, "equal", "to", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 5, 1, true, true}, true
	case effectWordsAt(tokens, index, "for", "each"):
		return dynamicAmountPrefix{EffectDynamicAmountFormForEach, index + 2, precedingEffectMultiplier(tokens[:index], atoms), false, true}, true
	case effectWordsAt(tokens, index, "equal", "to"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 2, 1, false, false}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 7, 2, true, true}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 6, 1, true, true}, true
	case effectWordsAt(tokens, index, "where", "X", "is"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 3, 1, false, false}, true
	default:
		return dynamicAmountPrefix{}, false
	}
}

func precedingEffectMultiplier(tokens []shared.Token, atoms Atoms) int {
	multiplier := 0
	for _, token := range tokens {
		value, ok := effectNumber(token, atoms)
		if !ok || value == 0 {
			continue
		}
		if multiplier != 0 && multiplier != value {
			return 0
		}
		multiplier = value
	}
	if multiplier == 0 {
		return 1
	}
	return multiplier
}

func parseDynamicAmountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	switch {
	case effectWordsAt(tokens, start, "your", "life", "total") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountControllerLife},
			end:    start + 3,
		}, true
	case effectWordsAt(tokens, start, "its", "power") && dynamicAmountBoundary(tokens, start+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: tokens[start].Span},
			end:    start + 2,
		}, true
	case effectWordsAt(tokens, start, "this", "creature") &&
		start+4 < len(tokens) && tokens[start+2].Kind == shared.Apostrophe &&
		equalWord(tokens[start+3], "s") && equalWord(tokens[start+4], "power") &&
		dynamicAmountBoundary(tokens, start+5):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 5,
		}, true
	case start+2 < len(tokens) && equalWord(tokens[start], "this") &&
		strings.EqualFold(tokens[start+1].Text, "creature's") &&
		equalWord(tokens[start+2], "power") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 3,
		}, true
	case effectWordsAt(tokens, start, "basic", "land", "type", "among", "lands", "you", "control") &&
		dynamicAmountBoundary(tokens, start+7):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountBasicLandTypes},
			end:    start + 7, count: true,
		}, true
	case effectWordsAt(tokens, start, "basic", "land", "types", "among", "lands", "you", "control") &&
		dynamicAmountBoundary(tokens, start+7):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountBasicLandTypes},
			end:    start + 7, count: true, plural: true,
		}, true
	}
	if subject, ok := parseDynamicCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[start].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	end := start
	for end < len(tokens) && tokens[end].Span.End.Offset <= nameSpan.End.Offset {
		end++
	}
	if end < len(tokens) && equalWord(tokens[end], "power") && dynamicAmountBoundary(tokens, end+1) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: nameSpan},
			end:    end + 1,
		}, true
	}
	if end+2 < len(tokens) && tokens[end].Kind == shared.Apostrophe &&
		equalWord(tokens[end+1], "s") && equalWord(tokens[end+2], "power") &&
		dynamicAmountBoundary(tokens, end+3) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: nameSpan},
			end:    end + 3,
		}, true
	}
	return dynamicAmountSubject{}, false
}

func parseDynamicCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if equalWord(tokens[start], "card") || equalWord(tokens[start], "cards") {
		if subject, ok := parseDynamicEventCardCountSubject(tokens, start); ok {
			return subject, true
		}
		if subject, ok := parseDynamicCardCountSubject(tokens, start, atoms); ok {
			return subject, true
		}
	}
	if subject, ok := parseDynamicObjectNounCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	return parseDynamicSelectionCountSubject(tokens, start, atoms)
}

// parseDynamicEventCardCountSubject recognizes "card[s] discarded this way" and
// "card[s] drawn this way" count subjects. In a draw or discard triggered
// ability these refer to the cards drawn or discarded in the triggering event,
// which the lowerer resolves only when the enclosing trigger matches.
func parseDynamicEventCardCountSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	end := start + 1
	if end >= len(tokens) ||
		(!equalWord(tokens[end], "discarded") && !equalWord(tokens[end], "drawn")) {
		return dynamicAmountSubject{}, false
	}
	end++
	if !effectWordsAt(tokens, end, "this", "way") || !dynamicAmountBoundary(tokens, end+2) {
		return dynamicAmountSubject{}, false
	}
	end += 2
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountEventCardCount},
		end:    end, count: true, plural: strings.EqualFold(tokens[start].Text, "cards"),
	}, true
}

func parseDynamicObjectNounCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	noun, ok := atoms.ObjectNounAt(tokens[start].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	plural := strings.HasSuffix(strings.ToLower(tokens[start].Text), "s")
	if noun == ObjectNounOpponent {
		end := start + 1
		if effectWordsAt(tokens, end, "you", "have") {
			end += 2
		}
		if dynamicAmountBoundary(tokens, end) {
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountOpponentCount},
				end:    end, count: true, plural: plural,
			}, true
		}
		return dynamicAmountSubject{}, false
	}
	if !slices.Contains([]ObjectNoun{
		ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand, ObjectNounPermanent,
	}, noun) {
		return dynamicAmountSubject{}, false
	}
	end := start + 1
	for _, suffix := range [][]string{{"you", "control"}, {"your", "opponents", "control"}, {"on", "the", "battlefield"}} {
		if !effectWordsAt(tokens, end, suffix...) || !dynamicAmountBoundary(tokens, end+len(suffix)) {
			continue
		}
		subjectEnd := end + len(suffix)
		selection := parseSelection(tokens[start:subjectEnd], atoms)
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// parseDynamicSelectionCountSubject recognizes "for each <selection> ..." count
// subjects led by a subtype, color, supertype, or color qualifier rather than a
// bare card-type noun (for example "Shrine you control", "colorless creature you
// control", "Elf card in your graveyard", or "card in your hand"). The leading
// run of tokens must all be recognized selection atoms; anything else fails
// closed so unsupported wordings stay rejected.
func parseDynamicSelectionCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	end, ok := scanDynamicCountSelectionTokens(tokens, start, atoms)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	plural := dynamicCountHeadPlural(tokens, end-1, atoms)
	for _, suffix := range [][]string{{"you", "control"}, {"your", "opponents", "control"}, {"on", "the", "battlefield"}} {
		if !effectWordsAt(tokens, end, suffix...) || !dynamicAmountBoundary(tokens, end+len(suffix)) {
			continue
		}
		subjectEnd := end + len(suffix)
		selection := buildDynamicCountSelection(tokens, start, subjectEnd, atoms)
		if selection.Zone != zone.None || !dynamicCountSelectionTypesFaithful(selection) {
			return dynamicAmountSubject{}, false
		}
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	for _, zoneSuffix := range []struct {
		words []string
		kind  zone.Type
	}{
		{[]string{"in", "your", "graveyard"}, zone.Graveyard},
		{[]string{"in", "your", "hand"}, zone.Hand},
	} {
		if !effectWordsAt(tokens, end, zoneSuffix.words...) || !dynamicAmountBoundary(tokens, end+len(zoneSuffix.words)) {
			continue
		}
		subjectEnd := end + len(zoneSuffix.words)
		selection := buildDynamicCountSelection(tokens, start, subjectEnd, atoms)
		if !dynamicCountSelectionTypesFaithful(selection) {
			return dynamicAmountSubject{}, false
		}
		selection.Controller = SelectionControllerYou
		selection.Zone = zoneSuffix.kind
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

func scanDynamicCountSelectionTokens(tokens []shared.Token, start int, atoms Atoms) (int, bool) {
	end := start
	for end < len(tokens) && isDynamicCountSelectionToken(tokens[end], atoms) {
		end++
	}
	if end == start {
		return start, false
	}
	return end, true
}

func isDynamicCountSelectionToken(token shared.Token, atoms Atoms) bool {
	if noun, ok := atoms.ObjectNounAt(token.Span); ok {
		return slices.Contains([]ObjectNoun{
			ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment,
			ObjectNounLand, ObjectNounPermanent, ObjectNounCard,
		}, noun)
	}
	if _, ok := atoms.CardTypeAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.ExcludedCardTypeAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.ColorAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.ExcludedColorAt(token.Span); ok {
		return true
	}
	if qualifier, ok := atoms.ColorQualifierAt(token.Span); ok {
		return qualifier == ColorQualifierColorless || qualifier == ColorQualifierMulticolored
	}
	if _, ok := atoms.SupertypeAt(token.Span); ok {
		return true
	}
	if _, ok := atoms.SubtypeAt(token.Span); ok {
		return true
	}
	return false
}

func buildDynamicCountSelection(tokens []shared.Token, start, end int, atoms Atoms) SelectionSyntax {
	selection := parseSelection(tokens[start:end], atoms)
	for i := start; i < end; i++ {
		qualifier, ok := atoms.ColorQualifierAt(tokens[i].Span)
		if !ok {
			continue
		}
		switch qualifier {
		case ColorQualifierColorless:
			selection.Colorless = true
		case ColorQualifierMulticolored:
			selection.Multicolored = true
		default:
		}
	}
	return selection
}

// dynamicCountSelectionTypesFaithful reports whether a count selection's parsed
// card types round-trip through the count lowering paths, which carry a single
// card type via the selection Kind and drop the redundant RequiredTypesAny.
// A selection is faithful only when RequiredTypesAny is empty or holds exactly
// the one card type the Kind already encodes; anything else (a type the Kind
// cannot represent, such as "instant card", or a multi-type conjunction such as
// "artifact creature") would silently mis-count, so it fails closed.
func dynamicCountSelectionTypesFaithful(selection SelectionSyntax) bool {
	switch len(selection.RequiredTypesAny) {
	case 0:
		return true
	case 1:
		return selection.RequiredTypesAny[0] == impliedCountCardType(selection.Kind)
	default:
		return false
	}
}

func impliedCountCardType(kind SelectionKind) CardType {
	switch kind {
	case SelectionArtifact:
		return CardTypeArtifact
	case SelectionCreature:
		return CardTypeCreature
	case SelectionEnchantment:
		return CardTypeEnchantment
	case SelectionLand:
		return CardTypeLand
	case SelectionPlaneswalker:
		return CardTypePlaneswalker
	case SelectionBattle:
		return CardTypeBattle
	default:
		return CardTypeUnknown
	}
}

func dynamicCountHeadPlural(tokens []shared.Token, headIndex int, atoms Atoms) bool {
	if headIndex < 0 || headIndex >= len(tokens) {
		return false
	}
	if _, ok := atoms.ObjectNounAt(tokens[headIndex].Span); !ok {
		return false
	}
	return strings.HasSuffix(strings.ToLower(tokens[headIndex].Text), "s")
}

func parseDynamicCardCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	end := start + 1
	if end >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	keyword, ok := atoms.KeywordSelectorStartingAt(tokens[end].Span)
	if !ok || keyword.Excluded || keyword.Keyword != KeywordCycling {
		return dynamicAmountSubject{}, false
	}
	for end < len(tokens) && tokens[end].Span.End.Offset <= keyword.Span.End.Offset {
		end++
	}
	if !effectWordsAt(tokens, end, "in", "your", "graveyard") || !dynamicAmountBoundary(tokens, end+3) {
		return dynamicAmountSubject{}, false
	}
	end += 3
	selection := parseSelection(tokens[start:end], atoms)
	selection.Kind = SelectionCard
	selection.Controller = SelectionControllerYou
	selection.Zone = zone.Graveyard
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
		end:    end, count: true, plural: strings.EqualFold(tokens[start].Text, "cards"),
	}, true
}

func dynamicAmountBoundary(tokens []shared.Token, end int) bool {
	if end >= len(tokens) {
		return true
	}
	if tokens[end].Kind == shared.Comma || tokens[end].Kind == shared.Period {
		return true
	}
	return equalWord(tokens[end], "to") || equalWord(tokens[end], "until")
}

func effectNumber(token shared.Token, atoms Atoms) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	return atoms.CardinalAt(token.Span)
}

func parseCounterPlacement(tokens []shared.Token, atoms Atoms) (counter.Kind, bool) {
	for _, token := range tokens {
		if equalWord(token, "and") {
			return counter.Kind(0), false
		}
	}
	span := shared.SpanOf(tokens)
	var kinds []counter.Kind
	for _, atom := range atoms.Counters() {
		if spanCovers(span, atom.Span) {
			kinds = append(kinds, atom.Kind)
		}
	}
	if len(kinds) != 1 {
		return counter.Kind(0), false
	}
	kind := kinds[0]
	return kind, kind.Valid() && kind != counter.Stun && kind != counter.Finality
}

func firstZone(atoms Atoms, span shared.Span, role ZoneRole) zone.Type {
	result := zone.None
	for _, atom := range atoms.Zones() {
		if atom.Role != role || !spanCovers(span, atom.Span) {
			continue
		}
		if result != zone.None && atom.Zone != result {
			return zone.None
		}
		result = atom.Zone
	}
	return result
}

func firstEffectSymbol(tokens []shared.Token) string {
	for _, token := range tokens {
		if token.Kind == shared.Symbol {
			return token.Text
		}
	}
	return ""
}

func referencesInSpan(atoms Atoms, span shared.Span) []Reference {
	var references []Reference
	for _, reference := range atoms.References() {
		if spanCovers(span, reference.Span) {
			references = append(references, reference)
		}
	}
	return references
}
