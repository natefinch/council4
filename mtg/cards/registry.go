// Package cards provides a registry indexing CardDef values by canonical card
// name. Card definitions live in letter-based sub-packages (a/, b/, ..., z/)
// and are aggregated here.
//
// # Adding a card
//
// New support uses authored Go over shared Engine operations, not Oracle
// compiler expansion. See docs/adr/0012-authored-go-card-implementations.md and
// mtg/cards/README.md for the migration prerequisites and legacy regeneration.
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

//go:generate go run github.com/natefinch/council4/cmd/genregistry

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
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

	count         int
	programs      cardprogram.Catalog
	oracleNames   map[string]string
	oracleIndices map[string]int
	identities    map[string][]cardset.Entry
}

func emptyRegistry() *Registry {
	return &Registry{
		entries:       make(map[string][]func() *game.CardDef),
		cache:         make(map[string][]*game.CardDef),
		oracleNames:   make(map[string]string),
		oracleIndices: make(map[string]int),
		identities:    make(map[string][]cardset.Entry),
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
			r.identities[card.Name] = append(r.identities[card.Name], cardset.Entry{Name: card.Name})
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
	registry, err := NewProgramRegistry(entrySets...)
	if err != nil {
		panic(err)
	}
	return registry
}

// NewProgramRegistry combines legacy definitions and explicit authored identities.
// Authored entries require explicit card-specific verification references; these
// are recorded evidence, not tests executed or correctness proven by admission.
// It constructs code catalogs eagerly but leaves CardDefs lazy.
func NewProgramRegistry(entrySets ...[]cardset.Entry) (*Registry, error) {
	r := emptyRegistry()
	var programs []cardprogram.Program
	var entries []cardset.Entry
	for _, set := range entrySets {
		entries = append(entries, set...)
	}
	// Identity selection is independent of which set the caller supplies first:
	// establish legacy names, then apply only explicit authored replacements.
	slices.SortStableFunc(entries, func(a, b cardset.Entry) int {
		if a.OracleID == "" && b.OracleID != "" {
			return -1
		}
		if a.OracleID != "" && b.OracleID == "" {
			return 1
		}
		return 0
	})
	for _, entry := range entries {
		if entry.Name == "" || entry.New == nil {
			return nil, fmt.Errorf("registry entry %q has no name or constructor", entry.Name)
		}
		if entry.OracleID != "" {
			verification := strings.TrimSpace(entry.Verification)
			if verification == "" || verification == "mtg/rules named-card scenarios; mixed production acceptance (#3278)" {
				return nil, fmt.Errorf("authored %s requires explicit card-specific verification evidence", entry.Name)
			}
			if old, exists := r.oracleNames[entry.OracleID]; exists {
				return nil, fmt.Errorf("duplicate Oracle identity %s (%s, %s)", entry.OracleID, old, entry.Name)
			}
			if existing := r.entries[entry.Name]; len(existing) > 0 {
				knownAuthored := false
				for _, name := range r.oracleNames {
					knownAuthored = knownAuthored || name == entry.Name
				}
				if knownAuthored {
					if entry.ReplaceLegacy {
						return nil, fmt.Errorf("%s cannot replace an already authored identity as legacy", entry.Name)
					}
				} else {
					if !entry.ReplaceLegacy || len(existing) != 1 {
						return nil, fmt.Errorf("authored %s collides with an existing name; explicit unique legacy replacement required", entry.Name)
					}
					r.count -= len(existing)
					delete(r.entries, entry.Name)
					delete(r.identities, entry.Name)
				}
			} else if entry.ReplaceLegacy {
				return nil, fmt.Errorf("authored %s requests a nonexistent legacy replacement", entry.Name)
			}
			r.oracleNames[entry.OracleID] = entry.Name
			r.oracleIndices[entry.OracleID] = len(r.entries[entry.Name])
		} else if len(entry.Programs) != 0 || entry.ReplaceLegacy {
			return nil, fmt.Errorf("authored entry %s requires an explicit Oracle identity", entry.Name)
		}
		r.entries[entry.Name] = append(r.entries[entry.Name], entry.New)
		entry.Programs = slices.Clone(entry.Programs)
		r.identities[entry.Name] = append(r.identities[entry.Name], entry)
		programs = append(programs, entry.Programs...)
		r.count++
	}
	var err error
	r.programs, err = cardprogram.NewCatalog(programs)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// NewDefaultRegistry returns a Registry over the full committed card corpus —
// every card in the letter sub-packages, built lazily. Token definitions are
// intentionally excluded: they are not real cards and must not resolve from a
// decklist.
//
// The aggregated set is maintained by genregistry (see registry_sets.go); run
// go generate ./mtg/cards/... after adding a new letter sub-package.
func NewDefaultRegistry() *Registry {
	return NewLazyRegistry(DefaultCardSets()...)
}

// DefaultCardSets returns the Cards entries from every committed letter
// sub-package. It is the data backing NewDefaultRegistry.
func DefaultCardSets() [][]cardset.Entry {
	return append(installedCardSets(), authored.Cards)
}

// Programs returns the immutable catalog installed with the definitions.
func (r *Registry) Programs() cardprogram.Catalog { return r.programs }

var defaultPrograms = sync.OnceValues(func() (cardprogram.Catalog, error) {
	registry, err := NewProgramRegistry(DefaultCardSets()...)
	if err != nil {
		return cardprogram.Catalog{}, err
	}
	return registry.Programs(), nil
})

// DefaultPrograms builds the immutable production catalog once without retaining
// per-game state. Custom registries supply their own Programs explicitly.
func DefaultPrograms() (cardprogram.Catalog, error) { return defaultPrograms() }

// LookupOracleID chooses an exact authored identity even when names are shared.
// Legacy name-only entries remain available by name, not a guessed Oracle ID.
func (r *Registry) LookupOracleID(oracleID string) (*game.CardDef, error) {
	name, exists := r.oracleNames[oracleID]
	if !exists {
		return nil, fmt.Errorf("oracle identity %s is not registered", oracleID)
	}
	built, err := r.LookupAllChecked(name)
	if err != nil {
		return nil, err
	}
	return built[r.oracleIndices[oracleID]], nil
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
	built, err := r.LookupAllChecked(name)
	if err != nil {
		panic(err)
	}
	return built
}

// LookupChecked reports a missing/version-mismatched program during deck loading.
func (r *Registry) LookupChecked(name string) (*game.CardDef, error) {
	built, err := r.LookupAllChecked(name)
	if err != nil || len(built) == 0 {
		return nil, err
	}
	return built[0], nil
}

// LookupAllChecked constructs definitions once and verifies their code bindings.
func (r *Registry) LookupAllChecked(name string) ([]*game.CardDef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if built, ok := r.cache[name]; ok {
		return append([]*game.CardDef(nil), built...), nil
	}
	constructors, ok := r.entries[name]
	if !ok {
		return nil, nil
	}
	built := make([]*game.CardDef, len(constructors))
	for i, construct := range constructors {
		built[i] = construct()
		if err := r.programs.ValidateCard(built[i]); err != nil {
			return nil, err
		}
	}
	r.cache[name] = built
	return append([]*game.CardDef(nil), built...), nil
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
