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
// Each letter sub-package exports a var Cards []cardset.Entry: a name paired with
// a constructor that builds the card's CardDef on demand. The Registry in this
// package combines them into a name lookup and constructs each CardDef the first
// time it is looked up, so registering the whole corpus (tens of thousands of
// cards) costs only a name and a function pointer per card rather than building
// every CardDef up front — a game touches only a few hundred of them.
package cards

//go:generate go run github.com/natefinch/council4/cardgen/cmd/genregistry

import (
	"sync"

	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
)

// Registry indexes card constructors by canonical card name and builds each
// CardDef lazily, caching the result so repeated lookups return the same value
// (CardDefs are immutable templates the engine instantiates from).
//
// The cache is populated on Lookup, so the Registry may be used concurrently
// (for example by parallel simulations sharing one corpus); a mutex guards the
// cache. The entries map is written only during construction and read-only
// afterward.
type Registry struct {
	entries map[string][]func() *game.CardDef

	mu    sync.Mutex
	cache map[string][]*game.CardDef

	count int
}

func emptyRegistry() *Registry {
	return &Registry{
		entries: make(map[string][]func() *game.CardDef),
		cache:   make(map[string][]*game.CardDef),
	}
}

// NewRegistry creates a Registry from already-constructed card slices. Distinct
// definitions may share a printed card name. It is the eager constructor, kept
// for tests and callers that already hold built CardDefs; the full corpus is
// registered lazily via NewLazyRegistry instead.
func NewRegistry(cardSets ...[]*game.CardDef) *Registry {
	r := emptyRegistry()
	for _, set := range cardSets {
		for _, card := range set {
			r.entries[card.Name] = append(r.entries[card.Name], func() *game.CardDef { return card })
			r.count++
		}
	}
	return r
}

// NewLazyRegistry creates a Registry from card-constructor entries, building each
// CardDef only when it is first looked up. Registering the full corpus this way
// allocates only a name and a function pointer per card, not the CardDef, so the
// large committed corpus does not sit in memory when only a handful of cards are
// ever used (the client-side WebAssembly playtester relies on this to stay within
// the browser's memory ceiling). Distinct definitions may share a printed name.
func NewLazyRegistry(entrySets ...[]cardset.Entry) *Registry {
	r := emptyRegistry()
	for _, set := range entrySets {
		for _, entry := range set {
			r.entries[entry.Name] = append(r.entries[entry.Name], entry.New)
			r.count++
		}
	}
	return r
}

// NewDefaultRegistry returns a Registry over the full committed card corpus —
// every card in the letter sub-packages, built lazily. Token definitions are
// intentionally excluded: they are not real cards and must not resolve from a
// decklist.
//
// The aggregated set is maintained by genregistry (see registry_sets.go); run
// go generate ./mtg/cards/... after adding a new letter sub-package.
func NewDefaultRegistry() *Registry {
	return NewLazyRegistry(defaultCardSets()...)
}

// DefaultCardSets returns the Cards entries from every committed letter
// sub-package. It is the data backing NewDefaultRegistry.
func DefaultCardSets() [][]cardset.Entry {
	return defaultCardSets()
}

// Lookup returns the first CardDef for the given card name, or nil if not found,
// constructing and caching it on first use.
func (r *Registry) Lookup(name string) *game.CardDef {
	matches := r.LookupAll(name)
	if len(matches) == 0 {
		return nil
	}
	return matches[0]
}

// LookupAll returns every CardDef with the given printed card name, constructing
// and caching them on first use. The returned slice is a fresh copy the caller
// may retain; the CardDefs themselves are shared, immutable templates.
func (r *Registry) LookupAll(name string) []*game.CardDef {
	r.mu.Lock()
	defer r.mu.Unlock()
	if built, ok := r.cache[name]; ok {
		return append([]*game.CardDef(nil), built...)
	}
	constructors, ok := r.entries[name]
	if !ok {
		return nil
	}
	built := make([]*game.CardDef, len(constructors))
	for i, construct := range constructors {
		built[i] = construct()
	}
	r.cache[name] = built
	return append([]*game.CardDef(nil), built...)
}

// All returns all registered card names. It does not construct any CardDef.
func (r *Registry) All() []string {
	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}
	return names
}

// Len returns the number of registered cards.
func (r *Registry) Len() int {
	return r.count
}
