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
		clauses, ok := parseActivationRestriction(sentences[i].Tokens)
		if !ok {
			break
		}
		// Sentences are walked back-to-front and the whole slice is reversed
		// below to restore source order; append each sentence's own clauses
		// reversed so that final reversal also restores their in-sentence order.
		for j := len(clauses) - 1; j >= 0; j-- {
			restrictions = append(restrictions, clauses[j])
		}
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

// parseActivationRestriction parses one "Activate only …" sentence into its
// typed restriction clauses. A single sentence may conjoin two restrictions
// ("Activate only as a sorcery and only once each turn."), so it returns a
// slice: each conjoined clause becomes its own ActivationRestriction sharing the
// full sentence span. The compiler combines the resulting clause kinds (e.g.
// sorcery + once-per-turn) and fails closed on any unrecognized pairing.
func parseActivationRestriction(tokens []shared.Token) ([]ActivationRestriction, bool) {
	fullSpan := shared.SpanOf(tokens)
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) < 2 || !syntaxWordsEqual(tokens[:2], "activate", "only") {
		return nil, false
	}
	if len(tokens) == 2 {
		return []ActivationRestriction{{Kind: ActivationRestrictionUnsupported, Span: fullSpan}}, true
	}
	if equalWord(tokens[2], "if") {
		return parseConditionalActivationRestriction(tokens, fullSpan)
	}
	clauses := splitActivationRestrictionConjunction(tokens[2:])
	restrictions := make([]ActivationRestriction, 0, len(clauses))
	for _, clause := range clauses {
		restriction := ActivationRestriction{Kind: ActivationRestrictionUnsupported, Span: fullSpan}
		matchActivationRestrictionBody(clause, &restriction)
		restrictions = append(restrictions, restriction)
	}
	return restrictions, true
}

// parseConditionalActivationRestriction handles an "Activate only if
// <condition> and only <timing>" sentence by peeling a trailing "and only
// <timing>" restriction tail off the condition. It returns the typed timing
// restriction(s) for the tail, each spanning only the tail (from the "and"
// through the sentence's end), while leaving the "Activate only if <condition>"
// prefix in the body for the condition parser, which removes the tail's
// activation-timing span before recognizing the gate. tokens excludes the
// trailing period; fullSpan covers the whole sentence including it. It returns
// ok=false when no recognized timing tail is present, so a bare "Activate only
// if <condition>" stays a pure activation condition.
func parseConditionalActivationRestriction(tokens []shared.Token, fullSpan shared.Span) ([]ActivationRestriction, bool) {
	for i := 3; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "and") || !equalWord(tokens[i+1], "only") {
			continue
		}
		clauses := splitActivationRestrictionConjunction(tokens[i+2:])
		restrictions := make([]ActivationRestriction, 0, len(clauses))
		recognized := true
		for _, clause := range clauses {
			restriction := ActivationRestriction{Kind: ActivationRestrictionUnsupported}
			matchActivationRestrictionBody(clause, &restriction)
			if restriction.Kind == ActivationRestrictionUnsupported {
				recognized = false
				break
			}
			restrictions = append(restrictions, restriction)
		}
		if !recognized || len(restrictions) == 0 {
			continue
		}
		tailSpan := shared.Span{Start: tokens[i].Span.Start, End: fullSpan.End}
		for j := range restrictions {
			restrictions[j].Span = tailSpan
		}
		return restrictions, true
	}
	return nil, false
}

// splitActivationRestrictionConjunction splits the body of an "Activate only …"
// sentence on a top-level "and" conjunction, dropping a leading "only" that
// reintroduces the restriction framing on the following clause ("… and only
// once each turn"). None of the supported restriction bodies (sorcery/instant
// timing, once-per-turn frequency, phase/step, player turn) contain an internal
// "and", so splitting there never fractures a single supported body; an
// unsupported split simply fails closed in the compiler.
func splitActivationRestrictionConjunction(tokens []shared.Token) [][]shared.Token {
	var clauses [][]shared.Token
	start := 0
	for i := 0; i < len(tokens); i++ {
		if !equalWord(tokens[i], "and") {
			continue
		}
		clauses = append(clauses, tokens[start:i])
		next := i + 1
		if next < len(tokens) && equalWord(tokens[next], "only") {
			next++
		}
		start = next
		i = next - 1
	}
	clauses = append(clauses, tokens[start:])
	return clauses
}

// matchActivationRestrictionBody recognizes a single restriction clause body and
// sets restriction.Kind (and any typed sub-structure) accordingly. An
// unrecognized body leaves Kind at ActivationRestrictionUnsupported.
func matchActivationRestrictionBody(remainder []shared.Token, restriction *ActivationRestriction) {
	if sorcerySpan, ok := parseActivationSorceryTiming(remainder); ok {
		restriction.Kind = ActivationRestrictionSorceryTiming
		restriction.SorcerySpan = sorcerySpan
		return
	}
	if parseActivationInstantTiming(remainder) {
		restriction.Kind = ActivationRestrictionInstantTiming
		return
	}
	if frequency, ok := parseActivationFrequencyRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionFrequency
		restriction.Frequency = frequency
		return
	}
	if phaseStep, ok := parseActivationPhaseStepRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionPhaseStep
		restriction.PhaseStep = phaseStep
		return
	}
	if playerTurn, ok := parseActivationPlayerTurnRestriction(remainder); ok {
		restriction.Kind = ActivationRestrictionPlayerTurn
		restriction.PlayerTurn = playerTurn
		return
	}
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
