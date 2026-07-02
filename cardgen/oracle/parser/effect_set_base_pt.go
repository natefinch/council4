package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseSetBasePowerToughnessEffect recognizes the one-shot continuous base
// power/toughness SET effect "[Until end of turn,] <subject> ha(s|ve) base power
// and toughness <N/N|X/X>[ and <gain all creature types|become every creature
// type>][ until end of turn]." (Mirror Entity, Square Up, Biomass Mutation).
//
// The leading or trailing "until end of turn" duration is required, which is
// what distinguishes this resolving effect from the permanent static
// declaration "Other creatures have base power and toughness 1/1." (Godhead of
// Awe), parsed elsewhere. The amount is a literal N/N or the variable X/X whose
// value is the cost's X. The optional rider grants every creature type. The
// subject is a controlled/battlefield creature group (recorded in
// StaticSubject), a single targeted creature (left for the target machinery), or
// the source permanent ("This creature ..."). Any richer shape — a keyword
// rider, a color or subtype change, "loses all abilities", a missing duration,
// or a "where X is ..." count — fails closed so those cards stay unsupported.
func parseSetBasePowerToughnessEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner := body[:len(body)-1]
	remaining, leadDuration := stripLeadingDurationClause(inner, atoms)
	if leadDuration != EffectDurationNone && leadDuration != EffectDurationUntilEndOfTurn {
		return nil, false
	}
	endOfTurn := leadDuration == EffectDurationUntilEndOfTurn
	remaining, trailingEOT := trimTrailingUntilEndOfTurn(remaining)
	endOfTurn = endOfTurn || trailingEOT
	if !endOfTurn {
		return nil, false
	}

	remaining, losesAllAbilities := stripLosesAllAbilitiesRider(remaining)

	anchor := setBasePowerToughnessAnchor(remaining)
	if anchor < 2 {
		return nil, false
	}
	if !equalWord(remaining[anchor-1], "has") && !equalWord(remaining[anchor-1], "have") {
		return nil, false
	}
	amount, ok := setBasePowerToughnessAmount(remaining, anchor)
	if !ok {
		return nil, false
	}
	cursor := anchor + 7

	everyCreatureType, cursor, ok := setBasePowerToughnessTypeRider(remaining, cursor)
	if !ok || cursor != len(remaining) {
		return nil, false
	}

	subjectWithVerb := remaining[:anchor]
	subjectTokens := remaining[:anchor-1]
	staticSubject := parseEffectStaticSubject(subjectWithVerb, atoms)
	source := false
	switch {
	case staticSubject.Kind != EffectStaticSubjectNone:
		// Group form; StaticSubject carries the affected group.
	case setBasePowerToughnessSourceSubject(subjectTokens, atoms):
		source = true
	case setBasePowerToughnessTargetSubject(subjectTokens):
		// Single-target form; the target machinery extracts the target.
	default:
		return nil, false
	}

	effect := EffectSyntax{
		Kind:                       EffectSetBasePT,
		Context:                    EffectContextController,
		Span:                       sentence.Span,
		ClauseSpan:                 sentence.Span,
		Text:                       sentence.Text,
		Tokens:                     append([]shared.Token(nil), body...),
		Duration:                   EffectDurationUntilEndOfTurn,
		StaticSubject:              staticSubject,
		SetBasePower:               amount.power,
		SetBaseToughness:           amount.toughness,
		SetBasePTVariableX:         amount.variableX,
		SetBasePTEveryCreatureType: everyCreatureType,
		SetBasePTSource:            source,
		SetBasePTLosesAllAbilities: losesAllAbilities,
	}
	return []EffectSyntax{effect}, true
}

// stripLosesAllAbilitiesRider removes the "lose(s) all abilities and" clause that
// can precede the "ha(s|ve) base power and toughness" verb ("<subject> loses all
// abilities and has base power and toughness N/N"). It returns the tokens with the
// rider removed — leaving the plain "<subject> ha(s|ve) base ..." shape the rest of
// the parser already handles — and whether the rider was present. Any other shape
// is returned unchanged.
func stripLosesAllAbilitiesRider(tokens []shared.Token) ([]shared.Token, bool) {
	anchor := setBasePowerToughnessAnchor(tokens)
	if anchor < 5 {
		return tokens, false
	}
	if !equalWord(tokens[anchor-1], "has") && !equalWord(tokens[anchor-1], "have") {
		return tokens, false
	}
	losesVerb := equalWord(tokens[anchor-5], "loses") || equalWord(tokens[anchor-5], "lose")
	if !losesVerb ||
		!equalWord(tokens[anchor-4], "all") ||
		!equalWord(tokens[anchor-3], "abilities") ||
		!equalWord(tokens[anchor-2], "and") {
		return tokens, false
	}
	stripped := append([]shared.Token(nil), tokens[:anchor-5]...)
	stripped = append(stripped, tokens[anchor-1:]...)
	return stripped, true
}

// setBasePowerToughnessAnchor returns the index of the "base power and toughness"
// phrase in tokens, or -1 when absent.
func setBasePowerToughnessAnchor(tokens []shared.Token) int {
	for i := 0; i+3 < len(tokens); i++ {
		if staticWordsAt(tokens, i, "base", "power", "and", "toughness") {
			return i
		}
	}
	return -1
}

// setBasePowerToughnessAmountResult holds the parsed N/N or X/X base
// power/toughness value: a literal power and toughness, or the variable X form.
type setBasePowerToughnessAmountResult struct {
	power     int
	toughness int
	variableX bool
}

// setBasePowerToughnessAmount parses the "N/N" or "X/X" value following the
// "base power and toughness" phrase anchored at index. Both characteristics must
// be literal non-negative integers or both the variable "X"; a mixed form fails.
func setBasePowerToughnessAmount(tokens []shared.Token, anchor int) (setBasePowerToughnessAmountResult, bool) {
	if anchor+6 >= len(tokens) || tokens[anchor+5].Kind != shared.Slash {
		return setBasePowerToughnessAmountResult{}, false
	}
	powerToken := tokens[anchor+4]
	toughnessToken := tokens[anchor+6]
	if equalWord(powerToken, "x") && equalWord(toughnessToken, "x") {
		return setBasePowerToughnessAmountResult{variableX: true}, true
	}
	power, powerOK := staticUnsignedInteger(powerToken)
	toughness, toughnessOK := staticUnsignedInteger(toughnessToken)
	if powerOK && toughnessOK {
		return setBasePowerToughnessAmountResult{power: power, toughness: toughness}, true
	}
	return setBasePowerToughnessAmountResult{}, false
}

// setBasePowerToughnessTypeRider consumes an optional "and <gain all creature
// types|become every creature type|are/is every creature type>" rider starting
// at cursor, returning whether the every-creature-type grant is present, the
// cursor past the consumed rider, and whether the tokens from cursor form either
// no rider or exactly the recognized rider. Any other trailing tokens fail.
func setBasePowerToughnessTypeRider(tokens []shared.Token, cursor int) (everyCreatureType bool, next int, ok bool) {
	if cursor == len(tokens) {
		return false, cursor, true
	}
	if !equalWord(tokens[cursor], "and") {
		return false, cursor, false
	}
	cursor++
	if cursor >= len(tokens) {
		return false, cursor, false
	}
	verbs := []string{"gain", "gains", "become", "becomes", "are", "is"}
	matched := false
	for _, verb := range verbs {
		if equalWord(tokens[cursor], verb) {
			matched = true
			break
		}
	}
	if !matched {
		return false, cursor, false
	}
	cursor++
	switch {
	case staticWordsAt(tokens, cursor, "all", "creature", "types"):
		return true, cursor + 3, true
	case staticWordsAt(tokens, cursor, "every", "creature", "type"):
		return true, cursor + 3, true
	default:
		return false, cursor, false
	}
}

// setBasePowerToughnessSourceSubject reports whether the subject tokens name the
// source permanent itself ("this creature"/"this permanent" or the card's own
// name).
func setBasePowerToughnessSourceSubject(tokens []shared.Token, atoms Atoms) bool {
	if staticWordsAt(tokens, 0, "this", "creature") && len(tokens) == 2 {
		return true
	}
	if staticWordsAt(tokens, 0, "this", "permanent") && len(tokens) == 2 {
		return true
	}
	if len(tokens) == 0 {
		return false
	}
	return slices.Contains(atoms.SelfNameSpans(), shared.SpanOf(tokens))
}

// setBasePowerToughnessTargetSubject reports whether the subject tokens name a
// single targeted creature, which the target machinery extracts separately.
func setBasePowerToughnessTargetSubject(tokens []shared.Token) bool {
	for i := range tokens {
		if equalWord(tokens[i], "target") {
			return true
		}
	}
	return false
}
