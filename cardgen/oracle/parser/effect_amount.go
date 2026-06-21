package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
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
	case effectContainsWords(words, "for", "as", "long", "as", "that", "creature", "is", "enchanted"):
		return EffectDurationWhileControlledCreatureEnchanted
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

func leadingDelayedTiming(tokens []shared.Token) DelayedTimingKind {
	if len(tokens) == 9 &&
		effectWordsAt(tokens, 0, "at", "the", "beginning", "of", "your", "next", "main", "phase") &&
		tokens[8].Kind == shared.Comma {
		return DelayedTimingNextMain
	}
	return DelayedTimingNone
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

// parseTokenKeywords returns, in source order, every recognized keyword name in a
// create clause ("with menace and reach" -> [Menace, Reach]). It scans only the
// effect clause, so each returned keyword is present in that clause's text; the
// create-token exactness recognizer reconstructs the "with <keyword> and
// <keyword> ..." rider from this list and fails closed on any byte mismatch. It
// returns nil for non-create effects and for create clauses with no keyword.
func parseTokenKeywords(kind EffectKind, tokens []shared.Token, atoms Atoms) []KeywordKind {
	if kind != EffectCreate {
		return nil
	}
	keywords := scanKeywords(tokens, atoms)
	if len(keywords) == 0 {
		return nil
	}
	kinds := make([]KeywordKind, 0, len(keywords))
	for _, keyword := range keywords {
		kinds = append(kinds, keyword.Kind)
	}
	return kinds
}

// parseTokenName captures a created creature token's explicit Oracle name from
// the trailing "named <Name>" tail of a create clause ("... Serpent creature
// tokens named Koma's Coil." -> "Koma's Coil"). It returns the name joined
// verbatim from the source tokens after the first "named" word that follows the
// token noun ("token"/"tokens"), through the clause end (excluding a trailing
// period). It returns "" for non-create effects and for clauses with no such
// tail. The create-token exactness recognizer reconstructs and byte-checks the
// "named <Name>" tail, so any spurious capture fails closed there.
func parseTokenName(kind EffectKind, tokens []shared.Token) string {
	if kind != EffectCreate {
		return ""
	}
	noun := -1
	for i, token := range tokens {
		if equalWord(token, "token") || equalWord(token, "tokens") {
			noun = i
		}
	}
	if noun < 0 {
		return ""
	}
	named := -1
	for i := noun + 1; i < len(tokens); i++ {
		if equalWord(tokens[i], "named") {
			named = i
			break
		}
	}
	if named < 0 {
		return ""
	}
	nameTokens := tokens[named+1:]
	if len(nameTokens) > 0 && nameTokens[len(nameTokens)-1].Kind == shared.Period {
		nameTokens = nameTokens[:len(nameTokens)-1]
	}
	if len(nameTokens) == 0 {
		return ""
	}
	// A "with" inside the name region marks a granted-ability rider ("... named X
	// with \"...\"") whose quoted body was stripped from the clause. Such tokens
	// carry an ability this recognizer cannot represent, so fail closed rather
	// than absorb the dangling "with" into the name.
	for _, token := range nameTokens {
		if equalWord(token, "with") {
			return ""
		}
	}
	return joinedEffectText(nameTokens)
}

// parseTokenChoice reports whether a create clause offers a choice between two
// complete token specs joined by "or" ("create a Food token or a Treasure
// token"). The signal is a "token"/"tokens" noun on each side of an "or": each
// alternative names its own token, so the effect creates one of the
// alternatives rather than a single multi-subtype token. The create-token
// exactness recognizer reconstructs and byte-checks the full "a <A> token or a
// <B> token" wording, so a spurious signal fails closed there. It returns false
// for non-create clauses and for every clause without that two-noun "or" shape.
func parseTokenChoice(kind EffectKind, tokens []shared.Token) bool {
	if kind != EffectCreate {
		return false
	}
	orIndex := -1
	for i, token := range tokens {
		if equalWord(token, "or") {
			orIndex = i
			break
		}
	}
	if orIndex < 0 {
		return false
	}
	nounBefore := slices.ContainsFunc(tokens[:orIndex], func(token shared.Token) bool {
		return equalWord(token, "token") || equalWord(token, "tokens")
	})
	nounAfter := slices.ContainsFunc(tokens[orIndex+1:], func(token shared.Token) bool {
		return equalWord(token, "token") || equalWord(token, "tokens")
	})
	return nounBefore && nounAfter
}

func parsePTChange(tokens []shared.Token) (power, toughness SignedAmountSyntax) {
	for i := 0; i+4 < len(tokens); i++ {
		if tokens[i+2].Kind != shared.Slash {
			continue
		}
		power, powerOK := parseSignedPTSide(tokens[i], tokens[i+1])
		toughness, toughnessOK := parseSignedPTSide(tokens[i+3], tokens[i+4])
		if powerOK && toughnessOK {
			return power, toughness
		}
	}
	return SignedAmountSyntax{}, SignedAmountSyntax{}
}

// parseSignedPTSide parses one side of a power/toughness delta written as a sign
// followed by either a fixed integer ("+2", "-1") or the variable "X" ("+X",
// "-X"). The X side carries VariableX with Known left false so its magnitude is
// supplied by the effect's dynamic amount.
func parseSignedPTSide(sign, amount shared.Token) (SignedAmountSyntax, bool) {
	if sign.Kind != shared.Plus && sign.Kind != shared.Minus {
		return SignedAmountSyntax{}, false
	}
	span := shared.Span{Start: sign.Span.Start, End: amount.Span.End}
	if amount.Kind == shared.Word && equalWord(amount, "X") {
		return SignedAmountSyntax{
			Span:      span,
			Negative:  sign.Kind == shared.Minus,
			VariableX: true,
		}, true
	}
	if amount.Kind != shared.Integer {
		return SignedAmountSyntax{}, false
	}
	value, err := strconv.Atoi(amount.Text)
	if err != nil {
		return SignedAmountSyntax{}, false
	}
	return SignedAmountSyntax{
		Span:     span,
		Value:    value,
		Known:    true,
		Negative: sign.Kind == shared.Minus,
	}, true
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
	if kind == EffectGain || kind == EffectLose {
		for i := 0; i+1 < len(tokens); i++ {
			if equalWord(tokens[i], "that") && equalWord(tokens[i+1], "much") {
				return EffectAmountSyntax{
					Span:        shared.SpanOf(tokens[i : i+2]),
					Text:        joinedEffectText(tokens[i : i+2]),
					DynamicKind: EffectDynamicAmountTriggeringLifeChange,
				}
			}
		}
	}
	if kind == EffectDraw || kind == EffectUntap {
		for i, token := range tokens {
			value, ok := effectNumber(token, atoms)
			if !ok || value < 1 || i < 2 ||
				!equalWord(tokens[i-2], "up") ||
				!equalWord(tokens[i-1], "to") {
				continue
			}
			return EffectAmountSyntax{
				Span:       token.Span,
				RangeKnown: true,
				Minimum:    0,
				Maximum:    value,
			}
		}
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

// parseCreateForEachAmount types a leading "for each <count subject>,"
// iteration prefix on a controller creature-token create effect. The "for each
// X" sits before the verb, so parseEffectAmount only sees the post-verb
// "a ... token" and models a fixed single token; this types the iterator as a
// dynamic count (the for-each subject) with the create clause's single token as
// the per-iteration multiplier, so the lowerer creates one token per counted
// object. It returns false unless the effect is a single-token controller
// creature-token create (tokenPTKnown) whose pre-verb tokens are exactly a
// recognized "for each <count subject>," prefix; in particular it never retypes
// a copy-of-target token create, which carries no fixed power/toughness.
func parseCreateForEachAmount(kind EffectKind, context EffectContextKind, tokenPTKnown bool, pre []shared.Token, clauseAmount EffectAmountSyntax, atoms Atoms) (EffectAmountSyntax, bool) {
	if kind != EffectCreate || context != EffectContextController || !tokenPTKnown ||
		!clauseAmount.Known || clauseAmount.Value != 1 {
		return EffectAmountSyntax{}, false
	}
	prefix, ok := parseDynamicAmountPrefix(pre, 0, atoms)
	if !ok || prefix.form != EffectDynamicAmountFormForEach {
		return EffectAmountSyntax{}, false
	}
	subject, ok := parseDynamicAmountSubject(pre, prefix.start, atoms)
	if !ok || !subject.count || subject.count != prefix.count ||
		subject.plural != prefix.plural {
		return EffectAmountSyntax{}, false
	}
	if subject.end != len(pre)-1 || pre[subject.end].Kind != shared.Comma {
		return EffectAmountSyntax{}, false
	}
	amount := subject.amount
	amount.DynamicForm = EffectDynamicAmountFormForEach
	amount.Multiplier = clauseAmount.Value
	amount.Span = shared.SpanOf(pre[:subject.end])
	amount.Text = joinedEffectText(pre[:subject.end])
	return amount, true
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
	if subject, ok := parseDynamicSourceCounterCount(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicGreatestCharacteristicSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicGreatestDiscardedThisWaySubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicDevotionSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicSpellsCastThisTurnSubject(tokens, start); ok {
		return subject, true
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
	case effectWordsAt(tokens, start, "its", "toughness") && dynamicAmountBoundary(tokens, start+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourceToughness, ReferenceSpan: tokens[start].Span},
			end:    start + 2,
		}, true
	case effectWordsAt(tokens, start, "its", "mana", "value") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourceManaValue, ReferenceSpan: tokens[start].Span},
			end:    start + 3,
		}, true
	case start+1 < len(tokens) && equalWord(tokens[start], "that") &&
		referencePossessiveObjectNoun(tokens[start+1]) &&
		effectWordsAt(tokens, start+2, "mana", "value") &&
		dynamicAmountBoundary(tokens, start+4):
		// "that <object>'s mana value" names the mana value of a referenced
		// permanent ("that permanent's mana value", "that card's mana value").
		// The reference spans the possessive object phrase ("that permanent's");
		// the collectReferences pass recognizes the same span so the amount's
		// referent binds to the antecedent the prior clause acted on.
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourceManaValue, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 4,
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
	case effectWordsAt(tokens, start, "the", "life", "lost", "this", "way") &&
		dynamicAmountBoundary(tokens, start+5):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountLifeLostThisWay},
			end:    start + 5,
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

// parseDynamicDevotionSubject recognizes "your devotion to <color>" and the
// two-color "your devotion to <color> and <color>" amount subjects (CR 700.5).
// The recognized colors are carried on the amount so the lowerer can rebuild the
// runtime devotion count. It fails closed for any unrecognized color word.
func parseDynamicDevotionSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "your", "devotion", "to") {
		return dynamicAmountSubject{}, false
	}
	colorStart := start + 3
	if colorStart >= len(tokens) || tokens[colorStart].Kind != shared.Word {
		return dynamicAmountSubject{}, false
	}
	first, ok := recognizeColorWord(tokens[colorStart].Text)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	colors := []Color{first}
	end := colorStart + 1
	if effectWordsAt(tokens, end, "and") &&
		end+1 < len(tokens) && tokens[end+1].Kind == shared.Word {
		second, ok := recognizeColorWord(tokens[end+1].Text)
		if !ok {
			return dynamicAmountSubject{}, false
		}
		colors = append(colors, second)
		end += 2
	}
	if !dynamicAmountBoundary(tokens, end) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountDevotion, Colors: colors},
		end:    end,
	}, true
}

// parseDynamicSpellsCastThisTurnSubject recognizes the storm-counter amount
// subject "spell[s] you've cast this turn" (controller-scoped, CR 608.2c). The
// triggering spell counts toward the total because its cast event precedes the
// resolving ability. It fails closed for any trailing qualifier ("from anywhere
// other than your hand", "other than the first") so only the plain count is
// matched.
func parseDynamicSpellsCastThisTurnSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	plural := false
	switch {
	case equalWord(tokens[start], "spell"):
	case equalWord(tokens[start], "spells"):
		plural = true
	default:
		return dynamicAmountSubject{}, false
	}
	end := start + 1
	switch {
	case effectWordsAt(tokens, end, "you've", "cast", "this", "turn"):
		end += 4
	case effectWordsAt(tokens, end, "you", "have", "cast", "this", "turn"):
		end += 5
	default:
		return dynamicAmountSubject{}, false
	}
	if !dynamicAmountBoundary(tokens, end) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSpellsCastThisTurn},
		end:    end, count: true, plural: plural,
	}, true
}

func parseDynamicSourceCounterCount(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	for _, atom := range atoms.Counters() {
		if atom.Span.Start.Offset != tokens[start].Span.Start.Offset {
			continue
		}
		counterNoun := start
		for counterNoun < len(tokens) && tokens[counterNoun].Span.End.Offset <= atom.Span.End.Offset {
			counterNoun++
		}
		if counterNoun >= len(tokens) ||
			(!equalWord(tokens[counterNoun], "counter") && !equalWord(tokens[counterNoun], "counters")) ||
			counterNoun+1 >= len(tokens) || !equalWord(tokens[counterNoun+1], "on") ||
			counterNoun+2 >= len(tokens) {
			continue
		}
		nameSpan, end, ok := sourceCounterReferenceSpan(tokens, counterNoun+2, atoms)
		if !ok || !dynamicAmountBoundary(tokens, end) {
			continue
		}
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{
				DynamicKind:   EffectDynamicAmountSourceCounterCount,
				ReferenceSpan: nameSpan,
				CounterKind:   atom.Kind,
			},
			end:    end,
			count:  true,
			plural: equalWord(tokens[counterNoun], "counters"),
		}, true
	}
	return dynamicAmountSubject{}, false
}

// sourceCounterReferenceSpan recognizes the object naming the source permanent
// that carries the counted counters in a "<kind> counter(s) on <object>" amount.
// The object is the card's own name ("burden counters on The One Ring"), a "this
// <permanent type>" self marker ("charge counter on this artifact"), or the
// pronoun "it". It returns the reference span and the token index just past it.
func sourceCounterReferenceSpan(tokens []shared.Token, idx int, atoms Atoms) (shared.Span, int, bool) {
	if idx >= len(tokens) {
		return shared.Span{}, 0, false
	}
	if nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[idx].Span); ok {
		end := idx
		for end < len(tokens) && tokens[end].Span.End.Offset <= nameSpan.End.Offset {
			end++
		}
		return nameSpan, end, true
	}
	if equalWord(tokens[idx], "this") && idx+1 < len(tokens) && referenceSelfMarkerNoun(tokens[idx+1]) {
		return shared.SpanOf(tokens[idx : idx+2]), idx + 2, true
	}
	if equalWord(tokens[idx], "it") {
		return shared.SpanOf(tokens[idx : idx+1]), idx + 1, true
	}
	return shared.Span{}, 0, false
}

// parseDynamicGreatestCharacteristicSubject recognizes "the greatest <power |
// toughness | mana value> among <group>" amount subjects. The group is any
// battlefield count subject (for example "creatures you control",
// "permanents you control", "Mutants you control"), parsed by reusing the
// count-subject scanners; the recognized selection is carried on the amount so
// the lowerer can rebuild the battlefield group. It fails closed for non-
// battlefield groups (a zone-qualified count) so unsupported wordings stay
// rejected.
func parseDynamicGreatestCharacteristicSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "the", "greatest") {
		return dynamicAmountSubject{}, false
	}
	var kind EffectDynamicAmountKind
	var groupStart int
	switch {
	case effectWordsAt(tokens, start+2, "power", "among"):
		kind, groupStart = EffectDynamicAmountGreatestPower, start+4
	case effectWordsAt(tokens, start+2, "toughness", "among"):
		kind, groupStart = EffectDynamicAmountGreatestToughness, start+4
	case effectWordsAt(tokens, start+2, "mana", "value", "among"):
		kind, groupStart = EffectDynamicAmountGreatestManaValue, start+5
	default:
		return dynamicAmountSubject{}, false
	}
	inner, ok := parseDynamicCountSubject(tokens, groupStart, atoms)
	if !ok || inner.amount.DynamicKind != EffectDynamicAmountCount || inner.amount.Selection == nil ||
		inner.amount.Selection.Zone != zone.None {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: kind, Selection: inner.amount.Selection},
		end:    inner.end,
	}, true
}

// parseDynamicGreatestDiscardedThisWaySubject recognizes "the greatest number
// of cards a player discarded this way", the maximum per-player discard count
// produced by a preceding "each player discards their hand" effect in the same
// ability. It backs the Windfall family draw amount and carries no selection;
// the lowerer rebuilds the amount from the published discard result.
func parseDynamicGreatestDiscardedThisWaySubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start,
		"the", "greatest", "number", "of", "cards", "a", "player", "discarded", "this", "way") ||
		!dynamicAmountBoundary(tokens, start+10) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountGreatestDiscardedThisWay},
		end:    start + 10,
	}, true
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
		if !effectWordsAt(tokens, end, suffix...) {
			continue
		}
		subjectEnd := end + len(suffix)
		selectionEnd := subjectEnd
		chosenType := false
		if !dynamicAmountBoundary(tokens, subjectEnd) {
			if _, qEnd, ok := counterQualifierKind(tokens, subjectEnd); ok && dynamicAmountBoundary(tokens, qEnd) {
				subjectEnd, selectionEnd = qEnd, qEnd
			} else if cEnd, ok := dynamicCharacteristicQualifierEnd(tokens, subjectEnd, atoms); ok && dynamicAmountBoundary(tokens, cEnd) {
				subjectEnd, selectionEnd = cEnd, cEnd
			} else if cEnd, ok := chosenTypeQualifierEnd(tokens, subjectEnd); ok && dynamicAmountBoundary(tokens, cEnd) {
				subjectEnd, chosenType = cEnd, true
			} else {
				continue
			}
		}
		selection := parseSelection(tokens[start:selectionEnd], atoms)
		selection.SubtypeFromEntryChoice = chosenType
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// dynamicCharacteristicQualifierEnd recognizes a trailing "with <characteristic>
// <comparison>" qualifier on a count subject (for example "creature you control
// with power 4 or greater") and returns the token index just past it. It mirrors
// the power, toughness, and mana-value comparisons parseSelection already
// understands so the count selection carries the same filter. It fails closed for
// any other "with" qualifier.
func dynamicCharacteristicQualifierEnd(tokens []shared.Token, start int, atoms Atoms) (int, bool) {
	if !effectWordsAt(tokens, start, "with") {
		return 0, false
	}
	idx := start + 1
	switch {
	case effectWordsAt(tokens, idx, "power"), effectWordsAt(tokens, idx, "toughness"):
		idx++
	case effectWordsAt(tokens, idx, "mana", "value"):
		idx += 2
	default:
		return 0, false
	}
	if idx >= len(tokens) {
		return 0, false
	}
	if _, ok := effectNumber(tokens[idx], atoms); ok {
		if effectWordsAt(tokens, idx+1, "or", "greater") || effectWordsAt(tokens, idx+1, "or", "less") {
			return idx + 3, true
		}
		return idx + 1, true
	}
	if effectWordsAt(tokens, idx, "equal", "to") && idx+2 < len(tokens) {
		if _, ok := effectNumber(tokens[idx+2], atoms); ok {
			return idx + 3, true
		}
	}
	return 0, false
}

// chosenTypeQualifierEnd recognizes a trailing "of the chosen type" qualifier on
// a count subject ("the number of creatures you control of the chosen type") and
// returns the token index just past it. The matched permanents must share the
// creature subtype the source permanent chose as it entered (Three Tree City);
// the caller records that as Selection.SubtypeFromEntryChoice.
func chosenTypeQualifierEnd(tokens []shared.Token, start int) (int, bool) {
	if effectWordsAt(tokens, start, "of", "the", "chosen", "type") {
		return start + 4, true
	}
	return 0, false
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
	if atoms.SelectionFlagIn(token.Span, SelectionFlagTapped) ||
		atoms.SelectionFlagIn(token.Span, SelectionFlagUntapped) {
		return true
	}
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
	if _, ok := atoms.ExcludedSubtypeAt(token.Span); ok {
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
	head := tokens[headIndex]
	if _, ok := atoms.ObjectNounAt(head.Span); ok {
		return strings.HasSuffix(strings.ToLower(head.Text), "s")
	}
	if sub, ok := atoms.SubtypeAt(head.Span); ok {
		return subtypeHeadIsPlural(head.Text, sub)
	}
	return false
}

// subtypeHeadIsPlural reports whether the source spelling of a subtype count
// head ("Goblins", "Wolves") is a plural form rather than the singular the atom
// normalizes to. It compares the spelling against the singular noun forms that
// resolve back to the same subtype identity, so a genuinely singular subtype
// whose spelling ends in "s" (such as "Pegasus") is not misread as plural. This
// lets "the number of <subtype> you control" (which requires a plural head)
// recognize subtype counts while keeping "for each <subtype>" singular.
func subtypeHeadIsPlural(text string, sub types.Sub) bool {
	for _, candidate := range SingularNounForms(text) {
		if strings.EqualFold(candidate, text) {
			continue
		}
		if identity, ok := recognizeSubtypePhrase(candidate); ok && identity == sub {
			return true
		}
	}
	return false
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
	return kind, kind.Valid() && kind != counter.Finality
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
