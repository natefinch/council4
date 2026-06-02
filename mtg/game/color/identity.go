package color

// Identity represents a set of colors, used in Commander format
// to define which colors a deck may contain based on the commander's
// color identity (CR 903.4).
type Identity struct {
	colors map[Color]bool
}

// NewIdentity creates an Identity from the given colors.
func NewIdentity(colors ...Color) Identity {
	ci := Identity{colors: make(map[Color]bool)}
	for _, c := range colors {
		ci.colors[c] = true
	}
	return ci
}

// Contains reports whether the identity includes the given color.
func (ci Identity) Contains(c Color) bool {
	return ci.colors[c]
}

// ContainsAll reports whether this identity is a superset of the other.
// Used to check if a card's color identity fits within a commander's.
func (ci Identity) ContainsAll(other Identity) bool {
	for c := range other.colors {
		if !ci.colors[c] {
			return false
		}
	}
	return true
}

// Colors returns the colors in this identity as a slice.
func (ci Identity) Colors() []Color {
	var result []Color
	for _, c := range AllColors() {
		if ci.colors[c] {
			result = append(result, c)
		}
	}
	return result
}

// NumColors returns the number of colors in this identity.
func (ci Identity) NumColors() int {
	return len(ci.colors)
}
