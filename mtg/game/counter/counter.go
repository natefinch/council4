// Package counter provides counter types and tracking for Magic: The Gathering
// permanents and players.
//
// Counters are named markers placed on permanents (or sometimes players) that
// modify their characteristics or track game state. The most common are +1/+1
// and -1/-1 counters on creatures, loyalty counters on planeswalkers, and
// poison counters on players.
package counter

import "maps"

// Kind identifies a specific type of counter.
type Kind int

// Kind values identify supported Magic counter names.
const (
	PlusOnePlusOne   Kind = iota // +1/+1 counter (CR 122.1)
	MinusOneMinusOne             // -1/-1 counter (CR 122.1)
	Loyalty                      // Loyalty counter (planeswalkers, CR 209)
	Charge                       // Charge counter
	Time                         // Time counter (suspend, vanishing)
	Defense                      // Defense counter (battles)
	Poison                       // Poison counter (on players)
	Lore                         // Lore counter (sagas)
	Verse                        // Verse counter
	Shield                       // Shield counter (damage prevention)
	Stun                         // Stun counter (prevents untapping)
	Finality                     // Finality counter
	Brick                        // Brick counter
	Page                         // Page counter
	Enlightened                  // Enlightened counter
	Oil                          // Oil counter
	Blood                        // Blood counter
	Indestructible               // Indestructible counter
	Deathtouch                   // Deathtouch counter
	Flying                       // Flying counter
	FirstStrike                  // First strike counter
	Hexproof                     // Hexproof counter
	Lifelink                     // Lifelink counter
	Menace                       // Menace counter
	Reach                        // Reach counter
	Trample                      // Trample counter
	Vigilance                    // Vigilance counter
	Energy                       // Energy counter (on players)
	Experience                   // Experience counter (on players)
	Burden                       // Burden counter
)

// Valid reports whether k is a recognized counter kind.
func (k Kind) Valid() bool {
	return k >= PlusOnePlusOne && k <= Burden
}

// PlayerOnly reports whether k may be placed only on players.
func (k Kind) PlayerOnly() bool {
	switch k {
	case Poison, Energy, Experience:
		return true
	default:
		return false
	}
}

// String returns the human-readable name of the counter kind.
func (k Kind) String() string {
	switch k {
	case PlusOnePlusOne:
		return "+1/+1"
	case MinusOneMinusOne:
		return "-1/-1"
	case Loyalty:
		return "loyalty"
	case Charge:
		return "charge"
	case Time:
		return "time"
	case Defense:
		return "defense"
	case Poison:
		return "poison"
	case Lore:
		return "lore"
	case Verse:
		return "verse"
	case Shield:
		return "shield"
	case Stun:
		return "stun"
	case Finality:
		return "finality"
	case Brick:
		return "brick"
	case Page:
		return "page"
	case Enlightened:
		return "enlightened"
	case Oil:
		return "oil"
	case Blood:
		return "blood"
	case Indestructible:
		return "indestructible"
	case Deathtouch:
		return "deathtouch"
	case Flying:
		return "flying"
	case FirstStrike:
		return "first strike"
	case Hexproof:
		return "hexproof"
	case Lifelink:
		return "lifelink"
	case Menace:
		return "menace"
	case Reach:
		return "reach"
	case Trample:
		return "trample"
	case Vigilance:
		return "vigilance"
	case Energy:
		return "energy"
	case Experience:
		return "experience"
	case Burden:
		return "burden"
	default:
		return "unknown"
	}
}

// Set tracks the counters on a single permanent or player.
type Set struct {
	counts map[Kind]int
}

// NewSet creates an empty counter set.
func NewSet() Set {
	return Set{counts: make(map[Kind]int)}
}

// Add adds n counters of the given kind. n must be positive.
func (s *Set) Add(k Kind, n int) {
	if n <= 0 {
		return
	}
	if s.counts == nil {
		s.counts = make(map[Kind]int)
	}
	s.counts[k] += n
}

// Remove removes up to n counters of the given kind. Returns the number
// actually removed (may be less than n if fewer exist).
func (s *Set) Remove(k Kind, n int) int {
	if s.counts == nil || n <= 0 {
		return 0
	}
	have := s.counts[k]
	if have <= n {
		delete(s.counts, k)
		return have
	}
	s.counts[k] -= n
	return n
}

// Get returns the number of counters of the given kind.
func (s *Set) Get(k Kind) int {
	if s.counts == nil {
		return 0
	}
	return s.counts[k]
}

// Has reports whether there is at least one counter of the given kind.
func (s *Set) Has(k Kind) bool {
	return s.Get(k) > 0
}

// CancelOpposites applies the state-based action that removes pairs of
// +1/+1 and -1/-1 counters (CR 704.5r). Returns the number of pairs removed.
func (s *Set) CancelOpposites() int {
	if s.counts == nil {
		return 0
	}
	plus := s.counts[PlusOnePlusOne]
	minus := s.counts[MinusOneMinusOne]
	if plus == 0 || minus == 0 {
		return 0
	}
	removed := min(plus, minus)
	s.counts[PlusOnePlusOne] -= removed
	s.counts[MinusOneMinusOne] -= removed
	if s.counts[PlusOnePlusOne] == 0 {
		delete(s.counts, PlusOnePlusOne)
	}
	if s.counts[MinusOneMinusOne] == 0 {
		delete(s.counts, MinusOneMinusOne)
	}
	return removed
}

// All returns a copy of all counter kinds and their counts.
func (s *Set) All() map[Kind]int {
	result := make(map[Kind]int, len(s.counts))
	maps.Copy(result, s.counts)
	return result
}

// Clone returns a deep copy of the set that shares no map state with the
// receiver, so mutating one set does not affect the other.
func (s *Set) Clone() Set {
	clone := Set{}
	if s.counts != nil {
		clone.counts = maps.Clone(s.counts)
	}
	return clone
}

// IsEmpty reports whether there are no counters.
func (s *Set) IsEmpty() bool {
	return len(s.counts) == 0
}
