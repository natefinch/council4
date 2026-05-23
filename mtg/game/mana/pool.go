package mana

// Pool represents a player's current mana pool, tracking available
// mana by color. Mana pools empty at the end of each step and phase.
type Pool struct {
	mana map[Color]int
}

// NewPool creates an empty mana pool.
func NewPool() Pool {
	return Pool{mana: make(map[Color]int)}
}

// Add adds mana of the given color to the pool.
func (p *Pool) Add(c Color, amount int) {
	if p.mana == nil {
		p.mana = make(map[Color]int)
	}
	p.mana[c] += amount
}

// Amount returns the amount of mana of the given color in the pool.
func (p *Pool) Amount(c Color) int {
	if p.mana == nil {
		return 0
	}
	return p.mana[c]
}

// Spend removes mana of the given color from the pool. It returns false
// if there is insufficient mana of that color.
func (p *Pool) Spend(c Color, amount int) bool {
	if p.mana == nil || p.mana[c] < amount {
		return false
	}
	p.mana[c] -= amount
	if p.mana[c] == 0 {
		delete(p.mana, c)
	}
	return true
}

// Total returns the total amount of mana in the pool across all colors.
func (p *Pool) Total() int {
	total := 0
	for _, v := range p.mana {
		total += v
	}
	return total
}

// Empty removes all mana from the pool.
func (p *Pool) Empty() {
	p.mana = make(map[Color]int)
}

// IsEmpty reports whether the pool has no mana.
func (p *Pool) IsEmpty() bool {
	return p.Total() == 0
}

// ColorIdentity represents a set of colors, used in Commander format
// to define which colors a deck may contain based on the commander's
// color identity (CR 903.4).
type ColorIdentity struct {
	colors map[Color]bool
}

// NewColorIdentity creates a ColorIdentity from the given colors.
func NewColorIdentity(colors ...Color) ColorIdentity {
	ci := ColorIdentity{colors: make(map[Color]bool)}
	for _, c := range colors {
		ci.colors[c] = true
	}
	return ci
}

// Contains reports whether the identity includes the given color.
func (ci ColorIdentity) Contains(c Color) bool {
	return ci.colors[c]
}

// ContainsAll reports whether this identity is a superset of the other.
// Used to check if a card's color identity fits within a commander's.
func (ci ColorIdentity) ContainsAll(other ColorIdentity) bool {
	for c := range other.colors {
		if !ci.colors[c] {
			return false
		}
	}
	return true
}

// Colors returns the colors in this identity as a slice.
func (ci ColorIdentity) Colors() []Color {
	var result []Color
	for _, c := range AllColors() {
		if ci.colors[c] {
			result = append(result, c)
		}
	}
	return result
}

// NumColors returns the number of colors in this identity.
func (ci ColorIdentity) NumColors() int {
	return len(ci.colors)
}
