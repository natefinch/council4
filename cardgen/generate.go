package cardgen

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// GenerateCardSource generates a Go source file for a CardDef from Scryfall data.
// The file belongs to the given package name (e.g., "l" for the l/ subdirectory).
func GenerateCardSource(card *ScryfallCard, pkgName string) (string, error) {
	var b strings.Builder

	needsMana := card.ManaCost != "" || len(card.Colors) > 0 || len(card.ColorIdentity) > 0

	b.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	b.WriteString("import (\n")
	b.WriteString("\t\"github.com/natefinch/council4/mtg/game\"\n")
	if needsMana {
		b.WriteString("\t\"github.com/natefinch/council4/mtg/game/mana\"\n")
	}
	b.WriteString(")\n\n")

	// Oracle text comment.
	b.WriteString(fmt.Sprintf("// %s\n", card.Name))
	b.WriteString("//\n")
	b.WriteString(fmt.Sprintf("// Type: %s\n", card.TypeLine))
	if card.ManaCost != "" {
		b.WriteString(fmt.Sprintf("// Cost: %s\n", card.ManaCost))
	}
	b.WriteString("//\n")
	b.WriteString("// Oracle text:\n")
	for _, line := range strings.Split(card.OracleText, "\n") {
		b.WriteString(fmt.Sprintf("//   %s\n", line))
	}
	b.WriteString("//\n")
	b.WriteString("// TODO: Fill in Abilities from oracle text.\n")

	varName := CardNameToVarName(card.Name)
	b.WriteString(fmt.Sprintf("\nvar %s = &game.CardDef{\n", varName))

	// Name
	b.WriteString(fmt.Sprintf("\tName: %q,\n", card.Name))

	// ManaCost
	costLiteral, err := ParseManaCostLiteral(card.ManaCost)
	if err != nil {
		return "", fmt.Errorf("parsing mana cost for %s: %w", card.Name, err)
	}
	if costLiteral != "" {
		b.WriteString(fmt.Sprintf("\tManaCost: %s,\n", costLiteral))
	}

	// ManaValue
	b.WriteString(fmt.Sprintf("\tManaValue: %d,\n", int(card.CMC)))

	// Colors
	if len(card.Colors) > 0 {
		var colorLiterals []string
		for _, c := range card.Colors {
			colorLiterals = append(colorLiterals, ColorToLiteral(c))
		}
		b.WriteString(fmt.Sprintf("\tColors: []mana.Color{%s},\n", strings.Join(colorLiterals, ", ")))
	}

	// ColorIdentity
	if len(card.ColorIdentity) > 0 {
		var ciLiterals []string
		for _, c := range card.ColorIdentity {
			ciLiterals = append(ciLiterals, ColorToLiteral(c))
		}
		b.WriteString(fmt.Sprintf("\tColorIdentity: mana.NewColorIdentity(%s),\n", strings.Join(ciLiterals, ", ")))
	}

	// Supertypes, Types, Subtypes
	parsed := ParseTypeLine(card.TypeLine)

	if len(parsed.Supertypes) > 0 {
		var stLiterals []string
		for _, st := range parsed.Supertypes {
			stLiterals = append(stLiterals, SupertypeToLiteral(st))
		}
		b.WriteString(fmt.Sprintf("\tSupertypes: []game.Supertype{%s},\n", strings.Join(stLiterals, ", ")))
	}

	if len(parsed.Types) > 0 {
		var tLiterals []string
		for _, t := range parsed.Types {
			tLiterals = append(tLiterals, CardTypeToLiteral(t))
		}
		b.WriteString(fmt.Sprintf("\tTypes: []game.CardType{%s},\n", strings.Join(tLiterals, ", ")))
	}

	if len(parsed.Subtypes) > 0 {
		var subLiterals []string
		for _, s := range parsed.Subtypes {
			subLiterals = append(subLiterals, fmt.Sprintf("%q", s))
		}
		b.WriteString(fmt.Sprintf("\tSubtypes: []string{%s},\n", strings.Join(subLiterals, ", ")))
	}

	// Power/Toughness
	if card.Power != nil {
		b.WriteString(fmt.Sprintf("\tPower: %s,\n", ptLiteral(*card.Power)))
	}
	if card.Toughness != nil {
		b.WriteString(fmt.Sprintf("\tToughness: %s,\n", ptLiteral(*card.Toughness)))
	}

	// Loyalty
	if card.Loyalty != nil {
		if n, err := strconv.Atoi(*card.Loyalty); err == nil {
			b.WriteString(fmt.Sprintf("\tLoyalty: ptrInt(%d),\n", n))
		}
	}

	// Defense
	if card.Defense != nil {
		if n, err := strconv.Atoi(*card.Defense); err == nil {
			b.WriteString(fmt.Sprintf("\tDefense: ptrInt(%d),\n", n))
		}
	}

	// OracleText
	b.WriteString(fmt.Sprintf("\tOracleText: %q,\n", card.OracleText))

	// Abilities placeholder
	b.WriteString("\t// Abilities: filled in by LLM from oracle text.\n")
	b.WriteString("\tAbilities: []game.AbilityDef{},\n")

	b.WriteString("}\n")

	// Emit ptrInt helper if needed by loyalty or defense fields.
	if card.Loyalty != nil || card.Defense != nil {
		b.WriteString("\nfunc ptrInt(n int) *int { return &n }\n")
	}

	return b.String(), nil
}

func ptLiteral(val string) string {
	if val == "*" {
		return "&game.PT{IsStar: true}"
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Sprintf("&game.PT{} /* unparseable: %q */", val)
	}
	return fmt.Sprintf("&game.PT{Value: %d}", n)
}

// CardNameToVarName converts a card name to a Go exported variable name.
// e.g., "Lightning Bolt" -> "LightningBolt", "Sol Ring" -> "SolRing"
func CardNameToVarName(name string) string {
	var b strings.Builder
	capitalize := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if capitalize {
			b.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CardNameToFileName converts a card name to a snake_case file name.
// e.g., "Lightning Bolt" -> "lightning_bolt", "Sol Ring" -> "sol_ring"
func CardNameToFileName(name string) string {
	var b strings.Builder
	prevWasUnderscore := false
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if !prevWasUnderscore && b.Len() > 0 {
				b.WriteRune('_')
				prevWasUnderscore = true
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
		prevWasUnderscore = false
	}
	result := b.String()
	return strings.TrimSuffix(result, "_")
}

// CardNameToPackageLetter returns the lowercase first letter of the card name.
func CardNameToPackageLetter(name string) string {
	for _, r := range name {
		if unicode.IsLetter(r) {
			return string(unicode.ToLower(r))
		}
	}
	return "other"
}
