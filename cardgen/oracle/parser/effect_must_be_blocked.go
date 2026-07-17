package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseTargetMustBeBlockedEffect recognizes the exact resolving requirement
// "Target <creature selection> must be blocked this combat if able." before the
// generic target parser can absorb the predicate into the target noun phrase.
func parseTargetMustBeBlockedEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := tokens
	if len(words) > 0 && words[len(words)-1].Kind == shared.Period {
		words = words[:len(words)-1]
	}
	must := -1
	for i := 1; i < len(words); i++ {
		if equalWord(words[i], "must") {
			must = i
			break
		}
	}
	if must < 2 ||
		!equalWord(words[0], "target") ||
		len(words)-must != 7 ||
		!effectWordsAt(words, must, "must", "be", "blocked", "this", "combat", "if", "able") {
		return nil, false
	}
	targetTokens := words[:must]
	selectionTokens := words[1:must]
	selection := parseSelection(selectionTokens, atoms)
	if targetSelectionHasUnsupportedQualifier(selectionTokens, atoms) {
		selection = SelectionSyntax{Span: selection.Span, Text: selection.Text}
	}
	if selection.Kind == SelectionUnknown && selectionIsBareTokenTarget(selection) {
		selection.Kind = SelectionPermanent
	}
	if conjunctiveTypeTarget(selection) {
		selection.ConjunctiveTypes = true
	}
	applyLeadingPowerToughnessPrefix(selectionTokens, &selection)
	cardinality := TargetCardinalitySyntax{Min: 1, Max: 1}
	target := TargetSyntax{
		Span:        shared.SpanOf(targetTokens),
		ChoiceSpan:  exactTargetChoiceSpan(words, 0, targetTokens, cardinality, selection),
		Text:        joinedEffectText(targetTokens),
		Cardinality: cardinality,
		Selection:   selection,
		Exact:       exactRuntimeTargetSyntax(targetTokens, cardinality, selection),
	}
	if !target.Exact {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectMustBeBlocked,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[must].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextTarget,
		Duration:   EffectDurationThisCombat,
		Targets:    []TargetSyntax{target},
		Exact:      true,
	}}, true
}
