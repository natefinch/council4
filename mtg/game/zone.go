package game

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game/id"
)

// ZoneType identifies which game zone a card is in (CR 400).
type ZoneType int

const (
	// ZoneNone means no zone applies, such as a newly created token before it
	// enters the battlefield.
	ZoneNone ZoneType = iota

	// ZoneLibrary is a player's draw deck — hidden and ordered (CR 401).
	ZoneLibrary

	// ZoneHand is a player's hand — hidden from opponents (CR 402).
	ZoneHand

	// ZoneBattlefield is the shared play area — public (CR 403).
	// Note: battlefield is not per-player in the Zone struct; it is
	// represented as []*Permanent in the Game struct.
	ZoneBattlefield

	// ZoneGraveyard is a player's discard pile — public, ordered (CR 404).
	ZoneGraveyard

	// ZoneStack is the shared stack — public, ordered LIFO (CR 405).
	// Note: the stack is represented by the Stack struct, not a Zone.
	ZoneStack

	// ZoneExile is the exile zone — usually public, sometimes face-down (CR 406).
	ZoneExile

	// ZoneCommand is the command zone — public (CR 408).
	// Commanders start here in Commander format.
	ZoneCommand
)

// String returns the zone name.
func (z ZoneType) String() string {
	switch z {
	case ZoneNone:
		return "None"
	case ZoneLibrary:
		return "Library"
	case ZoneHand:
		return "Hand"
	case ZoneBattlefield:
		return "Battlefield"
	case ZoneGraveyard:
		return "Graveyard"
	case ZoneStack:
		return "Stack"
	case ZoneExile:
		return "Exile"
	case ZoneCommand:
		return "Command"
	default:
		return "Unknown"
	}
}

// IsPublic reports whether cards in this zone are visible to all players.
func (z ZoneType) IsPublic() bool {
	switch z {
	case ZoneBattlefield, ZoneGraveyard, ZoneStack, ZoneExile, ZoneCommand:
		return true
	default:
		return false
	}
}

// Zone is an ordered collection of card instance IDs representing a
// game zone. The library and graveyard are ordered (top = index 0);
// the hand is conceptually unordered but stored as a slice for simplicity.
type Zone struct {
	// Type identifies which zone this is.
	Type ZoneType

	// cards stores card instance IDs in order. For library and graveyard,
	// index 0 is the "top."
	cards []id.ID

	// faceDown tracks which cards in this zone are face-down (relevant
	// for exile and morph/disguise).
	faceDown map[id.ID]bool
}

// NewZone creates an empty zone of the given type.
func NewZone(t ZoneType) Zone {
	return Zone{
		Type:     t,
		faceDown: make(map[id.ID]bool),
	}
}

// Add adds a card to the top (front) of the zone.
func (z *Zone) Add(cardID id.ID) {
	z.cards = append([]id.ID{cardID}, z.cards...)
}

// AddToBottom adds a card to the bottom (end) of the zone.
func (z *Zone) AddToBottom(cardID id.ID) {
	z.cards = append(z.cards, cardID)
}

// Remove removes a card from the zone. Returns true if found and removed.
func (z *Zone) Remove(cardID id.ID) bool {
	for i, c := range z.cards {
		if c == cardID {
			z.cards = append(z.cards[:i], z.cards[i+1:]...)
			delete(z.faceDown, cardID)
			return true
		}
	}
	return false
}

// Top returns the top card of the zone (index 0) and true, or zero
// and false if the zone is empty.
func (z *Zone) Top() (id.ID, bool) {
	if len(z.cards) == 0 {
		return 0, false
	}
	return z.cards[0], true
}

// Bottom returns the bottom card of the zone and true, or zero and
// false if the zone is empty.
func (z *Zone) Bottom() (id.ID, bool) {
	if len(z.cards) == 0 {
		return 0, false
	}
	return z.cards[len(z.cards)-1], true
}

// Shuffle randomizes the order of cards in the zone using rng.
func (z *Zone) Shuffle(rng *rand.Rand) {
	if rng == nil {
		panic("nil rng")
	}
	rng.Shuffle(len(z.cards), func(i, j int) {
		z.cards[i], z.cards[j] = z.cards[j], z.cards[i]
	})
}

// Size returns the number of cards in the zone.
func (z *Zone) Size() int {
	return len(z.cards)
}

// Contains reports whether the zone contains the given card.
func (z *Zone) Contains(cardID id.ID) bool {
	return slices.Contains(z.cards, cardID)
}

// All returns a copy of all card IDs in the zone, in order.
func (z *Zone) All() []id.ID {
	result := make([]id.ID, len(z.cards))
	copy(result, z.cards)
	return result
}

// SetFaceDown marks a card in this zone as face-down.
func (z *Zone) SetFaceDown(cardID id.ID, faceDown bool) {
	if z.faceDown == nil {
		z.faceDown = make(map[id.ID]bool)
	}
	if faceDown {
		z.faceDown[cardID] = true
	} else {
		delete(z.faceDown, cardID)
	}
}

// IsFaceDown reports whether a card in this zone is face-down.
func (z *Zone) IsFaceDown(cardID id.ID) bool {
	return z.faceDown[cardID]
}
