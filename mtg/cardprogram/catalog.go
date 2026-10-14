// Package cardprogram defines immutable authored programs and their execution seam.
// It depends on game data, never on the rules implementation or raw Game state.
package cardprogram

import (
	"errors"
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// Program associates a stable binding with synchronous Go resolution.
// Resolve must not capture mutable per-game state; use the supplied Context.
type Program struct {
	Reference      game.ProgramReference
	SourceRevision string
	RulesRevision  string
	Resolve        func(Context) error
	// Dependencies declares handlers scheduled or introduced by Go resolution,
	// whose bodies cannot be discovered by inspecting the Card Definition.
	Dependencies []game.ProgramReference
}

// Binding identifies installed executable code and its shared-rule dependency.
// Empty revisions on custom fixtures mean unbound source, not verified fidelity.
type Binding struct {
	Reference      game.ProgramReference `json:"reference"`
	SourceRevision string                `json:"source_revision"`
	RulesRevision  string                `json:"rules_revision"`
}

// Catalog is immutable after construction and safe to share between Engines.
// Its zero value is an empty catalog, not an implicit fallback implementation.
type Catalog struct {
	programs map[string]Program
}

// NewCatalog copies definitions and refuses ambiguous or missing implementations.
func NewCatalog(programs []Program) (Catalog, error) {
	catalog := Catalog{programs: make(map[string]Program, len(programs))}
	for _, program := range programs {
		if err := program.Reference.Validate(); err != nil {
			return Catalog{}, err
		}
		if program.Resolve == nil {
			return Catalog{}, fmt.Errorf("program %s has no resolver", program.Reference.ID)
		}
		if _, exists := catalog.programs[program.Reference.ID]; exists {
			return Catalog{}, fmt.Errorf("duplicate program ID %s", program.Reference.ID)
		}
		program.Dependencies = slices.Clone(program.Dependencies)
		catalog.programs[program.Reference.ID] = program
	}
	for _, program := range programs {
		for _, dependency := range program.Dependencies {
			if _, err := catalog.Lookup(dependency); err != nil {
				return Catalog{}, fmt.Errorf("program %s dependency: %w", program.Reference.ID, err)
			}
		}
	}
	return catalog, nil
}

// Bindings includes transitive Go-only dependencies in deterministic order.
func (c Catalog) Bindings(references []game.ProgramReference) ([]Binding, error) {
	var out []Binding
	seen := make(map[game.ProgramReference]bool)
	var visit func(game.ProgramReference) error
	visit = func(ref game.ProgramReference) error {
		if seen[ref] {
			return nil
		}
		seen[ref] = true
		p, err := c.Lookup(ref)
		if err != nil {
			return err
		}
		out = append(out, Binding{Reference: ref, SourceRevision: p.SourceRevision, RulesRevision: p.RulesRevision})
		for _, dependency := range p.Dependencies {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		return nil
	}
	for _, ref := range references {
		if err := visit(ref); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(out, func(a, b Binding) int {
		if a.Reference.ID < b.Reference.ID {
			return -1
		}
		if a.Reference.ID > b.Reference.ID {
			return 1
		}
		return 0
	})
	return out, nil
}

// Lookup checks the exact declared version rather than choosing a nearby handler.
func (c Catalog) Lookup(reference game.ProgramReference) (Program, error) {
	if err := reference.Validate(); err != nil {
		return Program{}, err
	}
	program, exists := c.programs[reference.ID]
	if !exists {
		return Program{}, fmt.Errorf("program %s is not installed", reference.ID)
	}
	if program.Reference.Version != reference.Version {
		return Program{}, fmt.Errorf("program %s version %s is not installed (have %s)",
			reference.ID, reference.Version, program.Reference.Version)
	}
	program.Dependencies = slices.Clone(program.Dependencies)
	return program, nil
}

// Run executes one binding against an Engine-owned context.
func (c Catalog) Run(reference game.ProgramReference, ctx Context) error {
	if ctx == nil {
		return errors.New("program execution requires a context")
	}
	program, err := c.Lookup(reference)
	if err != nil {
		return err
	}
	if err := program.Resolve(ctx); err != nil {
		return fmt.Errorf("program %s: %w", reference.ID, err)
	}
	return nil
}

// ValidateCard rejects missing/version-mismatched code before loading a deck.
func (c Catalog) ValidateCard(def *game.CardDef) error {
	if def == nil {
		return errors.New("cannot validate programs for a nil card")
	}
	references := def.ProgramReferences()
	if len(references) != 0 {
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			return fmt.Errorf("%s: invalid authored declaration: %s", def.Name, issues[0].Message)
		}
	}
	for _, reference := range references {
		if _, err := c.Lookup(reference); err != nil {
			return fmt.Errorf("%s: %w", def.Name, err)
		}
	}
	return nil
}

// MarshalJSON refuses to pretend that installed Go functions are portable data.
// Serialize game.ProgramReference values and load them against an installed catalog.
func (Catalog) MarshalJSON() ([]byte, error) {
	return nil, errors.New("program catalogs contain Go code and cannot be serialized; serialize versioned references instead")
}
