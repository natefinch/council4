package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeRepeatProcessSequence folds bounded repeat wording and the open-ended
// optional process "<action>. If you do, <payoff> and repeat this process." into
// one EffectRepeatProcess.
func recognizeRepeatProcessSequence(sentences []Sentence, atoms Atoms) bool {
	if recognizeRepeatUntilFailureSequence(sentences, atoms) {
		return true
	}
	if recognizeTrailingRepeatProcessSequence(sentences, atoms) {
		return true
	}
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

// recognizeRepeatUntilFailureSequence recognizes the composable resolving shape
// "<optional action>. If you do, <payoff> and repeat this process." The leading
// action and payoff are parsed through the ordinary effect grammar, preserving
// their typed optionality, selectors, zones, and entry riders. The wrapper merely
// records that the resulting body is open-ended; compiler and lowering do not
// inspect Oracle text.
func recognizeRepeatUntilFailureSequence(sentences []Sentence, atoms Atoms) bool {
	if len(sentences) != 2 || len(sentences[0].Effects) != 0 {
		return false
	}
	firstTokens := semanticEffectTokens(sentences[0].Tokens)
	firstTargets := parseTargets(firstTokens, atoms)
	firstEffects := parseEffects(sentences[0], firstTokens, atoms)
	if len(firstTargets) != 0 || len(firstEffects) != 1 || !firstEffects[0].Optional {
		return false
	}

	secondTokens := semanticEffectTokens(sentences[1].Tokens)
	payoffTokens, ok := cutRepeatThisProcessSuffix(secondTokens)
	if !ok || !leadingIfYouDo(payoffTokens) {
		return false
	}
	payoffTokens = stripLeadingIfYouDoClause(payoffTokens)
	payoffTargets := parseTargets(payoffTokens, atoms)
	payoffEffects := parseEffects(sentences[1], payoffTokens, atoms)
	if len(payoffTargets) != 0 || len(payoffEffects) == 0 {
		return false
	}

	body := make([]EffectSyntax, 0, len(firstEffects)+len(payoffEffects))
	body = append(body, firstEffects...)
	body = append(body, payoffEffects...)
	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[1].Span.End}
	for i := range sentences {
		sentences[i].Effects = nil
		sentences[i].Targets = nil
	}
	sentences[0].Effects = []EffectSyntax{{
		Kind:               EffectRepeatProcess,
		Context:            EffectContextController,
		Span:               span,
		ClauseSpan:         span,
		Text:               joinSentenceText(sentences),
		Tokens:             joinSentenceTokens(sentences),
		RepeatBody:         body,
		RepeatUntilFailure: true,
		Exact:              true,
	}}
	return true
}

func leadingIfYouDo(tokens []shared.Token) bool {
	return len(tokens) >= 4 &&
		effectWordsAt(tokens, 0, "if", "you", "do") &&
		tokens[3].Kind == shared.Comma
}

func cutRepeatThisProcessSuffix(tokens []shared.Token) ([]shared.Token, bool) {
	if len(tokens) < 5 {
		return nil, false
	}
	end := len(tokens)
	if tokens[end-1].Kind == shared.Period {
		end--
	}
	if end < 4 ||
		!effectWordsAt(tokens, end-4, "and", "repeat", "this", "process") {
		return nil, false
	}
	return tokens[:end-4], true
}

// recognizeTrailingRepeatProcessSequence folds "<process>. Repeat this process
// once." into a RepeatProcess whose count includes the initial execution. The
// process may span multiple preceding sentences; each must parse to effects and
// must not introduce targets.
func recognizeTrailingRepeatProcessSequence(sentences []Sentence, atoms Atoms) bool {
	if len(sentences) < 2 {
		return false
	}
	additional, ok := matchTrailingRepeatProcessClause(strings.TrimSpace(sentences[len(sentences)-1].Text))
	if !ok {
		return false
	}
	var body []EffectSyntax
	for i := 0; i < len(sentences)-1; i++ {
		tokens := semanticEffectTokens(sentences[i].Tokens)
		sentences[i].Targets = parseTargets(tokens, atoms)
		if len(sentences[i].Targets) != 0 {
			clearRepeatParse(sentences, i)
			return false
		}
		sentences[i].Effects = parseEffects(sentences[i], tokens, atoms)
		recognizeTargetOpponentHandManaSentence(&sentences[i])
		collapseManaSpendRiderSentence(&sentences[i], tokens)
		if len(sentences[i].Effects) == 0 {
			clearRepeatParse(sentences, i)
			return false
		}
		body = append(body, sentences[i].Effects...)
	}
	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[len(sentences)-1].Span.End}
	text := joinSentenceText(sentences)
	tokens := joinSentenceTokens(sentences)
	for i := range sentences {
		sentences[i].Effects = nil
		sentences[i].Targets = nil
	}
	sentences[0].Effects = []EffectSyntax{{
		Kind:       EffectRepeatProcess,
		Context:    EffectContextController,
		Span:       span,
		ClauseSpan: span,
		Text:       text,
		Tokens:     tokens,
		Amount:     EffectAmountSyntax{Value: additional + 1, Known: true},
		RepeatBody: body,
		Exact:      true,
	}}
	return true
}

func clearRepeatParse(sentences []Sentence, through int) {
	for i := 0; i <= through; i++ {
		sentences[i].Targets = nil
		sentences[i].Effects = nil
	}
}

func matchTrailingRepeatProcessClause(text string) (int, bool) {
	const prefix = "Repeat this process "
	if len(text) <= len(prefix) || !strings.EqualFold(text[:len(prefix)], prefix) {
		return 0, false
	}
	count := strings.TrimSuffix(strings.TrimSpace(text[len(prefix):]), ".")
	switch {
	case strings.EqualFold(count, "once"):
		return 1, true
	case strings.EqualFold(count, "twice"):
		return 2, true
	}
	const suffix = " times"
	if len(count) <= len(suffix) || !strings.EqualFold(count[len(count)-len(suffix):], suffix) {
		return 0, false
	}
	value, ok := CardinalWordValue(strings.TrimSpace(count[:len(count)-len(suffix)]))
	return value, ok && value >= 1
}

func joinSentenceText(sentences []Sentence) string {
	var builder strings.Builder
	for i := range sentences {
		if i > 0 {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(sentences[i].Text)
	}
	return builder.String()
}

func joinSentenceTokens(sentences []Sentence) []shared.Token {
	var tokens []shared.Token
	for i := range sentences {
		tokens = append(tokens, sentences[i].Tokens...)
	}
	return tokens
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
