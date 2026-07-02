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
		{[]string{"at", "end", "of", "combat"}, DelayedTimingEndOfCombat},
	} {
		start := end - len(suffix.words)
		if start >= 0 && effectWordsAt(tokens, start, suffix.words...) {
			return append(append([]shared.Token(nil), tokens[:start]...), tokens[end:]...), suffix.timing
		}
	}
	return tokens, DelayedTimingNone
}

func leadingDelayedTiming(tokens []shared.Token) DelayedTimingKind {
	if len(tokens) != 9 || tokens[8].Kind != shared.Comma {
		return DelayedTimingNone
	}
	for _, pattern := range []struct {
		words  []string
		timing DelayedTimingKind
	}{
		{[]string{"at", "the", "beginning", "of", "your", "next", "main", "phase"}, DelayedTimingNextMain},
		{[]string{"at", "the", "beginning", "of", "the", "next", "end", "step"}, DelayedTimingNextEndStep},
		{[]string{"at", "the", "beginning", "of", "your", "next", "end", "step"}, DelayedTimingNextEndStep},
	} {
		if effectWordsAt(tokens, 0, pattern.words...) {
			return pattern.timing
		}
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

// parseTokenPTVariableX reports a created token whose printed power and
// toughness are both the variable "X" ("an X/X ... token"): a word "X", a
// slash, and a word "X". The actual value of X is bound elsewhere in the
// ability ("where X is <dynamic>"); this only records that the token's P/T is
// the variable rather than fixed integers.
func parseTokenPTVariableX(kind EffectKind, tokens []shared.Token) bool {
	if kind != EffectCreate {
		return false
	}
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i+1].Kind != shared.Slash ||
			!equalWord(tokens[i], "X") ||
			!equalWord(tokens[i+2], "X") {
			continue
		}
		return true
	}
	return false
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

// parseTokenKeywordToxic returns the integer rank of a created token's toxic
// keyword ("with toxic 1" -> 1), the one parameterized creature keyword that
// appears on created tokens in the corpus. The bare keyword list (TokenKeywords)
// records that toxic is present but drops its integer; this captures the rank so
// the create-token exactness recognizer can reconstruct "toxic N" and lowering
// can grant the parameterized keyword ability. It returns 0 for non-create
// clauses and for clauses whose token carries no toxic keyword.
func parseTokenKeywordToxic(kind EffectKind, tokens []shared.Token, atoms Atoms) int {
	if kind != EffectCreate {
		return 0
	}
	for _, keyword := range scanKeywords(tokens, atoms) {
		if keyword.Kind != KeywordToxic {
			continue
		}
		if keyword.Parameter.Kind != KeywordParameterInteger {
			return 0
		}
		return keyword.Parameter.Integer()
	}
	return 0
}

// predefinedTokenNames lists the named tokens whose identity is a card name
// rather than a card subtype, so the create clause spells out neither their
// printed characteristics nor their abilities ("create a tapped Mutavault
// token."). Each name maps to a fixed token definition in lowering; the parser
// captures only these recognized names so other capitalized words in a create
// clause are unaffected. The map key is the lowercased source word; the value is
// the canonical token name.
var predefinedTokenNames = map[string]string{
	"mutavault": "Mutavault",
}

// parsePredefinedTokenName captures a created predefined named token's name from
// the noun slot of a create clause ("create a tapped Mutavault token." ->
// "Mutavault"). The name must be one of the recognized predefined-token names
// (predefinedTokenNames) and must sit immediately before the "token"/"tokens"
// noun. It returns "" for non-create clauses and for any clause whose pre-noun
// word is not a recognized predefined-token name. The create-token exactness
// recognizer reconstructs and byte-checks the "<Name> token" noun phrase, so a
// spurious capture fails closed there.
func parsePredefinedTokenName(kind EffectKind, tokens []shared.Token) string {
	if kind != EffectCreate {
		return ""
	}
	noun := -1
	for i, token := range tokens {
		if equalWord(token, "token") || equalWord(token, "tokens") {
			noun = i
			break
		}
	}
	if noun < 1 {
		return ""
	}
	if name, ok := predefinedTokenNames[strings.ToLower(tokens[noun-1].Text)]; ok {
		return name
	}
	return ""
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

// parseLeadingTokenName captures a created token's name from the leading
// "Create <Name>, a/an ... token ..." form ("Create Avacyn, a legendary 8/8
// white Angel creature token with flying, vigilance, and indestructible.") used
// by named tokens that print their name before the token description. The clause
// arrives with its "Create"/"creates" verb already stripped, so the leading name
// (when present) begins at the first token and runs to the comma that
// immediately precedes the "a"/"an" token spec. It returns "" for non-create
// clauses, for the leading-article form that names no token ("Create a ..."), and
// for any clause whose name region holds a token noun. The create-token exactness
// recognizer reconstructs and byte-checks the full clause, so any spurious capture
// fails closed there.
func parseLeadingTokenName(kind EffectKind, tokens []shared.Token) string {
	if kind != EffectCreate || len(tokens) == 0 {
		return ""
	}
	// A leading article opens the token spec directly, so no name precedes it.
	if equalWord(tokens[0], "a") || equalWord(tokens[0], "an") {
		return ""
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
		// A token noun before any comma is not the leading-name shape; fail closed
		// so other create wordings are unaffected.
		if equalWord(tokens[i], "token") || equalWord(tokens[i], "tokens") {
			return ""
		}
	}
	if comma <= 0 || comma+1 >= len(tokens) {
		return ""
	}
	if !equalWord(tokens[comma+1], "a") && !equalWord(tokens[comma+1], "an") {
		return ""
	}
	return joinedEffectText(tokens[:comma])
}

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
	if kind == EffectCounter {
		tokens = trimSpellTargetRestrictionTail(tokens)
	}
	if kind == EffectGainPlayerCounter {
		if symbols := energySymbolsAfter(tokens, 0); len(symbols) > 0 {
			return EffectAmountSyntax{
				Span:  shared.SpanOf(symbols),
				Value: len(symbols),
				Known: true,
			}
		}
		// Named player-counter word form ("an experience counter", "two poison
		// counters"): the count is an ordinary leading number/article handled by
		// the generic amount parsing below, defaulting to one.
	}
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
	if kind == EffectCreate {
		for i := 0; i+1 < len(tokens); i++ {
			if equalWord(tokens[i], "that") && equalWord(tokens[i+1], "many") {
				return EffectAmountSyntax{
					Span:        shared.SpanOf(tokens[i : i+2]),
					Text:        joinedEffectText(tokens[i : i+2]),
					DynamicKind: EffectDynamicAmountTriggeringCombatDamage,
				}
			}
		}
	}
	if kind == EffectDraw {
		for i := 0; i+2 < len(tokens); i++ {
			if equalWord(tokens[i], "that") && equalWord(tokens[i+1], "many") && equalWord(tokens[i+2], "cards") {
				return EffectAmountSyntax{
					Span:        shared.SpanOf(tokens[i : i+2]),
					Text:        joinedEffectText(tokens[i : i+2]),
					DynamicKind: EffectDynamicAmountTriggeringCounterCount,
				}
			}
		}
	}
	if kind == EffectMill {
		for i := 0; i+2 < len(tokens); i++ {
			if equalWord(tokens[i], "that") && equalWord(tokens[i+1], "many") && equalWord(tokens[i+2], "cards") {
				return EffectAmountSyntax{
					Span:        shared.SpanOf(tokens[i : i+2]),
					Text:        joinedEffectText(tokens[i : i+2]),
					DynamicKind: EffectDynamicAmountTriggeringCombatDamage,
				}
			}
		}
	}
	if kind == EffectPut {
		// "put that many <kind> counters" in a triggered ability reads the
		// quantity the trigger measured ("Whenever you discard one or more cards,
		// put that many +1/+1 counters on this creature." — Marauding Mako;
		// "Whenever a creature you control deals combat damage to a player, put
		// that many +1/+1 counters on it." — Necropolis Regent; "Whenever you gain
		// life, put that many +1/+1 counters on this creature." — Ageless Entity).
		// The generic scan below would otherwise misread the "+1" of "+1/+1" as a
		// fixed amount of one. The parser cannot see the trigger, so it records a
		// generic triggering-event amount that lowering resolves per event kind,
		// failing closed outside a measuring trigger.
		for i := 0; i+1 < len(tokens); i++ {
			if equalWord(tokens[i], "that") && equalWord(tokens[i+1], "many") {
				return EffectAmountSyntax{
					Span:        shared.SpanOf(tokens[i : i+2]),
					Text:        joinedEffectText(tokens[i : i+2]),
					DynamicKind: EffectDynamicAmountTriggeringEventAmount,
				}
			}
		}
	}
	if kind == EffectDealDamage {
		// "deals that much damage" in a counter-placement trigger reads the
		// number of counters added by the triggering event (Shalai and Hallar).
		// The "that much damage plus N" damage-increase replacement is a distinct
		// construct parsed elsewhere, so a trailing "plus" excludes it here. The
		// counter-placement gate lives in lowering (lowerEventCounterCountAmount),
		// keeping the amount closed for every other trigger event.
		for i := 0; i+2 < len(tokens); i++ {
			if !equalWord(tokens[i], "that") || !equalWord(tokens[i+1], "much") ||
				!equalWord(tokens[i+2], "damage") {
				continue
			}
			if i+3 < len(tokens) && equalWord(tokens[i+3], "plus") {
				break
			}
			return EffectAmountSyntax{
				Span:        shared.SpanOf(tokens[i : i+2]),
				Text:        joinedEffectText(tokens[i : i+2]),
				DynamicKind: EffectDynamicAmountTriggeringCounterCount,
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
			// A number that completes a "mana value N" filter qualifier on a
			// target or selection is the filter's bound, not the effect's amount
			// (e.g. "Counter target spell with mana value 1."). Skip it so the
			// counter/destroy/etc. effect keeps its unknown amount.
			if i >= 2 && equalWord(tokens[i-1], "value") && equalWord(tokens[i-2], "mana") {
				continue
			}
			// A number that completes a "with power N or less/greater" or "with
			// toughness N or less/greater" filter qualifier on a target or
			// selection is likewise the filter's bound, not the effect's amount
			// (e.g. "Return all creature cards with power 2 or less from your
			// graveyard to your hand.", Dusk // Dawn). Skip it so the return/put/
			// etc. effect keeps its unknown amount, mirroring the mana-value skip.
			if i >= 2 && equalWord(tokens[i-2], "with") &&
				(equalWord(tokens[i-1], "power") || equalWord(tokens[i-1], "toughness")) {
				continue
			}
			return EffectAmountSyntax{Span: token.Span, Value: value, Known: true}
		}
		if equalWord(token, "a") || equalWord(token, "an") || equalWord(token, "another") {
			if i > 0 && equalWord(tokens[i-1], "from") {
				continue
			}
			// The determiner "another" denotes a single object other than the
			// effect's source ("Sacrifice another creature."); its count is one,
			// exactly like "a"/"an". The "another"/"other" exclusion itself rides
			// on the selection's Another flag, consumed downstream.
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
	// addend is a fixed leading offset added to the counted amount, from the
	// "N plus the number of <count>" form ("X is 2 plus the number of artifacts
	// you control"). It is zero for every prefix without a leading addend.
	addend int
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
		end := subject.end
		if prefix.addend != 0 {
			match.Addend = prefix.addend
		}
		if addend, addendEnd, ok := parseDynamicAmountAddend(tokens, subject.end); ok {
			match.Addend = addend
			end = addendEnd
		}
		match.Span = shared.SpanOf(tokens[i:end])
		match.Text = joinedEffectText(tokens[i:end])
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
	if prefix, ok := parseLeadingAddendCountPrefix(tokens, index); ok {
		return prefix, true
	}
	switch {
	case effectWordsAt(tokens, index, "equal", "to", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 6, 2, true, true, 0}, true
	case effectWordsAt(tokens, index, "equal", "to", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 5, 1, true, true, 0}, true
	case effectWordsAt(tokens, index, "for", "each"):
		return dynamicAmountPrefix{EffectDynamicAmountFormForEach, index + 2, precedingEffectMultiplier(tokens[:index], atoms), false, true, 0}, true
	case effectWordsAt(tokens, index, "equal", "to"):
		return dynamicAmountPrefix{EffectDynamicAmountFormEqual, index + 2, 1, false, false, 0}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "twice", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 7, 2, true, true, 0}, true
	case effectWordsAt(tokens, index, "where", "X", "is", "the", "number", "of"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 6, 1, true, true, 0}, true
	case effectWordsAt(tokens, index, "where", "X", "is"):
		return dynamicAmountPrefix{EffectDynamicAmountFormWhereX, index + 3, 1, false, false, 0}, true
	default:
		return dynamicAmountPrefix{}, false
	}
}

// parseLeadingAddendCountPrefix recognizes a dynamic count amount whose count is
// offset by a fixed leading addend: "where X is N plus the number of <count>" or
// "equal to N plus the number of <count>" (Welding Sparks: "X is 3 plus the
// number of artifacts you control"; Galvanic Bombardment, Thunder Salvo). The
// "N plus" sits between the form keyword and the "the number of" count phrase,
// the leading mirror of the trailing "the number of <count> plus N" addend
// parseDynamicAmountAddend already consumes. It returns a prefix positioned just
// past "the number of", carrying the leading addend and the count's multiplier
// of one. It fails closed for every wording without the exact
// "<form> <N> plus the number of" run, so amounts without a leading addend keep
// their existing prefixes.
func parseLeadingAddendCountPrefix(tokens []shared.Token, index int) (dynamicAmountPrefix, bool) {
	var form EffectDynamicAmountForm
	var addendIndex int
	switch {
	case effectWordsAt(tokens, index, "where", "X", "is"):
		form = EffectDynamicAmountFormWhereX
		addendIndex = index + 3
	case effectWordsAt(tokens, index, "equal", "to"):
		form = EffectDynamicAmountFormEqual
		addendIndex = index + 2
	default:
		return dynamicAmountPrefix{}, false
	}
	if addendIndex >= len(tokens) {
		return dynamicAmountPrefix{}, false
	}
	addend, ok := addendCardinal(tokens[addendIndex])
	if !ok || addend < 1 {
		return dynamicAmountPrefix{}, false
	}
	if !effectWordsAt(tokens, addendIndex+1, "plus", "the", "number", "of") {
		return dynamicAmountPrefix{}, false
	}
	return dynamicAmountPrefix{
		form:       form,
		start:      addendIndex + 5,
		multiplier: 1,
		plural:     true,
		count:      true,
		addend:     addend,
	}, true
}

func precedingEffectMultiplier(tokens []shared.Token, atoms Atoms) int {
	multiplier := 0
	for i, token := range tokens {
		if tokenInCounterAtom(token, atoms) || powerToughnessDigit(tokens, i) {
			continue
		}
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

func tokenInCounterAtom(token shared.Token, atoms Atoms) bool {
	for _, atom := range atoms.Counters() {
		if spanCovers(atom.Span, token.Span) {
			return true
		}
	}
	return false
}

// powerToughnessDigit reports whether the token at index i is one of the two
// integers in a created token's printed "<power>/<toughness>" (an integer
// adjacent to a slash that joins two integers). Such digits describe the token's
// size, not a token count, so a "Create a 2/2 ... token for each <iterator>"
// clause must not mistake the printed "2" for a per-iteration multiplier.
func powerToughnessDigit(tokens []shared.Token, i int) bool {
	if tokens[i].Kind != shared.Integer {
		return false
	}
	if i+2 < len(tokens) && tokens[i+1].Kind == shared.Slash && tokens[i+2].Kind == shared.Integer {
		return true
	}
	if i-2 >= 0 && tokens[i-1].Kind == shared.Slash && tokens[i-2].Kind == shared.Integer {
		return true
	}
	return false
}

// parseDynamicMaxSubject recognizes the "whichever is greater" combinator over
// two rules-derived amounts: "<A> or <B>, whichever is greater" (Willowdusk,
// Essence Seer). It produces an EffectDynamicAmountMaxOf amount carrying the two
// operand subjects, each parsed as a complete standalone subject so it reuses
// every recognized amount form. It splits the region before the trailing
// "whichever is greater" at the "or" that yields two fully-recognized operands,
// requiring the operands to agree on whether they are counts. It fails closed
// when no trailing "whichever is greater" is present, when no split yields two
// recognized operands, or when the operands disagree on count shape, so any
// unrecognized combinator wording keeps the amount unrecognized.
func parseDynamicMaxSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	tail := -1
	for i := start; i+2 < len(tokens); i++ {
		if equalWord(tokens[i], "whichever") && equalWord(tokens[i+1], "is") &&
			equalWord(tokens[i+2], "greater") {
			tail = i
			break
		}
	}
	if tail < 0 {
		return dynamicAmountSubject{}, false
	}
	end := tail + 3
	if !dynamicAmountBoundary(tokens, end) {
		return dynamicAmountSubject{}, false
	}
	region := tail
	if region-1 > start && tokens[region-1].Kind == shared.Comma {
		region--
	}
	for split := start + 1; split < region-1; split++ {
		if !equalWord(tokens[split], "or") {
			continue
		}
		left := tokens[start:split]
		right := tokens[split+1 : region]
		leftSub, leftOK := parseDynamicAmountSubject(left, 0, atoms)
		if !leftOK || leftSub.end != len(left) {
			continue
		}
		rightSub, rightOK := parseDynamicAmountSubject(right, 0, atoms)
		if !rightOK || rightSub.end != len(right) {
			continue
		}
		if leftSub.count != rightSub.count || leftSub.plural != rightSub.plural {
			continue
		}
		leftAmount := leftSub.amount
		leftAmount.Multiplier = 1
		rightAmount := rightSub.amount
		rightAmount.Multiplier = 1
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{
				DynamicKind: EffectDynamicAmountMaxOf,
				Operands:    []EffectAmountSyntax{leftAmount, rightAmount},
			},
			end:    end,
			count:  leftSub.count,
			plural: leftSub.plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// parseDynamicAmountSubjectHelper runs the dedicated dynamic-amount subject
// recognizers in priority order, returning the first that matches. It is split
// out of parseDynamicAmountSubject so the dispatch chain stays within the
// maintainability budget; the remaining inline switch handles the shorter
// keyword-phrase forms.
func parseDynamicAmountSubjectHelper(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if subject, ok := parseDynamicMaxSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicSourceCounterCount(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicGreatestCharacteristicSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicSharedCreatureTypeCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicTotalCharacteristicSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicColorCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicColorsOfManaSpentSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicTimesKickedSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicGreatestDiscardedThisWaySubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicDestroyedThisWaySubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicCardsNamedSelfInGraveyardsSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicDevotionSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicSpellsCastThisTurnSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseDynamicLifeChangedThisTurnSubject(tokens, start); ok {
		return subject, true
	}
	if subject, ok := parseSacrificedCreatureCharacteristic(tokens, start); ok {
		return subject, true
	}
	return dynamicAmountSubject{}, false
}

func parseDynamicAmountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if start >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	if subject, ok := parseDynamicAmountSubjectHelper(tokens, start, atoms); ok {
		return subject, true
	}
	switch {
	case effectWordsAt(tokens, start, "your", "life", "total") && dynamicAmountBoundary(tokens, start+3):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountControllerLife},
			end:    start + 3,
		}, true
	case effectWordsAt(tokens, start, "your", "speed") && dynamicAmountBoundary(tokens, start+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountControllerSpeed},
			end:    start + 2,
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
		effectWordsAt(tokens, start+2, "power") &&
		dynamicAmountBoundary(tokens, start+3):
		// "that <object>'s power" names the power of a referenced permanent (the
		// prior clause's permanent, or the triggering creature of an enters
		// trigger). The reference spans the possessive object phrase ("that
		// creature's"); collectReferences recognizes the same span so the
		// amount's referent binds to the antecedent. It backs "deals damage
		// equal to that creature's power" (Terror of the Peaks).
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourcePower, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
			end:    start + 3,
		}, true
	case start+1 < len(tokens) && equalWord(tokens[start], "that") &&
		referencePossessiveObjectNoun(tokens[start+1]) &&
		effectWordsAt(tokens, start+2, "toughness") &&
		dynamicAmountBoundary(tokens, start+3):
		// "that <object>'s toughness" is the toughness sibling of the
		// "that <object>'s power" referenced-object amount above.
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSourceToughness, ReferenceSpan: shared.SpanOf(tokens[start : start+2])},
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
	case effectWordsAt(tokens, start, "the", "excess", "damage", "dealt", "this", "way") &&
		dynamicAmountBoundary(tokens, start+6):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountExcessDamageDealtThisWay},
			end:    start + 6,
		}, true
	case effectWordsAt(tokens, start, "the", "damage", "dealt", "this", "way") &&
		dynamicAmountBoundary(tokens, start+5):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountDamageDealtThisWay},
			end:    start + 5,
		}, true
	case effectWordsAt(tokens, start, "the", "result") &&
		dynamicAmountBoundary(tokens, start+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountDieRollResult},
			end:    start + 2,
		}, true
	case effectWordsAt(tokens, start, "basic", "land", "type", "among", "lands", "you", "control") &&
		dynamicAmountBoundary(tokens, start+7):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountBasicLandTypes},
			end:    start + 7, count: true,
		}, true
	case effectWordsAt(tokens, start, "time", "you've", "cast", "your", "commander", "from", "the", "command", "zone", "this", "game") &&
		dynamicAmountBoundary(tokens, start+11):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCommanderCastCount},
			end:    start + 11, count: true,
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
	if end < len(tokens) && dynamicAmountBoundary(tokens, end+1) {
		if kind, ok := selfNameCharacteristicKind(tokens[end]); ok {
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: kind, ReferenceSpan: nameSpan},
				end:    end + 1,
			}, true
		}
	}
	if end+2 < len(tokens) && tokens[end].Kind == shared.Apostrophe &&
		equalWord(tokens[end+1], "s") && dynamicAmountBoundary(tokens, end+3) {
		if kind, ok := selfNameCharacteristicKind(tokens[end+2]); ok {
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: kind, ReferenceSpan: nameSpan},
				end:    end + 3,
			}, true
		}
	}
	return dynamicAmountSubject{}, false
}

// selfNameCharacteristicKind maps the characteristic noun following a card's own
// name in a possessive amount ("<Name>'s power", "<Name>'s toughness") to its
// dynamic-amount kind. Power backs counter-scaled mana like Marwyn, the
// Nurturer; toughness backs group enters-with-counters quantities like Arwen,
// Weaver of Hope. Any other noun fails closed.
func selfNameCharacteristicKind(token shared.Token) (EffectDynamicAmountKind, bool) {
	switch {
	case equalWord(token, "power"):
		return EffectDynamicAmountSourcePower, true
	case equalWord(token, "toughness"):
		return EffectDynamicAmountSourceToughness, true
	default:
		return EffectDynamicAmountNone, false
	}
}

// parseSacrificedCreatureCharacteristic recognizes "the sacrificed creature's
// power/toughness/mana value" — a back-reference to the permanent sacrificed to
// pay the enclosing activated ability's cost (Altar of Dementia). The possessive
// is accepted both as a single "creature's" token and as the "creature" "'" "s"
// apostrophe split. It fails closed when the trailing characteristic is not one
// of the three recognized words.
func parseSacrificedCreatureCharacteristic(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "the", "sacrificed") {
		return dynamicAmountSubject{}, false
	}
	possessiveEnd := start + 2
	switch {
	case possessiveEnd < len(tokens) && strings.EqualFold(tokens[possessiveEnd].Text, "creature's"):
		possessiveEnd++
	case possessiveEnd+2 < len(tokens) && equalWord(tokens[possessiveEnd], "creature") &&
		tokens[possessiveEnd+1].Kind == shared.Apostrophe && equalWord(tokens[possessiveEnd+2], "s"):
		possessiveEnd += 3
	default:
		return dynamicAmountSubject{}, false
	}
	switch {
	case effectWordsAt(tokens, possessiveEnd, "power") && dynamicAmountBoundary(tokens, possessiveEnd+1):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSacrificedPower},
			end:    possessiveEnd + 1,
		}, true
	case effectWordsAt(tokens, possessiveEnd, "toughness") && dynamicAmountBoundary(tokens, possessiveEnd+1):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSacrificedToughness},
			end:    possessiveEnd + 1,
		}, true
	case effectWordsAt(tokens, possessiveEnd, "mana", "value") && dynamicAmountBoundary(tokens, possessiveEnd+2):
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountSacrificedManaValue},
			end:    possessiveEnd + 2,
		}, true
	default:
		return dynamicAmountSubject{}, false
	}
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

// parseDynamicLifeChangedThisTurnSubject recognizes the controller's total life
// gained or lost so far this turn as a dynamic amount subject ("the life you've
// lost this turn", "the (amount of )?life you (have )?(gained|lost) this turn").
// Damage to the controller counts toward the life-lost total because dealing
// damage to a player causes that player to lose that much life (CR 120.3). It
// backs Children of Korlis ("You gain life equal to the life you've lost this
// turn") and the life-tracking family. It is controller-scoped: the "you" names
// the resolving ability's controller, so the subject attaches no referent. It
// fails closed on any other wording.
func parseDynamicLifeChangedThisTurnSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "the") {
		return dynamicAmountSubject{}, false
	}
	idx := start + 1
	switch {
	case effectWordsAt(tokens, idx, "amount", "of", "life"):
		idx += 3
	case effectWordsAt(tokens, idx, "life"):
		idx++
	default:
		return dynamicAmountSubject{}, false
	}
	switch {
	case effectWordsAt(tokens, idx, "you've"):
		idx++
	case effectWordsAt(tokens, idx, "you", "have"):
		idx += 2
	case effectWordsAt(tokens, idx, "you"):
		idx++
	default:
		return dynamicAmountSubject{}, false
	}
	var kind EffectDynamicAmountKind
	switch {
	case effectWordsAt(tokens, idx, "lost"):
		kind = EffectDynamicAmountLifeLostThisTurn
	case effectWordsAt(tokens, idx, "gained"):
		kind = EffectDynamicAmountLifeGainedThisTurn
	default:
		return dynamicAmountSubject{}, false
	}
	idx++
	if !effectWordsAt(tokens, idx, "this", "turn") {
		return dynamicAmountSubject{}, false
	}
	idx += 2
	if !dynamicAmountBoundary(tokens, idx) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: kind},
		end:    idx,
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

// parseDynamicSharedCreatureTypeCountSubject recognizes the "other [attacking]
// creature <scope> that shares [at least one | a] creature type with it" count
// subject (Coat of Arms: "for each other creature on the battlefield that shares
// a creature type with it"; Shared Animosity: "for each other attacking creature
// that shares a creature type with it"), the number of other creatures in a
// group that share a creature type with the affected permanent. The scope after
// the creature noun selects the group ("on the battlefield" for every creature,
// "you control" for the controller's creatures); an "attacking" adjective scopes
// the group to attacking creatures and needs no suffix. The recognized selection
// is carried on the amount so the lowerer can rebuild the group. The "other"
// qualifier is intentionally dropped from the selection — the affected permanent
// is excluded at resolution, not by the group filter. It fails closed for any
// other wording so unsupported phrasings stay rejected.
func parseDynamicSharedCreatureTypeCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "other") {
		return dynamicAmountSubject{}, false
	}
	nounStart := start + 1
	matched := false
	// An "attacking" combat adjective defines the counted group by combat
	// involvement, so it stands in for a controller or battlefield scope suffix
	// (Shared Animosity: "for each other attacking creature that shares a
	// creature type with it"). The adjective is carried into the parsed selection
	// so the lowerer scopes the count to attacking creatures.
	if effectWordsAt(tokens, nounStart, "attacking") {
		nounStart++
		matched = true
	}
	if !effectWordsAt(tokens, nounStart, "creature") {
		return dynamicAmountSubject{}, false
	}
	scopeEnd := nounStart + 1
	for _, suffix := range [][]string{{"on", "the", "battlefield"}, {"you", "control"}} {
		if effectWordsAt(tokens, scopeEnd, suffix...) {
			scopeEnd += len(suffix)
			matched = true
			break
		}
	}
	if !matched || !effectWordsAt(tokens, scopeEnd, "that", "shares") {
		return dynamicAmountSubject{}, false
	}
	idx := scopeEnd + 2
	switch {
	case effectWordsAt(tokens, idx, "at", "least", "one"):
		idx += 3
	case effectWordsAt(tokens, idx, "a"):
		idx++
	default:
	}
	if !effectWordsAt(tokens, idx, "creature", "type", "with", "it") ||
		!dynamicAmountBoundary(tokens, idx+4) {
		return dynamicAmountSubject{}, false
	}
	selection := parseSelection(tokens[start+1:scopeEnd], atoms)
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{
			DynamicKind:   EffectDynamicAmountSharedCreatureTypeCount,
			Selection:     &selection,
			ReferenceSpan: tokens[idx+3].Span,
		},
		end: idx + 4, count: true,
	}, true
}

// parseDynamicTotalCharacteristicSubject recognizes "the total <power |
// toughness> of <group>" amount subjects (Ghalta, Primal Hunger: the total
// power of creatures you control). The group is any battlefield count subject,
// parsed by reusing the count-subject scanners; the recognized selection is
// carried on the amount so the lowerer can rebuild the battlefield group. It
// fails closed for non-battlefield groups (a zone-qualified count) so
// unsupported wordings stay rejected.
func parseDynamicTotalCharacteristicSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "the", "total") {
		return dynamicAmountSubject{}, false
	}
	var kind EffectDynamicAmountKind
	var groupStart int
	switch {
	case effectWordsAt(tokens, start+2, "power", "of"):
		kind, groupStart = EffectDynamicAmountTotalPower, start+4
	case effectWordsAt(tokens, start+2, "toughness", "of"):
		kind, groupStart = EffectDynamicAmountTotalToughness, start+4
	case effectWordsAt(tokens, start+2, "mana", "value", "of"):
		kind, groupStart = EffectDynamicAmountTotalManaValue, start+5
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

// parseDynamicColorsOfManaSpentSubject recognizes the Converge amount subject
// "color of mana spent to cast it" (Crystalline Crawler: "enters with a +1/+1
// counter on it for each color of mana spent to cast it"), the number of
// distinct colors of mana spent to cast the source spell. It carries no
// selection; the runtime records the colors of mana spent as the spell is cast.
// The singular "color" pairs with a "for each" prefix, matching the count-number
// agreement the dynamic-amount dispatcher enforces. It fails closed on any other
// wording so unrelated "mana spent" phrasings stay rejected.
func parseDynamicColorsOfManaSpentSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "color", "of", "mana", "spent", "to", "cast", "it") ||
		!dynamicAmountBoundary(tokens, start+7) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountColorsOfManaSpent},
		end:    start + 7, count: true,
	}, true
}

// parseDynamicTimesKickedSubject recognizes the Multikicker amount subject
// "time it was kicked" / "time this spell was kicked" (CR 702.32), the number
// of times the spell was kicked. It backs "for each time it was kicked" amounts
// such as Everflowing Chalice's enters-with-counters quantity and Wolfbriar
// Elemental's Wolf-token count. It carries no selection; the runtime records
// the kick count as the spell is cast. The singular "time" pairs with a
// "for each" prefix, matching the count-number agreement the dynamic-amount
// dispatcher enforces. It fails closed on any other wording.
func parseDynamicTimesKickedSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	if effectWordsAt(tokens, start, "time", "it", "was", "kicked") &&
		dynamicAmountBoundary(tokens, start+4) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountTimesKicked},
			end:    start + 4, count: true,
		}, true
	}
	if effectWordsAt(tokens, start, "time", "this", "spell", "was", "kicked") &&
		dynamicAmountBoundary(tokens, start+5) {
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountTimesKicked},
			end:    start + 5, count: true,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// parseDynamicColorCountSubject recognizes the "color among <group>" /
// "colors among <group>" amount subject (Faeburrow Elder: "+1/+1 for each color
// among permanents you control"), the number of distinct colors found among the
// permanents of a battlefield group. The group after "among" is parsed by
// reusing the count-subject scanners, so it stays generic over the permanent
// filter; the recognized selection is carried on the amount so the lowerer can
// rebuild the battlefield group. The singular "color" pairs with a "for each"
// prefix and the plural "colors" with a "number of" prefix, matching the count
// number agreement the dynamic-amount dispatcher enforces. It fails closed for a
// non-battlefield group (a zone-qualified count) so unsupported wordings stay
// rejected.
func parseDynamicColorCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	var plural bool
	var groupStart int
	switch {
	case effectWordsAt(tokens, start, "color", "among"):
		groupStart = start + 2
	case effectWordsAt(tokens, start, "colors", "among"):
		plural, groupStart = true, start+2
	default:
		return dynamicAmountSubject{}, false
	}
	inner, ok := parseDynamicCountSubject(tokens, groupStart, atoms)
	if !ok || inner.amount.DynamicKind != EffectDynamicAmountCount || inner.amount.Selection == nil ||
		inner.amount.Selection.Zone != zone.None {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountColorCount, Selection: inner.amount.Selection},
		end:    inner.end, count: true, plural: plural,
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

// parseDynamicDestroyedThisWaySubject recognizes "<permanent noun> destroyed
// this way", the count of permanents destroyed by a preceding destroy effect in
// the same ability ("for each creature destroyed this way", "for each permanent
// destroyed this way"). The noun names a permanent card type or the catch-all
// "permanent" and is descriptive of what the prior clause destroyed, so the
// count carries no selection; the lowerer reads the count the destroy effect
// publishes. It fails closed on any other noun or trailing text.
func parseDynamicDestroyedThisWaySubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	plural, ok := destroyedThisWayNounPlural(tokens[start])
	if !ok {
		return dynamicAmountSubject{}, false
	}
	end := start + 1
	if !effectWordsAt(tokens, end, "destroyed", "this", "way") || !dynamicAmountBoundary(tokens, end+3) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountDestroyedThisWay},
		end:    end + 3, count: true, plural: plural,
	}, true
}

// parseDynamicCardsNamedSelfInGraveyardsSubject recognizes the count subject
// "card named <this card> in each graveyard" (Rite of Flame: "add {R} for each
// card named Rite of Flame in each graveyard.") and its controller-scoped
// sibling "card named <this card> in your graveyard" (Compound Fracture, Growth
// Cycle). It accepts only the card's own name, matched through an Atoms
// self-name span beginning right after "named", so a literal other-card name
// fails closed. The subject is a singular count, so it pairs with the singular
// "for each" prefix. The "in each graveyard" wording counts every graveyard;
// "in your graveyard" counts only the controller's.
func parseDynamicCardsNamedSelfInGraveyardsSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "card", "named") {
		return dynamicAmountSubject{}, false
	}
	nameStart := start + 2
	if nameStart >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[nameStart].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	nameEnd := nameStart
	for nameEnd < len(tokens) && tokens[nameEnd].Span.End.Offset <= nameSpan.End.Offset {
		nameEnd++
	}
	if nameEnd == nameStart {
		return dynamicAmountSubject{}, false
	}
	var kind EffectDynamicAmountKind
	switch {
	case effectWordsAt(tokens, nameEnd, "in", "each", "graveyard"):
		kind = EffectDynamicAmountCardsNamedSelfInGraveyards
	case effectWordsAt(tokens, nameEnd, "in", "your", "graveyard"):
		kind = EffectDynamicAmountCardsNamedSelfInControllerGraveyard
	default:
		return dynamicAmountSubject{}, false
	}
	if !dynamicAmountBoundary(tokens, nameEnd+3) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: kind},
		end:    nameEnd + 3, count: true,
	}, true
}

// destroyedThisWayNounPlural reports whether token is a permanent-naming noun
// admissible in a "<noun> destroyed this way" count subject and whether it is
// the plural spelling. The recognized nouns are the permanent card types and the
// catch-all "permanent"; any other word fails closed.
func destroyedThisWayNounPlural(token shared.Token) (plural, ok bool) {
	switch strings.ToLower(token.Text) {
	case "permanent", "creature", "artifact", "enchantment", "land", "planeswalker":
		return false, true
	case "permanents", "creatures", "artifacts", "enchantments", "lands", "planeswalkers":
		return true, true
	default:
		return false, false
	}
}

func parseDynamicCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if equalWord(tokens[start], "card") || equalWord(tokens[start], "cards") {
		if subject, ok := parseDynamicEventCardCountSubject(tokens, start); ok {
			return subject, true
		}
		if subject, ok := parseDynamicCardsDrawnThisTurnSubject(tokens, start); ok {
			return subject, true
		}
		if subject, ok := parseDynamicCardCountSubject(tokens, start, atoms); ok {
			return subject, true
		}
	}
	if subject, ok := parseDynamicTypeUnionCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicInstantSorceryCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicObjectNounCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	if subject, ok := parseDynamicAttackingCreatureCountSubject(tokens, start, atoms); ok {
		return subject, true
	}
	return parseDynamicSelectionCountSubject(tokens, start, atoms)
}

// parseDynamicAttackingCreatureCountSubject recognizes the "[other] attacking
// <creature-noun-or-subtype>" count subject of an attack-triggered self-pump
// scaled by the attacking force ("for each other attacking Goblin", Goblin
// Piledriver; "for each other attacking Aurochs", Aurochs; "for each other
// attacking creature", Rampaging Classmate; "for each attacking creature",
// Charging Hooligan). The "attacking" combat adjective scopes the counted group
// to attacking creatures. An optional "you control" suffix narrows that combat
// group to the controller's attackers; no other suffix is permitted. The head
// must be the bare "creature" noun or a single creature subtype, so a battle,
// planeswalker, or richer trailing qualifier fails closed.
// The optional leading "other" self-exclusion is kept in the span handed to
// parseSelection so it records the Other flag, which excludes the attacking
// permanent itself at resolution rather than through the group filter.
func parseDynamicAttackingCreatureCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	scanStart := start
	if equalWord(tokens[start], "other") {
		scanStart = start + 1
	}
	if !effectWordsAt(tokens, scanStart, "attacking") {
		return dynamicAmountSubject{}, false
	}
	nounIndex := scanStart + 1
	if nounIndex >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	headWord := false
	if noun, ok := atoms.ObjectNounAt(tokens[nounIndex].Span); ok && noun == ObjectNounCreature {
		headWord = true
	} else if _, ok := atoms.SubtypeAt(tokens[nounIndex].Span); ok {
		headWord = true
	}
	end := nounIndex + 1
	if effectWordsAt(tokens, end, "you", "control") {
		end += 2
	}
	if !headWord || !dynamicAmountBoundary(tokens, end) {
		return dynamicAmountSubject{}, false
	}
	plural := dynamicCountHeadPlural(tokens, nounIndex, atoms)
	selection := parseSelection(tokens[start:end], atoms)
	if selection.Zone != zone.None || !selection.Attacking || selection.Blocking {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
		end:    end, count: true, plural: plural,
	}, true
}

// parseDynamicTypeUnionCountSubject recognizes a "for each" count subject whose
// matched permanents satisfy a disjunction of card types joined by "or" or
// "and/or" ("artifact and/or enchantment you control", "creature or artifact you
// control"). Each alternative must be a counting card-type noun (artifact,
// creature, enchantment, or land); the explicit connector distinguishes this
// union from a bare "artifact creature" type conjunction, which stays
// unsupported. The resulting count selection carries the types as a
// RequiredTypesAny union so the runtime counts a permanent once when it matches
// any listed type.
func parseDynamicTypeUnionCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if _, ok := dynamicUnionCardTypeAt(tokens, start, atoms); !ok {
		return dynamicAmountSubject{}, false
	}
	idx := start + 1
	connectors := 0
	for {
		next, ok := consumeDynamicUnionConnector(tokens, idx)
		if !ok {
			break
		}
		if _, ok := dynamicUnionCardTypeAt(tokens, next, atoms); !ok {
			return dynamicAmountSubject{}, false
		}
		idx = next + 1
		connectors++
	}
	if connectors == 0 {
		return dynamicAmountSubject{}, false
	}
	for _, suffix := range [][]string{{"you", "control"}, {"your", "opponents", "control"}, {"on", "the", "battlefield"}} {
		if !effectWordsAt(tokens, idx, suffix...) || !dynamicAmountBoundary(tokens, idx+len(suffix)) {
			continue
		}
		subjectEnd := idx + len(suffix)
		selection := parseSelection(tokens[start:subjectEnd], atoms)
		if len(selection.RequiredTypesAny) < 2 {
			return dynamicAmountSubject{}, false
		}
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// parseDynamicInstantSorceryCountSubject recognizes a "for each instant and
// sorcery card[s] in your graveyard/hand" count subject. "Instant and sorcery
// cards" is a fixed idiom for cards that are instants or sorceries (no card is
// both), so the bare "and" denotes a disjunction the runtime counts once per
// matching card through a RequiredTypesAny union. Only the controller's own
// graveyard and hand, where instant and sorcery cards exist, are recognized;
// other zones, type pairs, and connectors fail closed.
func parseDynamicInstantSorceryCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	if start+3 >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	first, ok := atoms.CardTypeAt(tokens[start].Span)
	if !ok || !equalWord(tokens[start+1], "and") {
		return dynamicAmountSubject{}, false
	}
	second, ok := atoms.CardTypeAt(tokens[start+2].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	if !instantSorceryPair(first, second) {
		return dynamicAmountSubject{}, false
	}
	headIndex := start + 3
	if headIndex >= len(tokens) ||
		(!equalWord(tokens[headIndex], "card") && !equalWord(tokens[headIndex], "cards")) {
		return dynamicAmountSubject{}, false
	}
	plural := strings.EqualFold(tokens[headIndex].Text, "cards")
	end := headIndex + 1
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
		if len(selection.RequiredTypesAny) != 2 {
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

// instantSorceryPair reports whether the two card types are exactly the instant
// and sorcery pair (in either order).
func instantSorceryPair(a, b CardType) bool {
	return (a == CardTypeInstant && b == CardTypeSorcery) ||
		(a == CardTypeSorcery && b == CardTypeInstant)
}

// dynamicUnionCardTypeAt returns the counting card type beginning at index when
// the token names artifact, creature, enchantment, or land. Other card types
// (which the battlefield count lowering cannot represent as a required type) and
// non-card-type tokens fail closed.
func dynamicUnionCardTypeAt(tokens []shared.Token, index int, atoms Atoms) (CardType, bool) {
	if index >= len(tokens) {
		return "", false
	}
	cardType, ok := atoms.CardTypeAt(tokens[index].Span)
	if !ok {
		return "", false
	}
	switch cardType {
	case CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypeLand:
		return cardType, true
	default:
		return "", false
	}
}

// consumeDynamicUnionConnector reports whether an "or" or "and/or" connector
// begins at index and returns the index just past it. The lexer splits "and/or"
// into the words "and" and "or" around a Slash symbol.
func consumeDynamicUnionConnector(tokens []shared.Token, index int) (int, bool) {
	if effectWordsAt(tokens, index, "or") {
		return index + 1, true
	}
	if index+2 < len(tokens) && equalWord(tokens[index], "and") &&
		tokens[index+1].Kind == shared.Slash && equalWord(tokens[index+2], "or") {
		return index + 3, true
	}
	return 0, false
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

// parseDynamicCardsDrawnThisTurnSubject recognizes the controller's total cards
// drawn so far this turn as a count subject ("card[s] you've drawn this turn",
// "card[s] you have drawn this turn"). It backs "the number of cards you've
// drawn this turn" amounts (Thundering Djinn's attack-trigger damage). It is
// controller-scoped: the "you" names the resolving ability's controller, so the
// subject attaches no referent. The triggering or just-resolved draw counts,
// since its draw event precedes the resolving ability. It fails closed on any
// other wording (an opponent's draws, a trailing qualifier).
func parseDynamicCardsDrawnThisTurnSubject(tokens []shared.Token, start int) (dynamicAmountSubject, bool) {
	plural := false
	switch {
	case equalWord(tokens[start], "card"):
	case equalWord(tokens[start], "cards"):
		plural = true
	default:
		return dynamicAmountSubject{}, false
	}
	end := start + 1
	switch {
	case effectWordsAt(tokens, end, "you've", "drawn", "this", "turn"):
		end += 4
	case effectWordsAt(tokens, end, "you", "have", "drawn", "this", "turn"):
		end += 5
	default:
		return dynamicAmountSubject{}, false
	}
	if !dynamicAmountBoundary(tokens, end) {
		return dynamicAmountSubject{}, false
	}
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCardsDrawnThisTurn},
		end:    end, count: true, plural: plural,
	}, true
}

func parseDynamicObjectNounCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	nounStart := start
	if equalWord(tokens[start], "other") {
		nounStart = start + 1
	}
	if nounStart >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	noun, ok := atoms.ObjectNounAt(tokens[nounStart].Span)
	if !ok {
		return dynamicAmountSubject{}, false
	}
	plural := strings.HasSuffix(strings.ToLower(tokens[nounStart].Text), "s")
	if noun == ObjectNounOpponent {
		end := nounStart + 1
		if effectWordsAt(tokens, end, "you", "attacked", "this", "combat") &&
			dynamicAmountBoundary(tokens, end+4) {
			// "opponent you attacked this combat" is the Melee count: the number
			// of the controller's opponents being attacked this combat. It is a
			// distinct controller-scoped, combat-state amount, not a board count.
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountOpponentsAttackedThisCombat},
				end:    end + 4, count: true, plural: plural,
			}, true
		}
		if effectWordsAt(tokens, end, "you", "have") {
			end += 2
		}
		if dynamicAmountBoundary(tokens, end) {
			return dynamicAmountSubject{
				amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountOpponentCount},
				end:    end, count: true, plural: plural,
			}, true
		}
		if subject, ok := parseOpponentControllingCountSubject(tokens, nounStart+1, plural, atoms); ok {
			return subject, true
		}
		return dynamicAmountSubject{}, false
	}
	if !slices.Contains([]ObjectNoun{
		ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand, ObjectNounPermanent,
	}, noun) {
		return dynamicAmountSubject{}, false
	}
	end := nounStart + 1
	for _, suffix := range [][]string{{"you", "control"}, {"your", "opponents", "control"}, {"on", "the", "battlefield"}} {
		if !effectWordsAt(tokens, end, suffix...) {
			continue
		}
		subjectEnd := end + len(suffix)
		selectionEnd := subjectEnd
		chosenType := false
		chosenResolutionType := false
		if !dynamicAmountBoundary(tokens, subjectEnd) {
			if match, ok := counterQualifierKind(tokens, subjectEnd); ok && dynamicAmountBoundary(tokens, match.End) {
				subjectEnd, selectionEnd = match.End, match.End
			} else if cEnd, ok := dynamicCharacteristicQualifierEnd(tokens, subjectEnd, atoms); ok && dynamicAmountBoundary(tokens, cEnd) {
				subjectEnd, selectionEnd = cEnd, cEnd
			} else if cEnd, ok := chosenTypeQualifierEnd(tokens, subjectEnd); ok && dynamicAmountBoundary(tokens, cEnd) {
				subjectEnd, chosenType = cEnd, true
			} else if cEnd, ok := thatTypeQualifierEnd(tokens, subjectEnd); ok && dynamicAmountBoundary(tokens, cEnd) {
				subjectEnd, chosenResolutionType = cEnd, true
			} else {
				continue
			}
		}
		selection := parseSelection(tokens[start:selectionEnd], atoms)
		selection.SubtypeFromEntryChoice = chosenType
		selection.SubtypeFromChosenType = chosenResolutionType
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountCount, Selection: &selection},
			end:    subjectEnd, count: true, plural: plural,
		}, true
	}
	return dynamicAmountSubject{}, false
}

// parseOpponentControllingCountSubject recognizes the per-opponent control
// predicate "opponents who control <selection>" ("opponents who control a
// creature with power 4 or greater", Summon: Yojimbo chapter IV), counting the
// controller's opponents that control at least one matching permanent. start is
// the token index just past "opponents". The controlled selection is an
// optionally articled permanent noun with an optional trailing characteristic
// qualifier, parsed by parseSelection and scoped to "you control" so it resolves
// relative to each counted opponent. It fails closed for any other trailing
// wording.
func parseOpponentControllingCountSubject(tokens []shared.Token, start int, plural bool, atoms Atoms) (dynamicAmountSubject, bool) {
	if !effectWordsAt(tokens, start, "who", "control") {
		return dynamicAmountSubject{}, false
	}
	selStart := start + 2
	nounIdx := selStart
	if effectWordsAt(tokens, nounIdx, "a") || effectWordsAt(tokens, nounIdx, "an") {
		nounIdx++
	}
	if nounIdx >= len(tokens) {
		return dynamicAmountSubject{}, false
	}
	noun, ok := atoms.ObjectNounAt(tokens[nounIdx].Span)
	if !ok || !slices.Contains([]ObjectNoun{
		ObjectNounArtifact, ObjectNounCreature, ObjectNounEnchantment, ObjectNounLand, ObjectNounPermanent,
	}, noun) {
		return dynamicAmountSubject{}, false
	}
	selEnd := nounIdx + 1
	if !dynamicAmountBoundary(tokens, selEnd) {
		cEnd, ok := dynamicCharacteristicQualifierEnd(tokens, selEnd, atoms)
		if !ok || !dynamicAmountBoundary(tokens, cEnd) {
			return dynamicAmountSubject{}, false
		}
		selEnd = cEnd
	}
	selection := parseSelection(tokens[selStart:selEnd], atoms)
	selection.Controller = SelectionControllerYou
	return dynamicAmountSubject{
		amount: EffectAmountSyntax{DynamicKind: EffectDynamicAmountOpponentControllingCount, Selection: &selection},
		end:    selEnd, count: true, plural: plural,
	}, true
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

// thatTypeQualifierEnd recognizes a trailing "of that type" qualifier on a count
// subject ("each permanent you control of that type") and returns the token index
// just past it. The matched permanents must share the creature subtype chosen
// earlier in the same resolution by a "Choose a creature type." effect (Distant
// Melody); the caller records that as Selection.SubtypeFromChosenType.
func thatTypeQualifierEnd(tokens []shared.Token, start int) (int, bool) {
	if effectWordsAt(tokens, start, "of", "that", "type") {
		return start + 3, true
	}
	return 0, false
}

// parseDynamicSelectionCountSubject recognizes "for each <selection> ..." count
// subjects led by a subtype, color, supertype, or color qualifier rather than a
// bare card-type noun (for example "Shrine you control", "colorless creature you
// control", "Elf card in your graveyard", or "card in your hand"). An optional
// leading "other" self-exclusion qualifier ("for each other Elf you control") is
// skipped while scanning the selection atoms but kept in the span handed to
// buildDynamicCountSelection so parseSelection records the Other flag. The
// leading run of tokens must all be recognized selection atoms; anything else
// fails closed so unsupported wordings stay rejected.
func parseDynamicSelectionCountSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	scanStart := start
	if equalWord(tokens[start], "other") {
		scanStart = start + 1
	}
	end, ok := scanDynamicCountSelectionTokens(tokens, scanStart, atoms)
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
	// "historic" (artifact, legendary, or Saga; CR 702.61b) is a bare card
	// qualifier the lexer does not atomize, so recognize the word directly. It
	// lets a count subject such as "historic card in your graveyard" carry the
	// Historic flag that parseSelection sets from the same word.
	if equalWord(token, "historic") {
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
	// A trailing "plus N" rider ("the number of cards in your hand plus one.")
	// adds a fixed addend to the counted amount; the count subject itself ends at
	// "plus", and parseDynamicAmountAddend consumes the rider.
	if equalWord(tokens[end], "plus") {
		return true
	}
	// A conjoined keyword-grant rider ("… for each enchantment you control and
	// has first strike.") ends the count subject: the "and has/have/gain(s)"
	// clause is a separate static keyword grant, not part of the counted phrase.
	if equalWord(tokens[end], "and") && end+1 < len(tokens) &&
		(equalWord(tokens[end+1], "has") || equalWord(tokens[end+1], "have") ||
			equalWord(tokens[end+1], "gains") || equalWord(tokens[end+1], "gain")) {
		return true
	}
	return equalWord(tokens[end], "to") || equalWord(tokens[end], "until")
}

// parseDynamicAmountAddend consumes a trailing "plus N" rider that follows a
// dynamic count subject ("the number of cards in your hand plus one"). N is a
// spelled-out cardinal or integer; the rider must end at a clause boundary so a
// partial match fails closed. It returns the addend and the token index just past
// the consumed rider.
func parseDynamicAmountAddend(tokens []shared.Token, start int) (addend int, end int, ok bool) {
	if !effectWordsAt(tokens, start, "plus") || start+1 >= len(tokens) {
		return 0, 0, false
	}
	value, valid := addendCardinal(tokens[start+1])
	if !valid || value < 1 {
		return 0, 0, false
	}
	riderEnd := start + 2
	if !dynamicAmountBoundary(tokens, riderEnd) {
		return 0, 0, false
	}
	return value, riderEnd, true
}

// addendCardinal reads a small "plus N" addend count from a token, accepting an
// integer literal or a spelled-out cardinal one through ten.
func addendCardinal(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	if token.Kind != shared.Word {
		return 0, false
	}
	words := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	value, ok := words[strings.ToLower(token.Text)]
	return value, ok
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

// effectFromZone resolves an effect's origin zone. It prefers an explicit "from
// <zone>" atom. When none exists, a shuffle whose direct object is a graveyard
// and whose destination is a library ("shuffle your graveyard into your
// library") draws its source from the graveyard, which carries no "from"
// preposition and is therefore not tagged as a ZoneRoleFrom atom.
func effectFromZone(kind EffectKind, clause []shared.Token, atoms Atoms, span shared.Span, toZone zone.Type) zone.Type {
	if from := firstZone(atoms, span, ZoneRoleFrom); from != zone.None {
		return from
	}
	if kind == EffectShuffle && toZone == zone.Library && graveyardZonePhrase(clause) {
		return zone.Graveyard
	}
	return zone.None
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

// parseAmassSubtype returns the creature subtype named by an amass keyword
// action ("amass Orcs N" -> Orc, "amass Zombies N" -> Zombie). The untyped
// "amass N" form names no subtype in its text and defaults to Zombie, the
// rules-defined Army token type (CR 701.44c). The clause begins after the amass
// verb, so the subtype word, when present, is its leading token. It returns ""
// for non-amass effects.
func parseAmassSubtype(kind EffectKind, clause []shared.Token) types.Sub {
	if kind != EffectAmass {
		return ""
	}
	if len(clause) > 0 && clause[0].Kind == shared.Word {
		if sub, ok := recognizeSubtypePhrase(clause[0].Text); ok {
			return sub
		}
	}
	return types.Zombie
}
