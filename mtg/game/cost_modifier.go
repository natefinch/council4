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
// may combine with MatchCardType ("black creature spells"). MatchSubtypes and
// MatchColors are each mutually exclusive with MatchCardType.
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

	// MatchColors constrains a spell cost modifier to spells carrying any one of
	// these colors ("... that's red or green ..."): the modifier applies when
	// the spell has at least one of the listed colors. It holds two or more real
	// colors and is mutually exclusive with MatchColor and MatchCardType.
	MatchColors []color.Color

	// MatchSubtypes constrains a spell cost modifier to spells carrying any one
	// of these subtypes ("Aura and Equipment spells ..."): the modifier applies
	// when the spell has at least one of the listed subtypes. It may combine with
	// MatchColor and is mutually exclusive with MatchCardType and MatchColors.
	MatchSubtypes []types.Sub
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
	// RuleEffectCastSpellsAsThoughFlash lets the affected player cast spells as
	// though they had flash, i.e. at instant speed ("You may cast spells this
	// turn as though they had flash.", Borne Upon a Wind, Emergence Zone; "You
	// may cast spells as though they had flash.", Vedalken Orrery, Leyline of
	// Anticipation; CR 702.8 / 601.3e). It is a timing permission only and does
	// not bypass other casting restrictions. SpellTypes and SpellSubtypes
	// optionally narrow the grant to spells of those card types ("sorcery
	// spells", Hypersonic Dragon) or subtypes ("Aura and Equipment spells",
	// Sigarda's Aid); empty filters permit every spell.
	RuleEffectCastSpellsAsThoughFlash
	// RuleEffectPlayLandsFromZone grants the affected player a continuous
	// permission to play land cards from CastFromZone ("You may play lands from
	// your graveyard.", Ramunap Excavator, Crucible of Worlds; "You may play lands
	// from the top of your library.", Oracle of Mul Daya, Courser of Kruphix).
	// PermanentTypes carries the played card's required type (Land). Unlike
	// RuleEffectPlayFromZone it is a continuous static keyed on the zone and type
	// rather than a single AffectedCardID, so it applies to every matching card in
	// that zone. TopCardOnly restricts the permission to the top card of the source
	// zone (the top of the affected player's library).
	RuleEffectPlayLandsFromZone
	// RuleEffectPlayWithTopCardRevealed makes the affected player play with the top
	// card of their library revealed to all players ("Play with the top card of
	// your library revealed.", Oracle of Mul Daya, Courser of Kruphix, Future
	// Sight). It is a visibility static and grants no play permission on its own.
	RuleEffectPlayWithTopCardRevealed
	// RuleEffectCastSpellsFromZone grants the affected player a continuous
	// permission to cast spells from CastFromZone ("You may cast spells from the
	// top of your library.", Bolas's Citadel, Future Sight). SpellTypes filters the
	// castable spells by card type (any one of the listed types); an empty
	// SpellTypes permits casting any spell. Like RuleEffectPlayLandsFromZone it is
	// a continuous static keyed on the zone and type rather than a single
	// AffectedCardID, so it applies to every matching card in that zone.
	// TopCardOnly restricts the permission to the top card of the source zone (the
	// top of the affected player's library).
	RuleEffectCastSpellsFromZone
	// RuleEffectCantCastFromZones forbids the affected players (AffectedPlayer)
	// from casting spells from any of the zones in CantCastFromZones ("Your
	// opponents can't cast spells from anywhere other than their hands." expands
	// to the non-hand cast zones; "Players can't cast spells from graveyards or
	// libraries."). A "can't" restriction overrides any casting permission, so a
	// matching source zone makes the cast illegal regardless of other effects.
	RuleEffectCantCastFromZones
	// RuleEffectCantEnterFromZones forbids cards from entering the battlefield
	// out of any of the zones in EnterFromZones ("Creature cards in graveyards
	// and libraries can't enter the battlefield.", Grafdigger's Cage; "Permanent
	// cards in graveyards can't enter the battlefield.", Soulless Jailer). The
	// restriction is global (it affects every player). PermanentTypes filters the
	// affected entering cards by card type (any one of the listed types); an empty
	// PermanentTypes restricts every permanent card. EnterExcludeLandCards exempts
	// land cards, expressing the "nonland permanent" forms.
	RuleEffectCantEnterFromZones
	// RuleEffectLookAtTopCardAnyTime lets the affected player look at the top card
	// of their library at any time ("You may look at the top card of your library
	// any time.", Bolas's Citadel, Vizier of the Menagerie, Sphinx of Jwar Isle).
	// It is a private-visibility static (only the affected player sees the card)
	// and grants no play or cast permission on its own.
	RuleEffectLookAtTopCardAnyTime
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
		RuleEffectUntapDuringOtherPlayersUntapStep,
		RuleEffectCastSpellsAsThoughFlash,
		RuleEffectPlayLandsFromZone,
		RuleEffectPlayWithTopCardRevealed,
		RuleEffectCastSpellsFromZone,
		RuleEffectCantCastFromZones,
		RuleEffectCantEnterFromZones,
		RuleEffectLookAtTopCardAnyTime:
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

	// AppliesToNextSpellOnly limits a RuleEffectCantBeCountered effect to the
	// single next spell its controller casts ("The next spell you cast this turn
	// can't be countered.", Mistrise Village). When the controller casts a
	// matching spell, the effect attaches to that spell and is consumed, so later
	// spells are unaffected. When false the effect applies to every matching
	// spell for its duration ("Spells you cast this turn can't be countered.").
	AppliesToNextSpellOnly bool

	// TopCardOnly restricts a RuleEffectPlayLandsFromZone permission to the top
	// card of CastFromZone, i.e. the top of the affected player's library ("You
	// may play lands from the top of your library."). It is unused for every other
	// kind and for zones without a meaningful top card.
	TopCardOnly bool

	// CantCastFromZones lists the zones a RuleEffectCantCastFromZones restriction
	// forbids the affected players from casting spells out of. It is unused for
	// every other kind.
	CantCastFromZones []zone.Type

	// SpellSubtypes optionally narrows a RuleEffectCastSpellsAsThoughFlash grant
	// to spells carrying any one of these subtypes ("Aura and Equipment spells
	// ...", Sigarda's Aid). An empty list applies no subtype filter. It is unused
	// for every other kind.
	SpellSubtypes []types.Sub

	// EnterFromZones lists the zones a RuleEffectCantEnterFromZones restriction
	// forbids cards from entering the battlefield out of. It is unused for every
	// other kind.
	EnterFromZones []zone.Type

	// EnterExcludeLandCards exempts land cards from a RuleEffectCantEnterFromZones
	// restriction, expressing the "nonland permanent cards" forms (Weathered
	// Runestone). It is unused for every other kind.
	EnterExcludeLandCards bool
}
