package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseSwitchPowerToughnessEffect recognizes the one-shot continuous effect
// "[Until end of turn,] switch <subject>'s power and toughness[ until end of
// turn]." (Aeromoeba, Aquamoeba, Crag Puca, Twisted Image), which exchanges the
// affected creature's power and toughness for the turn (CR 613.4e, layer 7e).
//
// The leading or trailing "until end of turn" duration is required. The subject
// is a possessive naming either the source permanent ("this creature's", "this
// permanent's", or the card's own name) — recorded by SwitchPTSource — or a
// single targeted creature ("target creature's"), left for the target
// machinery. Any richer subject (a group, "its", "each creature's", or the
// "power and toughness of <X>" phrasing) fails closed so those cards stay
// unsupported.
func parseSwitchPowerToughnessEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
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
	if len(remaining) < 5 || !equalWord(remaining[0], "switch") {
		return nil, false
	}
	anchor := len(remaining) - 3
	if !staticWordsAt(remaining, anchor, "power", "and", "toughness") {
		return nil, false
	}
	subject := remaining[1:anchor]
	if !switchPowerToughnessSourceSubject(subject, atoms) {
		return nil, false
	}

	effect := EffectSyntax{
		Kind:           EffectSwitchPT,
		Context:        EffectContextController,
		Span:           sentence.Span,
		ClauseSpan:     sentence.Span,
		Text:           sentence.Text,
		Tokens:         append([]shared.Token(nil), body...),
		Duration:       EffectDurationUntilEndOfTurn,
		SwitchPTSource: true,
	}
	return []EffectSyntax{effect}, true
}

// switchPowerToughnessSourceSubject reports whether the possessive subject
// preceding "power and toughness" names the source permanent itself ("this
// creature's", "this permanent's", or the card's own possessive name). A single
// targeted creature ("target creature's") and any group subject fail closed
// because the target and group machinery does not yet recognize the possessive
// noun.
func switchPowerToughnessSourceSubject(subject []shared.Token, atoms Atoms) bool {
	if len(subject) == 0 {
		return false
	}
	if len(subject) == 2 && equalWord(subject[0], "this") &&
		(equalWord(subject[1], "creature's") || equalWord(subject[1], "permanent's")) {
		return true
	}
	return slices.Contains(atoms.SelfNameSpans(), shared.SpanOf(subject))
}
