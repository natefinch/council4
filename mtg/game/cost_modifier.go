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
	Kind             CostModifierKind
	Controller       PlayerID
	CardType         types.Card
	Color            color.Color
	AbilityKeyword   Keyword
	GenericIncrease  int
	GenericReduction int
	SetGeneric       opt.V[int]
	SetManaCost      opt.V[cost.Mana]
	MinimumGeneric   int

	MatchCardType bool
	MatchColor    bool
	// ChosenSubtypeFromEntryChoice constrains a creature spell cost modifier to
	// spells whose subtype matches the source permanent's entry-time
	// creature-type choice (see EntryTypeChoiceKey). It is meaningful only on a
	// CostModifierSpell that matches creatures by card type.
	ChosenSubtypeFromEntryChoice bool
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
	// is non-zero. It is a pointer so CostModifier stays cheap to copy; a nil
	// pointer means the modifier counts nothing.
	CountSelection *Selection
	// DynamicReduction is a generic cost reduction whose amount is evaluated as
	// the spell is cast ("This spell costs {X} less to cast, where X is <dynamic
	// amount>", e.g. the greatest power among creatures you control, or the number
	// of a kind of permanent you control). It is set only on an AffectedSource
	// CostModifierSpell; the rules layer evaluates the dynamic amount at cost time
	// and resolves it into a plain generic reduction, which never touches colored
	// requirements and never drops a cost below zero. It is mutually exclusive
	// with PerObjectReduction. A nil pointer means the modifier has no dynamic
	// reduction; it is a pointer so CostModifier stays cheap to copy.
	DynamicReduction *DynamicAmount
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
	// RuleEffectAdditionalLandPlays raises the number of lands the affected
	// player may play during their turn by AdditionalLandPlays ("You may play an
	// additional land this turn.", "You may play two additional lands on each of
	// your turns."). It is additive with other such effects and with the
	// one-land-per-turn baseline.
	RuleEffectAdditionalLandPlays
	// RuleEffectCantCastSpells forbids the affected players (AffectedPlayer) from
	// casting spells ("Your opponents can't cast spells."). When SpellTypes is
	// non-empty only spells of those card types are forbidden; an empty SpellTypes
	// forbids every spell. RestrictedDuringControllerTurn scopes the prohibition
	// to the source controller's turn ("During your turn, ...").
	RuleEffectCantCastSpells
	// RuleEffectCantActivateAbilities forbids the affected players (AffectedPlayer)
	// from activating abilities of permanents whose card type is in PermanentTypes
	// ("... activate abilities of artifacts, creatures, or enchantments."). An
	// empty PermanentTypes forbids activating abilities of any permanent.
	// RestrictedDuringControllerTurn scopes the prohibition to the source
	// controller's turn.
	RuleEffectCantActivateAbilities
	// RuleEffectAdditionalTriggerForEnteringPermanent makes a triggered ability of
	// a permanent controlled by this effect's controller trigger one additional
	// time when an entering permanent caused it to trigger ("If an artifact or
	// creature entering causes a triggered ability of a permanent you control to
	// trigger, that ability triggers an additional time.", Panharmonicon, Yarok,
	// Ancient Greenwarden). PermanentTypes filters the entering permanent's card
	// type (any of the listed types); an empty PermanentTypes matches any
	// entering permanent.
	RuleEffectAdditionalTriggerForEnteringPermanent
	// RuleEffectUntapDuringOtherPlayersUntapStep untaps a set of the source
	// controller's permanents during every other player's untap step ("Untap all
	// permanents you control during each other player's untap step.", Seedborn
	// Muse; "Untap all creatures you control ...", Drumbellower). AffectedSource
	// scopes it to the source permanent itself ("Untap this artifact during each
	// other player's untap step.", Unwinding Clock-style self forms); otherwise
	// PermanentTypes filters the controller's permanents (empty PermanentTypes
	// untaps every permanent the controller controls). The runtime applies it
	// during the active player's untap step whenever the active player is not the
	// effect's controller, so it serves both the "each other player's" and "each
	// opponent's" wordings.
	RuleEffectUntapDuringOtherPlayersUntapStep
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
		RuleEffectAdditionalTriggerForChosenCreatureType,
		RuleEffectAdditionalLandPlays,
		RuleEffectCantCastSpells,
		RuleEffectCantActivateAbilities,
		RuleEffectAdditionalTriggerForEnteringPermanent,
		RuleEffectUntapDuringOtherPlayersUntapStep:
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

	// AdditionalLandPlays is the number of extra land plays granted by a
	// RuleEffectAdditionalLandPlays effect. It is unused for every other kind.
	AdditionalLandPlays int

	// RestrictedDuringControllerTurn scopes a RuleEffectCantCastSpells or
	// RuleEffectCantActivateAbilities prohibition to the source controller's turn
	// ("During your turn, ..."). When false the prohibition applies on every turn.
	RestrictedDuringControllerTurn bool
}
