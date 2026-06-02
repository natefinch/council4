package cost

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/mana"
)

// SymbolKind classifies a mana symbol by how it can be paid.
type SymbolKind int

const (
	// ColoredSymbol is a single colored mana symbol (e.g., {W}, {U}).
	ColoredSymbol SymbolKind = iota
	// GenericSymbol is generic mana payable with any color (e.g., {3}).
	GenericSymbol
	// ColorlessSymbol requires specifically colorless mana ({C}).
	ColorlessSymbol
	// VariableSymbol is X mana — variable, chosen on cast.
	VariableSymbol
	// HybridSymbol can be paid with either of two colors (e.g., {W/U}).
	HybridSymbol
	// TwobridSymbol can be paid with a color or 2 generic (e.g., {2/W}).
	TwobridSymbol
	// PhyrexianSymbol can be paid with a color or 2 life (e.g., {W/P}).
	PhyrexianSymbol
	// SnowSymbol requires mana from a snow source ({S}).
	SnowSymbol
)

// Symbol represents a single mana symbol in a mana cost.
type Symbol struct {
	Kind SymbolKind

	// Color is the primary color for ColoredSymbol, PhyrexianSymbol,
	// MonoHybridSymbol, and HybridSymbol.
	Color mana.Color

	// AltColor is the second color for HybridSymbol (e.g., {W/U} has
	// Color=White, AltColor=Blue).
	AltColor mana.Color

	// Generic is the numeric value for GenericSymbol (e.g., 3 for {3}).
	Generic int
}

// IsColored returns true if the symbol is a colored mana symbol.
func (s Symbol) IsColored() bool {
	return s.Kind == ColoredSymbol || s.Kind == HybridSymbol || s.Kind == PhyrexianSymbol
}

// Colors returns the set of colors present in this symbol.
func (s Symbol) Colors() []mana.Color {
	switch s.Kind {
	case ColoredSymbol, PhyrexianSymbol, TwobridSymbol:
		return []mana.Color{s.Color}
	case HybridSymbol:
		return []mana.Color{s.Color, s.AltColor}
	default:
		return nil
	}
}

// Mana Symbols for costs.
var (
	W = Symbol{Kind: ColoredSymbol, Color: mana.W}
	U = Symbol{Kind: ColoredSymbol, Color: mana.U}
	B = Symbol{Kind: ColoredSymbol, Color: mana.B}
	R = Symbol{Kind: ColoredSymbol, Color: mana.R}
	G = Symbol{Kind: ColoredSymbol, Color: mana.G}

	C = Symbol{Kind: ColorlessSymbol, Color: mana.C} // Colorless mana cost. Should be ◇ but that's not a valid Go identifier.
	X = Symbol{Kind: VariableSymbol}                 // Generic X cost.
	S = Symbol{Kind: SnowSymbol}                     // Generic snow mana cost.
)

// O creates a generic mana cost with the given value.
func O(n int) Symbol {
	return Symbol{Kind: GenericSymbol, Generic: n}
}

// HybridMana creates a hybrid mana symbol (e.g., {W/U}).
func HybridMana(a, b mana.Color) Symbol {
	return Symbol{Kind: HybridSymbol, Color: a, AltColor: b}
}

// Twobrid creates a mono-hybrid mana symbol (e.g., {2/W}).
func Twobrid(c mana.Color) Symbol {
	return Symbol{Kind: TwobridSymbol, Color: c}
}

// PhyrexianMana creates a Phyrexian mana symbol (e.g., {W/P}).
func PhyrexianMana(c mana.Color) Symbol {
	return Symbol{Kind: PhyrexianSymbol, Color: c}
}

// SnowMana creates a snow mana symbol ({S}).
func SnowMana() Symbol {
	return Symbol{Kind: SnowSymbol}
}

// String returns a human-readable representation of the symbol.
func (s Symbol) String() string {
	switch s.Kind {
	case ColoredSymbol:
		return fmt.Sprintf("{%s}", s.Color)
	case GenericSymbol:
		return fmt.Sprintf("{%d}", s.Generic)
	case ColorlessSymbol:
		return "{C}"
	case VariableSymbol:
		return "{X}"
	case HybridSymbol:
		return fmt.Sprintf("{%s/%s}", s.Color, s.AltColor)
	case TwobridSymbol:
		return fmt.Sprintf("{2/%s}", s.Color)
	case PhyrexianSymbol:
		return fmt.Sprintf("{%s/P}", s.Color)
	case SnowSymbol:
		return "{S}"
	default:
		return "{?}"
	}
}
