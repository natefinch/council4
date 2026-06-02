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
// into Go source code that constructs a cost.Mana value.
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

	return "cost.Mana{\n\t\t\t" + strings.Join(symbols, ",\n\t\t\t") + ",\n\t\t}", nil
}

func symbolToLiteral(sym string) (string, error) {
	switch sym {
	case "X", "C", "S", "W", "U", "B", "R", "G":
		return "cost." + sym, nil

	default:
		// continue
	}

	// Phyrexian: W/P, U/P, etc.
	if before, ok := strings.CutSuffix(sym, "/P"); ok {
		color := before
		return fmt.Sprintf("cost.PhyrexianMana(mana.%s)", color), nil
	}
	// Hybrid: W/U, B/R, etc.
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		// Twobrid: 2/W
		if _, err := strconv.Atoi(parts[0]); err == nil {
			color := parts[1]
			return fmt.Sprintf("cost.Twobrid(mana.%s)", color), nil
		}
		// Hybrid: W/U
		return fmt.Sprintf("cost.HybridMana(mana.%s, mana.%s)", parts[0], parts[1]), nil
	}
	// Generic: a number
	if n, err := strconv.Atoi(sym); err == nil {
		return fmt.Sprintf("cost.O(%d)", n), nil
	}
	return "", fmt.Errorf("unsupported mana symbol: %s", sym)
}
