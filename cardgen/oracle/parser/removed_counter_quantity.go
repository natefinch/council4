package parser

import (
	"fmt"
	"slices"
	"strings"
)

func exactRemoveAllOfKindCountersEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown {
		return false
	}
	body := *effect
	body.References = slices.DeleteFunc(slices.Clone(effect.References), func(reference Reference) bool {
		return reference.Span.Start.Offset < effect.VerbSpan.End.Offset
	})
	object, ok := exactRemoveCounterObjectText(&body)
	return ok && strings.EqualFold(exactEffectClauseText(effect),
		fmt.Sprintf("Remove all %s counters from %s.", effect.CounterKind.String(), object))
}

func emitRemovedCounterQuantityOwnership(abilities []Ability) {
	for i := range abilities {
		emitBodyRemovedCounterQuantities(abilities[i].Sentences)
		if abilities[i].Modal != nil {
			for j := range abilities[i].Modal.Options {
				emitBodyRemovedCounterQuantities(abilities[i].Modal.Options[j].Sentences)
			}
		}
	}
}

func emitBodyRemovedCounterQuantities(sentences []Sentence) {
	var previous *EffectSyntax
	var removal *EffectSyntax
	ambiguous := false
	consumed := 0
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			amount := &effect.Amount
			explicit := amount.DynamicKind == EffectDynamicAmountRemovedCounterCount
			anaphor := strings.EqualFold(amount.Text, "that much") || strings.EqualFold(amount.Text, "that many")
			if explicit || anaphor && removal != nil {
				amount.DynamicKind = EffectDynamicAmountRemovedCounterCount
				amount.Multiplier = max(1, amount.Multiplier)
				if removal != nil && (explicit || previous == removal) && !ambiguous && removal.Exact && !removal.Negated &&
					removal.DelayedTiming == DelayedTimingNone {
					amount.ProducerClauseID = removal.ClauseID
					consumed = removal.ClauseID
				}
			}
			if effect.Kind == EffectRemoveCounter {
				ambiguous = removal != nil && removal.ClauseID != consumed
				removal = effect
			}
			previous = effect
		}
	}
}
