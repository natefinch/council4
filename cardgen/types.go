package cardgen

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/natefinch/council4/mtg/game"
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
		for _, st := range strings.Fields(subtypePart) {
			result.Subtypes = append(result.Subtypes, st)
		}
	}

	for _, word := range strings.Fields(mainPart) {
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
		return "game.Legendary"
	case "Basic":
		return "game.Basic"
	case "Snow":
		return "game.Snow"
	case "World":
		return "game.World"
	case "Ongoing":
		return "game.Ongoing"
	default:
		return "/* unknown supertype: " + name + " */"
	}
}

// CardTypeToLiteral converts a card type name to its Go constant name.
func CardTypeToLiteral(name string) string {
	switch name {
	case "Land":
		return "game.TypeLand"
	case "Creature":
		return "game.TypeCreature"
	case "Artifact":
		return "game.TypeArtifact"
	case "Enchantment":
		return "game.TypeEnchantment"
	case "Instant":
		return "game.TypeInstant"
	case "Sorcery":
		return "game.TypeSorcery"
	case "Planeswalker":
		return "game.TypePlaneswalker"
	case "Battle":
		return "game.TypeBattle"
	case "Kindred":
		return "game.TypeKindred"
	default:
		return "/* unknown type: " + name + " */"
	}
}

var subtypeLiteralTypes = map[string]struct {
	cardType game.CardType
	prefix   string
}{
	"Artifact":    {cardType: game.TypeArtifact, prefix: "ArtifactSubtype"},
	"Creature":    {cardType: game.TypeCreature, prefix: "CreatureSubtype"},
	"Enchantment": {cardType: game.TypeEnchantment, prefix: "EnchantmentSubtype"},
	"Kindred":     {cardType: game.TypeKindred, prefix: "CreatureSubtype"},
	"Land":        {cardType: game.TypeLand, prefix: "LandSubtype"},
}

// SubtypeToLiteral converts a subtype name to its Go constant for the card type
// family where that subtype is used. Unknown subtypes fall back to a string
// literal so generation can continue while the central subtype list is updated.
func SubtypeToLiteral(name string, types []string) string {
	for _, typ := range types {
		info, ok := subtypeLiteralTypes[typ]
		if !ok {
			continue
		}
		if game.KnownSubtypeForType(info.cardType, name) {
			return "game." + info.prefix + goIdentifierSuffix(name)
		}
	}
	return strconv.Quote(name)
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
		b.WriteRune(r)
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
