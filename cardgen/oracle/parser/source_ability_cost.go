package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func emitSourceAbilityCostReduction(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal != nil || len(ability.Sentences) < 2 {
			continue
		}
		sentence := ability.Sentences[len(ability.Sentences)-1]
		if ability.Kind == AbilityActivated {
			amount, selection, ok := sourceAbilityCostReduction(sentence.Tokens, ability.Atoms)
			if !ok {
				continue
			}
			ability.SourceAbilityCostReduction = &SourceAbilityCostReductionSyntax{
				Span:           sentence.Span,
				Amount:         amount,
				CountSelection: selection,
			}
			continue
		}
		// A keyword ability (Equip) carries a self-referential flat reduction on
		// its trailing sentence: "This ability costs {N} less to activate" with an
		// optional gating condition ("... if you're the monarch."). The keyword
		// lowering builds the conditional cost modifier from the amount and the
		// ability's condition clause. The flat form is recognized only for keyword
		// (static) abilities: activated-ability lowering does not yet consume it,
		// so leaving it unrecognized there fails closed rather than silently
		// dropping the reduction.
		if ability.Kind == AbilityStatic {
			if amount, span, ok := sourceAbilityFlatCostReduction(sentence, ability); ok {
				ability.SourceAbilityCostReduction = &SourceAbilityCostReductionSyntax{
					Span:   span,
					Amount: amount,
				}
			}
		}
	}
}

// sourceAbilityFlatCostReduction recognizes the flat "This ability costs {N}
// less to activate" reduction, optionally followed by a gating condition. It
// returns the reduction amount and the source span of the "This ability costs
// {N} less to activate" clause. A trailing condition is excluded from that span
// because the ability carries it as its own condition clause, which lowering
// consumes and covers separately.
func sourceAbilityFlatCostReduction(sentence Sentence, ability *Ability) (int, shared.Span, bool) {
	tokens := sentence.Tokens
	if len(tokens) < 8 ||
		!effectWordsAt(tokens, 0, "this", "ability", "costs") ||
		tokens[3].Kind != shared.Symbol {
		return 0, shared.Span{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[3].Text)
	if !ok || amount <= 0 || !effectWordsAt(tokens, 4, "less", "to", "activate") {
		return 0, shared.Span{}, false
	}
	span := shared.Span{Start: tokens[0].Span.Start, End: tokens[6].Span.End}
	rest := tokens[7:]
	// Plain flat form: nothing but the closing period follows the clause.
	if len(rest) == 1 && rest[0].Kind == shared.Period {
		return amount, span, true
	}
	// Conditional flat form: the remaining tokens form a gating condition the
	// ability carries as its own condition clause beginning at the same offset.
	condStart := tokens[7].Span.Start.Offset
	for j := range ability.ConditionClauses {
		if ability.ConditionClauses[j].Span.Start.Offset == condStart {
			return amount, span, true
		}
	}
	return 0, shared.Span{}, false
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
