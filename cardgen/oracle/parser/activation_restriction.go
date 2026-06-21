package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func parseTrailingActivationRestrictions(
	source string,
	tokens []shared.Token,
	reminders, quoted []Delimited,
) []ActivationRestriction {
	sentences := ParseSentences(source, activationRestrictionSemanticTokens(tokens, reminders, quoted))
	var restrictions []ActivationRestriction
	for i := len(sentences) - 1; i >= 0; i-- {
		restriction, ok := parseActivationRestriction(sentences[i].Tokens)
		if !ok {
			break
		}
		restrictions = append(restrictions, restriction)
	}
	for left, right := 0, len(restrictions)-1; left < right; left, right = left+1, right-1 {
		restrictions[left], restrictions[right] = restrictions[right], restrictions[left]
	}
	return restrictions
}

func activationRestrictionSemanticTokens(
	tokens []shared.Token,
	reminders, quoted []Delimited,
) []shared.Token {
	excluded := append(append([]Delimited(nil), reminders...), quoted...)
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		var skip bool
		for _, delimiter := range excluded {
			if token.Span.Start.Offset >= delimiter.Span.Start.Offset &&
				token.Span.End.Offset <= delimiter.Span.End.Offset {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, token)
		}
	}
	return result
}

func parseActivationRestriction(tokens []shared.Token) (ActivationRestriction, bool) {
	fullSpan := shared.SpanOf(tokens)
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) < 2 || !syntaxWordsEqual(tokens[:2], "activate", "only") {
		return ActivationRestriction{}, false
	}
	restriction := ActivationRestriction{
		Kind: ActivationRestrictionUnsupported,
		Span: fullSpan,
	}
	if len(tokens) == 2 {
		return restriction, true
	}
	if equalWord(tokens[2], "if") {
		return ActivationRestriction{}, false
	}
	remainder := tokens[2:]
	if sorcerySpan, ok := parseActivationSorceryTiming(remainder); ok {
		restriction.Kind = ActivationRestrictionSorceryTiming
		restriction.SorcerySpan = sorcerySpan
		return restriction, true
	}
	if parseActivationInstantTiming(remainder) {
		restriction.Kind = ActivationRestrictionInstantTiming
		return restriction, true
	}
	if frequency, ok := parseActivationFrequencyRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionFrequency
		restriction.Frequency = frequency
		return restriction, true
	}
	if phaseStep, ok := parseActivationPhaseStepRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionPhaseStep
		restriction.PhaseStep = phaseStep
		return restriction, true
	}
	if playerTurn, ok := parseActivationPlayerTurnRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionPlayerTurn
		restriction.PlayerTurn = playerTurn
		return restriction, true
	}
	return restriction, true
}

// parseActivationInstantTiming reports whether the restriction names instant
// timing ("as an instant", "at instant speed", or "any time you could cast an
// instant"). Instant timing is the default for activated abilities, so this
// restriction lowers to no timing restriction; recognizing it keeps the
// otherwise no-op sentence from blocking the ability.
func parseActivationInstantTiming(tokens []shared.Token) bool {
	if rest, ok := cutSyntaxWords(tokens, "as"); ok && syntaxWordsEqual(rest, "an", "instant") {
		return true
	}
	if rest, ok := cutSyntaxWords(tokens, "at"); ok && syntaxWordsEqual(rest, "instant", "speed") {
		return true
	}
	if rest, ok := cutSyntaxWords(tokens, "any", "time", "you", "could", "cast"); ok &&
		syntaxWordsEqual(rest, "an", "instant") {
		return true
	}
	return false
}

func parseActivationSorceryTiming(tokens []shared.Token) (shared.Span, bool) {
	if rest, ok := cutSyntaxWords(tokens, "as"); ok && syntaxWordsEqual(rest, "a", "sorcery") {
		return shared.SpanOf(tokens), true
	}
	if rest, ok := cutSyntaxWords(tokens, "at"); ok && syntaxWordsEqual(rest, "sorcery", "speed") {
		return shared.SpanOf(tokens), true
	}
	if rest, ok := cutSyntaxWords(tokens, "any", "time", "you", "could", "cast"); ok &&
		syntaxWordsEqual(rest, "a", "sorcery") {
		return shared.SpanOf(tokens), true
	}
	return shared.Span{}, false
}

func parseActivationFrequencyRestriction(tokens []shared.Token) (ActivationFrequencyRestriction, bool) {
	count, rest, ok := parseActivationFrequencyCount(tokens)
	if !ok {
		return ActivationFrequencyRestriction{}, false
	}
	period, ok := parseActivationFrequencyPeriod(rest)
	if !ok {
		return ActivationFrequencyRestriction{}, false
	}
	return ActivationFrequencyRestriction{
		Span:   shared.SpanOf(tokens),
		Count:  count,
		Period: period,
	}, true
}

func parseActivationFrequencyCount(tokens []shared.Token) (ActivationFrequencyCount, []shared.Token, bool) {
	if len(tokens) > 0 && equalWord(tokens[0], "once") {
		return ActivationFrequencyCount{
			Kind: ActivationFrequencyCountOnce,
			Span: tokens[0].Span,
		}, tokens[1:], true
	}
	if len(tokens) >= 2 && syntaxWordsEqual(tokens[:2], "one", "time") {
		return ActivationFrequencyCount{
			Kind: ActivationFrequencyCountOnce,
			Span: shared.SpanOf(tokens[:2]),
		}, tokens[2:], true
	}
	return ActivationFrequencyCount{}, nil, false
}

func parseActivationFrequencyPeriod(tokens []shared.Token) (ActivationFrequencyPeriod, bool) {
	if len(tokens) != 2 || !equalWord(tokens[1], "turn") {
		return ActivationFrequencyPeriod{}, false
	}
	if !equalWord(tokens[0], "each") &&
		!equalWord(tokens[0], "every") &&
		!equalWord(tokens[0], "per") {
		return ActivationFrequencyPeriod{}, false
	}
	return ActivationFrequencyPeriod{
		Kind: ActivationFrequencyPeriodTurn,
		Span: shared.SpanOf(tokens),
	}, true
}

func parseActivationPhaseStepRestriction(tokens []shared.Token) (ActivationPhaseStepRestriction, bool) {
	remainder, ok := cutSyntaxWords(tokens, "during")
	if !ok || len(remainder) == 0 {
		return ActivationPhaseStepRestriction{}, false
	}
	if name, ok := parsePhaseStepName(remainder, false); ok {
		return ActivationPhaseStepRestriction{
			Span:       shared.SpanOf(tokens),
			Quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierNone},
			Player:     TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny},
			Name:       name,
		}, true
	}
	determiner, ok := parsePhaseStepDeterminer(remainder)
	if !ok {
		return ActivationPhaseStepRestriction{}, false
	}
	name, ok := parsePhaseStepName(
		determiner.remainder,
		determiner.quantifier.Kind == PhaseStepQuantifierEachOf,
	)
	if !ok {
		return ActivationPhaseStepRestriction{}, false
	}
	return ActivationPhaseStepRestriction{
		Span:       shared.SpanOf(tokens),
		Quantifier: determiner.quantifier,
		Player:     determiner.player,
		Name:       name,
	}, true
}

// parseActivationPlayerTurnRestriction reconstructs a "during <player>'s turn"
// restriction, e.g. "Activate only during your turn." The possessive player
// selector is captured so the compiler can fail closed on selectors it does not
// yet model; only "your turn" is reduced to a typed timing restriction today.
func parseActivationPlayerTurnRestriction(tokens []shared.Token) (ActivationPlayerTurnRestriction, bool) {
	remainder, ok := cutSyntaxWords(tokens, "during")
	if !ok || len(remainder) == 0 {
		return ActivationPlayerTurnRestriction{}, false
	}
	parsed := parseTriggerPlayerSelector(remainder)
	if !parsed.ok || parsed.form != triggerPlayerSelectorPossessive {
		return ActivationPlayerTurnRestriction{}, false
	}
	if len(parsed.remainder) != 1 || !equalWord(parsed.remainder[0], "turn") {
		return ActivationPlayerTurnRestriction{}, false
	}
	return ActivationPlayerTurnRestriction{
		Span:   shared.SpanOf(tokens),
		Player: parsed.player,
	}, true
}
