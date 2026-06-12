package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

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

	// ControlsMatching is the Selection-based successor to ControllerControls.
	// When present, the context controller must control at least MinCount
	// objects matching the Selection (MinCount defaults to 1), optionally
	// constrained by TotalPower. ControllerControls and ControlsMatching must
	// not both be specified.
	ControlsMatching opt.V[SelectionCount]

	// ControllerLifeAtLeast requires the context controller's current life total
	// to meet the threshold. AnyPlayerLifeAtMost checks every non-eliminated
	// player. Zero values disable these predicates.
	ControllerLifeAtLeast     int
	ControllerHandSizeAtLeast int
	AnyPlayerLifeAtMost       int

	// OpponentCountAtLeast requires this many non-eliminated opponents.
	OpponentCountAtLeast int

	// ControllerHandEmpty and the controller-relative thresholds model
	// live game-state predicates used by ability words such as threshold,
	// delirium, domain, hellbent, and coven.
	ControllerHandEmpty                     bool
	ControllerGraveyardCardCountAtLeast     int
	ControllerGraveyardCardTypeCountAtLeast int
	ControllerBasicLandTypeCountAtLeast     int
	ControllerCreaturePowerDiversityAtLeast int

	// AnyOpponentControls checks each opponent independently. OpponentsControl
	// counts matching permanents controlled by all opponents collectively.
	AnyOpponentControls opt.V[SelectionCount]
	OpponentsControl    opt.V[SelectionCount]

	// Object tests a referenced object in the current condition context, such as
	// a triggering event permanent. It may use last-known information.
	// ObjectMatches, when present, applies the shared Selection semantics to that
	// object. An empty ObjectMatches Selection is a wildcard existence check.
	Object                                                       opt.V[ObjectReference]
	ObjectMatches                                                opt.V[Selection]
	Types                                                        []types.Card
	EventPermanentNameUniqueAmongControlledAndGraveyardCreatures bool
	SourceClassLevelAtLeast                                      int
	SourceClassLevelLessThan                                     int
	SourceNotMonstrous                                           bool
	ControllerHasMaxSpeed                                        bool
	TargetEnteredThisTurn                                        opt.V[int]
	CastFromZone                                                 opt.V[zone.Type]

	// EventHistory is satisfied when the selected turn's event history contains
	// at least one event matching the stored pattern. When Condition.Negate is
	// true the predicate is inverted (e.g. "if no spells were cast last turn").
	EventHistory opt.V[EventHistoryCondition]
}

// PermanentFilter matches permanents for reusable condition predicates. Empty
// fields are wildcards. Types and Supertypes are all required; SubtypesAny and
// ColorsAny match when any listed value is present.
type PermanentFilter struct {
	Types          []types.Card
	Supertypes     []types.Super
	SubtypesAny    []types.Sub
	ColorsAny      []color.Color
	ExcludedColors []color.Color

	// MinCount defaults to 1 when any other filter field is set.
	MinCount int

	Power      opt.V[compare.Int]
	Toughness  opt.V[compare.Int]
	TotalPower opt.V[compare.Int]

	// ExcludeSource ignores the condition source permanent when counting
	// matches, for conditions that ask for "another" permanent.
	ExcludeSource bool
}

// Empty reports whether the filter contains no active predicate.
func (f PermanentFilter) Empty() bool {
	return len(f.Types) == 0 &&
		len(f.Supertypes) == 0 &&
		len(f.SubtypesAny) == 0 &&
		len(f.ColorsAny) == 0 &&
		len(f.ExcludedColors) == 0 &&
		f.MinCount == 0 &&
		!f.Power.Exists &&
		!f.Toughness.Exists &&
		!f.TotalPower.Exists &&
		!f.ExcludeSource
}

// Empty reports whether the condition contains no active predicate.
func (c *Condition) Empty() bool {
	return c.ControllerControls.Empty() &&
		!c.ControlsMatching.Exists &&
		c.ControllerLifeAtLeast == 0 &&
		c.ControllerHandSizeAtLeast == 0 &&
		c.AnyPlayerLifeAtMost == 0 &&
		c.OpponentCountAtLeast == 0 &&
		!c.ControllerHandEmpty &&
		c.ControllerGraveyardCardCountAtLeast == 0 &&
		c.ControllerGraveyardCardTypeCountAtLeast == 0 &&
		c.ControllerBasicLandTypeCountAtLeast == 0 &&
		c.ControllerCreaturePowerDiversityAtLeast == 0 &&
		!c.AnyOpponentControls.Exists &&
		!c.OpponentsControl.Exists &&
		!c.Object.Exists &&
		!c.ObjectMatches.Exists &&
		len(c.Types) == 0 &&
		!c.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures &&
		c.SourceClassLevelAtLeast == 0 &&
		c.SourceClassLevelLessThan == 0 &&
		!c.SourceNotMonstrous &&
		!c.ControllerHasMaxSpeed &&
		!c.TargetEnteredThisTurn.Exists &&
		!c.CastFromZone.Exists &&
		!c.EventHistory.Exists
}

// EventHistoryWindow selects which turn's event log an EventHistoryCondition
// searches.
type EventHistoryWindow uint8

// Event history window values.
const (
	// EventHistoryCurrentTurn checks events emitted during the current turn.
	EventHistoryCurrentTurn EventHistoryWindow = iota
	// EventHistoryPreviousTurn checks events emitted during the immediately
	// preceding turn.
	EventHistoryPreviousTurn
)

// EventHistoryCondition checks that the chosen turn's event log contains at
// least one event matching Pattern. Negate on the enclosing Condition inverts
// the result (e.g. "if no spells were cast last turn").
type EventHistoryCondition struct {
	Pattern TriggerPattern
	Window  EventHistoryWindow
}
