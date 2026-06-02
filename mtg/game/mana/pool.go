package mana

import "maps"

// Pool represents a player's current mana pool. It tracks mana by spendable
// units so rules can distinguish provenance such as snow mana while preserving
// simple color-count APIs.
type Pool struct {
	mana map[Unit]int
}

// NewPool creates an empty mana pool.
func NewPool() Pool {
	return Pool{mana: make(map[Unit]int)}
}

// Add adds mana of the given color to the pool.
func (p *Pool) Add(c Color, amount int) {
	p.AddUnit(Unit{Color: c}, amount)
}

// AddSnow adds snow mana of the given color to the pool.
func (p *Pool) AddSnow(c Color, amount int) {
	p.AddUnit(Unit{Color: c, Snow: true}, amount)
}

// AddUnit adds mana units to the pool.
func (p *Pool) AddUnit(unit Unit, amount int) {
	if amount <= 0 {
		return
	}
	if p.mana == nil {
		p.mana = make(map[Unit]int)
	}
	p.mana[unit] += amount
}

// Amount returns the amount of mana of the given color in the pool.
func (p *Pool) Amount(c Color) int {
	if p.mana == nil {
		return 0
	}
	total := 0
	for unit, amount := range p.mana {
		if unit.Color == c {
			total += amount
		}
	}
	return total
}

// SnowAmount returns the amount of snow mana in the pool, regardless of color.
func (p *Pool) SnowAmount() int {
	if p.mana == nil {
		return 0
	}
	total := 0
	for unit, amount := range p.mana {
		if unit.Snow {
			total += amount
		}
	}
	return total
}

// Units returns a copy of the pool's mana unit counts.
func (p *Pool) Units() map[Unit]int {
	units := make(map[Unit]int)
	maps.Copy(units, p.mana)
	return units
}

// Spend removes mana of the given color from the pool. It returns false
// if there is insufficient mana of that color.
func (p *Pool) Spend(c Color, amount int) bool {
	return p.SpendMatching(amount, func(unit Unit) bool {
		return unit.Color == c
	})
}

// SpendSnow removes snow mana from the pool. It returns false if there is
// insufficient snow mana of any color.
func (p *Pool) SpendSnow(amount int) bool {
	return p.SpendMatching(amount, func(unit Unit) bool {
		return unit.Snow
	})
}

// CanSpendMatching reports whether the pool contains at least amount mana
// units that satisfy matches.
func (p *Pool) CanSpendMatching(amount int, matches func(Unit) bool) bool {
	if amount <= 0 {
		return true
	}
	if p.mana == nil || matches == nil {
		return false
	}
	total := 0
	for _, unit := range spendOrder() {
		if !matches(unit) {
			continue
		}
		total += p.mana[unit]
		if total >= amount {
			return true
		}
	}
	return false
}

// SpendMatching removes amount mana units that satisfy matches. It prefers
// non-snow mana before snow mana so simple colored payments preserve snow
// provenance when possible.
func (p *Pool) SpendMatching(amount int, matches func(Unit) bool) bool {
	if !p.CanSpendMatching(amount, matches) {
		return false
	}
	remaining := amount
	for _, unit := range spendOrder() {
		if remaining == 0 {
			break
		}
		if !matches(unit) {
			continue
		}
		spent := min(p.mana[unit], remaining)
		p.mana[unit] -= spent
		remaining -= spent
		if p.mana[unit] == 0 {
			delete(p.mana, unit)
		}
	}
	return remaining == 0
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
	p.mana = make(map[Unit]int)
}

// IsEmpty reports whether the pool has no mana.
func (p *Pool) IsEmpty() bool {
	return p.Total() == 0
}

func spendOrder() []Unit {
	var units []Unit
	for _, color := range AllColors() {
		units = append(units, Unit{Color: color})
	}
	units = append(units, Unit{Color: Colorless})
	for _, color := range AllColors() {
		units = append(units, Unit{Color: color, Snow: true})
	}
	units = append(units, Unit{Color: Colorless, Snow: true})
	return units
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
