// Package cards provides a registry mapping canonical card names to CardDef
// values. Card definitions live in letter-based sub-packages (a/, b/, ..., z/)
// and are aggregated here.
//
// # Adding a card
//
// Use the Oracle compiler to generate supported card definitions:
//
//	go run ./cardgen/oracle/cmd/compilecards -in oracle-cards.json -out mtg/cards
//
// After adding or editing cards manually, run go generate to update the
// sub-package card lists:
//
//	go generate ./mtg/cards/...
//
// # Architecture
//
// Each letter sub-package exports a var Cards []*game.CardDef containing all
// cards in that package. The Registry in this package combines them into a
// single name→CardDef lookup.
package cards

import "github.com/natefinch/council4/mtg/game"

// Registry maps canonical card names to their CardDef values.
type Registry struct {
	cards map[string]*game.CardDef
}

// NewRegistry creates a Registry from the given card slices.
// Duplicate card names cause a panic.
func NewRegistry(cardSets ...[]*game.CardDef) *Registry {
	r := &Registry{cards: make(map[string]*game.CardDef)}
	for _, set := range cardSets {
		for _, card := range set {
			if _, exists := r.cards[card.Name]; exists {
				panic("cards: duplicate card name: " + card.Name)
			}
			r.cards[card.Name] = card
		}
	}
	return r
}

// Lookup returns the CardDef for the given card name, or nil if not found.
func (r *Registry) Lookup(name string) *game.CardDef {
	return r.cards[name]
}

// All returns all registered card names.
func (r *Registry) All() []string {
	names := make([]string, 0, len(r.cards))
	for name := range r.cards {
		names = append(names, name)
	}
	return names
}

// Len returns the number of registered cards.
func (r *Registry) Len() int {
	return len(r.cards)
}
