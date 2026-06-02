package mana

// Color represents one of the five colors of Magic mana, plus colorless.
type Color int

// Color values identify Magic's mana colors and true colorless mana.
const (
	White     Color = iota // {W}
	Blue                   // {U}
	Black                  // {B}
	Red                    // {R}
	Green                  // {G}
	Colorless              // {C} — true colorless mana (e.g., from Wastes)
)

// String returns the conventional single-letter abbreviation for the color.
func (c Color) String() string {
	switch c {
	case White:
		return "W"
	case Blue:
		return "U"
	case Black:
		return "B"
	case Red:
		return "R"
	case Green:
		return "G"
	case Colorless:
		return "C"
	default:
		return "?"
	}
}

// AllColors returns all five colors of Magic (not including colorless).
func AllColors() []Color {
	return []Color{White, Blue, Black, Red, Green}
}
