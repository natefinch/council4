package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// Type conditions reuse selection atoms, but require every token to belong to
// the noun phrase: parseSelection alone deliberately tolerates other qualifiers.
func parseConditionTypeNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	if leftEnd, rightStart, ok := conditionAlternativeConnector(tokens); ok {
		if len(selection.Supertypes) > 0 {
			return ConditionSelection{}, false
		}
		left, leftOK := parseConditionTypeNoun(tokens[:leftEnd], atoms, ConditionSelection{})
		rightTokens := tokens[rightStart:]
		if rest, ok := cutTokenPrefix(rightTokens, "a"); ok {
			rightTokens = rest
		} else if rest, ok := cutTokenPrefix(rightTokens, "an"); ok {
			rightTokens = rest
		}
		right, rightOK := parseConditionTypeNoun(rightTokens, atoms, ConditionSelection{})
		if !leftOK || !rightOK {
			return ConditionSelection{}, false
		}
		if conditionTypeSelectionEmpty(left) || conditionTypeSelectionEmpty(right) {
			return ConditionSelection{}, false
		}
		alternatives := []ConditionSelection{left, right}
		if conditionTypeUnionMembers(left) && conditionTypeUnionMembers(right) {
			for _, side := range alternatives {
				selection.RequiredTypesAny = append(selection.RequiredTypesAny, side.RequiredTypes...)
				selection.RequiredTypesAny = append(selection.RequiredTypesAny, side.RequiredTypesAny...)
			}
		} else {
			selection.AnyOf = append(selection.AnyOf, alternatives...)
		}
		return selection, true
	}
	nounTokens := tokens
	cardNoun := equalWord(tokens[len(tokens)-1], "card") || equalWord(tokens[len(tokens)-1], "cards")
	if cardNoun {
		tokens = tokens[:len(tokens)-1]
	}
	permanentNoun := len(tokens) > 0 && (equalWord(tokens[len(tokens)-1], "permanent") || equalWord(tokens[len(tokens)-1], "permanents"))
	if permanentNoun {
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) == 0 && !permanentNoun {
		if cardNoun && len(selection.Supertypes) > 0 {
			return selection, true
		}
		return ConditionSelection{}, false
	}
	required := make([]TriggerCardType, 0, len(tokens))
	excluded := make([]TriggerCardType, 0, len(tokens))
	for i, token := range tokens {
		if token.Kind == shared.Comma && i > 0 && i+1 < len(tokens) {
			_, before := atoms.ExcludedCardTypeAt(tokens[i-1].Span)
			_, after := atoms.ExcludedCardTypeAt(tokens[i+1].Span)
			if before && after {
				continue
			}
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok {
			mapped := triggerCardTypeFromAtom(cardType)
			if mapped == TriggerCardTypeUnknown || slices.Contains(required, mapped) {
				return ConditionSelection{}, false
			}
			required = append(required, mapped)
		} else if cardType, ok := atoms.ExcludedCardTypeAt(token.Span); ok {
			mapped := triggerCardTypeFromAtom(cardType)
			if mapped == TriggerCardTypeUnknown || slices.Contains(excluded, mapped) {
				return ConditionSelection{}, false
			}
			excluded = append(excluded, mapped)
		} else {
			return ConditionSelection{}, false
		}
	}
	if len(required) == 0 && !cardNoun && !permanentNoun {
		return ConditionSelection{}, false
	}
	parsed := parseSelection(nounTokens, atoms)
	// The strict atom pass above supplies conjunction semantics; shared
	// parseSelection supplies the canonical exclusion vocabulary.
	if len(parsed.RequiredTypesAny) != len(required) || len(parsed.ExcludedTypes) != len(excluded) {
		return ConditionSelection{}, false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, required...)
	for _, cardType := range parsed.ExcludedTypes {
		selection.ExcludedTypes = append(selection.ExcludedTypes, triggerCardTypeFromAtom(cardType))
	}
	if permanentNoun && cardNoun {
		selection.RequiredTypesAny = append(selection.RequiredTypesAny,
			TriggerCardTypeArtifact, TriggerCardTypeBattle, TriggerCardTypeCreature,
			TriggerCardTypeEnchantment, TriggerCardTypeLand, TriggerCardTypePlaneswalker)
	}
	return selection, true
}

func conditionTypeUnionMembers(selection ConditionSelection) bool {
	return len(selection.AnyOf) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		(len(selection.RequiredTypes) == 1 && len(selection.RequiredTypesAny) == 0 ||
			len(selection.RequiredTypes) == 0 && len(selection.RequiredTypesAny) > 0)
}

func conditionTypeSelectionEmpty(selection ConditionSelection) bool {
	return len(selection.RequiredTypes) == 0 && len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 && len(selection.AnyOf) == 0
}
