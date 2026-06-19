package game

import (
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
type CostModifier struct {
	Kind               CostModifierKind
	Controller         PlayerID
	MatchCardType      bool
	CardType           types.Card
	AbilityKeyword     Keyword
	GenericIncrease    int
	GenericReduction   int
	SetGeneric         opt.V[int]
	SetManaCost        opt.V[cost.Mana]
	MinimumGeneric     int
	FirstCycleEachTurn bool
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
)

// BlockerRestriction bounds which blockers a restricted block prohibition stops.
// Power is the threshold for the power-comparison kinds and is unused for the
// flying kind.
type BlockerRestriction struct {
	Kind  BlockerRestrictionKind
	Power int
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
