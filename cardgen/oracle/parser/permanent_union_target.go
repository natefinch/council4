package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parsePermanentUnionTarget proves both complete permanent predicates before
// assigning the union's domain. It does not reuse parseSelection's cross-zone
// or flattened union guesses; each branch must reconstruct independently.
func parsePermanentUnionTarget(
	tokens []shared.Token,
	atoms Atoms,
	start, targetIndex int,
	cardinality TargetCardinalitySyntax,
) (TargetSyntax, bool) {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) || start != targetIndex {
		return TargetSyntax{}, false
	}
	end := targetSyntaxEnd(tokens, atoms, targetIndex+1)
	phrase := tokens[targetIndex+1 : end]
	if !slices.ContainsFunc(phrase, func(t shared.Token) bool { return equalWord(t, "or") }) {
		return TargetSyntax{}, false
	}
	targetTokens := tokens[start:end]
	parsed := parseSelection(phrase, atoms)
	if exactRuntimeTargetSyntax(targetTokens, cardinality, parsed) {
		// Keep already exact type/subtype unions and conjunctions unchanged.
		return TargetSyntax{}, false
	}
	controller := parsed.Controller
	switch controller {
	case SelectionControllerAny:
	case SelectionControllerYou, SelectionControllerOpponent, SelectionControllerNotYou:
		words, _ := targetControllerWords(controller)
		if !endsWithWords(phrase, words...) {
			return TargetSyntax{}, false
		}
		phrase = phrase[:len(phrase)-len(words)]
	default:
		return TargetSyntax{}, false
	}
	for i, token := range phrase {
		if !equalWord(token, "or") || i == 0 || i == len(phrase)-1 {
			continue
		}
		left, leftOK := permanentUnionMember(phrase[:i], atoms)
		right, rightOK := permanentUnionMember(phrase[i+1:], atoms)
		if !leftOK || !rightOK {
			continue
		}
		selection := SelectionSyntax{
			Span: shared.SpanOf(tokens[targetIndex+1 : end]),
			Text: joinedEffectText(tokens[targetIndex+1 : end]),
		}
		exact := permanentUnionScopeProven(left, right)
		if exact {
			selection.Kind = SelectionPermanent
			selection.Controller = controller
			selection.Alternatives = []SelectionSyntax{left, right}
		}
		return TargetSyntax{
			Span:        shared.SpanOf(targetTokens),
			Text:        joinedEffectText(targetTokens),
			Cardinality: cardinality,
			Selection:   selection,
			Exact:       exact,
		}, true
	}
	return TargetSyntax{}, false
}

func permanentUnionMember(tokens []shared.Token, atoms Atoms) (SelectionSyntax, bool) {
	member := parseSelection(tokens, atoms)
	switch member.Kind {
	case SelectionArtifact, SelectionCreature, SelectionEnchantment, SelectionLand,
		SelectionPermanent, SelectionPlaneswalker, SelectionBattle:
	case SelectionUnknown:
		if len(member.SubtypesAny) != 1 {
			return SelectionSyntax{}, false
		}
	default:
		return SelectionSyntax{}, false
	}
	if member.Controller != SelectionControllerAny || len(member.Alternatives) != 0 ||
		member.Another || member.Other || member.ManaValueX {
		return SelectionSyntax{}, false
	}
	member.ConjunctiveTypes = conjunctiveTypeTarget(member)
	words, ok := permanentSelectionQualifierWords(member)
	if !ok || !strings.EqualFold(joinedEffectText(tokens), strings.Join(words, " ")) {
		return SelectionSyntax{}, false
	}
	return disjunctMemberSelection(member), true
}

func permanentUnionScopeProven(left, right SelectionSyntax) bool {
	if permanentUnionSharesNoun(left, right) || permanentUnionSharesNoun(right, left) {
		return false
	}
	// Every canonical with/without dimension must be explicit on both
	// branches; a right-only tail may restrict the whole union.
	if (right.Keyword != KeywordUnknown || right.ExcludedKeyword != KeywordUnknown) &&
		left.Keyword == KeywordUnknown && left.ExcludedKeyword == KeywordUnknown ||
		right.MatchManaValue && !left.MatchManaValue ||
		right.MatchPower && !right.PowerToughnessPrefix && (!left.MatchPower || left.PowerToughnessPrefix) ||
		right.MatchToughness && !right.PowerToughnessPrefix && (!left.MatchToughness || left.PowerToughnessPrefix) ||
		(right.PowerLessThanSource || right.PowerGreaterThanSource) &&
			!left.PowerLessThanSource && !left.PowerGreaterThanSource {
		return false
	}
	// Unlike a branch-local color, leading state/determiner adjectives can
	// govern the entire list. Independent scope requires a second explicit
	// adjective in that dimension, not an unqualified branch.
	if (left.Tapped || left.Untapped || left.Attacking || left.Blocking) &&
		!right.Tapped && !right.Untapped && !right.Attacking && !right.Blocking ||
		(left.TokenOnly || left.NonToken) && !right.TokenOnly && !right.NonToken ||
		(left.Colorless || left.Multicolored) && !right.Colorless && !right.Multicolored ||
		len(left.Supertypes) != 0 && len(right.Supertypes) == 0 {
		return false
	}
	return true
}

func permanentUnionSharesNoun(bare, typed SelectionSyntax) bool {
	if bare.Kind != SelectionUnknown || len(bare.SubtypesAny) != 1 || len(typed.SubtypesAny) != 1 {
		return false
	}
	for _, kind := range typed.RequiredTypesAny {
		if SubtypeMatchesCardType(bare.SubtypesAny[0], kind) &&
			SubtypeMatchesCardType(typed.SubtypesAny[0], kind) {
			return true
		}
	}
	return false
}

func permanentUnionScopeDiagnostics(abilities []Ability) []shared.Diagnostic {
	var diagnostics []shared.Diagnostic
	appendSentences := func(sentences []Sentence, atoms Atoms) {
		for _, sentence := range sentences {
			for _, target := range sentence.Targets {
				if target.Exact || target.Cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
					continue
				}
				tokens := tokensWithinParserSpan(sentence.Tokens, target.Span)
				if len(tokens) < 2 || !equalWord(tokens[0], "target") {
					continue
				}
				if parsed, recognized := parsePermanentUnionTarget(tokens, atoms, 0, 0, target.Cardinality); recognized && !parsed.Exact {
					diagnostics = append(diagnostics, shared.Diagnostic{
						Severity: shared.SeverityWarning,
						Summary:  "unsupported permanent union scope",
						Detail:   "shared noun or qualifier scope is not proven by independent permanent predicates",
						Span:     target.Span,
					})
				}
			}
		}
	}
	for _, ability := range abilities {
		appendSentences(ability.Sentences, ability.Atoms)
		if ability.Modal != nil {
			for _, mode := range ability.Modal.Options {
				appendSentences(mode.Sentences, mode.Atoms)
			}
		}
		if ability.DiceTable != nil {
			for _, row := range ability.DiceTable.Rows {
				appendSentences(row.Sentences, row.Atoms)
			}
		}
	}
	return diagnostics
}
