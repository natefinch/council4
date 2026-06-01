package game

import "github.com/natefinch/council4/opt"

// Condition is a reusable rules predicate evaluated by mtg/rules in an explicit
// context such as a static ability, activation restriction, trigger, effect, or
// replacement event.
type Condition struct {
	// Text preserves the printed condition for diagnostics and generated-card
	// review.
	Text string

	// Negate inverts the whole condition, e.g. "unless you control...".
	Negate bool

	// ControllerControls requires the context controller to control matching
	// permanents. It is ignored when the filter is empty.
	ControllerControls PermanentFilter
}

// PermanentFilter matches permanents for reusable condition predicates. Empty
// fields are wildcards. Types and Supertypes are all required; SubtypesAny
// matches when any listed subtype is present.
type PermanentFilter struct {
	Types       []CardType
	Supertypes  []Supertype
	SubtypesAny []string

	// MinCount defaults to 1 when any other filter field is set.
	MinCount int

	Power     opt.V[IntComparison]
	Toughness opt.V[IntComparison]
}

// Empty reports whether the filter contains no active predicate.
func (f PermanentFilter) Empty() bool {
	return len(f.Types) == 0 &&
		len(f.Supertypes) == 0 &&
		len(f.SubtypesAny) == 0 &&
		f.MinCount == 0 &&
		!f.Power.Exists &&
		!f.Toughness.Exists
}

// Empty reports whether the condition contains no active predicate.
func (c Condition) Empty() bool {
	return c.ControllerControls.Empty()
}
