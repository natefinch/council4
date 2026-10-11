package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
)

// parseConditionSelection parses a permanent noun phrase into a typed selection,
// consuming card-type, subtype, color, and supertype atoms by span. It fails
// closed unless every token belongs to a recognized production.
func parseConditionSelection(tokens []shared.Token, atoms Atoms) (ConditionSelection, bool) {
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	var selection ConditionSelection
	// Trailing "with <qualifier>" clause: either "with power <n> or greater" or
	// "with <keyword>" (e.g. "a creature with flying").
	if idx := tokenWordIndex(tokens, "with"); idx >= 0 {
		qualifier := tokens[idx+1:]
		if !parseConditionPowerQualifier(qualifier, &selection) &&
			!parseConditionKeywordQualifier(qualifier, &selection) {
			return ConditionSelection{}, false
		}
		tokens = tokens[:idx]
	}
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	// Leading tapped/untapped state.
	switch {
	case equalWord(tokens[0], "tapped"):
		selection.Tapped = ConditionTappedTrue
		tokens = tokens[1:]
	case equalWord(tokens[0], "untapped"):
		selection.Tapped = ConditionTappedFalse
		tokens = tokens[1:]
	default:
	}
	// Leading supertypes (basic/snow/legendary).
	for len(tokens) > 0 {
		supertype, ok := conditionSupertypeAtom(tokens[0].Span, atoms)
		if !ok {
			break
		}
		selection.Supertypes = append(selection.Supertypes, supertype)
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return selection, len(selection.Supertypes) > 0
	}
	return parseConditionNoun(tokens, atoms, selection)
}

func parseConditionNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if clause, ok := parseConditionTypeNoun(tokens, atoms, selection); ok {
		return clause, true
	}
	for _, token := range tokens {
		if token.Kind == shared.Comma {
			return parseConditionSubtypeList(tokens, atoms, selection)
		}
	}
	if leftEnd, rightStart, ok := conditionAlternativeConnector(tokens); ok {
		return parseConditionAlternativeNoun(tokens[:leftEnd], tokens[rightStart:], atoms, selection)
	}
	// Color-qualified "<colors> creature|permanent".
	if clause, ok := parseConditionColorQualified(tokens, atoms, selection); ok {
		return clause, true
	}
	// A bare permanent (no required type), e.g. "permanent" or "permanents".
	if tokenWordsEqual(tokens, "permanent") || tokenWordsEqual(tokens, "permanents") {
		return selection, true
	}
	// A bare token, e.g. "you control a token".
	if tokenWordsEqual(tokens, "token") || tokenWordsEqual(tokens, "tokens") {
		selection.TokenOnly = true
		return selection, true
	}
	// A subtype noun: creature, land, or "<name> planeswalker".
	return parseConditionSubtypeNoun(tokens, atoms, selection)
}

func parseConditionSubtypeNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(tokens) >= 3 &&
		(equalWord(tokens[len(tokens)-1], "card") || equalWord(tokens[len(tokens)-1], "cards")) &&
		(equalWord(tokens[len(tokens)-2], "permanent") || equalWord(tokens[len(tokens)-2], "permanents")) {
		subtype, subtypeOK := atoms.SubtypeAt(shared.SpanOf(tokens[:len(tokens)-2]))
		qualified, qualifierOK := parseConditionTypeNoun(tokens[len(tokens)-2:], atoms, selection)
		if subtypeOK && qualifierOK {
			qualified.SubtypesAny = append(qualified.SubtypesAny, subtype)
			return qualified, true
		}
		return ConditionSelection{}, false
	}
	if len(tokens) > 1 && (equalWord(tokens[len(tokens)-1], "card") || equalWord(tokens[len(tokens)-1], "cards")) {
		tokens = tokens[:len(tokens)-1]
	}
	span := shared.SpanOf(tokens)
	if subtype, ok := atoms.SubtypeAt(span); ok {
		selection.SubtypesAny = append(selection.SubtypesAny, subtype)
		return selection, true
	}
	if len(tokens) >= 2 && equalWord(tokens[len(tokens)-1], "planeswalker") {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypePlaneswalker); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypePlaneswalker)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	// A typed subtype noun "<name> creature", e.g. "a Griffin creature".
	if len(tokens) >= 2 &&
		(equalWord(tokens[len(tokens)-1], "creature") || equalWord(tokens[len(tokens)-1], "creatures")) {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypeCreature); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	return ConditionSelection{}, false
}

// conditionAlternativeConnector locates a two-member "or" or "and/or" union.
// The lexer splits "and/or" into Word("and"), Slash, Word("or").
func conditionAlternativeConnector(tokens []shared.Token) (leftEnd, rightStart int, ok bool) {
	for i := range tokens {
		if equalWord(tokens[i], "or") {
			return i, i + 1, i > 0 && i+1 < len(tokens)
		}
		if i+2 < len(tokens) &&
			equalWord(tokens[i], "and") &&
			tokens[i+1].Kind == shared.Slash &&
			equalWord(tokens[i+2], "or") {
			return i, i + 3, i > 0 && i+3 < len(tokens)
		}
	}
	return 0, 0, false
}

func parseConditionAlternativeNoun(left, right []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(left) == 0 || len(right) == 0 {
		return ConditionSelection{}, false
	}
	combined := make([]shared.Token, 0, len(left)+1+len(right))
	combined = append(combined, left...)
	combined = append(combined, shared.Token{Kind: shared.Word, Text: "or"})
	combined = append(combined, right...)
	if clause, ok := parseConditionColorQualified(combined, atoms, selection); ok {
		return clause, true
	}
	if trimmed, ok := cutTokenPrefix(right, "a"); ok {
		right = trimmed
	} else if trimmed, ok := cutTokenPrefix(right, "an"); ok {
		right = trimmed
	}
	// Land subtype disjunction ("a Forest or an Island") carries the Land card
	// type so the matched permanent must be a land of either basic type.
	leftLand, leftLandOK := conditionSubtypeAtom(shared.SpanOf(left), atoms, CardTypeLand)
	rightLand, rightLandOK := conditionSubtypeAtom(shared.SpanOf(right), atoms, CardTypeLand)
	if leftLandOK && rightLandOK {
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeLand)
		selection.SubtypesAny = append(selection.SubtypesAny, leftLand, rightLand)
		return selection, true
	}
	// Generic subtype disjunction ("another Wolf or Werewolf"). Each side names a
	// subtype of any card type and the match constrains only the subtype, exactly
	// like the single-subtype noun production, so a permanent matches if it has
	// either named subtype.
	leftSub, leftOK := atoms.SubtypeAt(shared.SpanOf(left))
	rightSub, rightOK := atoms.SubtypeAt(shared.SpanOf(right))
	if !leftOK || !rightOK {
		return ConditionSelection{}, false
	}
	selection.SubtypesAny = append(selection.SubtypesAny, leftSub, rightSub)
	return selection, true
}

// parseConditionColorQualified handles a color-qualified creature, permanent,
// or card noun, including "colorless" and "multicolored".
func parseConditionColorQualified(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(tokens) < 2 {
		return ConditionSelection{}, false
	}
	last := tokens[len(tokens)-1]
	colorTokens := tokens[:len(tokens)-1]
	if (equalWord(last, "card") || equalWord(last, "cards")) && len(tokens) > 2 {
		if cardType, ok := atoms.CardTypeAt(tokens[len(tokens)-2].Span); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, triggerCardTypeFromAtom(cardType))
			colorTokens = tokens[:len(tokens)-2]
		}
	}
	switch {
	case equalWord(last, "creature"), equalWord(last, "creatures"):
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
	case equalWord(last, "permanent"), equalWord(last, "permanents"),
		equalWord(last, "card"), equalWord(last, "cards"):
	default:
		return ConditionSelection{}, false
	}
	if tokenWordsEqual(colorTokens, "colorless") {
		selection.Colorless = true
		return selection, true
	}
	if tokenWordsEqual(colorTokens, "multicolored") {
		selection.Multicolored = true
		return selection, true
	}
	for _, token := range colorTokens {
		if equalWord(token, "or") {
			continue
		}
		color, ok := atoms.ColorAt(token.Span)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, triggerColorFromAtom(color))
	}
	if len(selection.ColorsAny) == 0 {
		return ConditionSelection{}, false
	}
	return selection, true
}

func parseConditionPowerQualifier(tokens []shared.Token, selection *ConditionSelection) bool {
	attribute, rest, ok := cutConditionAttributeWord(tokens)
	if !ok {
		return false
	}
	comparison, ok := conditionFixedAttributeComparison(attribute, rest)
	if !ok {
		return false
	}
	if attribute == ConditionAttributePower && comparison.Op == compare.GreaterOrEqual {
		selection.PowerAtLeast = comparison.Value
		selection.MatchPowerAtLeast = true
		return true
	}
	selection.AttributeCompare = comparison
	return true
}

// parseConditionKeywordQualifier recognizes a single keyword name following
// "with" (e.g. "a creature with flying"). The qualifier tokens must form exactly
// one keyword name; trailing text fails closed.
func parseConditionKeywordQualifier(tokens []shared.Token, selection *ConditionSelection) bool {
	kind, length, ok := recognizeKeywordNameAt(tokens, 0)
	if !ok || length != len(tokens) {
		return false
	}
	selection.Keyword = kind
	return true
}
