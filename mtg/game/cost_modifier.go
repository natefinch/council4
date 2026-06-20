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
	Kind          CostModifierKind
	Controller    PlayerID
	MatchCardType bool
	CardType      types.Card
	MatchColor    bool
	Color         color.Color
	// ChosenSubtypeFromEntryChoice constrains a creature spell cost modifier to
	// spells whose subtype matches the source permanent's entry-time
	// creature-type choice (see EntryTypeChoiceKey). It is meaningful only on a
	// CostModifierSpell that matches creatures by card type.
	ChosenSubtypeFromEntryChoice bool
	AbilityKeyword               Keyword
	GenericIncrease              int
	GenericReduction             int
	SetGeneric                   opt.V[int]
	SetManaCost                  opt.V[cost.Mana]
	MinimumGeneric               int
	FirstCycleEachTurn           bool

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
	// RuleEffectPlayerProtection grants the affected player protection from
	// sources matching Protection.
	RuleEffectPlayerProtection
	// RuleEffectAttackTax adds AttackTaxGeneric generic mana to the declaration
	// cost of each creature attacking the affected player.
	RuleEffectAttackTax
	// RuleEffectLifeTotalCantChange prevents the affected player's life total
	// from increasing or decreasing, including life payments.
	RuleEffectLifeTotalCantChange
	// RuleEffectPlayFromZone permits playing a specific card from a non-hand
	// zone, including either casting it as a spell or playing it as a land.
	RuleEffectPlayFromZone
	// RuleEffectAdditionalTriggerForChosenCreatureType makes a triggered ability
	// of another creature controlled by this effect's controller trigger one
	// additional time when that creature has the subtype chosen by the source as
	// it entered.
	RuleEffectAdditionalTriggerForChosenCreatureType
)

// Valid reports whether k identifies a supported rule effect.
func (k RuleEffectKind) Valid() bool {
	switch k {
	case RuleEffectCantGainLife,
		RuleEffectCantAttack,
		RuleEffectCantBlock,
		RuleEffectCostModifier,
		RuleEffectCastFromZone,
		RuleEffectCantBeCountered,
		RuleEffectCantBeBlocked,
		RuleEffectMustBeBlocked,
		RuleEffectMustAttack,
		RuleEffectGrantHandCardAbility,
		RuleEffectDoesntUntap,
		RuleEffectCantBeBlockedByMoreThanOne,
		RuleEffectNoMaximumHandSize,
		RuleEffectCantBeBlockedByCreaturesWith,
		RuleEffectPlayerProtection,
		RuleEffectAttackTax,
		RuleEffectLifeTotalCantChange,
		RuleEffectPlayFromZone,
		RuleEffectAdditionalTriggerForChosenCreatureType:
		return true
	default:
		return false
	}
}

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
	Protection         ProtectionKeyword
	AttackTaxGeneric   int

	CostModifier CostModifier

	CardSelection  Selection
	GrantedAbility ActivatedAbility

	CastFromZone   zone.Type
	AffectedCardID id.ID
	CastFace       opt.V[FaceIndex]
	ExpiresFor     PlayerID
}
