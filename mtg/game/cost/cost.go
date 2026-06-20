// Package cost defines declarative mana and non-mana costs for spells and abilities.
package cost

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Mana represents the mana cost of a spell or ability as an ordered
// list of mana symbols. The order matches the printed card (e.g.,
// {2}{W}{U} for a spell costing 2 generic, 1 white, 1 blue).
type Mana []Symbol

// ManaValue returns the total mana value (formerly "converted mana cost")
// of the cost. Variable (X) symbols contribute 0 when not on the stack.
func (m Mana) ManaValue() int {
	total := 0
	for _, s := range m {
		switch s.Kind {
		case ColoredSymbol, ColorlessSymbol, PhyrexianSymbol, SnowSymbol, HybridSymbol:
			total++
		case GenericSymbol:
			total += s.Generic
		case TwobridSymbol:
			total += 2
		default:
			// X = 0 except on the stack, and unknown symbols add no mana value.
		}
	}
	return total
}

// Colors returns the set of colors present in this cost.
// Generic and colorless symbols do not contribute colors.
func (m Mana) Colors() []mana.Color {
	seen := make(map[mana.Color]bool)
	var colors []mana.Color
	for _, s := range m {
		switch s.Kind {
		case ColoredSymbol, PhyrexianSymbol, TwobridSymbol:
			if !seen[s.Color] {
				seen[s.Color] = true
				colors = append(colors, s.Color)
			}
		case HybridSymbol:
			if !seen[s.Color] {
				seen[s.Color] = true
				colors = append(colors, s.Color)
			}
			if !seen[s.AltColor] {
				seen[s.AltColor] = true
				colors = append(colors, s.AltColor)
			}
		default:
		}
	}
	return colors
}

// String returns the cost in conventional MTG notation (e.g., "{2}{W}{U}").
func (m Mana) String() string {
	var b strings.Builder
	for _, s := range m {
		_, _ = b.WriteString(s.String())
	}
	return b.String()
}

// Multiply returns the exact mana requirements repeated count times. Generic
// requirements are combined into one symbol; non-generic symbols retain their
// relative order. A nonpositive count is the explicit zero cost.
func (m Mana) Multiply(count int) Mana {
	if count <= 0 {
		return Mana{O(0)}
	}
	generic := 0
	nonGeneric := make(Mana, 0, len(m)*count)
	for range count {
		for _, symbol := range m {
			if symbol.Kind == GenericSymbol {
				generic += symbol.Generic
				continue
			}
			nonGeneric = append(nonGeneric, symbol)
		}
	}
	if generic == 0 {
		if len(nonGeneric) == 0 {
			return Mana{O(0)}
		}
		return nonGeneric
	}
	result := make(Mana, 0, len(nonGeneric)+1)
	result = append(result, O(generic))
	return append(result, nonGeneric...)
}

// ManaForColor returns the mana color corresponding to the given color.
// We convert this way since color is a subset of mana colors (colorless is a valid mana color).
func ManaForColor(c color.Color) mana.Color {
	switch c {
	case color.White:
		return mana.W
	case color.Blue:
		return mana.U
	case color.Black:
		return mana.B
	case color.Red:
		return mana.R
	case color.Green:
		return mana.G
	default:
		panic("invalid color:" + string(c))
	}
}
