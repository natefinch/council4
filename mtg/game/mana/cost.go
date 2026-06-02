package mana

import "github.com/natefinch/council4/mtg/game/color"

import "strings"

// Cost represents the mana cost of a spell or ability as an ordered
// list of mana symbols. The order matches the printed card (e.g.,
// {2}{W}{U} for a spell costing 2 generic, 1 white, 1 blue).
type Cost []Symbol

// ManaValue returns the total mana value (formerly "converted mana cost")
// of the cost. Variable (X) symbols contribute 0 when not on the stack.
func (c Cost) ManaValue() int {
	total := 0
	for _, s := range c {
		switch s.Kind {
		case ColoredSymbol, ColorlessSymbol, PhyrexianSymbol, SnowSymbol, HybridSymbol:
			total++
		case GenericSymbol:
			total += s.Generic
		case MonoHybridSymbol:
			total += 2
		default:
			// X = 0 except on the stack, and unknown symbols add no mana value.
		}
	}
	return total
}

// Colors returns the set of colors present in this cost.
// Generic and colorless symbols do not contribute colors.
func (c Cost) Colors() []color.Color {
	seen := make(map[color.Color]bool)
	var colors []color.Color
	for _, s := range c {
		switch s.Kind {
		case ColoredSymbol, PhyrexianSymbol, MonoHybridSymbol:
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
func (c Cost) String() string {
	var b strings.Builder
	for _, s := range c {
		_, _ = b.WriteString(s.String())
	}
	return b.String()
}
