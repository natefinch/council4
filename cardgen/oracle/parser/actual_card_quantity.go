package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseActualCardResultSubject(tokens []shared.Token, start int, atoms Atoms) (dynamicAmountSubject, bool) {
	for i := start + 1; i+2 < len(tokens); i++ {
		outcome, cardDomain, ok := resultThisWayOutcomeKind(tokens[i].Text)
		if !ok || !cardDomain || !effectWordsAt(tokens, i+1, "this", "way") ||
			!dynamicAmountBoundary(tokens, i+3) {
			continue
		}
		noun := tokens[start:i]
		if !validResultNounAtoms(noun, atoms) || !validResultThisWayNoun(noun, true, atoms) ||
			!equalWord(noun[len(noun)-1], "card") && !equalWord(noun[len(noun)-1], "cards") {
			return dynamicAmountSubject{}, false
		}
		selection := parseSelection(noun, atoms)
		if !exactActualCardResultNoun(noun, selection) {
			return dynamicAmountSubject{}, false
		}
		return dynamicAmountSubject{
			amount: EffectAmountSyntax{
				DynamicKind:   EffectDynamicAmountActualCardResultCount,
				ResultOutcome: outcome, Selection: &selection,
			},
			end: i + 3, count: true, plural: equalWord(noun[len(noun)-1], "cards"),
		}, true
	}
	return dynamicAmountSubject{}, false
}

func exactActualCardResultNoun(noun []shared.Token, selection SelectionSyntax) bool {
	if len(selection.RequiredTypesAny) > 1 && !selection.ConjunctiveTypes &&
		(len(selection.ColorsAny) != 0 || len(selection.ExcludedColors) != 0 ||
			len(selection.Supertypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
			len(selection.SubtypesAny) != 0 || len(selection.ExcludedTypes) != 0 ||
			selection.Colorless || selection.Multicolored || selection.Colored) {
		return false
	}
	expected, ok := graveyardCardNoun(selection, false)
	if !ok {
		return false
	}
	if equalWord(noun[len(noun)-1], "cards") {
		expected += "s"
	}
	return strings.EqualFold(joinedEffectText(noun), expected)
}

func parseLeadingActualCardQuantity(context EffectContextKind, pre []shared.Token, perIteration EffectAmountSyntax, atoms Atoms) (EffectAmountSyntax, bool) {
	if context != EffectContextController || !perIteration.Known || perIteration.Value < 1 ||
		!effectWordsAt(pre, 0, "for", "each") {
		return EffectAmountSyntax{}, false
	}
	subject, ok := parseActualCardResultSubject(pre, 2, atoms)
	if !ok || subject.plural || subject.end >= len(pre) || pre[subject.end].Kind != shared.Comma {
		return EffectAmountSyntax{}, false
	}
	rest := subject.end + 1
	if rest < len(pre) && equalWord(pre[rest], "you") {
		rest++
	}
	if rest != len(pre) {
		return EffectAmountSyntax{}, false
	}
	amount := subject.amount
	amount.DynamicForm = EffectDynamicAmountFormForEach
	amount.Multiplier = perIteration.Value
	amount.Span = shared.SpanOf(pre[:subject.end])
	amount.Text = joinedEffectText(pre[:subject.end])
	amount.actualCardPerIteration = &perIteration
	return amount, true
}

// Consume the iterator once, then reuse the action's ordinary exactness check.
func exactLeadingActualCardQuantity(effect *EffectSyntax) (exact, handled bool) {
	if unconsumedLeadingActualResultIterator(effect) {
		return false, true
	}
	if effect.Amount.DynamicKind != EffectDynamicAmountActualCardResultCount ||
		effect.Amount.Span.Start.Offset >= effect.VerbSpan.Start.Offset {
		return false, false
	}
	end := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span.End == effect.Amount.Span.End
	})
	if end < 0 || end+1 >= len(effect.Tokens) || effect.Tokens[end+1].Kind != shared.Comma {
		return false, true
	}
	body := *effect
	body.Tokens = effect.Tokens[end+2:]
	verb := slices.IndexFunc(body.Tokens, func(token shared.Token) bool { return token.Span == effect.VerbSpan })
	if verb < 0 {
		return false, true
	}
	if effect.Amount.actualCardPerIteration == nil {
		return false, true
	}
	body.Amount = *effect.Amount.actualCardPerIteration
	return exactEffectSyntax(&body), true
}

func unconsumedLeadingActualResultIterator(effect *EffectSyntax) bool {
	if effect.Amount.DynamicKind == EffectDynamicAmountActualCardResultCount || !createHasLeadingIterator(effect) {
		return false
	}
	for i, token := range effect.Tokens {
		if token.Span.Start.Offset >= effect.VerbSpan.Start.Offset {
			break
		}
		_, cardDomain, ok := resultThisWayOutcomeKind(token.Text)
		if ok && cardDomain && effectWordsAt(effect.Tokens, i+1, "this", "way") {
			if effect.Context != EffectContextReferencedObjectController {
				return true
			}
			return (effect.Kind != EffectCreate || !exactCreateTokenForEachExiledThisWayEffectSyntax(effect)) &&
				(effect.Kind != EffectDraw || !exactDrawForEachExiledThisWayEffectSyntax(effect)) &&
				(effect.Kind != EffectCloak || !exactCloakForEachExiledThisWayEffectSyntax(effect))
		}
	}
	return false
}

func emitActualCardQuantityOwnership(abilities []Ability) {
	for i := range abilities {
		emitBodyActualCardQuantities(abilities[i].Sentences)
		if abilities[i].Modal != nil {
			for j := range abilities[i].Modal.Options {
				emitBodyActualCardQuantities(abilities[i].Modal.Options[j].Sentences)
			}
		}
	}
}

func emitBodyActualCardQuantities(sentences []Sentence) {
	producers := make(map[EffectKind]*EffectSyntax)
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			if effect.Amount.actualCardPerIteration != nil && ei+1 < len(sentences[si].Effects) {
				effect.Exact = false
			}
			if effect.Amount.DynamicKind == EffectDynamicAmountEventCardCount &&
				effect.Amount.ResultOutcome == EffectDiscard && producers[EffectDiscard] != nil {
				effect.Amount.DynamicKind = EffectDynamicAmountActualCardResultCount
				effect.Amount.Selection = &SelectionSyntax{Kind: SelectionCard}
			}
			if effect.Amount.DynamicKind == EffectDynamicAmountActualCardResultCount {
				producer := producers[effect.Amount.ResultOutcome]
				switch {
				case producer != nil && producer.Exact && !producer.Negated &&
					producer.DelayedTiming == DelayedTimingNone:
					effect.Amount.ProducerClauseID = producer.ClauseID
				case producer == nil && effect.Amount.ResultOutcome == EffectDiscard &&
					effect.Amount.Selection != nil &&
					(strings.EqualFold(effect.Amount.Selection.Text, "card") ||
						strings.EqualFold(effect.Amount.Selection.Text, "cards")):
					effect.Amount.DynamicKind = EffectDynamicAmountEventCardCount
					effect.Amount.Selection = nil
				default:
					effect.Exact = false
				}
			}
			producers[effect.Kind] = effect
		}
	}
}
