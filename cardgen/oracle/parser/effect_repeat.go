package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeRepeatProcessSequence folds a "Repeat the following process <count>
// times." sentence and the single sentence that follows it into one
// EffectRepeatProcess whose RepeatBody holds the following sentence's effects.
// The count is the spell's {X} ("X times", recorded as a variable amount) or a
// fixed cardinal ("three times"). It fails closed unless the loop sentence is
// unrecognized on its own, the body sentence parses to at least one effect, and
// the body has no targets (back-references across the loop boundary are not
// modeled). It returns true when it claims the sentences so the caller can stop.
func recognizeRepeatProcessSequence(sentences []Sentence, atoms Atoms) bool {
	if len(sentences) != 2 || len(sentences[0].Effects) != 0 {
		return false
	}
	amount, ok := matchRepeatProcessClause(strings.TrimSpace(sentences[0].Text))
	if !ok {
		return false
	}
	tokens := semanticEffectTokens(sentences[1].Tokens)
	sentences[1].Targets = parseTargets(tokens, atoms)
	if len(sentences[1].Targets) != 0 {
		sentences[1].Targets = nil
		return false
	}
	sentences[1].Effects = parseEffects(sentences[1], tokens, atoms)
	recognizeTargetOpponentHandManaSentence(&sentences[1])
	collapseManaSpendRiderSentence(&sentences[1], tokens)
	body := sentences[1].Effects
	if len(body) == 0 {
		return false
	}
	sentences[1].Effects = nil
	sentences[1].Targets = nil
	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[1].Span.End}
	sentences[0].Effects = []EffectSyntax{{
		Kind:       EffectRepeatProcess,
		Context:    EffectContextController,
		Span:       span,
		ClauseSpan: span,
		Text:       sentences[0].Text + " " + sentences[1].Text,
		Tokens:     append(append([]shared.Token(nil), sentences[0].Tokens...), sentences[1].Tokens...),
		Amount:     amount,
		RepeatBody: body,
		Exact:      true,
	}}
	return true
}

// matchRepeatProcessClause recognizes "Repeat the following process <count>
// times." and "Repeat this process <count> times.", where <count> is "X" (the
// spell's variable amount) or a cardinal word ("two".."ten"). It returns the
// repeat count as an effect amount.
func matchRepeatProcessClause(text string) (EffectAmountSyntax, bool) {
	const suffix = " times."
	for _, prefix := range []string{"Repeat the following process ", "Repeat this process "} {
		if len(text) <= len(prefix)+len(suffix) ||
			!strings.EqualFold(text[:len(prefix)], prefix) ||
			!strings.EqualFold(text[len(text)-len(suffix):], suffix) {
			continue
		}
		count := strings.TrimSpace(text[len(prefix) : len(text)-len(suffix)])
		if strings.EqualFold(count, "X") {
			return EffectAmountSyntax{VariableX: true}, true
		}
		if value, ok := CardinalWordValue(count); ok && value >= 1 {
			return EffectAmountSyntax{Value: value, Known: true}, true
		}
	}
	return EffectAmountSyntax{}, false
}
