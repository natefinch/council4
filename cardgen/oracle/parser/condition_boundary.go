package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// ConditionBoundary marks where a condition introducer begins in an ability's or
// mode's token stream. The parser owns the Oracle vocabulary that recognizes the
// introducer, classifies duration phrases, and locates the triggered
// intervening-if position; the compiler consumes these typed boundaries by token
// start position instead of inspecting "if"/"unless"/"only if"/"as long as"
// spelling.
type ConditionBoundary struct {
	// Start is the source position of the introducer's first token. The compiler
	// matches a boundary to a token by this position.
	Start shared.Position `json:"-"`
	// NodeID is a stable typed identity the parser assigns to each boundary in
	// source order. Every condition segment emitted for this boundary, across the
	// semantic-body and raw-trigger token streams, carries this NodeID so the
	// compiler matches "the same condition" by identity instead of span equality.
	NodeID int `json:"-"`
	// Kind is the grammatical introducer that opens the clause.
	Kind ConditionIntroKind `json:",omitempty"`
	// DurationSkip reports that this "as long as" introducer opens a source
	// duration phrase ("for as long as ..." or "as long as this [type] remains/is
	// on the battlefield") that compileDuration already captures, so the compiler
	// must not also treat it as a standalone condition.
	DurationSkip bool `json:",omitempty"`
	// Intervening reports that this introducer is a triggered ability's
	// intervening-if: an "if" clause immediately following the trigger event
	// comma. It is only ever set for triggered abilities.
	Intervening bool `json:",omitempty"`
	// Resolving reports that this condition is introduced by sentence-leading
	// "Then if" and therefore gates a later resolving effect rather than
	// restricting whether an activated ability may be activated.
	Resolving bool `json:",omitempty"`
	// ActivationKeyword is the source span of an "Activate" keyword that
	// immediately precedes an "only if" introducer ("Activate only if ..."). It
	// is the zero span when absent. The parser owns the recognition of this
	// activation-restriction keyword so the compiler and lowering can account
	// for its consumed source span without inspecting token spelling.
	ActivationKeyword shared.Span `json:"-"`
}

// emitConditionBoundaries fills each ability's and mode's typed condition
// boundaries from its raw tokens. Boundaries are keyed by absolute source
// position, so each downstream scan stream (semantic body tokens, raw trigger
// tokens, or mode tokens) consumes exactly the boundaries whose tokens it walks.
func emitConditionBoundaries(abilities []Ability, cardName string) {
	for i := range abilities {
		ability := &abilities[i]
		body := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
		semantic := eventHistorySemanticTokens(body, ability.Reminders, ability.Quoted)
		ability.ConditionBoundaries = conditionBoundaries(
			ability.Tokens,
			ability.Kind == AbilityTriggered,
			conditionForcedAttackIfAble(semantic),
			cardName,
		)
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			semantic := eventHistorySemanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
			mode.ConditionBoundaries = conditionBoundaries(
				mode.Tokens,
				false,
				conditionForcedAttackIfAble(semantic),
				cardName,
			)
		}
	}
}

// conditionBoundaries enumerates every condition introducer in tokens, in source
// order. triggered selects whether intervening-if positions are recorded. When
// ifAbleExcluded is set (the ability's semantic body says "attacks each combat if
// able"), an introducer that opens a standalone "if able" clause is dropped,
// because that restriction is captured structurally and must not emit a
// condition.
func conditionBoundaries(tokens []shared.Token, triggered, ifAbleExcluded bool, cardName string) []ConditionBoundary {
	intervening := -1
	if triggered {
		if comma := triggerBodyComma(tokens, cardName); comma >= 0 {
			intervening = comma + 1
		}
	}
	var boundaries []ConditionBoundary
	for i := 0; i < len(tokens); i++ {
		intro, _ := conditionIntroAt(tokens, i)
		if intro == ConditionIntroUnknown {
			continue
		}
		end := conditionClauseEnd(tokens, i)
		if ifAbleExcluded && conditionClauseIsIfAble(tokens[i:end]) {
			i = end - 1
			continue
		}
		if effectWordsAt(tokens, i, creatureSpellHasteConditionWords...) {
			i = end - 1
			continue
		}
		if entersAsCopyCounterRiderConditionAt(tokens, i) || punisherUnlessClauseAt(tokens, i) {
			i = end - 1
			continue
		}
		if playFromTopPayLifeRiderConditionAt(tokens, i) {
			i = end - 1
			continue
		}
		if conditionLeaveBattlefieldExileReplacementAt(tokens, i) {
			i = end - 1
			continue
		}
		boundaries = append(boundaries, ConditionBoundary{
			Start:             tokens[i].Span.Start,
			NodeID:            len(boundaries),
			Kind:              intro,
			DurationSkip:      intro == ConditionIntroAsLongAs && conditionAsLongAsIsDuration(tokens, i),
			Intervening:       triggered && intro == ConditionIntroIf && i == intervening,
			Resolving:         conditionIsResolvingThenIf(tokens, i, intro),
			ActivationKeyword: conditionActivationKeyword(tokens, i, intro),
		})
		i = end - 1
	}

	return boundaries
}

func conditionIsResolvingThenIf(tokens []shared.Token, index int, intro ConditionIntroKind) bool {
	if intro != ConditionIntroIf || index < 0 {
		return false
	}
	if index > 0 && equalWord(tokens[index-1], "then") &&
		(index == 1 || tokenEndsSentence(tokens[index-2])) {
		return true
	}
	return conditionIsResolvingCreatedToken(tokens, index)
}

// conditionIsResolvingCreatedToken reports whether a sentence-leading "If the
// token ..." introduces a resolving condition that inspects a token a prior
// effect in the same ability created (Yenna, Redtooth Regent: "If the token is
// an Aura, ..."). The token only exists during resolution, so the gate is
// evaluated then rather than as an activation restriction.
func conditionIsResolvingCreatedToken(tokens []shared.Token, index int) bool {
	sentenceLeading := index == 0 || tokenEndsSentence(tokens[index-1])
	return sentenceLeading && index+2 < len(tokens) &&
		equalWord(tokens[index+1], "the") && equalWord(tokens[index+2], "token")
}

func tokenEndsSentence(token shared.Token) bool {
	return token.Kind == shared.Period
}

// conditionActivationKeyword returns the span of an "Activate" keyword
// immediately preceding an "only if" introducer at index, or the zero span when
// the introducer is not an "only if" or is not preceded by "Activate". This is
// the "Activate only if ..." activation restriction; capturing the keyword span
// here lets the compiler and lowering account for its consumed source without
// re-inspecting token spelling.
func conditionActivationKeyword(tokens []shared.Token, index int, intro ConditionIntroKind) shared.Span {
	if intro != ConditionIntroOnlyIf || index == 0 {
		return shared.Span{}
	}
	if !equalWord(tokens[index-1], "activate") {
		return shared.Span{}
	}
	return tokens[index-1].Span
}

// conditionAttacksEachCombatIfAble reports whether the semantic tokens spell
// "attack[s] each combat if able", the restriction whose trailing "if able" must
// not become a standalone condition. The singular self form ("This creature
// attacks each combat if able."), the plural group form ("Creatures you
// control attack each combat if able."), and the directed two-player form ("each
// creature they control attacks the other chosen player each combat if able.",
// The Brothers' War chapter II) are recognized.
func conditionAttacksEachCombatIfAble(semantic []shared.Token) bool {
	return conditionContainsSequence(semantic, 0, "attacks", "each", "combat", "if", "able") ||
		conditionContainsSequence(semantic, 0, "attack", "each", "combat", "if", "able") ||
		conditionContainsSequence(semantic, 0, "attacks", "the", "other", "chosen", "player", "each", "combat", "if", "able")
}

// conditionForcedAttackIfAble reports whether the semantic tokens spell a
// forced-attack restriction whose trailing "if able" must not become a
// standalone condition: the static "attacks each combat if able" form or the
// one-shot group "attack this turn if able" form (Bident of Thassa: "Creatures
// your opponents control attack this turn if able.").
func conditionForcedAttackIfAble(semantic []shared.Token) bool {
	return conditionAttacksEachCombatIfAble(semantic) ||
		conditionContainsSequence(semantic, 0, "attack", "this", "turn", "if", "able")
}

// conditionAsLongAsIsDuration reports whether an "as long as" introducer at index
// opens a duration phrase rather than a standalone condition: either it is
// preceded by "for" ("for as long as ..."), or it is the
// "as long as this [type] remains/is on the battlefield" source-duration form.
func conditionAsLongAsIsDuration(tokens []shared.Token, index int) bool {
	if index > 0 && equalWord(tokens[index-1], "for") {
		return true
	}
	return conditionSourceOnBattlefieldPhrase(tokens, index)
}

// conditionSourceOnBattlefieldPhrase reports whether the tokens starting at index
// spell "as long as this [type] remains on the battlefield" or
// "as long as this [type] is on the battlefield" — specifically the
// source-on-battlefield duration pattern, not other "as long as this [type] is
// [state]" conditions.
func conditionSourceOnBattlefieldPhrase(tokens []shared.Token, index int) bool {
	words := normalizedWords(tokens[index:])
	return wordsContainSequence(words, "as", "long", "as", "this") &&
		(wordsContainSequence(words, "remains", "on", "the", "battlefield") ||
			wordsContainSequence(words, "is", "on", "the", "battlefield"))
}

// conditionClauseIsIfAble reports whether clause is exactly the words "if able".
func conditionClauseIsIfAble(clause []shared.Token) bool {
	return wordsEqual(normalizedWords(clause), "if", "able")
}

// conditionContainsSequence reports whether the normalized words of tokens, from
// start, contain the given consecutive words.
func conditionContainsSequence(tokens []shared.Token, start int, words ...string) bool {
	return wordsContainSequence(normalizedWords(tokens[start:]), words...)
}

func wordsContainSequence(words []string, expected ...string) bool {
	for i := 0; i+len(expected) <= len(words); i++ {
		if startsWords(words[i:], expected...) {
			return true
		}
	}
	return false
}

func wordsEqual(words []string, expected ...string) bool {
	return len(words) == len(expected) && startsWords(words, expected...)
}
