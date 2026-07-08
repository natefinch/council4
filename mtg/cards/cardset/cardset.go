// Package cardset defines the lazy card-registration entry shared by the
// generated letter packages (a/, b/, ..., z/) and the cards registry.
//
// It exists as a separate leaf package to break an import cycle: the cards
// package imports every letter package (to aggregate the default corpus), so a
// letter package cannot import cards. Both sides can import this package, which
// depends only on game.
package cardset

import "github.com/natefinch/council4/mtg/game"

// Entry pairs a card's canonical (printed) name with a constructor that builds
// its CardDef on demand. A generated letter package exports its cards as a
// []Entry so a registry can index every card by name without constructing all of
// them: only the cards a caller actually looks up are ever built. The name is
// emitted at generation time (it is the card's first-face Name), so indexing
// costs nothing but a string and a function pointer per card.
type Entry struct {
	// Name is the canonical printed card name, matching the Name of the CardDef
	// that New returns (the front face for a multi-face card).
	Name string
	// New constructs the card's CardDef. A registry calls it at most once per
	// registered entry and caches the result.
	New func() *game.CardDef
}
