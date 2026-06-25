package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitCoinFlipSequences folds the "Flip a coin." outcome family onto each
// ability that resolves a flip and one or both of its win/lose branches. It runs
// after resolving syntax and semantic accessors are emitted so the branch
// clauses are already classified; the recognizer re-parses each branch clause in
// isolation (dropping the "If you win/lose the flip," prefix) and stores the
// typed branch sentences on ability.CoinFlip. The consumed sentences keep their
// source text for coverage but shed their effects and condition wording so the
// downstream stages read the typed coin-flip structure instead of the condition
// segments, which the compiler does not model as a state predicate.
func emitCoinFlipSequences(abilities []Ability) {
	for i := range abilities {
		recognizeControllerCoinFlipSequence(&abilities[i])
	}
}

// recognizeControllerCoinFlipSequence matches an ability whose first resolving
// sentence is exactly "Flip a coin." and whose remaining sentences are each a
// distinct "If you win the flip, <effect>." or "If you lose the flip, <effect>."
// branch. It records the freshly parsed branch sentences on ability.CoinFlip and
// strips the consumed sentences' effects and condition semantics so coverage
// credits the whole construct and the compiler lowers it as a coin flip. It
// fails closed (leaving the ability untouched) for any other shape: a missing
// flip line, a non-branch trailing sentence, a duplicated branch, a branch whose
// clause does not parse cleanly, or a flip that already carries effects.
func recognizeControllerCoinFlipSequence(ability *Ability) {
	if ability.Modal != nil || ability.DiceTable != nil || len(ability.Sentences) < 2 {
		return
	}
	flip := &ability.Sentences[0]
	if flip.StaticRule != nil || len(flip.Effects) != 0 || !sentenceIsFlipACoin(flip) {
		return
	}

	var win, lose []Sentence
	for i := 1; i < len(ability.Sentences); i++ {
		sentence := &ability.Sentences[i]
		outcome, branch, ok := parseCoinFlipBranch(ability, sentence)
		if !ok {
			return
		}
		switch outcome {
		case coinFlipOutcomeWin:
			if win != nil {
				return
			}
			win = branch
		case coinFlipOutcomeLose:
			if lose != nil {
				return
			}
			lose = branch
		default:
			return
		}
	}
	if win == nil && lose == nil {
		return
	}

	spans := make([]shared.Span, 0, len(ability.Sentences))
	construct := ability.Sentences[0].Span
	for i := range ability.Sentences {
		spans = append(spans, ability.Sentences[i].Span)
		ability.Sentences[i].Effects = nil
		if ability.Sentences[i].Span.End.Offset > construct.End.Offset {
			construct.End = ability.Sentences[i].Span.End
		}
	}
	ability.CoinFlip = &CoinFlip{Win: win, Lose: lose, Spans: spans, ConstructSpan: construct}
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
	ability.TriggerConditionSegments = nil
}

type coinFlipOutcome int

const (
	coinFlipOutcomeWin coinFlipOutcome = iota
	coinFlipOutcomeLose
)

// sentenceIsFlipACoin reports whether the sentence's semantic tokens are exactly
// "Flip a coin." with no trailing wording, so the flip is a standalone clause.
func sentenceIsFlipACoin(sentence *Sentence) bool {
	tokens := semanticEffectTokens(sentence.Tokens)
	return len(tokens) == 4 &&
		effectWordsAt(tokens, 0, "flip", "a", "coin") &&
		tokens[3].Kind == shared.Period
}

// parseCoinFlipBranch recognizes a single "If you win the flip, <effect>." or
// "If you lose the flip, <effect>." branch sentence. It confirms the leading
// condition boundary is an "if" intro, reconstructs the consequence clause after
// the comma from the sentence source text, and parses it through the full
// pipeline. The branch is accepted only when its clause parses without
// diagnostics into a single ability that carries at least one resolving effect.
func parseCoinFlipBranch(ability *Ability, sentence *Sentence) (coinFlipOutcome, []Sentence, bool) {
	tokens := semanticEffectTokens(sentence.Tokens)
	if len(tokens) < 7 || tokens[5].Kind != shared.Comma {
		return 0, nil, false
	}
	var outcome coinFlipOutcome
	switch {
	case effectWordsAt(tokens, 0, "if", "you", "win", "the", "flip"):
		outcome = coinFlipOutcomeWin
	case effectWordsAt(tokens, 0, "if", "you", "lose", "the", "flip"):
		outcome = coinFlipOutcomeLose
	default:
		return 0, nil, false
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, tokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return 0, nil, false
	}

	relStart := tokens[6].Span.Start.Offset - sentence.Span.Start.Offset
	if relStart < 0 || relStart > len(sentence.Text) {
		return 0, nil, false
	}
	// Re-parse the consequence clause in isolation so its effects are classified
	// without the "If you win/lose the flip," condition wording the parser would
	// otherwise misread. Pad the clause with leading spaces equal to the
	// consequence's card offset so the lexer assigns the branch tokens, effects,
	// and references their absolute card-source offsets, keeping every downstream
	// span (body-span computation, coverage) in the card's coordinate system.
	clause := strings.Repeat(" ", tokens[6].Span.Start.Offset) +
		titleFirstEffectText(sentence.Text[relStart:])
	document, diagnostics := Parse(clause, Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		return 0, nil, false
	}
	branch := document.Abilities[0]
	if branch.Kind != AbilityStatic || branch.Modal != nil || branch.DiceTable != nil ||
		branch.CoinFlip != nil || !sentencesHaveResolvingEffect(branch.Sentences) {
		return 0, nil, false
	}
	return outcome, branch.Sentences, true
}

// sentencesHaveResolvingEffect reports whether any sentence carries at least one
// resolving effect, so an empty or purely static branch fails closed.
func sentencesHaveResolvingEffect(sentences []Sentence) bool {
	for i := range sentences {
		if len(sentences[i].Effects) != 0 {
			return true
		}
	}
	return false
}
