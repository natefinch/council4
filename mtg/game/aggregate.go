package game

import "github.com/natefinch/council4/mtg/game/compare"

// AggregateKind identifies a player- or board-derived quantity that a
// Condition can compare against a threshold. It replaces the family of
// dedicated comparator fields (AtLeast/AtMost/Exactly) that previously encoded
// both the quantity and the comparator direction separately. A single quantity
// is described once by an AggregateKind; the comparator direction is carried by
// compare.Op, so adding support for "at least", "at most", and "exactly" for a
// quantity requires no new Condition fields.
type AggregateKind int

const (
	// AggregateNone is the zero value and identifies no quantity.
	AggregateNone AggregateKind = iota

	// AggregateControllerLife is the context controller's current life total.
	AggregateControllerLife
)

// AggregateComparison compares a player- or board-derived quantity against a
// threshold using a typed comparator. Conditions hold a slice of these so
// multiple quantities can be ANDed without one Condition field per quantity.
type AggregateComparison struct {
	// Aggregate identifies which quantity to evaluate.
	Aggregate AggregateKind

	// Op is the comparator direction applied to the evaluated quantity.
	Op compare.Op

	// Value is the threshold the quantity is compared against.
	Value int
}
