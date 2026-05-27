package game

import "github.com/natefinch/council4/opt"

import "github.com/natefinch/council4/mtg/game/mana"

// TargetAllow identifies broad categories a target spec can choose from. The
// targetability test starts with the word "target" and the target kind chosen as
// part of casting/activating/trigger placement (CR 115, CR 601.2c, CR 603.3d).
type TargetAllow int

const (
	TargetAllowUnspecified TargetAllow = 0
	TargetAllowPermanent   TargetAllow = 1 << 0
	TargetAllowPlayer      TargetAllow = 1 << 1
	TargetAllowStackObject TargetAllow = 1 << 2
)

// ControllerRelation constrains a permanent by its controller relative to the
// player choosing targets.
type ControllerRelation int

const (
	ControllerAny ControllerRelation = iota
	ControllerYou
	ControllerOpponent
	ControllerNotYou
)

// PlayerRelation constrains a player target relative to the player choosing targets.
type PlayerRelation int

const (
	PlayerAny PlayerRelation = iota
	PlayerYou
	PlayerOpponent
	PlayerNotYou
)

// TriState represents an optional boolean predicate.
type TriState int

const (
	TriAny TriState = iota
	TriTrue
	TriFalse
)

// CombatStateFilter constrains a permanent by current combat involvement.
type CombatStateFilter int

const (
	CombatStateAny CombatStateFilter = iota
	CombatStateAttacking
	CombatStateBlocking
	CombatStateAttackingOrBlocking
)

// ComparisonOp identifies an integer comparison operation.
type ComparisonOp int

const (
	CompareAny ComparisonOp = iota
	CompareEqual
	CompareLessOrEqual
	CompareGreaterOrEqual
	CompareLessThan
	CompareGreaterThan
)

// IntComparison is a simple comparison against a fixed integer value.
type IntComparison struct {
	Op    ComparisonOp
	Value int
}

// TargetPredicate carries structured target legality predicates parsed from
// common oracle text. Empty fields are wildcards. These predicates model target
// restrictions that must be legal when chosen and again on resolution
// (CR 115.1, CR 608.2b).
type TargetPredicate struct {
	PermanentTypes []CardType
	ExcludedTypes  []CardType

	Colors         []mana.Color
	ExcludedColors []mana.Color

	Controller ControllerRelation
	Player     PlayerRelation

	Tapped      TriState
	CombatState CombatStateFilter

	Keyword         Keyword
	ExcludedKeyword Keyword

	ManaValue opt.V[IntComparison]
	Power     opt.V[IntComparison]
	Toughness opt.V[IntComparison]

	Another bool
}
