// Package deck parses Magic: The Gathering Commander decklists in the standard
// Moxfield/MTGO text export format into structured data the engine can load.
//
// A Decklist is the file specification — card names with quantities, plus the
// commander. It is distinct from the in-game Library zone: a Decklist describes
// what a player registered, while the Library is where those cards live during
// a game.
package deck

// Entry is one decklist line: a card name and how many copies it specifies.
type Entry struct {
	// Quantity is the number of copies; always positive in a parsed Entry.
	Quantity int

	// Name is the card name with leading quantity and trailing set/collector
	// or foil annotations removed.
	Name string
}

// Decklist is a parsed Commander decklist: the commander(s) plus the main deck.
type Decklist struct {
	// Commander holds the commander entries. It is usually one card, or two for
	// partner/background pairings. It is empty when the decklist text has no
	// commander section; callers may then designate the commander separately.
	Commander []Entry

	// Cards holds the main-deck entries, excluding the commander.
	Cards []Entry
}

// Count returns the total number of cards in the decklist, summing the
// quantities of every commander and main-deck entry.
func (d *Decklist) Count() int {
	total := 0
	for i := range d.Commander {
		total += d.Commander[i].Quantity
	}
	for i := range d.Cards {
		total += d.Cards[i].Quantity
	}
	return total
}
