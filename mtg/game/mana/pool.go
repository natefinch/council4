package mana

import (
	"maps"
)

// Pool represents a player's current mana pool. It tracks mana by spendable
// units so rules can distinguish provenance such as snow mana while preserving
// simple color-count APIs.
type Pool struct {
	mana map[Unit]int
	// persist records the subset of mana that does not empty as steps and phases
	// end (the CR 500.4 exception used by "Until end of turn, you don't lose this
	// mana as steps and phases end", Grand Warlord Radha). Empty restores the
	// pool to this reserved mana instead of clearing it; spending reconciles it so
	// spent persistent mana does not reappear on the next Empty; ClearPersistent
	// drops the reservation at end-of-turn cleanup so the following Empty removes
	// it (CR 514.2). It is nil for pools that never received persistent mana, in
	// which case Empty clears the pool exactly as before.
	persist map[Unit]int
	// spent is the running total of mana removed from this pool to pay costs.
	// It is never reset by Empty (emptying discards unspent mana, which is not
	// spent), so a caller can measure mana spent over a span by differencing it.
	spent int
}

// NewPool creates an empty mana pool.
func NewPool() Pool {
	return Pool{mana: make(map[Unit]int)}
}

// Clone returns a deep copy of the pool that shares no map state with the
// receiver, so mutating one pool does not affect the other.
func (p *Pool) Clone() Pool {
	clone := Pool{spent: p.spent}
	if p.mana != nil {
		clone.mana = maps.Clone(p.mana)
	}
	if p.persist != nil {
		clone.persist = maps.Clone(p.persist)
	}
	return clone
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

// AddPersistent adds mana of the given color that does not empty as steps and
// phases end (CR 500.4 exception). See Pool.persist.
func (p *Pool) AddPersistent(c Color, amount int) {
	p.AddPersistentUnit(Unit{Color: c}, amount)
}

// AddPersistentSnow adds snow mana of the given color that does not empty as
// steps and phases end (CR 500.4 exception). See Pool.persist.
func (p *Pool) AddPersistentSnow(c Color, amount int) {
	p.AddPersistentUnit(Unit{Color: c, Snow: true}, amount)
}

// AddPersistentUnit adds mana units that do not empty as steps and phases end.
// The units are spendable like any other mana, but Empty preserves them until
// ClearPersistent is called at end-of-turn cleanup. See Pool.persist.
func (p *Pool) AddPersistentUnit(unit Unit, amount int) {
	if amount <= 0 {
		return
	}
	p.AddUnit(unit, amount)
	if p.persist == nil {
		p.persist = make(map[Unit]int)
	}
	p.persist[unit] += amount
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
			delete(p.persist, unit)
		} else if p.persist != nil && p.persist[unit] > p.mana[unit] {
			p.persist[unit] = p.mana[unit]
		}
	}
	p.spent += amount
	return remaining == 0
}

// Spent returns the running total of mana removed from this pool to pay costs
// over its lifetime. Difference two readings to measure mana spent across a
// span such as a turn.
func (p *Pool) Spent() int {
	return p.spent
}

// Total returns the total amount of mana in the pool across all colors.
func (p *Pool) Total() int {
	total := 0
	for _, v := range p.mana {
		total += v
	}
	return total
}

// Empty removes all mana from the pool, except mana reserved by AddPersistent
// that has not yet been released by ClearPersistent (CR 500.4 exception). For a
// pool that never received persistent mana this clears the pool entirely, as
// before.
func (p *Pool) Empty() {
	if len(p.persist) == 0 {
		p.mana = make(map[Unit]int)
		return
	}
	p.mana = maps.Clone(p.persist)
}

// ClearPersistent releases any mana reservation created by AddPersistent so the
// next Empty removes it. It models the "until end of turn" duration expiring at
// end-of-turn cleanup (CR 514.2), after which the reserved mana empties like any
// other mana as the following step or phase ends.
func (p *Pool) ClearPersistent() {
	p.persist = nil
}

// IsEmpty reports whether the pool has no mana.
func (p *Pool) IsEmpty() bool {
	return p.Total() == 0
}

func spendOrder() []Unit {
	return []Unit{
		{Color: W},
		{Color: U},
		{Color: B},
		{Color: R},
		{Color: G},
		{Color: C},
		{Color: W, Snow: true},
		{Color: U, Snow: true},
		{Color: B, Snow: true},
		{Color: R, Snow: true},
		{Color: G, Snow: true},
		{Color: C, Snow: true},
	}
}
