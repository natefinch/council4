package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func emitSourceAbilityCostReduction(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Kind != AbilityActivated || ability.Modal != nil ||
			len(ability.Sentences) < 2 {
			continue
		}
		sentence := ability.Sentences[len(ability.Sentences)-1]
		amount, selection, ok := sourceAbilityCostReduction(sentence.Tokens, ability.Atoms)
		if !ok {
			continue
		}
		ability.SourceAbilityCostReduction = &SourceAbilityCostReductionSyntax{
			Span:           sentence.Span,
			Amount:         amount,
			CountSelection: selection,
		}
	}
}

func sourceAbilityCostReduction(tokens []shared.Token, atoms Atoms) (int, SelectionSyntax, bool) {
	if len(tokens) < 11 || tokens[len(tokens)-1].Kind != shared.Period ||
		!effectWordsAt(tokens, 0, "this", "ability", "costs") ||
		tokens[3].Kind != shared.Symbol {
		return 0, SelectionSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[3].Text)
	if !ok || amount <= 0 ||
		!effectWordsAt(tokens, 4, "less", "to", "activate", "for", "each") {
		return 0, SelectionSyntax{}, false
	}
	subject, ok := parseDynamicCountSubject(tokens, 9, atoms)
	if !ok || !subject.count || subject.end != len(tokens)-1 ||
		subject.amount.Selection == nil ||
		subject.amount.DynamicKind != EffectDynamicAmountCount {
		return 0, SelectionSyntax{}, false
	}
	selection := *subject.amount.Selection
	if selection.Zone != 0 {
		return 0, SelectionSyntax{}, false
	}
	return amount, selection, true
}
