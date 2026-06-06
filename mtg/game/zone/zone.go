// Package zone defines Magic game zones and their ordered card collections.
package zone

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game/id"
)

// Type identifies which game zone a card is in (CR 400).
type Type int

const (
	// None means no zone applies, such as a newly created token before it
	// enters the battlefield.
	None Type = iota

	// Library is a player's draw deck — hidden and ordered (CR 401).
	Library

	// Hand is a player's hand — hidden from opponents (CR 402).
	Hand

	// Battlefield is the shared play area — public (CR 403).
	// The battlefield is represented separately from Zone by game.Game.
	Battlefield

	// Graveyard is a player's discard pile — public, ordered (CR 404).
	Graveyard

	// Stack is the shared stack — public, ordered LIFO (CR 405).
	// The stack is represented by game.Stack, not Zone.
	Stack

	// Exile is the exile zone — usually public, sometimes face-down (CR 406).
	Exile

	// Command is the command zone — public (CR 408).
	// Commanders start here in Commander format.
	Command
)

// String returns the zone name.
func (t Type) String() string {
	switch t {
	case None:
		return "None"
	case Library:
		return "Library"
	case Hand:
		return "Hand"
	case Battlefield:
		return "Battlefield"
	case Graveyard:
		return "Graveyard"
	case Stack:
		return "Stack"
	case Exile:
		return "Exile"
	case Command:
		return "Command"
	default:
		return "Unknown"
	}
}

// IsPublic reports whether cards in this zone are visible to all players.
func (t Type) IsPublic() bool {
	switch t {
	case Battlefield, Graveyard, Stack, Exile, Command:
		return true
	default:
		return false
	}
}

// Zone is an ordered collection of card instance IDs. The library and
// graveyard are ordered with the top at index 0. The hand is conceptually
// unordered but stored as a slice.
type Zone struct {
	// Type identifies which zone this is.
	Type Type

	// cards stores card instance IDs in order.
	cards []id.ID

	// faceDown tracks which cards in this zone are face-down.
	faceDown map[id.ID]bool
}

// New creates an empty zone of the given type.
func New(t Type) Zone {
	return Zone{
		Type:     t,
		faceDown: make(map[id.ID]bool),
	}
}

// Add adds a card to the top of the zone.
func (z *Zone) Add(cardID id.ID) {
	z.cards = append([]id.ID{cardID}, z.cards...)
}

// AddToBottom adds a card to the bottom of the zone.
func (z *Zone) AddToBottom(cardID id.ID) {
	z.cards = append(z.cards, cardID)
}

// Remove removes a card from the zone. It reports whether the card was found.
func (z *Zone) Remove(cardID id.ID) bool {
	for i, card := range z.cards {
		if card == cardID {
			z.cards = append(z.cards[:i], z.cards[i+1:]...)
			delete(z.faceDown, cardID)
			return true
		}
	}
	return false
}

// Top returns the top card and reports whether the zone is non-empty.
func (z *Zone) Top() (id.ID, bool) {
	if len(z.cards) == 0 {
		return 0, false
	}
	return z.cards[0], true
}

// Bottom returns the bottom card and reports whether the zone is non-empty.
func (z *Zone) Bottom() (id.ID, bool) {
	if len(z.cards) == 0 {
		return 0, false
	}
	return z.cards[len(z.cards)-1], true
}

// Shuffle randomizes the card order using rng.
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

// Contains reports whether the zone contains cardID.
func (z *Zone) Contains(cardID id.ID) bool {
	return slices.Contains(z.cards, cardID)
}

// All returns a copy of all card IDs in order.
func (z *Zone) All() []id.ID {
	result := make([]id.ID, len(z.cards))
	copy(result, z.cards)
	return result
}

// SetFaceDown records whether a card in this zone is face-down.
func (z *Zone) SetFaceDown(cardID id.ID, faceDown bool) {
	if z.faceDown == nil {
		z.faceDown = make(map[id.ID]bool)
	}
	if faceDown {
		z.faceDown[cardID] = true
		return
	}
	delete(z.faceDown, cardID)
}

// IsFaceDown reports whether a card in this zone is face-down.
func (z *Zone) IsFaceDown(cardID id.ID) bool {
	return z.faceDown[cardID]
}
