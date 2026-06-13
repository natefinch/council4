package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TriggerCardType identifies a literal card type in trigger syntax.
type TriggerCardType uint8

// Literal card types recognized in trigger syntax.
const (
	TriggerCardTypeUnknown TriggerCardType = iota
	TriggerCardTypeArtifact
	TriggerCardTypeBattle
	TriggerCardTypeCreature
	TriggerCardTypeEnchantment
	TriggerCardTypeInstant
	TriggerCardTypeLand
	TriggerCardTypePlaneswalker
	TriggerCardTypeSorcery
)

// TriggerColor identifies a literal color in trigger syntax.
type TriggerColor uint8

// Literal colors recognized in trigger syntax.
const (
	TriggerColorUnknown TriggerColor = iota
	TriggerColorWhite
	TriggerColorBlue
	TriggerColorBlack
	TriggerColorRed
	TriggerColorGreen
)

// TriggerSubtype is a literal subtype in trigger syntax.
type TriggerSubtype string

// TriggerSupertype identifies a literal supertype in trigger syntax.
type TriggerSupertype uint8

// Literal supertypes recognized in trigger syntax.
const (
	TriggerSupertypeUnknown TriggerSupertype = iota
	TriggerSupertypeLegendary
	TriggerSupertypeSnow
)

// TriggerController identifies a literal controller relation in trigger syntax.
type TriggerController uint8

// Literal controller relations recognized in trigger syntax.
const (
	ControllerAny TriggerController = iota
	ControllerYou
	ControllerOpponent
)

// TriggerSelection is typed syntax for a permanent noun phrase in a trigger.
type TriggerSelection struct {
	RequiredTypes    []TriggerCardType
	RequiredTypesAny []TriggerCardType
	ExcludedTypes    []TriggerCardType
	Supertypes       []TriggerSupertype
	SubtypesAny      []TriggerSubtype
	ColorsAny        []TriggerColor
	ExcludedColors   []TriggerColor
	Colorless        bool
	Multicolored     bool
	NonToken         bool
	TokenOnly        bool
	Controller       TriggerController
}

func parseTriggerSelection(tokens []shared.Token) (TriggerSelection, bool) {
	words := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if i+2 < len(tokens) &&
			equalWord(token, "and") &&
			tokens[i+1].Kind == shared.Slash &&
			equalWord(tokens[i+2], "or") {
			words = append(words, "and/or")
			i += 2
			continue
		}
		if token.Kind != shared.Word {
			return TriggerSelection{}, false
		}
		words = append(words, strings.ToLower(token.Text))
	}
	if len(words) == 0 {
		return TriggerSelection{}, false
	}
	selection := TriggerSelection{}
	words, selection.Controller = cutTriggerController(words)
	for len(words) > 0 {
		switch words[0] {
		case "nontoken":
			selection.NonToken = true
		case "token":
			selection.TokenOnly = true
		case "legendary":
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeLegendary)
		case "snow":
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeSnow)
		case "white", "blue", "black", "red", "green":
			selection.ColorsAny = append(selection.ColorsAny, triggerColor(words[0]))
		case "nonwhite", "nonblue", "nonblack", "nonred", "nongreen":
			selection.ExcludedColors = append(selection.ExcludedColors, triggerColor(strings.TrimPrefix(words[0], "non")))
		case "colorless":
			selection.Colorless = true
		case "multicolored":
			selection.Multicolored = true
		default:
			goto noun
		}
		words = words[1:]
	}

noun:
	if len(words) == 0 {
		return TriggerSelection{}, false
	}
	if len(words) == 3 && (words[1] == "and/or" || words[1] == "or") {
		left, leftOK := triggerCardType(words[0])
		right, rightOK := triggerCardType(words[2])
		if leftOK && rightOK {
			selection.RequiredTypesAny = []TriggerCardType{left, right}
			return selection, true
		}
		if leftOK || rightOK || !looksLikeTriggerSubtype(words[0]) || !looksLikeTriggerSubtype(words[2]) {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{TriggerSubtype(singularTriggerWord(words[0])), TriggerSubtype(singularTriggerWord(words[2]))}
		return selection, true
	}
	var subtypeWords []string
	for _, word := range words {
		if rest, prefixed := strings.CutPrefix(word, "non"); prefixed {
			if cardType, ok := triggerCardType(rest); ok {
				selection.ExcludedTypes = append(selection.ExcludedTypes, cardType)
				continue
			}
		}
		if cardType, ok := triggerCardType(word); ok {
			if cardType != TriggerCardTypeUnknown && !slices.Contains(selection.RequiredTypes, cardType) {
				selection.RequiredTypes = append(selection.RequiredTypes, cardType)
			}
			continue
		}
		subtypeWords = append(subtypeWords, word)
	}
	if len(subtypeWords) > 0 {
		subtype := strings.Join(subtypeWords, " ")
		if !looksLikeTriggerSubtype(subtype) {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{TriggerSubtype(singularTriggerWord(subtype))}
	}
	return selection, true
}

func cutTriggerController(words []string) ([]string, TriggerController) {
	for _, suffix := range []struct {
		words      []string
		controller TriggerController
	}{
		{[]string{"you", "control"}, ControllerYou},
		{[]string{"an", "opponent", "controls"}, ControllerOpponent},
		{[]string{"your", "opponents", "control"}, ControllerOpponent},
		{[]string{"you", "don't", "control"}, ControllerOpponent},
	} {
		if len(words) >= len(suffix.words) && slices.Equal(words[len(words)-len(suffix.words):], suffix.words) {
			return words[:len(words)-len(suffix.words)], suffix.controller
		}
	}
	return words, ControllerAny
}

func triggerColor(word string) TriggerColor {
	switch word {
	case "white":
		return TriggerColorWhite
	case "blue":
		return TriggerColorBlue
	case "black":
		return TriggerColorBlack
	case "red":
		return TriggerColorRed
	case "green":
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
	}
}

func triggerCardType(word string) (TriggerCardType, bool) {
	switch singularTriggerWord(word) {
	case "artifact":
		return TriggerCardTypeArtifact, true
	case "battle":
		return TriggerCardTypeBattle, true
	case "creature":
		return TriggerCardTypeCreature, true
	case "enchantment":
		return TriggerCardTypeEnchantment, true
	case "instant":
		return TriggerCardTypeInstant, true
	case "land":
		return TriggerCardTypeLand, true
	case "permanent":
		return TriggerCardTypeUnknown, true
	case "planeswalker":
		return TriggerCardTypePlaneswalker, true
	case "sorcery":
		return TriggerCardTypeSorcery, true
	default:
		return TriggerCardTypeUnknown, false
	}
}

func singularTriggerWord(word string) string {
	return strings.TrimSuffix(word, "s")
}

func looksLikeTriggerSubtype(subject string) bool {
	fields := strings.Fields(subject)
	if len(fields) == 0 || len(fields) > 2 {
		return false
	}
	for _, word := range fields {
		if _, ok := triggerCardType(word); ok ||
			strings.HasPrefix(word, "non") ||
			slices.Contains([]string{"an", "a", "the", "you", "your", "opponent", "or", "and", "but"}, word) {
			return false
		}
	}
	for _, r := range subject {
		if (r >= 'a' && r <= 'z') || r == ' ' || r == '-' || r == '\'' {
			continue
		}
		return false
	}
	return true
}
