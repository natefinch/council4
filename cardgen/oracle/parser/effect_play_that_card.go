package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePlayThatCardEffect recognizes the turn-scoped play permission "you may
// play <object> this turn." (and the "until end of turn", "until the end of your
// next turn", and "until your next end step" windows), where <object> is a
// back-reference to a just-exiled card ("it"/"that card" for a single card,
// "them"/"those cards" for several). It is the land-inclusive sibling of the
// "you may cast it this turn." cast permission produced by the generic effect
// parser: "play" grants both a land play and a spell cast of the referenced
// card.
//
// The clause is the "if you do" consequence of a discard trigger's "you may
// exile that card from your graveyard." reflexive optional (Containment
// Construct); the leading "If you do," condition is already stripped from the
// tokens by the caller. The recognized effect carries the play window duration,
// the optional ("you may") marker, and the back-reference so the exile clause
// and this permission lower together into a single exile-for-play primitive. Any
// other wording fails closed and flows through the generic effect parser.
func parsePlayThatCardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	tokens = stripLeadingIfYouDoClause(tokens)
	rest, leadingDuration := stripLeadingDurationClause(tokens, atoms)
	words := wordsOnly(rest)
	if len(words) < 5 || !equalWord(words[0], "you") || !equalWord(words[1], "may") ||
		!equalWord(words[2], "play") {
		return nil, false
	}
	object, after, ok := matchBackReferenceObjectWords(words[3:])
	if !ok {
		return nil, false
	}
	// A trailing "without paying its mana cost" makes the granted play a free
	// cast ("You may play it this turn without paying its mana cost.", Dauthi
	// Voidwalker). It sits after the play window, so it is stripped before the
	// duration is matched; the rider lowers onto the play effect. A played land
	// has no mana cost, so the flag only affects a spell cast of the card.
	freeCast := false
	if trimmed, cutOK := cutTokenSuffix(after, "without", "paying", "its", "mana", "cost"); cutOK {
		after = trimmed
		freeCast = true
	}
	duration, ok := matchPlayPermissionDurationWords(after, leadingDuration)
	if !ok {
		return nil, false
	}
	verbSpan := words[2].Span
	references := referencesInSpan(atoms, shared.SpanOf(object))
	if len(references) != 1 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:                      EffectPlay,
		Span:                      sentence.Span,
		ClauseSpan:                sentence.Span,
		VerbSpan:                  verbSpan,
		Text:                      sentence.Text,
		Tokens:                    append([]shared.Token(nil), tokens...),
		Context:                   EffectContextController,
		Duration:                  duration,
		Optional:                  true,
		CastWithoutPayingManaCost: freeCast,
		OptionalSpan: shared.Span{
			Start: words[0].Span.Start,
			End:   words[1].Span.End,
		},
		References: references,
		Exact:      true,
	}}, true
}

// stripLeadingIfYouDoClause removes a leading "If you do," reflexive condition
// clause from the sentence tokens, returning the remaining tokens. The "you do"
// predicate is the reflexive gate of a "you may X. If you do, Y." structure;
// stripLeadingConditionClause leaves it in place (it strips only as-long-as and
// source-counter-state intros), so this recognizer drops it itself before
// matching the trailing play-permission clause. Tokens that do not begin with a
// comma-terminated "if" condition clause are returned unchanged.
func stripLeadingIfYouDoClause(tokens []shared.Token) []shared.Token {
	if len(tokens) == 0 || !equalWord(tokens[0], "if") {
		return tokens
	}
	intro, _ := conditionIntroAt(tokens, 0)
	if intro != ConditionIntroIf {
		return tokens
	}
	end := conditionClauseEnd(tokens, 0)
	if end >= len(tokens) || tokens[end].Kind != shared.Comma {
		return tokens
	}
	return tokens[end+1:]
}

// matchBackReferenceObjectWords matches a leading back-reference object phrase
// ("it", "that card", "them", or "those cards") and returns the object's tokens
// and the remaining words after it.
func matchBackReferenceObjectWords(words []shared.Token) (object, after []shared.Token, ok bool) {
	switch {
	case len(words) >= 1 && equalWord(words[0], "it"):
		return words[:1], words[1:], true
	case len(words) >= 1 && equalWord(words[0], "them"):
		return words[:1], words[1:], true
	case len(words) >= 2 && equalWord(words[0], "that") && equalWord(words[1], "card"):
		return words[:2], words[2:], true
	case len(words) >= 2 && equalWord(words[0], "those") && equalWord(words[1], "cards"):
		return words[:2], words[2:], true
	default:
		return nil, nil, false
	}
}

// matchPlayPermissionDurationWords maps the trailing play-window words ("this
// turn", "until end of turn", "until the end of your next turn", "until your
// next end step") to a duration kind. A non-empty leadingDuration (the clause
// began "Until end of turn, ...") supplies the window instead, in which case the
// trailing words must be empty.
func matchPlayPermissionDurationWords(words []shared.Token, leadingDuration EffectDurationKind) (EffectDurationKind, bool) {
	if leadingDuration != EffectDurationNone {
		if len(words) != 0 {
			return EffectDurationNone, false
		}
		return leadingDuration, true
	}
	switch {
	case tokenWordsEqual(words, "this", "turn"):
		return EffectDurationThisTurn, true
	case tokenWordsEqual(words, "until", "end", "of", "turn"):
		return EffectDurationUntilEndOfTurn, true
	case tokenWordsEqual(words, "until", "the", "end", "of", "your", "next", "turn"):
		return EffectDurationUntilEndOfYourNextTurn, true
	case tokenWordsEqual(words, "until", "your", "next", "end", "step"):
		return EffectDurationUntilYourNextEndStep, true
	default:
		return EffectDurationNone, false
	}
}

// cutTokenSuffix removes a trailing run of words from tokens, returning the
// tokens before it. It reports false when the tokens do not end with the words.
func cutTokenSuffix(tokens []shared.Token, words ...string) ([]shared.Token, bool) {
	if len(tokens) < len(words) {
		return nil, false
	}
	offset := len(tokens) - len(words)
	for i, word := range words {
		if !equalWord(tokens[offset+i], word) {
			return nil, false
		}
	}
	return tokens[:offset], true
}
