package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitEachPlayerChooseDestroySequences folds the "Starting with you, each player
// may choose <permanent>. Destroy each permanent chosen this way." construct
// (Druid of Purification) onto each ability that opens it. It runs after
// resolving syntax and semantic accessors are emitted so both sentences are
// classified. The choose sentence yields no effect on its own and the destroy
// sentence yields a bare destroy whose "chosen this way" back-reference the
// backend cannot bind, so the recognizer re-parses the choose sentence's
// candidate filter in isolation, stores the typed pool on the ability, and sheds
// both sentences' effects and body references.
func emitEachPlayerChooseDestroySequences(abilities []Ability) {
	for i := range abilities {
		recognizeEachPlayerChooseDestroySequence(&abilities[i])
	}
}

// recognizeEachPlayerChooseDestroySequence matches an ability holding, as two
// consecutive resolving sentences, a "Starting with you, each player may choose
// <candidate>." sentence immediately followed by "Destroy each permanent chosen
// this way." It records the typed candidate pool on ability.EachPlayerChooseDestroy
// and strips both sentences' effects so the construct lowers to a single
// EachPlayerChooseDestroy interaction. It fails closed (leaving the ability
// untouched) for any other shape, so a card already carrying a vote, modal, dice,
// or coin-flip construct, a missing or unrecognized candidate filter, or a
// destroy sentence that is not the exact "chosen this way" wording is unaffected.
func recognizeEachPlayerChooseDestroySequence(ability *Ability) {
	if ability.Vote != nil || ability.Modal != nil || ability.DiceTable != nil ||
		ability.CoinFlip != nil || ability.EachPlayerChooseDestroy != nil {
		return
	}
	for i := 0; i+1 < len(ability.Sentences); i++ {
		choose := &ability.Sentences[i]
		destroy := &ability.Sentences[i+1]
		pool, optional, ok := eachPlayerChoosePoolSelection(choose)
		if !ok || !eachPlayerChosenThisWayDestroy(destroy) {
			continue
		}
		construct := choose.Span
		if destroy.Span.End.Offset > construct.End.Offset {
			construct.End = destroy.Span.End
		}
		choose.Effects = nil
		destroy.Effects = nil
		ability.EachPlayerChooseDestroy = &EachPlayerChooseDestroyClause{
			Pool:          pool,
			Optional:      optional,
			Spans:         []shared.Span{choose.Span, destroy.Span},
			ConstructSpan: construct,
		}
		ability.SemanticReferences = nil
		return
	}
}

// eachPlayerChoosePoolSelection reports whether sentence is exactly "Starting
// with you, each player may choose <candidate>." and, when so, returns the typed
// candidate pool selection and that the choice is optional (the "may"). The
// candidate filter is re-parsed as the selection of a synthetic "Destroy target
// <candidate>." clause, so any filter the single-target destroy grammar already
// types is accepted and every other candidate wording fails closed.
func eachPlayerChoosePoolSelection(sentence *Sentence) (pool SelectionSyntax, optional, ok bool) {
	if sentence.StaticRule != nil || len(sentence.Effects) != 0 {
		return SelectionSyntax{}, false, false
	}
	tokens := semanticEffectTokens(sentence.Tokens)
	// "starting" "with" "you" "," "each" "player" "may" "choose" <pool...> "."
	if len(tokens) < 10 {
		return SelectionSyntax{}, false, false
	}
	if !effectWordsAt(tokens, 0, "starting", "with", "you") ||
		tokens[3].Kind != shared.Comma ||
		!effectWordsAt(tokens, 4, "each", "player", "may", "choose") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return SelectionSyntax{}, false, false
	}
	poolTokens := tokens[8 : len(tokens)-1]
	if len(poolTokens) < 2 || (!equalWord(poolTokens[0], "a") && !equalWord(poolTokens[0], "an")) {
		return SelectionSyntax{}, false, false
	}
	selection, matched := parsePoolTargetSelection("Destroy target " + joinedEffectText(poolTokens[1:]) + ".")
	if !matched {
		return SelectionSyntax{}, false, false
	}
	return selection, true, true
}

// parsePoolTargetSelection re-parses a synthetic single-target destroy clause and
// returns its target selection, the typed candidate pool of an
// EachPlayerChooseDestroy construct. It accepts only a clause that parses without
// diagnostics into a single destroy effect naming exactly one target, so any
// candidate filter the target grammar cannot type fails closed.
func parsePoolTargetSelection(clause string) (SelectionSyntax, bool) {
	document, diagnostics := Parse(clause, Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		return SelectionSyntax{}, false
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
		return SelectionSyntax{}, false
	}
	effect := ability.Sentences[0].Effects[0]
	if effect.Kind != EffectDestroy || len(effect.Targets) != 1 {
		return SelectionSyntax{}, false
	}
	return effect.Targets[0].Selection, true
}

// eachPlayerChosenThisWayDestroy reports whether sentence is exactly "Destroy
// each permanent chosen this way." — the destroy half of an
// EachPlayerChooseDestroy construct, acting on the set the choose half populated.
func eachPlayerChosenThisWayDestroy(sentence *Sentence) bool {
	if len(sentence.Effects) != 1 || sentence.Effects[0].Kind != EffectDestroy {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(sentence.Text), "Destroy each permanent chosen this way.")
}
