package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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
// CardSelection is the sole card-subject filter for a spell cost modifier: it
// names the spells the modifier affects by card type, excluded card type, color,
// color disjunction, subtype, and minimum power, matched through the shared
// card-subject Selection matcher. An empty CardSelection affects every spell the
// caster filter admits.
type CostModifier struct {
	Kind             CostModifierKind
	Controller       PlayerID
	AbilityKeyword   Keyword
	GenericIncrease  int
	GenericReduction int

	// ColoredIncrease lists colored mana symbols added to the affected cost on
	// top of any GenericIncrease ("Black spells you cast cost {B} more to
	// cast.", Derelor; the mono-color Leech cycle). Each entry is one basic
	// colored mana symbol (W, U, B, R, or G); the rules layer appends one such
	// symbol to the cost for each entry, raising a colored requirement the
	// caster must satisfy. It is meaningful only on a CostModifierSpell. An
	// empty slice adds no colored mana. It is a slice so CostModifier stays a
	// plain value; nil and empty are equivalent.
	ColoredIncrease []mana.Color
	SetGeneric      opt.V[int]
	SetManaCost     opt.V[cost.Mana]
	MinimumGeneric  int

	// LifeIncrease, when positive on a CostModifierSpell, adds an additional
	// "pay N life" cost to each affected spell ("Spells your opponents cast that
	// target this creature cost an additional 3 life to cast.", Terror of the
	// Peaks). Unlike GenericIncrease, it is paid in life rather than mana: the
	// rules layer appends a pay-life additional cost of this many life to the
	// spell, so the caster must have at least that much life to cast. It is
	// meaningful only on a CostModifierSpell and combines with the
	// affected-player caster filter and the TargetsSource predicate. A zero
	// value adds no life tax.
	LifeIncrease int

	// LifePayableTaxInstances, when positive on a CostModifierSpell, adds that
	// many "{2} or 2 life" generic Phyrexian symbols to the spell's cost instead
	// of plain generic mana. The rules layer sets it for the command-zone
	// commander tax of a spell whose static lets the caster pay 2 life rather
	// than each {2} of that tax (Liesa, Shroud of Dusk).
	LifePayableTaxInstances int

	// ChosenSubtypeFromEntryChoice constrains a creature spell cost modifier to
	// spells whose subtype matches the source permanent's entry-time
	// creature-type choice (see EntryTypeChoiceKey). It is meaningful only on a
	// CostModifierSpell whose CardSelection matches creatures by card type.
	ChosenSubtypeFromEntryChoice  bool
	ChosenCardTypeFromEntryChoice bool
	FirstCycleEachTurn            bool

	// PerObjectReduction is a dynamic generic cost reduction scoped to spells
	// ("This spell costs {N} less to cast for each <object>"; "[<filter>] spells
	// you cast cost {N} less to cast for each <permanent> you control"): the
	// affected spell costs this many generic mana less for each battlefield
	// permanent matching CountSelection. It is set on an AffectedSource spell
	// cost modifier (the SELF reduction) and on a controller-scoped group spell
	// modifier supplied by another permanent's static; in both cases the rules
	// layer counts the matching permanents at cost time and resolves the
	// reduction into a plain generic reduction, which never touches colored
	// requirements and never drops a cost below zero. A non-zero value requires
	// Kind CostModifierSpell.
	PerObjectReduction int
	// CountSelection bounds the battlefield permanents counted for a
	// PerObjectReduction modifier. It is meaningful only when PerObjectReduction
	// is non-zero. It is a pointer so CostModifier stays cheap to copy; a nil
	// pointer means the modifier counts nothing.
	CountSelection *Selection
	// CountZone, when present, scopes a PerObjectReduction count to the caster's
	// cards in this zone matching CountSelection rather than to battlefield
	// permanents ("This spell costs {N} less to cast for each <card> in your
	// graveyard."). It names a real card zone the caster owns (graveyard or
	// hand). It is meaningful only when PerObjectReduction is non-zero on a
	// CostModifierSpell; when absent the count is over battlefield permanents.
	CountZone opt.V[zone.Type]
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

	// ReductionCondition gates a flat GenericReduction scoped to the spell that
	// carries it ("This spell costs {N} less to cast if <condition>.", Wizard's
	// Lightning, Squash, Draconic Lore): the spell costs GenericReduction generic
	// mana less only while the condition holds. It is set only on an
	// AffectedSource CostModifierSpell; the rules layer evaluates the board- and
	// player-state condition against the caster as the spell is cast and applies
	// the reduction, which never touches colored requirements and never drops a
	// cost below zero, only when it is satisfied. It is mutually exclusive with
	// PerObjectReduction and DynamicReduction. An absent option leaves the
	// reduction ungated.
	ReductionCondition opt.V[Condition]
	// TargetsTappedCreature gates a source spell's flat reduction on at least one
	// chosen target being a tapped creature.
	TargetsTappedCreature bool

	// SourceZone constrains a spell cost modifier to spells being cast from a
	// single zone ("Spells you cast from your graveyard cost {N} less to cast.",
	// Gravebreaker Lamia, Patrician Geist): the modifier applies only when the
	// spell is cast from this zone. When the option is absent the modifier
	// applies to spells cast from any zone. It is meaningful only on a
	// CostModifierSpell and combines with the CardSelection card filter.
	SourceZone opt.V[zone.Type]

	// SourceZones generalizes SourceZone to a set: when non-empty it constrains a
	// spell cost modifier to spells being cast from any one of the listed zones
	// ("Spells you cast from anywhere other than your hand cost {N} less to
	// cast.", Sage of the Beyond, which expands to the non-hand cast zones
	// graveyard, exile, library, and command). The modifier applies only when the
	// spell is cast from one of these zones. It is mutually exclusive with the
	// single-zone SourceZone option and, like it, is meaningful only on a
	// CostModifierSpell and combines with the CardSelection card filter. It is a
	// slice so CostModifier stays a plain value; nil and empty are equivalent and
	// impose no zone filter.
	SourceZones []zone.Type

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

	// SharedExiledCardTypeReduction, when positive on a CostModifierSpell,
	// reduces the affected spell's generic cost by this much for each card type
	// the spell shares with the cards exiled with the source permanent ("Spells
	// you cast cost {N} less to cast for each card type they share with cards
	// exiled with this creature.", Cemetery Prowler). The rules layer reads the
	// source permanent's linked-exile set named by ExiledLinkKey, takes the
	// distinct card types among those exiled cards, intersects them with the
	// casting spell's card types, and resolves the shared count times this
	// amount into a plain generic reduction, which never touches colored
	// requirements and never drops a cost below zero. It is meaningful only on a
	// CostModifierSpell supplied by another permanent's static (not
	// AffectedSource) and pairs with a non-empty ExiledLinkKey.
	SharedExiledCardTypeReduction int
	// SharedExiledCardTypeReductionOnce applies the reduction once when at least
	// one type is shared, rather than once per shared type.
	SharedExiledCardTypeReductionOnce bool
	// ExiledLinkKey names the linked-exile set whose exiled cards a
	// SharedExiledCardTypeReduction reads, keyed by the source permanent's card
	// identity. It is meaningful only when that reduction is positive.
	ExiledLinkKey LinkedKey
	// ExiledLinkObjectScoped reads ExiledLinkKey under the source permanent's
	// current object identity, so a new incarnation does not inherit an imprint.
	ExiledLinkObjectScoped bool

	// CardSelection is the canonical card-subject filter for a spell cost
	// modifier, mirroring how triggers and additional costs describe the card
	// they match: card type, excluded card type, color, color disjunction,
	// subtype, and minimum power. The rules layer matches it through the shared
	// card-subject Selection matcher. An empty CardSelection affects every spell
	// the caster filter admits. Cost-specific context (source zone, target
	// relation, entry-choice subtype, linked-exile and dynamic reductions) stays
	// on its own fields, outside the Selection.
	CardSelection Selection
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
	// RuleEffectCantBeBlockedExceptBy is the complementary restricted block
	// prohibition "can't be blocked except by ...": the affected attacker can be
	// blocked only by creatures matching the carried BlockerRestriction ("can't
	// be blocked except by creatures with flying", "... except by black
	// creatures", "... except by artifact creatures", "... except by creatures
	// with defender", "... except by legendary creatures"). Every blocker that
	// does not match the restriction is prohibited; matching blockers are
	// allowed. Unlike RuleEffectCantBeBlockedByCreaturesWith, which stops only
	// matching blockers, this stops all non-matching ones.
	RuleEffectCantBeBlockedExceptBy
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
	// permanent's type, supertype, and subtype filter. A plain "a ... you
	// control" filter includes the source object itself; an "another ... you
	// control" filter sets ExcludeSource so the doubler does not double its own
	// triggers (Twinflame Travelers), and AnyOf models an "or"-joined filter
	// whose branches may differ in self-exclusion ("a Shaman or another Wizard
	// you control", Harmonic Prodigy). An empty AffectedSelection matches any
	// controlled permanent.
	RuleEffectAdditionalTriggerForControlledPermanent
	// RuleEffectMustBeBlockedByAllAble is the true-lure requirement: every
	// creature able to block the affected attacker must do so ("All creatures
	// able to block this creature do so.", Taunting Elf; "All creatures able to
	// block enchanted creature do so.", Lure; "... equipped creature ...",
	// Nemesis Mask). Unlike RuleEffectMustBeBlocked, which only forces at least
	// one blocker, this forces every able blocker onto the attacker (CR 509.1c).
	// AffectedSource scopes it to the source creature; AffectedAttached scopes it
	// to the creature an Aura or Equipment is attached to.
	RuleEffectMustBeBlockedByAllAble
	// RuleEffectAssignCombatDamageAsThoughUnblocked lets the affected attacker
	// assign its combat damage to the player, planeswalker, or battle it is
	// attacking as though it weren't blocked ("You may have this creature assign
	// its combat damage as though it weren't blocked.", Lone Wolf, Thorn
	// Elemental, Rhox). While blocked, the attacker still takes its blockers'
	// damage, but it deals its own combat damage to its attack target rather than
	// to the blockers. AffectedSource scopes it to the source creature;
	// AffectedAttached scopes it to the attached creature.
	RuleEffectAssignCombatDamageAsThoughUnblocked
	// RuleEffectCantTransform prevents the affected permanents from transforming
	// ("Non-Human Werewolves you control can't transform.", Immerwolf). Like the
	// other group prohibitions it scopes the affected permanents with
	// AffectedController, PermanentTypes, and AffectedSelection (or AffectedSource
	// / AffectedAttached for a self- or attached-scoped form). A matching
	// permanent's transform is prevented (CR 701.28), so any attempt to transform
	// it does nothing.
	RuleEffectCantTransform
	// RuleEffectSuppressOpponentEnteringTriggers prevents a permanent entering the
	// battlefield from causing triggered abilities of permanents controlled by
	// the effect controller's opponents to trigger ("Permanents entering don't
	// cause abilities of permanents your opponents control to trigger.", Elesh
	// Norn, Mother of Machines). A pending triggered ability is suppressed when
	// its triggering event is a permanent entering the battlefield and its source
	// permanent is controlled by a player the effect controller treats as an
	// opponent (CR 614 / the entering-trigger interaction). The effect is global;
	// it carries no filters.
	RuleEffectSuppressOpponentEnteringTriggers
	// RuleEffectAttackTaxPerCreature taxes each creature attacking the effect's
	// controller a per-attacker generic cost. When
	// AttackTaxIncludesPlaneswalkers is set the protection also covers attacks on
	// any planeswalker that controller controls ("Creatures can't attack you or
	// planeswalkers you control ...", Baird, Archon of Absolution, Sphere of
	// Safety); otherwise it covers only direct attacks on the controller
	// ("Creatures can't attack you ...", Collective Restraint). Exactly one
	// per-attacker amount source is set:
	//   - AttackTaxGeneric, a fixed generic value ("... pays {1} for each of those
	//     creatures.", Baird, Archon of Absolution);
	//   - CardSelection, the number of permanents the controller controls matching
	//     it ("... where X is the number of enchantments you control.", Sphere of
	//     Safety);
	//   - AttackTaxScaledAmount, a board-derived aggregate ("... where X is the
	//     number of basic land types among lands you control.", Collective
	//     Restraint, domain).
	// Unlike the fixed-generic RuleEffectAttackTax (Propaganda), the amount is
	// evaluated from the battlefield as attackers are declared. AffectedPlayer
	// scopes the protected defending player to the controller.
	RuleEffectAttackTaxPerCreature
	// RuleEffectManaProductionMultiplier multiplies the mana produced whenever the
	// effect's controller taps a permanent for mana, scaling each such production
	// by ManaProductionMultiplier ("If you tap a permanent for mana, it produces
	// twice as much of that mana instead.", Mana Reflection, factor 2; "... three
	// times as much ...", Nyxbloom Ancient, factor 3). It is a controller-scoped
	// mana-production replacement (CR 605 / 106): it applies only when a permanent
	// the controller controls is tapped to produce that mana, so untap-cost and
	// other non-tap mana sources are unaffected. Multiple such effects compound
	// multiplicatively. It carries no filters beyond the factor.
	RuleEffectManaProductionMultiplier
	// RuleEffectSkipDrawStep makes the affected player skip their draw step
	// ("Skip your draw step.", Necropotence, Yawgmoth's Bargain). While the
	// effect applies, the affected player's draw step does not happen at all: no
	// beginning-of-draw-step triggers fire, no turn-based draw occurs, and no
	// priority is given during that step (CR 500.8, CR 504). It is a
	// controller-scoped turn-structure static carrying no payload beyond the
	// affected player.
	RuleEffectSkipDrawStep
	// RuleEffectCanBlockOnlyCreaturesWith is the blocker-side permission
	// restriction "can block only creatures with flying" (Cloud Sprite,
	// Gloomwidow): the affected creature may block an attacker only when that
	// attacker matches the carried BlockerRestriction (currently flying). Unlike
	// RuleEffectCantBeBlockedByCreaturesWith, the restriction characteristic
	// describes the attacker being blocked rather than the affected creature's own
	// blockers.
	RuleEffectCanBlockOnlyCreaturesWith
	// RuleEffectCantAttackAlone prohibits the affected creature from attacking
	// unless at least one other creature also attacks that combat ("This creature
	// can't attack alone.", Mogg Flunkies, Trusty Companion). A lone attack that
	// would declare only this creature is illegal; declaring it alongside another
	// attacker satisfies the restriction (CR 508.1a).
	RuleEffectCantAttackAlone
	// RuleEffectCantBlockAlone prohibits the affected creature from blocking
	// unless at least one other creature also blocks that combat ("This creature
	// can't block alone.", Craven Hulk). A block declaration that would leave this
	// creature as the only blocker is illegal; blocking alongside another blocker
	// satisfies the restriction (CR 509.1a).
	RuleEffectCantBlockAlone
	// RuleEffectCanAttackAsThoughDefender permits the affected creature to attack
	// even though it has defender ("This creature can attack this turn as though
	// it didn't have defender.", Krotiq Nestguard, Skyclave Squid, Returned
	// Phalanx). It is a combat permission scoped to the source creature
	// (AffectedSource); while it applies, the defender keyword no longer prevents
	// that creature from being declared as an attacker (CR 508.1a). It grants no
	// other ability and never makes a non-defender creature unable to attack.
	RuleEffectCanAttackAsThoughDefender
	// RuleEffectAssignCombatDamageUsingToughness makes the affected creatures
	// assign combat damage equal to their toughness rather than their power
	// ("<subject> assigns combat damage equal to its toughness rather than its
	// power.", Doran, the Siege Tower; Assault Formation; Belligerent Brontodon).
	// While it applies, each matching creature uses its toughness in place of its
	// power when assigning combat damage in any combat damage step (CR 510.1a /
	// the combat-damage replacement), affecting both blocked and unblocked
	// assignments. AffectedSource scopes it to the source creature; an
	// AffectedController plus PermanentTypes/AffectedSelection scopes it to a
	// creature group ("each creature you control", "each creature"). Added last so
	// existing kinds keep their wire values.
	RuleEffectAssignCombatDamageUsingToughness
	// RuleEffectCantActivateAbilitiesOfPermanent forbids any player from
	// activating the affected permanent's own activated abilities ("Enchanted
	// creature can't attack or block, and its activated abilities can't be
	// activated.", Arrest; Pacifism's pinning siblings). It scopes to a single
	// permanent (AffectedAttached resolves to the enchanted permanent, or
	// AffectedSource to the source itself), so it ignores the affected-player
	// relation that the player-scoped RuleEffectCantActivateAbilities uses. When
	// ExemptManaAbilities is set the prohibition spares the permanent's mana
	// abilities ("... can't be activated unless they're mana abilities.", Faith's
	// Fetters, Realmbreaker's Grasp). Added last so existing kinds keep their wire
	// values.
	RuleEffectCantActivateAbilitiesOfPermanent
	// RuleEffectGoaded makes the affected creature goaded by this effect's
	// controller for as long as the effect applies ("Enchanted creature gets
	// +2/+2 and is goaded.", Psychic Impetus; "Equipped creature ... and is
	// goaded.", Bloodthirsty Blade). A goaded creature attacks each combat if
	// able and attacks a player other than the goading player if able (CR 701.38).
	// Unlike the one-shot goad keyword action, which records a turn-limited entry
	// in the permanent's Goaded map, this is a continuous static contribution:
	// AffectedAttached scopes it to the creature an Aura or Equipment is attached
	// to, and the effect's Controller is the goading player. Added last so
	// existing kinds keep their wire values.
	RuleEffectGoaded
	// RuleEffectPlayerHexproof grants the affected player hexproof ("You have
	// hexproof.", Aegis of the Gods, Leyline of Sanctity, Spirit of the Hearth):
	// that player can't be the target of spells or abilities opponents control.
	// AffectedPlayer scopes it to the controller. Added last so existing kinds
	// keep their wire values.
	RuleEffectPlayerHexproof
	// RuleEffectPlayerShroud grants the affected player shroud ("You have
	// shroud.", Ivory Mask, True Believer): that player can't be the target of
	// spells or abilities at all. AffectedPlayer scopes it to the controller.
	// Added last so existing kinds keep their wire values.
	RuleEffectPlayerShroud
	// RuleEffectCanBlockAdditional raises the number of creatures the affected
	// creature may block by AdditionalBlockCount ("This creature can block an
	// additional creature each combat.", Brave the Sands, Coastline Chimera;
	// "Each creature you control can block an additional creature each combat.").
	// A creature with no such effect blocks at most one creature (CR 509.1a); each
	// active effect matching the blocker adds its AdditionalBlockCount to that
	// limit. Added last so existing kinds keep their wire values.
	RuleEffectCanBlockAdditional
	// RuleEffectDamageDoesntCauseLifeLoss stops damage dealt to the affected
	// player from reducing that player's life total ("As long as you're the
	// monarch, damage doesn't cause you to lose life.", Archon of Coronation). The
	// damage is still dealt — it is marked, combat-damage triggers still fire, and
	// the source's controller still becomes the monarch — but the life loss step
	// is skipped. Added last so existing kinds keep their wire values.
	RuleEffectDamageDoesntCauseLifeLoss
	// RuleEffectRedirectDamageToSource redirects all damage that would be dealt to
	// the affected player to the effect's source permanent instead ("All damage
	// that would be dealt to you is dealt to this creature instead.", Protector of
	// the Crown). The runtime resolves the redirect target from the rule effect's
	// SourceObjectID. Added last so existing kinds keep their wire values.
	RuleEffectRedirectDamageToSource
	// RuleEffectCantBeSacrificed prevents the affected permanents from being
	// sacrificed ("Creatures you control but don't own ... can't be sacrificed.",
	// Garland, Royal Kidnapper). AffectedController plus AffectedSelection scope
	// the protected permanents (or AffectedSource for a self form). The runtime
	// excludes a matching permanent from every sacrifice choice and refuses to
	// sacrifice it, whether as a cost or by an effect. Added last so existing
	// kinds keep their wire values.
	RuleEffectCantBeSacrificed
	// RuleEffectCastLinkedExileForFree lets the affected player cast one spell
	// without paying its mana cost from among the cards exiled under the source's
	// ExiledLinkKey set ("until end of turn, you may cast a spell from among cards
	// exiled with this enchantment without paying its mana cost.", Court of
	// Locthwain). AffectedPlayer scopes it to the controller and ExiledLinkKey
	// names the source-keyed linked-exile pool. It is a one-shot permission: the
	// rules layer removes it once its player casts a spell under it, matching the
	// singular "a spell". Added last so existing kinds keep their wire values.
	RuleEffectCastLinkedExileForFree
	// RuleEffectActivateAbilitiesAsThoughHaste lets the affected player activate
	// abilities of creatures they control as though those creatures had haste
	// ("You may activate abilities of creatures you control as though those
	// creatures had haste.", Thousand-Year Elixir, Shang-Chi, Tyvar). It is a
	// controller-scoped activation permission: it removes the summoning-sickness
	// restriction (CR 302.6) that would otherwise stop a creature that hasn't been
	// under its controller's control continuously since their most recent turn
	// began from paying a {T} or {Q} cost in one of its own activated abilities
	// (CR 702.10c). It is an activation permission only — it does not let a
	// summoning-sick creature attack. AffectedPlayer scopes it to the controller.
	// Added last so existing kinds keep their wire values.
	RuleEffectActivateAbilitiesAsThoughHaste
	// RuleEffectGrantSpellKeyword grants GrantedKeyword (a cost-affecting spell
	// keyword such as Improvise) to spells the effect's controller-relative
	// caster casts ("Nonartifact spells you cast have improvise.", Inspiring
	// Statuary; "Noncreature spells you cast have improvise.", Ironheart, Clever
	// Champion; "The next spell you cast this turn has improvise.", Archway of
	// Innovation). AffectedController scopes the caster relative to the source
	// controller (ControllerYou for "you cast"). CardSelection filters the
	// granted spells by printed characteristics (nonartifact = ExcludedTypes
	// [Artifact]); an empty selection grants to every spell that player casts.
	// AppliesToNextSpellOnly limits the grant to the single next matching spell
	// its caster casts and is then consumed (the Archway one-shot); when false
	// the grant is a static that applies to every matching spell for its
	// duration (Inspiring Statuary). The grant only affects cost payment, which
	// the payment planner reads before costs are paid; it is never snapshotted
	// onto the spell. Added last so existing kinds keep their wire values.
	RuleEffectGrantSpellKeyword
	// RuleEffectAscend is the permanent ascend static ability (CR 702.131b): "Any
	// time you control ten or more permanents and you don't have the city's
	// blessing, you get the city's blessing for the rest of the game." It is a
	// marker rule effect carrying no payload beyond its Controller (set while its
	// source permanent is on the battlefield). The runtime continuously checks
	// each controller of an active ascend rule effect during state-based-action
	// processing and grants the city's blessing once that player controls ten or
	// more permanents; the blessing is player-level persistent state that is
	// never removed. Added last so existing kinds keep their wire values.
	RuleEffectAscend
	// RuleEffectCantBeTargetedByControllerOpponents prevents players other than
	// Controller from targeting the affected permanent with spells or abilities.
	// Unlike hexproof, the relation is anchored to this rule effect's source
	// controller rather than the affected permanent's controller.
	RuleEffectCantBeTargetedByControllerOpponents
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
		RuleEffectCantBeBlockedExceptBy,
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
		RuleEffectAdditionalTriggerForControlledPermanent,
		RuleEffectMustBeBlockedByAllAble,
		RuleEffectAssignCombatDamageAsThoughUnblocked,
		RuleEffectCantTransform,
		RuleEffectSuppressOpponentEnteringTriggers,
		RuleEffectAttackTaxPerCreature,
		RuleEffectManaProductionMultiplier,
		RuleEffectSkipDrawStep,
		RuleEffectCanBlockOnlyCreaturesWith,
		RuleEffectCantAttackAlone,
		RuleEffectCantBlockAlone,
		RuleEffectCanAttackAsThoughDefender,
		RuleEffectAssignCombatDamageUsingToughness,
		RuleEffectCantActivateAbilitiesOfPermanent,
		RuleEffectGoaded,
		RuleEffectPlayerHexproof,
		RuleEffectPlayerShroud,
		RuleEffectCanBlockAdditional,
		RuleEffectDamageDoesntCauseLifeLoss,
		RuleEffectRedirectDamageToSource,
		RuleEffectCantBeSacrificed,
		RuleEffectCastLinkedExileForFree,
		RuleEffectActivateAbilitiesAsThoughHaste,
		RuleEffectGrantSpellKeyword,
		RuleEffectAscend,
		RuleEffectCantBeTargetedByControllerOpponents:
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
	// BlockerRestrictionFlyingOrReach matches blockers with either keyword.
	BlockerRestrictionFlyingOrReach
	BlockerRestrictionPowerLessOrEqual
	BlockerRestrictionPowerGreaterOrEqual
	// BlockerRestrictionColor stops blockers of the BlockerRestriction's Color
	// ("can't be blocked by white creatures").
	BlockerRestrictionColor
	// BlockerRestrictionArtifact stops artifact-creature blockers ("can't be
	// blocked by artifact creatures").
	BlockerRestrictionArtifact
	// BlockerRestrictionDefender stops blockers with defender ("can't be blocked
	// except by creatures with defender").
	BlockerRestrictionDefender
	// BlockerRestrictionLegendary stops legendary-creature blockers ("can't be
	// blocked except by legendary creatures").
	BlockerRestrictionLegendary
	// BlockerRestrictionControlledByMonarch stops blockers controlled by the
	// player who currently holds the monarch designation ("can't be blocked by
	// creatures the monarch controls.", Azure Fleet Admiral).
	BlockerRestrictionControlledByMonarch
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
	// SpellColors optionally narrows a cast permission to spells carrying any of
	// these colors.
	SpellColors []color.Color
	// ExcludedSpellTypes exempts spells carrying any of these card types from a
	// RuleEffectCantCastSpells prohibition, expressing the "noncreature spells"
	// family ("Your opponents can't cast noncreature spells this turn.").
	ExcludedSpellTypes []types.Card
	DefendingPlayer    PlayerRelation
	// DefendingPlayerDirectOnly narrows a DefendingPlayer-scoped
	// RuleEffectCantAttack to direct attacks on the defending player, leaving
	// that player's planeswalkers and battles attackable ("can't attack you",
	// Champions of Minas Tirith; CR 508.1 treats a planeswalker/battle as a
	// distinct attack target). When false the restriction also covers attacks
	// aimed at planeswalkers or battles the defending player controls ("can't
	// attack you or planeswalkers you control", the Vow cycle).
	DefendingPlayerDirectOnly bool

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
	// GrantedKeyword is the keyword a keyword-grant rule effect confers:
	// RuleEffectGrantGraveyardCardKeyword confers it on the affected player's
	// matching graveyard cards, and RuleEffectGrantSpellKeyword confers it (a
	// cost-affecting spell keyword such as Improvise) on the affected caster's
	// matching spells. It is unused for every other kind.
	GrantedKeyword Keyword

	CastFromZone   zone.Type
	AffectedCardID id.ID
	CastFace       opt.V[FaceIndex]
	ExpiresFor     PlayerID

	// AffectToOwner, on a RuleEffectPlayFromZone permission, scopes the permission
	// to the owner of AffectedCardID rather than to AffectedPlayer relative to the
	// controller ("its owner may play it", Prowl, Stoic Strategist). The owner may
	// be an opponent of the effect's controller, which no controller-relative
	// PlayerRelation can express, so the rules layer resolves the affected player
	// by AffectedCardID's ownership when this is set. It is false for every other
	// kind.
	AffectToOwner bool

	// SpendAnyMana, on a RuleEffectPlayFromZone permission, lets the affected
	// player spend mana of any type to cast the permitted card ("mana of any
	// type can be spent to cast it.", Court of Locthwain, Court of the Grim
	// Captain). The colored and colorless requirements of the cast card's mana
	// cost stay the same size but become payable with mana of any color, so the
	// rules layer casts it under an alternative cost whose symbols are all
	// generic. It is false for every other kind.
	SpendAnyMana bool

	// ExiledLinkKey, on a RuleEffectCastLinkedExileForFree permission, names the
	// source-keyed linked-exile set whose cards the affected player may cast for
	// free ("cast a spell from among cards exiled with this enchantment", Court of
	// Locthwain). The rules layer keys the pool by the effect's SourceCardID and
	// this link key, matching how the source's ImpulseExile published the cards.
	// It is unused for every other kind.
	ExiledLinkKey LinkedKey

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

	// ExemptManaAbilities spares mana abilities from a
	// RuleEffectCantActivateAbilitiesOfPermanent prohibition ("Enchanted permanent
	// can't attack or block, and its activated abilities can't be activated unless
	// they're mana abilities.", Faith's Fetters, Realmbreaker's Grasp). When false
	// the prohibition stops every activated ability including mana abilities
	// (Arrest). It is unused for every other kind.
	ExemptManaAbilities bool

	// AppliesToNextSpellOnly limits a RuleEffectCantBeCountered or
	// RuleEffectGrantSpellKeyword effect to the single next spell its controller
	// casts ("The next spell you cast this turn can't be countered.", Mistrise
	// Village; "The next spell you cast this turn has improvise.", Archway of
	// Innovation). When the controller casts a matching spell, the effect is
	// consumed, so later spells are unaffected. When false the effect applies to
	// every matching spell for its duration ("Spells you cast this turn can't be
	// countered."; "Nonartifact spells you cast have improvise.").
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

	// AttackDefenderControlsSelection turns a RuleEffectCantAttack restriction into
	// a conditional "can't attack unless defending player controls ..." permission
	// gate ("This creature can't attack unless defending player controls an
	// Island.", Sea Monster). When the Selection is non-empty the affected creature
	// may attack a defending player (or a planeswalker/battle whose controller is
	// that player) only if that player controls at least one permanent matching the
	// Selection; otherwise the attack is prohibited. An empty Selection leaves the
	// RuleEffectCantAttack unconditional. It is unused for every other kind.
	AttackDefenderControlsSelection Selection

	// AttackDefenderIsMonarch turns a RuleEffectCantAttack restriction into a
	// conditional "can't attack unless defending player is the monarch" permission
	// gate ("This creature can't attack unless defending player is the monarch.",
	// Crown-Hunter Hireling). When true the affected creature may attack a
	// defending player (or a planeswalker/battle whose controller is that player)
	// only if that player currently holds the monarch designation; otherwise the
	// attack is prohibited. It is false for every other kind and is mutually
	// exclusive with a non-empty AttackDefenderControlsSelection.
	AttackDefenderIsMonarch bool

	// UntapUnlessControllerIsMonarch turns a RuleEffectDoesntUntap prohibition
	// into a conditional "doesn't untap during its controller's untap step unless
	// that player is the monarch" gate ("Enchanted creature doesn't untap during
	// its controller's untap step unless that player is the monarch.", Fall from
	// Favor). When true the affected permanent untaps normally while its controller
	// currently holds the monarch designation, and is prohibited from untapping
	// otherwise. It is false for every other kind.
	UntapUnlessControllerIsMonarch bool

	// AdditionalBlockCount is the number of extra creatures a
	// RuleEffectCanBlockAdditional lets the affected creature block, beyond the
	// default limit of one ("can block an additional creature" sets it to 1). It
	// is zero for every other kind.
	AdditionalBlockCount int

	// AffectedPlayerRef binds a group-scoped rule effect's affected permanents to
	// a specific player chosen at resolution rather than to the AffectedController
	// relation, expressing "each creature <a chosen target player> controls"
	// (The Brothers' War chapter II). createRuleEffectTemplates resolves it to
	// AffectedSpecificPlayer; a template whose reference does not resolve is
	// dropped. It is PlayerReferenceNone for every other effect.
	AffectedPlayerRef PlayerReference

	// AffectedSpecificPlayer is the resolved player whose creatures a
	// RuleEffectMustAttack effect affects. When set, ruleEffectMatchesPermanent
	// matches only permanents that player controls and ignores AffectedController.
	// It is the resolved form of AffectedPlayerRef and is unset for every other
	// effect.
	AffectedSpecificPlayer opt.V[PlayerID]

	// RequiredAttackTargetRef binds a RuleEffectMustAttack effect's directed
	// attack target to a specific player chosen at resolution ("attacks the other
	// chosen player ... each combat if able", The Brothers' War chapter II).
	// createRuleEffectTemplates resolves it to RequiredAttackTarget; a template
	// whose reference does not resolve is dropped. It is PlayerReferenceNone for
	// every other effect.
	RequiredAttackTargetRef PlayerReference

	// RequiredAttackTarget is the resolved player an affected creature must attack
	// (or a planeswalker or battle that player controls) each combat if able. When
	// set, the forced-attack requirement is directed: the creature is forced only
	// while it can attack that player, and an attack it makes must be aimed there.
	// It is the resolved form of RequiredAttackTargetRef and is unset for every
	// other effect.
	RequiredAttackTarget opt.V[PlayerID]

	// AttackTaxIncludesPlaneswalkers extends a RuleEffectAttackTaxPerCreature tax
	// to cover attacks on planeswalkers the effect's controller controls, not only
	// direct attacks on the controller ("Creatures can't attack you or
	// planeswalkers you control ...", Baird, Archon of Absolution, Sphere of
	// Safety). When false the tax covers only direct attacks on the controller
	// ("Creatures can't attack you ...", Collective Restraint). It is unused for
	// every other kind.
	AttackTaxIncludesPlaneswalkers bool

	// AttackTaxScaledAmount sets a RuleEffectAttackTaxPerCreature per-attacker
	// amount to a board-derived aggregate evaluated from the controller's
	// perspective ("... where X is the number of basic land types among lands you
	// control.", Collective Restraint, domain). It is AggregateNone when the
	// amount is the fixed AttackTaxGeneric or the CardSelection permanent count,
	// and is unused for every other kind.
	AttackTaxScaledAmount AggregateKind

	// ManaProductionMultiplier scales the mana produced when a
	// RuleEffectManaProductionMultiplier effect's controller taps a permanent for
	// mana ("If you tap a permanent for mana, it produces twice as much of that
	// mana instead.", Mana Reflection, 2; Nyxbloom Ancient, 3). It is at least 2
	// for that kind and unused (zero) for every other kind.
	ManaProductionMultiplier int

	// ExileCounterFilter narrows a RuleEffectPlayLandsFromZone or
	// RuleEffectCastSpellsFromZone permission whose CastFromZone is the exile zone
	// to cards carrying the named marker counter it holds ("... from among cards
	// you own in exile with croak counters on them.", Grolnok, the Omnivore). The
	// rules layer only permits playing or casting an exiled card that has at least
	// one such counter recorded in Game.ExileCounters. It is unset for every
	// permission with no exile-counter filter.
	ExileCounterFilter opt.V[counter.Kind]

	// WithoutPayingManaCost, on a per-card RuleEffectPlayFromZone permission, lets
	// the affected player cast AffectedCardID's spell without paying its mana cost
	// ("You may play it this turn without paying its mana cost.", Dauthi
	// Voidwalker). A land played under the same permission has no mana cost, so the
	// flag only affects the spell cast. The rules layer casts the permitted card
	// under a free alternative cost. It is false for every other kind, including
	// the paying ImpulseExile and Prowl play-from-exile grants.
	WithoutPayingManaCost bool

	// ExileCounterExiledByController narrows a RuleEffectPlayLandsFromZone or
	// RuleEffectCastSpellsFromZone exile-counter permission to cards that were
	// exiled by an ability the effect's controller controlled ("... if it was
	// exiled by an ability you controlled", Evelyn, the Covetous). The rules layer
	// checks the exiling controller recorded in Game.ExileCounterExiledBy, so one
	// player's permission can't reach cards another player's ability exiled with
	// the same counter kind. It is false for every permission with no provenance
	// filter and is meaningful only alongside ExileCounterFilter.
	ExileCounterExiledByController bool

	// OncePerTurn caps a RuleEffectPlayLandsFromZone or RuleEffectCastSpellsFromZone
	// permission to one use per turn per source permanent ("Once each turn, you may
	// play a card from exile ...", Evelyn, the Covetous). A source that grants both
	// a land-play and a spell-cast permission shares the single per-turn use, keyed
	// by SourceObjectID in Game.ExilePlayPermissionUsedThisTurn. It is false for
	// every permission with no per-turn cap.
	OncePerTurn bool
}
