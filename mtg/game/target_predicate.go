package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TargetAllow identifies broad categories a target spec can choose from. The
// targetability test starts with the word "target" and the target kind chosen as
// part of casting/activating/trigger placement (CR 115, CR 601.2c, CR 603.3d).
type TargetAllow int

// Target allow values identify broad categories a target can choose from.
const (
	TargetAllowUnspecified TargetAllow = 0
	TargetAllowPermanent   TargetAllow = 1 << 0
	TargetAllowPlayer      TargetAllow = 1 << 1
	TargetAllowStackObject TargetAllow = 1 << 2
	TargetAllowCard        TargetAllow = 1 << 3
)

// ControllerRelation constrains a permanent by its controller relative to the
// player choosing targets.
type ControllerRelation int

// Controller relation values compare controllers to the choosing player.
const (
	ControllerAny ControllerRelation = iota
	ControllerYou
	ControllerOpponent
	ControllerNotYou
)

// PlayerRelation constrains a player target relative to the player choosing targets.
type PlayerRelation int

// Player relation values compare players to the choosing player.
const (
	PlayerAny PlayerRelation = iota
	PlayerYou
	PlayerOpponent
	PlayerNotYou
)

// TriState represents an optional boolean predicate.
type TriState int

// Tri-state values express any, true, or false predicates.
const (
	TriAny TriState = iota
	TriTrue
	TriFalse
)

// CombatStateFilter constrains a permanent by current combat involvement.
type CombatStateFilter int

// Combat state filter values match current combat involvement.
const (
	CombatStateAny CombatStateFilter = iota
	CombatStateAttacking
	CombatStateBlocking
	CombatStateAttackingOrBlocking
)

// TargetPredicate carries structured target legality predicates parsed from
// common oracle text. Empty fields are wildcards. These predicates model target
// restrictions that must be legal when chosen and again on resolution
// (CR 115.1, CR 608.2b).
type TargetPredicate struct {
	PermanentTypes []types.Card
	ExcludedTypes  []types.Card

	SpellCardTypes         []types.Card
	ExcludedSpellCardTypes []types.Card
	StackObjectKinds       []StackObjectKind

	Colors         []color.Color
	ExcludedColors []color.Color

	Controller ControllerRelation
	Player     PlayerRelation

	Tapped      TriState
	CombatState CombatStateFilter

	Keyword         Keyword
	ExcludedKeyword Keyword

	ManaValue opt.V[compare.Int]
	Power     opt.V[compare.Int]
	Toughness opt.V[compare.Int]

	Another bool
}
