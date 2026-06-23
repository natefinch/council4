package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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

	// LifePayableTaxInstances, when positive on a CostModifierSpell, adds that
	// many "{2} or 2 life" generic Phyrexian symbols to the spell's cost instead
	// of plain generic mana. The rules layer sets it for the command-zone
	// commander tax of a spell whose static lets the caster pay 2 life rather
	// than each {2} of that tax (Liesa, Shroud of Dusk).
	LifePayableTaxInstances int

	MatchCardType bool
	MatchColor    bool
	// MatchExcludedCardType narrows a spell cost modifier to spells that do NOT
	// have ExcludedCardType ("Noncreature spells your opponents cast cost {2}
	// more to cast ...", Elspeth Conquers Death chapter II). It is meaningful
	// only on a CostModifierSpell and is mutually exclusive with the positive
	// card-type match (MatchCardType).
	MatchExcludedCardType bool
	ExcludedCardType      types.Card
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

	// SourceZone constrains a spell cost modifier to spells being cast from a
	// single zone ("Spells you cast from your graveyard cost {N} less to cast.",
	// Gravebreaker Lamia, Patrician Geist): the modifier applies only when the
	// spell is cast from this zone. When the option is absent the modifier
	// applies to spells cast from any zone. It is meaningful only on a
	// CostModifierSpell and combines with the card-type, color, and subtype
	// filters.
	SourceZone opt.V[zone.Type]

	// MinPower constrains a spell cost modifier to spells whose base printed
	// power is at least this threshold ("Creature spells you cast with power 4
	// or greater cost {2} less to cast.", Goreclaw): the modifier applies only
	// when the spell card has a numeric printed power greater than or equal to
	// the threshold. A spell with no printed power, or a star (*) power, never
	// satisfies the threshold. When the option is absent the modifier applies
	// regardless of power. It is meaningful only on a CostModifierSpell and
	// combines with the card-type, color, subtype, and zone filters.
	MinPower opt.V[int]

	// TargetsSource constrains a spell cost modifier to spells that target the
	// permanent whose static ability carries the modifier ("Spells your
	// opponents cast that target this creature cost {2} more to cast.", Boreal
	// Elemental; "Spells you cast that target this creature cost {2} less to
	// cast.", Elderwood Scion). The rules layer applies the modifier only when
	// one of the casting spell's chosen targets is exactly the source
	// permanent. It is meaningful only on a CostModifierSpell and combines with
	// the affected-player caster filter (PlayerOpponent for the defensive tax,
	// PlayerYou for the controller's discount).
	TargetsSource bool
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
	// RuleEffectGrantGraveyardCardKeyword grants GrantedKeyword to the affected
	// player's graveyard cards that match CardSelection ("[During your turn,]
	// nonland permanent cards in your graveyard have retrace.", Six, Wrenn and
	// Six Emblem). RestrictedDuringControllerTurn scopes the grant to the
	// controller's turn.
	RuleEffectGrantGraveyardCardKeyword
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
	// non-empty only spells of those card types are forbidden ("can't cast
	// creature spells"); when ExcludedSpellTypes is non-empty spells carrying any
	// of those card types are exempt ("can't cast noncreature spells"). An empty
	// SpellTypes and ExcludedSpellTypes forbids every spell.
	// RestrictedDuringControllerTurn scopes the prohibition to the source
	// controller's turn ("During your turn, ...").
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
	// RuleEffectPayLifeForColoredMana lets the affected player pay 2 life rather
	// than the mana for each ManaColor symbol in any cost they pay ("For each {B}
	// in a cost, you may pay 2 life rather than pay that mana.", K'rrik, Son of
	// Yawgmoth). It makes every colored mana symbol of ManaColor in the affected
	// player's costs payable like a Phyrexian symbol of that color.
	RuleEffectPayLifeForColoredMana
	// RuleEffectPayLifeForCommanderTax lets the caster pay 2 life rather than
	// each {2} of the command-zone commander tax when casting the source card
	// itself ("Rather than pay {2} for each previous time you've cast this spell
	// from the command zone this game, pay 2 life that many times.", Liesa,
	// Shroud of Dusk). It is a self-scoped static read from the card being cast;
	// the rules layer emits the commander tax as generic Phyrexian symbols so
	// each {2} instance is independently payable with 2 life.
	RuleEffectPayLifeForCommanderTax
	// RuleEffectDrawLimitPerTurn caps the number of cards the affected players
	// (AffectedPlayer) may draw each turn at DrawLimitPerTurn ("Each opponent
	// can't draw more than one card each turn.", Narset, Parter of Veils, Leovold;
	// "Each player can't draw more than one card each turn.", Spirit of the
	// Labyrinth). It is a continuous draw restriction (CR 120.3): once an affected
	// player has drawn DrawLimitPerTurn cards this turn, each further draw is
	// replaced by drawing nothing.
	RuleEffectDrawLimitPerTurn
	// RuleEffectCastLimitPerTurn caps the number of spells the affected players
	// (AffectedPlayer) may cast each turn at CastLimitPerTurn ("Each player can't
	// cast more than one spell each turn.", Rule of Law, Eidolon of Rhetoric,
	// Arcane Laboratory; "You can't cast more than one spell each turn.",
	// Moderation). It is a continuous casting restriction: once an affected player
	// has cast CastLimitPerTurn spells this turn, they can't begin to cast
	// another. SpellTypes and ExcludedSpellTypes optionally narrow the cap to
	// spells of, or other than, those card types; empty filters count every spell.
	RuleEffectCastLimitPerTurn
	// RuleEffectAdditionalTriggerForControlledPermanent makes a triggered ability
	// of a permanent controlled by this effect's controller trigger one
	// additional time when that permanent matches AffectedSelection ("If a
	// triggered ability of a legendary creature you control triggers, that
	// ability triggers an additional time.", Annie Joins Up; "... of an Ally you
	// control ...", Katara, the Fearless; "... of a Ninja creature you control
	// ...", Splinter, Radical Rat). AffectedSelection carries the source
	// permanent's type, supertype, and subtype filter; unlike the chosen-type and
	// entering-permanent doublers it includes the source object itself ("a ... you
	// control", not "another"). An empty AffectedSelection matches any controlled
	// permanent.
	RuleEffectAdditionalTriggerForControlledPermanent
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
		RuleEffectGrantGraveyardCardKeyword,
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
		RuleEffectLookAtTopCardAnyTime,
		RuleEffectPayLifeForColoredMana,
		RuleEffectPayLifeForCommanderTax,
		RuleEffectDrawLimitPerTurn,
		RuleEffectCastLimitPerTurn,
		RuleEffectAdditionalTriggerForControlledPermanent:
		return true
	default:
		return false
	}
}

// manaColorValid reports whether c is one of the five real colors of mana
// (colorless is rejected), backing RuleEffectPayLifeForColoredMana validation.
func manaColorValid(c mana.Color) bool {
	switch c {
	case mana.W, mana.U, mana.B, mana.R, mana.G:
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
	// ExcludedSpellTypes exempts spells carrying any of these card types from a
	// RuleEffectCantCastSpells prohibition, expressing the "noncreature spells"
	// family ("Your opponents can't cast noncreature spells this turn.").
	ExcludedSpellTypes []types.Card
	DefendingPlayer    PlayerRelation

	BlockerRestriction BlockerRestriction
	Protection         ProtectionKeyword
	AttackTaxGeneric   int

	// ManaColor names the colored mana symbol a RuleEffectPayLifeForColoredMana
	// effect lets the affected player pay 2 life for instead ("For each {B} in a
	// cost, ...", K'rrik). It is unused for every other kind.
	ManaColor mana.Color

	CostModifier CostModifier

	CardSelection  Selection
	GrantedAbility ActivatedAbility
	// GrantedKeyword is the keyword a RuleEffectGrantGraveyardCardKeyword effect
	// confers on the affected player's matching graveyard cards. It is unused for
	// every other kind.
	GrantedKeyword Keyword

	CastFromZone   zone.Type
	AffectedCardID id.ID
	CastFace       opt.V[FaceIndex]
	ExpiresFor     PlayerID

	// AdditionalLandPlays is the number of extra land plays granted by a
	// RuleEffectAdditionalLandPlays effect. It is unused for every other kind.
	AdditionalLandPlays int

	// DrawLimitPerTurn caps how many cards the affected players may draw each turn
	// for a RuleEffectDrawLimitPerTurn effect ("Each opponent can't draw more than
	// one card each turn."). It is a positive count and unused for every other
	// kind.
	DrawLimitPerTurn int

	// CastLimitPerTurn caps how many spells the affected players may cast each
	// turn for a RuleEffectCastLimitPerTurn effect ("Each player can't cast more
	// than one spell each turn."). It is a positive count and unused for every
	// other kind.
	CastLimitPerTurn int

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

	// SpellColorless additionally permits a RuleEffectCastSpellsFromZone permission
	// to cast colorless spells, so a spell qualifies when it matches the SpellTypes
	// filter or is colorless ("artifact spells and colorless spells from the top of
	// your library.", Mystic Forge). It is unused for every other kind.
	SpellColorless bool

	// SpellChosenSubtypeFrom narrows a RuleEffectCastSpellsFromZone permission to
	// spells sharing the creature subtype the source permanent chose as it entered,
	// reading the choice stored under this key in the source's EntryChoices
	// ("creature spells of the chosen type from the top of your library.",
	// Realmwalker). The empty key applies no chosen-type filter. It is unused for
	// every other kind.
	SpellChosenSubtypeFrom ChoiceKey

	// PayLifeEqualToManaValue makes spells cast under a
	// RuleEffectCastSpellsFromZone permission cost life equal to the cast spell's
	// mana value instead of its mana cost ("If you cast a spell this way, pay life
	// equal to its mana value rather than pay its mana cost.", Bolas's Citadel,
	// Gwenom, Remorseless). It is unused for every other kind.
	PayLifeEqualToManaValue bool

	// AffectedSelection optionally narrows the permanents a group-scoped rule
	// effect applies to beyond the AffectedController and PermanentTypes filters,
	// expressing the filtered controlled-creature mass statics ("Blue creatures
	// you control can't be blocked.", "Creatures you control with +1/+1 counters
	// on them can't be blocked."). An empty Selection imposes no extra filter.
	AffectedSelection Selection

	// BlockedSelection scopes a RuleEffectCantBlock restriction to a protected
	// group of attackers the affected blockers can't block ("Creatures with power
	// less than this creature's power can't block creatures you control."). A
	// non-empty Selection makes the prohibition conditional: an affected blocker
	// may still block attackers outside the selection. The selection is matched
	// from the effect's controller, so its controller relation resolves "you" to
	// the source controller. An empty Selection with BlockedSource false leaves the
	// can't-block prohibition unconditional.
	BlockedSelection Selection

	// BlockedSource scopes a RuleEffectCantBlock restriction to the effect's
	// source permanent ("Creatures with power less than this creature's power
	// can't block it."). When true the affected blockers can't block only the
	// source object; they may still block any other attacker. It is mutually
	// exclusive with a non-empty BlockedSelection.
	BlockedSource bool
}
