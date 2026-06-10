// Package cards provides a registry indexing CardDef values by canonical card
// name. Card definitions live in letter-based sub-packages (a/, b/, ..., z/)
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
// name lookup.
package cards

import "github.com/natefinch/council4/mtg/game"

// Registry indexes CardDef values by canonical card name.
type Registry struct {
	cards map[string][]*game.CardDef
	count int
}

// NewRegistry creates a Registry from the given card slices.
// Distinct definitions may share a printed card name.
func NewRegistry(cardSets ...[]*game.CardDef) *Registry {
	r := &Registry{cards: make(map[string][]*game.CardDef)}
	for _, set := range cardSets {
		for _, card := range set {
			r.cards[card.Name] = append(r.cards[card.Name], card)
			r.count++
		}
	}
	return r
}

// Lookup returns the first CardDef for the given card name, or nil if not found.
func (r *Registry) Lookup(name string) *game.CardDef {
	matches := r.cards[name]
	if len(matches) == 0 {
		return nil
	}
	return matches[0]
}

// LookupAll returns every CardDef with the given printed card name.
func (r *Registry) LookupAll(name string) []*game.CardDef {
	return append([]*game.CardDef(nil), r.cards[name]...)
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
	return r.count
}
