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

// Valid reports whether the relation is a recognized closed-enum value.
func (r PlayerRelation) Valid() bool {
	return r >= PlayerAny && r <= PlayerNotYou
}

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

// TargetPredicate carries the stack-object and spell-only qualifiers a target
// spec needs that are not permanent/card characteristics. Permanent, card, and
// player characteristic predicates live on the spec's Selection; this struct
// retains only the genuinely stack/spell-specific filters: which stack-object
// kinds the target accepts, the stack object's controller relation, the spell
// mana-value comparison, the matched spell's card types/colors/supertypes, and
// the matched stack object's source card types. Empty fields are wildcards.
// These restrictions must be legal when chosen and again on resolution
// (CR 115.1, CR 608.2b).
type TargetPredicate struct {
	// SpellCardTypes and SpellCardTypesAny restrict a matched spell stack object
	// by its card types (all of, respectively any of). ExcludedSpellCardTypes
	// names card types a matched spell must not carry. StackObjectKinds lists the
	// stack-object kinds the target accepts (CR 115.4).
	SpellCardTypes         []types.Card
	SpellCardTypesAny      []types.Card
	ExcludedSpellCardTypes []types.Card
	StackObjectKinds       []StackObjectKind

	// SpellSupertypes and SpellColorless qualify the spell kind within a
	// stack-object target that also allows abilities. They restrict only matched
	// spells (CR 115.4); abilities ignore them, so "target activated ability,
	// triggered ability, or legendary spell" requires the supertype only of the
	// spell choice.
	SpellSupertypes []types.Super
	SpellColorless  bool

	// SpellColors, SpellExcludedColors, and SpellMulticolored qualify a matched
	// spell stack object by its current face colors (CR 105, CR 115.4). A spell
	// matches when it has every color in SpellColors, none of the colors in
	// SpellExcludedColors (so "nonblue" also matches colorless spells), and, when
	// SpellMulticolored is set, two or more colors. Like SpellColorless they
	// restrict only matched spells; abilities ignore them.
	SpellColors         []color.Color
	SpellExcludedColors []color.Color
	SpellMulticolored   bool

	// StackObjectSourceTypes requires the matched stack object's source to have
	// all listed card types, modeling "from an artifact source" restrictions on
	// ability-counter targets (CR 113.7).
	StackObjectSourceTypes []types.Card

	// Controller constrains the matched stack object by the controller of the
	// spell or ability relative to the player choosing targets ("counter target
	// spell you don't control").
	Controller ControllerRelation

	// ManaValue compares the matched spell's mana value ("counter target spell
	// with mana value 3 or less"). It applies only to spell stack objects.
	ManaValue opt.V[compare.Int]

	// SpellTargets restricts a matched spell stack object to one whose chosen
	// targets include at least one target satisfying any of these requirements
	// ("counter target spell that targets a permanent you control"). Each
	// requirement matches either a permanent (by card types and controller
	// relation) or a player (by relation), relative to the player choosing the
	// counter target. It applies only to spell stack objects; abilities never
	// match. An empty slice imposes no restriction (CR 115.4, CR 608.2b).
	SpellTargets []SpellTargetRequirement
}

// SpellTargetRequirementKind classifies one alternative of a "that targets <X>"
// spell-target restriction as matching a permanent or a player.
type SpellTargetRequirementKind int

// Spell-target requirement kinds distinguish permanent and player requirements.
const (
	SpellTargetRequirementPermanent SpellTargetRequirementKind = iota
	SpellTargetRequirementPlayer
)

// Valid reports whether the kind is a recognized closed-enum value.
func (k SpellTargetRequirementKind) Valid() bool {
	return k == SpellTargetRequirementPermanent || k == SpellTargetRequirementPlayer
}

// SpellTargetRequirement is one acceptable target of a matched spell's chosen
// targets, for "Counter target spell that targets <X>". A permanent requirement
// matches a targeted permanent that has every type in RequiredTypes (an empty
// list matches any permanent) and whose controller satisfies Controller. A
// player requirement matches a targeted player satisfying Player. Controller and
// Player are relative to the player choosing the counter target.
type SpellTargetRequirement struct {
	Kind          SpellTargetRequirementKind
	RequiredTypes []types.Card
	Controller    ControllerRelation
	Player        PlayerRelation
}

// Empty reports whether the predicate carries no stack-object or spell qualifier
// and therefore imposes no stack-side restriction.
func (p TargetPredicate) Empty() bool {
	return len(p.SpellCardTypes) == 0 &&
		len(p.SpellCardTypesAny) == 0 &&
		len(p.ExcludedSpellCardTypes) == 0 &&
		len(p.StackObjectKinds) == 0 &&
		len(p.SpellSupertypes) == 0 &&
		!p.SpellColorless &&
		len(p.SpellColors) == 0 &&
		len(p.SpellExcludedColors) == 0 &&
		!p.SpellMulticolored &&
		len(p.StackObjectSourceTypes) == 0 &&
		p.Controller == ControllerAny &&
		!p.ManaValue.Exists &&
		len(p.SpellTargets) == 0
}
