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
	Age                          // Age counter (cumulative upkeep)
	Quest                        // Quest counter (Ascension cycle)
	Level                        // Level counter (leveler cards, CR 711)

	// Asymmetric power/toughness counters (CR 122.1). Unlike the symmetric
	// +1/+1 and -1/-1 counters these modify power and toughness by different
	// amounts. They are ordered after Level so the discrete kind values that
	// precede them, and the generated source identifiers, are unchanged.

	PlusOnePlusZero   // +1/+0 counter
	PlusTwoPlusTwo    // +2/+2 counter
	MinusZeroMinusOne // -0/-1 counter
	PlusZeroPlusOne   // +0/+1 counter
	MinusZeroMinusTwo // -0/-2 counter
	MinusTwoMinusTwo  // -2/-2 counter
	PlusOnePlusTwo    // +1/+2 counter
	PlusZeroPlusTwo   // +0/+2 counter
	MinusTwoMinusOne  // -2/-1 counter
	MinusOneMinusZero // -1/-0 counter

	// Named marker counters with no inherent rules meaning beyond being a
	// counter of a particular name (CR 122). They are placed and removed like
	// any other counter and back the many cards whose costs or abilities track
	// a uniquely named counter. Ordered last so the discrete kind values that
	// precede them, and the generated source identifiers, are unchanged.

	Spore      // Spore counter (Thallid/Saproling cycle)
	Fade       // Fade counter (Fading)
	Divinity   // Divinity counter (Myojin cycle)
	Healing    // Healing counter (Fylgja)
	Wish       // Wish counter (Wishclaw Talisman, Ring of Three Wishes)
	Study      // Study counter (Pursuit of Knowledge)
	Dream      // Dream counter (Rasputin Dreamweaver)
	Supply     // Supply counter (Stocking the Pantry)
	Story      // Story counter (Staff of the Storyteller)
	Film       // Film counter (Peter Parker's Camera)
	Hoofprint  // Hoofprint counter (Hoofprints of the Stag)
	Suspect    // Suspect counter (Investigator's Journal)
	Javelin    // Javelin counter (Icatian Javelineers)
	Cube       // Cube counter (Delif's Cube)
	Polyp      // Polyp counter (Coral Reef)
	Component  // Component counter (Component Pouch)
	Eon        // Eon counter (Magosi, the Waterveil)
	Incubation // Incubation counter (Drake Hatcher)
	Devotion   // Devotion counter (Pious Kitsune)
	Foreshadow // Foreshadow counter (Ominous Seas)
	Arrowhead  // Arrowhead counter (Serrated Arrows)
	Carrion    // Carrion counter (Osai Vultures)
	Corpse     // Corpse counter (Scavenging Ghoul)
	Loot       // Loot counter (Bandit's Haul)
	Net        // Net counter (Braided Net)
	Gold       // Gold counter (Dragon's Hoard)
	Currency   // Currency counter (Trade Caravan)
	Book       // Book counter (Spell Satchel)
	Blaze      // Blaze counter (Five-Alarm Fire)
	Palliation // Palliation counter (Palliation Accord)
	Gem        // Gem counter (Briber's Purse)
	Pressure   // Pressure counter (Hellion Crucible)
	Flame      // Flame counter (Flame Channeler)
	Ice        // Ice counter (Iceberg)
	Coin       // Coin counter (Noble's Purse)
	Depletion  // Depletion counter (Hickory Woodlot)
	Croak      // Croak counter (Grolnok, the Omnivore)
	Void       // Void counter (Dauthi Voidwalker, Sphere of Annihilation)
	Intel      // Intel counter (Flamewar, Streetwise Operative)
	Collection // Collection counter (Evelyn, the Covetous, Charitable Levy)
)

// Valid reports whether k is a recognized counter kind.
func (k Kind) Valid() bool {
	return k >= PlusOnePlusOne && k <= Collection
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
	case Age:
		return "age"
	case Quest:
		return "quest"
	case Level:
		return "level"
	case PlusOnePlusZero:
		return "+1/+0"
	case PlusTwoPlusTwo:
		return "+2/+2"
	case MinusZeroMinusOne:
		return "-0/-1"
	case PlusZeroPlusOne:
		return "+0/+1"
	case MinusZeroMinusTwo:
		return "-0/-2"
	case MinusTwoMinusTwo:
		return "-2/-2"
	case PlusOnePlusTwo:
		return "+1/+2"
	case PlusZeroPlusTwo:
		return "+0/+2"
	case MinusTwoMinusOne:
		return "-2/-1"
	case MinusOneMinusZero:
		return "-1/-0"
	case Spore:
		return "spore"
	case Fade:
		return "fade"
	case Divinity:
		return "divinity"
	case Healing:
		return "healing"
	case Wish:
		return "wish"
	case Study:
		return "study"
	case Dream:
		return "dream"
	case Supply:
		return "supply"
	case Story:
		return "story"
	case Film:
		return "film"
	case Hoofprint:
		return "hoofprint"
	case Suspect:
		return "suspect"
	case Javelin:
		return "javelin"
	case Cube:
		return "cube"
	case Polyp:
		return "polyp"
	case Component:
		return "component"
	case Eon:
		return "eon"
	case Incubation:
		return "incubation"
	case Devotion:
		return "devotion"
	case Foreshadow:
		return "foreshadow"
	case Arrowhead:
		return "arrowhead"
	case Carrion:
		return "carrion"
	case Corpse:
		return "corpse"
	case Loot:
		return "loot"
	case Net:
		return "net"
	case Gold:
		return "gold"
	case Currency:
		return "currency"
	case Book:
		return "book"
	case Blaze:
		return "blaze"
	case Palliation:
		return "palliation"
	case Gem:
		return "gem"
	case Pressure:
		return "pressure"
	case Flame:
		return "flame"
	case Ice:
		return "ice"
	case Coin:
		return "coin"
	case Depletion:
		return "depletion"
	case Croak:
		return "croak"
	case Void:
		return "void"
	case Intel:
		return "intel"
	case Collection:
		return "collection"
	default:
		return "unknown"
	}
}

// powerToughness returns the power and toughness a power/toughness counter of
// this kind modifies (CR 122.1, 613.4c), and whether kind is a power/toughness
// counter at all. The symmetric +1/+1 and -1/-1 counters report equal power and
// toughness; the asymmetric kinds report their printed dimensions.
func (k Kind) powerToughness() (power, toughness int, ok bool) {
	switch k {
	case PlusOnePlusOne:
		return 1, 1, true
	case MinusOneMinusOne:
		return -1, -1, true
	case PlusOnePlusZero:
		return 1, 0, true
	case PlusTwoPlusTwo:
		return 2, 2, true
	case MinusZeroMinusOne:
		return 0, -1, true
	case PlusZeroPlusOne:
		return 0, 1, true
	case MinusZeroMinusTwo:
		return 0, -2, true
	case MinusTwoMinusTwo:
		return -2, -2, true
	case PlusOnePlusTwo:
		return 1, 2, true
	case PlusZeroPlusTwo:
		return 0, 2, true
	case MinusTwoMinusOne:
		return -2, -1, true
	case MinusOneMinusZero:
		return -1, 0, true
	default:
		return 0, 0, false
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

// RemoveAll removes every counter of every kind and returns the total number
// removed, modeling the kind-agnostic "remove all counters from <permanent>"
// wording (Vampire Hexmage).
func (s *Set) RemoveAll() int {
	total := 0
	for _, n := range s.counts {
		total += n
	}
	s.counts = nil
	return total
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

// PowerToughnessDelta returns the total power and toughness modification from
// every power/toughness counter on the set (CR 613.4c). The symmetric +1/+1 and
// -1/-1 counters contribute equal power and toughness; the asymmetric kinds
// contribute their printed dimensions. Non-power/toughness counters are ignored.
func (s *Set) PowerToughnessDelta() (power, toughness int) {
	for kind, count := range s.counts {
		if p, t, ok := kind.powerToughness(); ok {
			power += p * count
			toughness += t * count
		}
	}
	return power, toughness
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
