package game

import "github.com/natefinch/council4/mtg/game/id"

// CostModifierKind identifies which costs a modifier applies to.
type CostModifierKind int

const (
	CostModifierSpell CostModifierKind = iota
	CostModifierAbility
	CostModifierAttack
)

// CostModifier is a generic-cost increase/reduction/set effect.
type CostModifier struct {
	Kind             CostModifierKind
	Controller       PlayerID
	MatchCardType    bool
	CardType         CardType
	GenericIncrease  int
	GenericReduction int
	SetGeneric       *int
	MinimumGeneric   int
}

// AttackTax is an additional generic mana cost to attack a player.
type AttackTax struct {
	DefendingPlayer PlayerID
	Amount          int
}

// RuleEffectKind identifies non-layer continuous rules effects such as
// prohibitions, permissions, and cost changes.
type RuleEffectKind int

const (
	RuleEffectNone RuleEffectKind = iota
	RuleEffectCantGainLife
	RuleEffectCantAttack
	RuleEffectCantBlock
	RuleEffectCostModifier
	RuleEffectCastFromZone
)

// RuleEffect models static or runtime effects that change game rules rather
// than permanent characteristics. mtg/rules owns matching and application.
type RuleEffect struct {
	ID             id.ID
	Kind           RuleEffectKind
	Controller     PlayerID
	SourceObjectID id.ID
	SourceCardID   id.ID
	Duration       EffectDuration
	CreatedTurn    int

	AffectedPlayer     PlayerRelation
	AffectedController ControllerRelation
	PermanentTypes     []CardType
	DefendingPlayer    PlayerRelation

	CostModifier CostModifier

	CastFromZone ZoneType
}
