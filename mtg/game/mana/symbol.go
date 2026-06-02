package mana

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/color"
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
	// MonoHybridSymbol can be paid with a color or 2 generic (e.g., {2/W}).
	MonoHybridSymbol
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
	Color color.Color

	// AltColor is the second color for HybridSymbol (e.g., {W/U} has
	// Color=White, AltColor=Blue).
	AltColor color.Color

	// Generic is the numeric value for GenericSymbol (e.g., 3 for {3}).
	Generic int
}

// Mana Symbols for costs.
var (
	R = Symbol{Kind: ColoredSymbol, Color: color.Red}
	U = Symbol{Kind: ColoredSymbol, Color: color.Blue}
	B = Symbol{Kind: ColoredSymbol, Color: color.Black}
	W = Symbol{Kind: ColoredSymbol, Color: color.White}
	G = Symbol{Kind: ColoredSymbol, Color: color.Green}
)

// ColoredMana creates a colored mana symbol.
func ColoredMana(c color.Color) Symbol {
	return Symbol{Kind: ColoredSymbol, Color: c}
}

// ColorlessMana creates a colorless mana symbol ({C}).
func ColorlessMana() Symbol {
	return Symbol{Kind: ColorlessSymbol, Color: color.Colorless}
}

// GenericMana creates a generic mana symbol with the given value.
func GenericMana(n int) Symbol {
	return Symbol{Kind: GenericSymbol, Generic: n}
}

// VariableMana creates an {X} mana symbol.
func VariableMana() Symbol {
	return Symbol{Kind: VariableSymbol}
}

// HybridMana creates a hybrid mana symbol (e.g., {W/U}).
func HybridMana(a, b color.Color) Symbol {
	return Symbol{Kind: HybridSymbol, Color: a, AltColor: b}
}

// MonoHybridMana creates a mono-hybrid mana symbol (e.g., {2/W}).
func MonoHybridMana(c color.Color) Symbol {
	return Symbol{Kind: MonoHybridSymbol, Color: c}
}

// PhyrexianMana creates a Phyrexian mana symbol (e.g., {W/P}).
func PhyrexianMana(c color.Color) Symbol {
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
	case MonoHybridSymbol:
		return fmt.Sprintf("{2/%s}", s.Color)
	case PhyrexianSymbol:
		return fmt.Sprintf("{%s/P}", s.Color)
	case SnowSymbol:
		return "{S}"
	default:
		return "{?}"
	}
}
