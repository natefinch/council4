package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// emitActivationCostReduction recognizes the exact trailing rider "This
// ability costs {N} less to activate for each <battlefield object>." The rider
// is represented separately from resolving effects because it changes the
// activation cost while the ability is announced.
func emitActivationCostReduction(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Kind != AbilityActivated || ability.Modal != nil || len(ability.Sentences) != 2 {
			continue
		}
		first := &ability.Sentences[0]
		rider := &ability.Sentences[1]
		if len(first.Effects) == 0 || len(rider.Effects) != 0 {
			continue
		}
		reduction, ok := parseActivationCostReduction(*rider, ability.Atoms)
		if !ok {
			continue
		}
		rider.ActivationCostReduction = &reduction
		for j := range first.Effects {
			first.Effects[j].HasUnrecognizedSibling = false
			first.Effects[j].Exact = exactEffectSyntax(&first.Effects[j])
		}
	}
}

func parseActivationCostReduction(sentence Sentence, atoms Atoms) (ActivationCostReductionSyntax, bool) {
	tokens := eventHistorySemanticTokens(sentence.Tokens, nil, nil)
	if len(tokens) < 11 || tokens[len(tokens)-1].Kind != shared.Period {
		return ActivationCostReductionSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, "this", "ability", "costs") || tokens[3].Kind != shared.Symbol {
		return ActivationCostReductionSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[3].Text)
	if !ok || amount <= 0 || !effectWordsAt(tokens, 4, "less", "to", "activate", "for", "each") {
		return ActivationCostReductionSyntax{}, false
	}
	subject, ok := parseDynamicCountSubject(tokens, 9, atoms)
	if !ok || !subject.count || subject.end != len(tokens)-1 ||
		subject.amount.DynamicKind != EffectDynamicAmountCount ||
		subject.amount.Selection == nil ||
		subject.amount.Selection.Zone != zone.None {
		return ActivationCostReductionSyntax{}, false
	}
	subject.amount.Span = shared.SpanOf(tokens[9 : len(tokens)-1])
	subject.amount.Text = shared.SliceSpan(sentence.Text, costRelativeSpan(subject.amount.Span, sentence.Span.Start.Offset))
	subject.amount.DynamicForm = EffectDynamicAmountFormForEach
	subject.amount.Multiplier = 1
	return ActivationCostReductionSyntax{
		Span:               sentence.Span,
		PerObjectReduction: amount,
		Amount:             subject.amount,
	}, true
}
