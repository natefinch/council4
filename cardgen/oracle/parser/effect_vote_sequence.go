package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitVoteSequences folds the "Starting with you, each player votes for <A> or
// <B>." voting family (CR 701.32) onto each ability that opens a vote and gates
// one or more consequences on its tally. It runs after resolving syntax and
// semantic accessors are emitted so the arm clauses are already classified; the
// recognizer re-parses each arm clause in isolation (dropping the "If <option>
// gets more votes," prefix) and stores the typed arms on ability.Vote. The
// consumed sentences keep their source text for coverage but shed their effects
// and condition wording so the downstream stages read the typed vote structure
// instead of the condition segments, which the compiler does not model as a
// state predicate.
func emitVoteSequences(abilities []Ability) {
	for i := range abilities {
		recognizeVoteSequence(&abilities[i])
	}
}

// recognizeVoteSequence matches an ability whose first resolving sentence is
// exactly "Starting with you, each player votes for <A> or <B>." and whose
// remaining sentences are each a distinct "If <option> gets more votes, ..." or
// "If <option> gets more votes or the vote is tied, ..." arm naming one of the
// two options. It records the freshly parsed arm sentences on ability.Vote and
// strips the consumed sentences' effects and condition semantics so coverage
// credits the whole construct and the compiler lowers it as a vote. It fails
// closed (leaving the ability untouched) for any other shape: a missing voting
// sentence, more than two options, a non-arm trailing sentence, an arm naming
// an unknown option, a duplicated arm, an arm whose clause does not parse
// cleanly, or a voting sentence that already carries effects.
func recognizeVoteSequence(ability *Ability) {
	if ability.Modal != nil || ability.DiceTable != nil || ability.CoinFlip != nil || len(ability.Sentences) < 2 {
		return
	}
	vote := &ability.Sentences[0]
	if vote.StaticRule != nil || len(vote.Effects) != 0 {
		return
	}
	options, ok := voteSentenceOptions(vote)
	if !ok {
		return
	}

	arms := make([]VoteArm, 0, len(ability.Sentences)-1)
	seen := make([]bool, len(options))
	for i := 1; i < len(ability.Sentences); i++ {
		arm, ok := parseVoteArm(ability, &ability.Sentences[i], options)
		if !ok {
			return
		}
		if seen[arm.Option] {
			return
		}
		seen[arm.Option] = true
		arms = append(arms, arm)
	}
	if len(arms) == 0 {
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
	ability.Vote = &VoteClause{Options: options, Arms: arms, Spans: spans, ConstructSpan: construct}
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
	ability.TriggerConditionSegments = nil
}

// voteSentenceOptions reports whether the sentence's semantic tokens are exactly
// "Starting with you, each player votes for <A> or <B>." with two single-word
// options, returning the option labels in printed order.
func voteSentenceOptions(sentence *Sentence) ([]string, bool) {
	tokens := semanticEffectTokens(sentence.Tokens)
	// "Starting" "with" "you" "," "each" "player" "votes" "for" A "or" B "."
	if len(tokens) != 12 {
		return nil, false
	}
	if !effectWordsAt(tokens, 0, "starting", "with", "you") ||
		tokens[3].Kind != shared.Comma ||
		!effectWordsAt(tokens, 4, "each", "player", "votes", "for") ||
		tokens[8].Kind != shared.Word ||
		!equalWord(tokens[9], "or") ||
		tokens[10].Kind != shared.Word ||
		tokens[11].Kind != shared.Period {
		return nil, false
	}
	return []string{tokens[8].Text, tokens[10].Text}, true
}

// parseVoteArm recognizes a single "If <option> gets more votes, <effect>." or
// "If <option> gets more votes or the vote is tied, <effect>." arm sentence. It
// confirms the leading condition boundary is an "if" intro, identifies the named
// option, reconstructs the consequence clause after the comma from the sentence
// source text, and parses it through the full pipeline. The arm is accepted only
// when the named option is one of the two vote options and its clause parses
// without diagnostics into a single ability that carries at least one resolving
// effect.
func parseVoteArm(ability *Ability, sentence *Sentence, options []string) (VoteArm, bool) {
	tokens := semanticEffectTokens(sentence.Tokens)
	// "If" <option> "gets" "more" "votes" [...] "," <effect>
	if len(tokens) < 7 || !equalWord(tokens[0], "if") || tokens[1].Kind != shared.Word ||
		!effectWordsAt(tokens, 2, "gets", "more", "votes") {
		return VoteArm{}, false
	}
	optionIndex := -1
	for i, option := range options {
		if strings.EqualFold(tokens[1].Text, option) {
			optionIndex = i
			break
		}
	}
	if optionIndex < 0 {
		return VoteArm{}, false
	}
	var tieInclusive bool
	commaIndex := 5
	switch {
	case tokens[5].Kind == shared.Comma:
		tieInclusive = false
	case effectWordsAt(tokens, 5, "or", "the", "vote", "is", "tied") && len(tokens) > 10 &&
		tokens[10].Kind == shared.Comma:
		tieInclusive = true
		commaIndex = 10
	default:
		return VoteArm{}, false
	}
	if commaIndex+1 >= len(tokens) {
		return VoteArm{}, false
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, tokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return VoteArm{}, false
	}

	consequence := tokens[commaIndex+1]
	relStart := consequence.Span.Start.Offset - sentence.Span.Start.Offset
	if relStart < 0 || relStart > len(sentence.Text) {
		return VoteArm{}, false
	}
	// Re-parse the consequence clause in isolation so its effects are classified
	// without the "If <option> gets more votes," condition wording the parser
	// would otherwise misread. Pad the clause with leading spaces equal to the
	// consequence's card offset so the lexer assigns the arm tokens, effects, and
	// references their absolute card-source offsets, keeping every downstream span
	// (body-span computation, coverage) in the card's coordinate system.
	clause := strings.Repeat(" ", consequence.Span.Start.Offset) +
		titleFirstEffectText(sentence.Text[relStart:])
	document, diagnostics := Parse(clause, Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		return VoteArm{}, false
	}
	arm := document.Abilities[0]
	if arm.Kind != AbilityStatic || arm.Modal != nil || arm.DiceTable != nil ||
		arm.CoinFlip != nil || arm.Vote != nil || !sentencesHaveResolvingEffect(arm.Sentences) {
		return VoteArm{}, false
	}
	return VoteArm{Option: optionIndex, TieInclusive: tieInclusive, Sentences: arm.Sentences}, true
}
