package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoveTopOfLibrary moves the top Amount cards of a referenced player's library
// to Destination. It is the single "move the top of a library to a zone"
// primitive: the destination is a parameter rather than a distinct primitive
// type, so a new destination, or a new combination of destination and rider,
// costs no new runtime code.
//
// It replaces two primitives that differed only in where the cards landed: Mill
// (graveyard) and ExileTopOfLibrary (exile). Each had picked up a rider the
// other could not express — Mill alone could publish a linked key, while
// ExileTopOfLibrary alone could place a marker counter or exile face down — so
// "mill three cards, then act on exactly those cards" and "exile the top card
// with a counter on it" each required their own primitive. Here they are
// combinations of existing fields, and every rider is available at every
// destination the runtime can honor.
//
// It puts cards from the top of a referenced player's library into Destination,
// or does so for every player in a referenced group ("each player mills", "each
// opponent exiles"). Exactly one of Player or PlayerGroup is set.
type MoveTopOfLibrary struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set

	// Destination is the zone the cards move to. Validation admits only
	// zone.Graveyard (mill) and zone.Exile (exile the top of the library),
	// because those are the only destinations the runtime implements; a new
	// destination means a new runtime path in handleMoveTopOfLibrary.
	Destination zone.Type

	// PublishLinked, when set, remembers every card that actually reaches
	// Destination as a card-scoped linked object on the source permanent so a
	// later instruction can act on exactly those cards ("mill three cards. ...
	// put a card from among those cards into your hand"). It is meaningful only
	// for the single Player form; the group form publishes nothing.
	PublishLinked LinkedKey

	// Counter names a named marker counter placed on each card once it reaches
	// exile ("exile the top card of each player's library with a collection
	// counter on it.", Evelyn, the Covetous). The counter is recorded in
	// Game.ExileCounters, which exists only for cards in exile, so validation
	// rejects a Counter with any non-exile Destination. It is unset for every
	// move that places no counter.
	Counter opt.V[counter.Kind]

	// FaceDown moves each card face down ("exile that many cards from the top of
	// your library face down.", Flamewar, Streetwise Operative). A face-down
	// card in exile hides its identity from every observer (CR 713); the exile
	// zone records the face-down state and clears it when the card leaves. Only
	// the exile zone tracks face-down state, so validation rejects FaceDown with
	// any non-exile Destination. It is false for the ordinary face-up move.
	FaceDown bool
}
