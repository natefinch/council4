// Package color defines Magic card colors.
package color

// Color represents one of the five colors of Magic.
type Color string

// Color values identify Magic's colors.
const (
	White Color = "White"
	Blue  Color = "Blue"
	Black Color = "Black"
	Red   Color = "Red"
	Green Color = "Green"
)

// Abbreviation returns the conventional single-letter abbreviation for the color.
func (c Color) Abbreviation() string {
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
	default:
		return "?"
	}
}

// AllColors returns all five colors of Magic (not including colorless).
func AllColors() []Color {
	return []Color{White, Blue, Black, Red, Green}
}
