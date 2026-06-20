package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CostModifierKind identifies which costs a modifier applies to.
type CostModifierKind int

// Cost modifier kind values identify affected cost categories.
const (
	CostModifierSpell CostModifierKind = iota
	CostModifierAbility
	CostModifierAttack
)

// CostModifier is a generic-cost increase/reduction/set effect.
//
// MatchColor constrains a spell cost modifier to spells of a single color. When
// MatchColor is set, Color names the required color; an empty Color is the
// colorless sentinel, constraining the modifier to colorless spells. MatchColor
// and MatchCardType are mutually exclusive.
type CostModifier struct {
	Kind               CostModifierKind
	Controller         PlayerID
	MatchCardType      bool
	CardType           types.Card
	MatchColor         bool
	Color              color.Color
	AbilityKeyword     Keyword
	GenericIncrease    int
	GenericReduction   int
	SetGeneric         opt.V[int]
	SetManaCost        opt.V[cost.Mana]
	MinimumGeneric     int
	FirstCycleEachTurn bool

	// PerObjectReduction is a dynamic generic cost reduction scoped to the spell
	// that carries it ("This spell costs {N} less to cast for each <object>"):
	// the spell costs this many generic mana less for each battlefield permanent
	// matching CountSelection. It is set only on an AffectedSource spell cost
	// modifier; the rules layer counts the matching permanents at cost time and
	// resolves the reduction into a plain generic reduction, which never touches
	// colored requirements and never drops a cost below zero. A non-zero value
	// requires Kind CostModifierSpell.
	PerObjectReduction int
	// CountSelection bounds the battlefield permanents counted for a
	// PerObjectReduction modifier. It is meaningful only when PerObjectReduction
	// is non-zero.
	CountSelection Selection
}

// AttackTax is an additional generic mana cost to attack a player.
type AttackTax struct {
	DefendingPlayer PlayerID
	Amount          int
}

// RuleEffectKind identifies non-layer continuous rules effects such as
// prohibitions, permissions, and cost changes.
type RuleEffectKind int

// Rule effect kind values identify supported non-layer rules effects.
const (
	RuleEffectNone RuleEffectKind = iota
	RuleEffectCantGainLife
	RuleEffectCantAttack
	RuleEffectCantBlock
	RuleEffectCostModifier
	RuleEffectCastFromZone
	RuleEffectCantBeCountered
	RuleEffectCantBeBlocked
	RuleEffectMustBeBlocked
	RuleEffectMustAttack
	RuleEffectGrantHandCardAbility
	RuleEffectDoesntUntap
	RuleEffectCantBeBlockedByMoreThanOne
	// RuleEffectNoMaximumHandSize removes the maximum hand size of the affected
	// player ("You have no maximum hand size."), so that player never discards
	// down to a hand-size limit during their cleanup step (CR 402.2).
	RuleEffectNoMaximumHandSize
	// RuleEffectCantBeBlockedByCreaturesWith is a restricted block prohibition:
	// the affected attacker can't be blocked by creatures matching the carried
	// BlockerRestriction ("can't be blocked by creatures with flying", "... with
	// power N or less", "... with power N or greater"). Unlike
	// RuleEffectCantBeBlocked it does not prohibit all blockers.
	RuleEffectCantBeBlockedByCreaturesWith
)

// BlockerRestrictionKind identifies the blocker characteristic that a restricted
// "can't be blocked by creatures with ..." prohibition stops.
type BlockerRestrictionKind int

// Blocker restriction kind values identify the supported blocker characteristics.
const (
	BlockerRestrictionNone BlockerRestrictionKind = iota
	BlockerRestrictionFlying
	BlockerRestrictionPowerLessOrEqual
	BlockerRestrictionPowerGreaterOrEqual
	// BlockerRestrictionColor stops blockers of the BlockerRestriction's Color
	// ("can't be blocked by white creatures").
	BlockerRestrictionColor
	// BlockerRestrictionArtifact stops artifact-creature blockers ("can't be
	// blocked by artifact creatures").
	BlockerRestrictionArtifact
)

// BlockerRestriction bounds which blockers a restricted block prohibition stops.
// Power is the threshold for the power-comparison kinds; Color names the stopped
// blocker color for BlockerRestrictionColor. Both are unused for kinds that do
// not need them.
type BlockerRestriction struct {
	Kind  BlockerRestrictionKind
	Power int
	Color color.Color
}

// RuleEffect models static or runtime effects that change game rules rather
// than permanent characteristics. mtg/rules owns matching and application.
type RuleEffect struct {
	ID               id.ID
	Kind             RuleEffectKind
	Controller       PlayerID
	SourceObjectID   id.ID
	SourceCardID     id.ID
	AffectedObjectID id.ID
	AffectedSource   bool
	AffectedAttached bool
	Duration         EffectDuration
	CreatedTurn      int

	AffectedPlayer     PlayerRelation
	AffectedController ControllerRelation
	PermanentTypes     []types.Card
	SpellTypes         []types.Card
	DefendingPlayer    PlayerRelation

	BlockerRestriction BlockerRestriction

	CostModifier CostModifier

	CardSelection  Selection
	GrantedAbility ActivatedAbility

	CastFromZone   zone.Type
	AffectedCardID id.ID
	CastFace       opt.V[FaceIndex]
	ExpiresFor     PlayerID
}
