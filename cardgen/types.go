package cardgen

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/natefinch/council4/mtg/game/types"
)

// ParsedTypeLine holds the parsed components of a Scryfall type line.
type ParsedTypeLine struct {
	Supertypes []string
	Types      []string
	Subtypes   []string
}

var knownSupertypes = map[string]bool{
	"Legendary": true,
	"Basic":     true,
	"Snow":      true,
	"World":     true,
	"Ongoing":   true,
}

var knownTypes = map[string]bool{
	"Land":         true,
	"Creature":     true,
	"Artifact":     true,
	"Enchantment":  true,
	"Instant":      true,
	"Sorcery":      true,
	"Planeswalker": true,
	"Battle":       true,
	"Kindred":      true,
}

// ParseTypeLine splits a Scryfall type line (e.g., "Legendary Creature — Angel")
// into supertypes, types, and subtypes.
func ParseTypeLine(typeLine string) ParsedTypeLine {
	var result ParsedTypeLine

	// Split on em-dash (—) to separate main types from subtypes.
	parts := strings.SplitN(typeLine, "—", 2)
	mainPart := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		subtypePart := strings.TrimSpace(parts[1])
		result.Subtypes = append(result.Subtypes, strings.Fields(subtypePart)...)
	}

	for word := range strings.FieldsSeq(mainPart) {
		if knownSupertypes[word] {
			result.Supertypes = append(result.Supertypes, word)
		} else if knownTypes[word] {
			result.Types = append(result.Types, word)
		}
		// Skip "//" for double-faced card type lines if encountered.
	}

	return result
}

// SupertypeToLiteral converts a supertype name to its Go constant name.
func SupertypeToLiteral(name string) string {
	switch name {
	case "Legendary":
		return "types.Legendary"
	case "Basic":
		return "types.Basic"
	case "Snow":
		return "types.Snow"
	case "World":
		return "types.World"
	case "Ongoing":
		return "types.Ongoing"
	default:
		return "/* unknown supertype: " + name + " */"
	}
}

// CardTypeToLiteral converts a card type name to its Go constant name.
func CardTypeToLiteral(name string) string {
	switch name {
	case "Land":
		return "types.Land"
	case "Creature":
		return "types.Creature"
	case "Artifact":
		return "types.Artifact"
	case "Enchantment":
		return "types.Enchantment"
	case "Instant":
		return "types.Instant"
	case "Sorcery":
		return "types.Sorcery"
	case "Planeswalker":
		return "types.Planeswalker"
	case "Battle":
		return "types.Battle"
	case "Kindred":
		return "types.Kindred"
	default:
		return "/* unknown type: " + name + " */"
	}
}

var subtypeLiteralTypes = map[string]struct {
	cardType types.Card
}{
	"Artifact":    {cardType: types.Artifact},
	"Creature":    {cardType: types.Creature},
	"Enchantment": {cardType: types.Enchantment},
	"Kindred":     {cardType: types.Kindred},
	"Land":        {cardType: types.Land},
}

// SubtypeToLiteral converts a subtype name to its Go constant for the card type
// family where that subtype is used. Unknown subtypes fall back to a string
// literal so generation can continue while the central subtype list is updated.
func SubtypeToLiteral(name string, cardTypes []string) string {
	for _, typ := range cardTypes {
		info, ok := subtypeLiteralTypes[typ]
		if !ok {
			continue
		}
		if types.KnownSubtypeForType(info.cardType, types.Sub(name)) {
			return "types." + goIdentifierSuffix(name)
		}
	}
	return "types.Sub(" + strconv.Quote(name) + ")"
}

func goIdentifierSuffix(name string) string {
	var b strings.Builder
	capitalizeNext := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			r = unicode.ToUpper(r)
			capitalizeNext = false
		}
		_, _ = b.WriteRune(r)
	}
	return b.String()
}

// ColorToLiteral converts a Scryfall single-letter color to a Go mana.Color name.
func ColorToLiteral(letter string) string {
	switch letter {
	case "W":
		return "mana.White"
	case "U":
		return "mana.Blue"
	case "B":
		return "mana.Black"
	case "R":
		return "mana.Red"
	case "G":
		return "mana.Green"
	default:
		return "/* unknown color: " + letter + " */"
	}
}
