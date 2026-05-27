package cardgen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// manaSymbolRe matches a single mana symbol like {W}, {2}, {W/U}, {2/W}, {W/P}, {X}, {C}, {S}.
var manaSymbolRe = regexp.MustCompile(`\{([^}]+)\}`)

// ParseManaCostLiteral converts a Scryfall mana cost string (e.g., "{2}{W}{U}")
// into Go source code that constructs a mana.Cost value.
// Returns empty string and nil error if the input is empty (e.g., lands).
// Returns an error if an unsupported mana symbol is encountered.
func ParseManaCostLiteral(cost string) (string, error) {
	if cost == "" {
		return "", nil
	}

	matches := manaSymbolRe.FindAllStringSubmatch(cost, -1)
	if len(matches) == 0 {
		return "", nil
	}

	var symbols []string
	for _, m := range matches {
		sym := m[1]
		literal, err := symbolToLiteral(sym)
		if err != nil {
			return "", fmt.Errorf("unsupported mana symbol {%s} in cost %q: %w", sym, cost, err)
		}
		symbols = append(symbols, literal)
	}

	return "mana.Cost{\n\t\t\t" + strings.Join(symbols, ",\n\t\t\t") + ",\n\t\t}", nil
}

func symbolToLiteral(sym string) (string, error) {
	// Variable: X
	if sym == "X" {
		return "mana.VariableMana()", nil
	}
	// Colorless: C
	if sym == "C" {
		return "mana.ColorlessMana()", nil
	}
	// Snow: S
	if sym == "S" {
		return "mana.SnowMana()", nil
	}
	// Phyrexian: W/P, U/P, etc.
	if strings.HasSuffix(sym, "/P") {
		color := strings.TrimSuffix(sym, "/P")
		goColor, err := colorLetter(color)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("mana.PhyrexianMana(mana.%s)", goColor), nil
	}
	// Hybrid: W/U, B/R, etc.
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		// Mono-hybrid: 2/W
		if _, err := strconv.Atoi(parts[0]); err == nil {
			goColor, err := colorLetter(parts[1])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("mana.MonoHybridMana(mana.%s)", goColor), nil
		}
		goA, err := colorLetter(parts[0])
		if err != nil {
			return "", err
		}
		goB, err := colorLetter(parts[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("mana.HybridMana(mana.%s, mana.%s)", goA, goB), nil
	}
	// Generic: a number
	if n, err := strconv.Atoi(sym); err == nil {
		return fmt.Sprintf("mana.GenericMana(%d)", n), nil
	}
	// Colored: W, U, B, R, G
	goColor, err := colorLetter(sym)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("mana.ColoredMana(mana.%s)", goColor), nil
}

func colorLetter(s string) (string, error) {
	switch s {
	case "W":
		return "White", nil
	case "U":
		return "Blue", nil
	case "B":
		return "Black", nil
	case "R":
		return "Red", nil
	case "G":
		return "Green", nil
	default:
		return "", fmt.Errorf("unknown color: %s", s)
	}
}
