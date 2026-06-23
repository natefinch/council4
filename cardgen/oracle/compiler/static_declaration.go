package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// StaticDeclarationKind identifies a declaration category that never resolves.
type StaticDeclarationKind uint8

// Static declaration categories.
const (
	StaticDeclarationUnknown StaticDeclarationKind = iota
	StaticDeclarationContinuous
	StaticDeclarationRule
	StaticDeclarationCostModifier
	StaticDeclarationCardAbilityGrant
	StaticDeclarationPlayerRule
	StaticDeclarationOpponentActionRestriction
	StaticDeclarationSpellUncounterable
	StaticDeclarationEnteringTriggerMultiplier
	StaticDeclarationControlledTriggerMultiplier
	StaticDeclarationUntapStep
	StaticDeclarationCharacteristicPowerToughness
	StaticDeclarationEnterBattlefieldRestriction
	StaticDeclarationCastAsThoughFlash
	StaticDeclarationGraveyardCardKeywordGrant
	StaticDeclarationDrawLimit
	StaticDeclarationCastLimit
)

// StaticDeclarationBlocker identifies exact static wording whose declaration
// category is understood but whose semantic vocabulary is not yet representable.
type StaticDeclarationBlocker uint8

// Exact static declaration blockers.
const (
	StaticDeclarationBlockerNone StaticDeclarationBlocker = iota
	StaticDeclarationBlockerHistoricCardSelection
	StaticDeclarationBlockerCondition
	StaticDeclarationBlockerDuration
	StaticDeclarationBlockerGroup
	StaticDeclarationBlockerOperation
	StaticDeclarationBlockerShell
)

// StaticContinuousLayer identifies a semantic continuous-effect layer.
type StaticContinuousLayer uint8

// Static continuous-effect layers currently recognized by Card Generation.
const (
	StaticLayerUnknown StaticContinuousLayer = iota
	StaticLayerAbility
	StaticLayerPowerToughnessModify
	StaticLayerPowerToughnessSet
	StaticLayerType
	StaticLayerColor
	StaticLayerControl
)

// StaticContinuousOperation identifies a characteristic operation.
type StaticContinuousOperation uint8

// Static continuous-effect operations.
const (
	StaticContinuousUnknown StaticContinuousOperation = iota
	StaticContinuousModifyPowerToughness
	StaticContinuousSetBasePowerToughness
	StaticContinuousGrantKeywords
	StaticContinuousGrantManaAbility
	StaticContinuousAddTypes
	StaticContinuousAddSubtypeFromEntryChoice
	StaticContinuousSetTypes
	StaticContinuousSetSubtypes
	StaticContinuousAddColors
	StaticContinuousSetColors
	StaticContinuousChangeControl
	StaticContinuousRemoveAllAbilities
	StaticContinuousGrantAbility
)

// StaticRuleKind identifies a non-layer rules declaration.
type StaticRuleKind uint8

// StaticRuleDomain identifies the rules action constrained by a declaration.
type StaticRuleDomain uint8

// Static rule domains. Operations are added only when the runtime can represent
// them, while the closed domains keep recognition independent of wording.
const (
	StaticRuleDomainUnknown StaticRuleDomain = iota
	StaticRuleDomainAttack
	StaticRuleDomainBlock
	StaticRuleDomainCast
	StaticRuleDomainActivate
	StaticRuleDomainTarget
	StaticRuleDomainCountering
	StaticRuleDomainAttackBlock
	StaticRuleDomainUntap
	StaticRuleDomainTrigger
)

// Static rule declarations currently recognized by Card Generation.
const (
	StaticRuleUnknown StaticRuleKind = iota
	StaticRuleCantBlock
	StaticRuleCantBeBlocked
	StaticRuleCantAttack
	StaticRuleMustAttack
	StaticRuleMustBeBlocked
	StaticRuleCantBeCountered
	StaticRuleCantAttackOrBlock
	StaticRuleDoesntUntap
	// StaticRuleCantAttackYou prohibits attacking the source's controller or
	// their planeswalkers ("can't attack you or planeswalkers you control").
	StaticRuleCantAttackYou
	// StaticRuleCantBeBlockedByMoreThanOne bounds blocking the subject to at
	// most one creature ("can't be blocked by more than one creature").
	StaticRuleCantBeBlockedByMoreThanOne
	// StaticRuleCantBeBlockedByCreaturesWith is a restricted block prohibition
	// bounded by a blocker characteristic ("can't be blocked by creatures with
	// flying", "... power N or less", "... power N or greater"); the bounding
	// characteristic travels in StaticRuleDeclaration.Blocker.
	StaticRuleCantBeBlockedByCreaturesWith
	StaticRuleAdditionalTriggerForChosenCreatureType
	// StaticRuleCantBlockAndCantBeBlocked combines the active "can't block" and
	// passive "can't be blocked" prohibitions printed as one sentence; it lowers
	// to both the can't-block and can't-be-blocked runtime rule effects.
	StaticRuleCantBlockAndCantBeBlocked
)

// StaticBlockerRestrictionKind identifies the blocker characteristic bounding a
// restricted "can't be blocked by creatures with ..." prohibition.
type StaticBlockerRestrictionKind uint8

// Static blocker restriction kinds.
const (
	StaticBlockerRestrictionNone StaticBlockerRestrictionKind = iota
	StaticBlockerRestrictionFlying
	StaticBlockerRestrictionPowerOrLess
	StaticBlockerRestrictionPowerOrGreater
	// StaticBlockerRestrictionColor bounds the prohibition to blockers of the
	// restriction's Color ("can't be blocked by white creatures").
	StaticBlockerRestrictionColor
	// StaticBlockerRestrictionArtifact bounds the prohibition to artifact-creature
	// blockers ("can't be blocked by artifact creatures").
	StaticBlockerRestrictionArtifact
)

// StaticBlockerRestriction is the closed blocker characteristic bounding a
// restricted block prohibition. Amount is the power threshold for the
// power-comparison kinds; Color names the stopped blocker color for the color
// kind. Both are unused for kinds that do not need them.
type StaticBlockerRestriction struct {
	Kind   StaticBlockerRestrictionKind
	Amount int
	Color  color.Color
}

// StaticZone identifies where a static declaration functions.
type StaticZone uint8

// Static declaration zones.
const (
	StaticZoneBattlefield StaticZone = iota
	StaticZoneStack
	StaticZoneHand
)

// StaticGroupDomain identifies the closed candidate domain of an affected group.
type StaticGroupDomain uint8

// Static affected-group domains.
const (
	StaticGroupUnknown StaticGroupDomain = iota
	StaticGroupSource
	StaticGroupBattlefield
	StaticGroupAttachedObject
	StaticGroupSourceControllerPermanents
	StaticGroupControllerHandCards
	StaticGroupControllerSpells
	// StaticGroupControllerEquipment is the controller's Equipment permanents on
	// the battlefield, the affected group of "Equipment you control have equip
	// {N}." Its members are matched at runtime by the Equip activated ability, so
	// the lowered cost modifier targets the Equip keyword directly.
	StaticGroupControllerEquipment
	// StaticGroupControllerGraveyardCards is the set of the controller's
	// graveyard cards a "[During your turn,] <filter> cards in your graveyard
	// have <keyword>." declaration grants a keyword to (Six, Wrenn and Six
	// Emblem). Its members are matched at runtime by the lowered rule effect's
	// card selection.
	StaticGroupControllerGraveyardCards
)

// StaticCombatState constrains a static group's members by combat involvement.
type StaticCombatState uint8

// Static combat-state filters. The zero value applies no combat constraint.
const (
	StaticCombatStateAny StaticCombatState = iota
	StaticCombatStateAttacking
	StaticCombatStateBlocking
)

// StaticTapState constrains a static group's members by tapped state.
type StaticTapState uint8

// Static tap-state filters. The zero value applies no tap constraint.
const (
	StaticTapStateAny StaticTapState = iota
	StaticTapStateTapped
	StaticTapStateUntapped
)

// StaticSelection is source-independent semantic data describing WHAT objects
// in a static declaration's group match.
type StaticSelection struct {
	RequiredTypes []types.Card
	Supertypes    []types.Super
	// ExcludedSupertypes lists supertypes a member must NOT carry (the
	// "nonlegendary creatures you control" exclusion). Lowering routes the first
	// entry onto the runtime Selection.ExcludedSupertype scalar.
	ExcludedSupertypes []types.Super
	SubtypesAny        []types.Sub
	ColorsAny          []color.Color
	Colorless          bool
	Multicolored       bool
	Controller         ControllerKind
	CombatState        StaticCombatState
	TapState           StaticTapState
	Keyword            parser.KeywordKind
	ExcludedKeyword    parser.KeywordKind
	TokenOnly          bool
	NonToken           bool
	// MatchCounter, when true, restricts the group to permanents carrying a
	// counter of RequiredCounter's kind ("creature you control with a +1/+1
	// counter on it"). A bool flag distinguishes "no counter requirement" from
	// "requires a +1/+1 counter" because counter.Kind's zero value names the
	// +1/+1 counter.
	MatchCounter    bool
	RequiredCounter counter.Kind
	// MatchAnyCounter, when true, restricts the group to permanents carrying a
	// counter of any kind ("creature you control with a counter on it",
	// Rishkar), independent of RequiredCounter.
	MatchAnyCounter bool
	// SubtypeFromEntryChoice constrains the group to permanents whose creature
	// subtype matches the source permanent's entry-time creature-type choice
	// ("creatures you control of the chosen type"). Lowering routes it to the
	// runtime Selection.SubtypeFromSourceEntryChoice predicate.
	SubtypeFromEntryChoice bool
	// Modified, when true, restricts the group to permanents that are modified:
	// carrying a counter or having an Aura or Equipment attached ("modified
	// creatures you control"). Lowering routes it to the runtime
	// Selection.MatchModified predicate.
	Modified bool
	// Commander, when true, restricts the group to permanents that are a
	// commander ("commander creatures you control"). Lowering routes it to the
	// runtime Selection.MatchCommander predicate.
	Commander bool
	// ColorFromEntryChoice constrains the group to permanents whose color
	// matches the source permanent's entry-time color choice ("creatures you
	// control of the chosen color"). Lowering routes it to the runtime
	// Selection.ColorChoice = ColorChoiceSourceEntry predicate.
	ColorFromEntryChoice bool
	// Power and Toughness carry an optional numeric power/toughness comparison
	// constraining the group ("creatures you control with power 1 or less").
	// MatchPower and MatchToughness mark each comparison active. PowerOrToughness
	// marks the disjunctive "power or toughness N or less" form (Tetsuko): a
	// member matches if EITHER its power OR its toughness satisfies the bound;
	// lowering emits a Selection.AnyOf of the two single-characteristic
	// alternatives rather than ANDing the two thresholds.
	Power            compare.Int
	MatchPower       bool
	Toughness        compare.Int
	MatchToughness   bool
	PowerOrToughness bool
	// PowerLessThanSource and PowerGreaterThanSource compare each member's power
	// to the static's SOURCE permanent's power ("with power greater than
	// <source>'s power", Champion of Lambholt). They are source-relative, so they
	// carry no fixed comparison and lower to Selection.PowerLessThanSource /
	// Selection.PowerGreaterThanSource.
	PowerLessThanSource    bool
	PowerGreaterThanSource bool
	// ExcludedTypes lists card types a member must NOT carry, set by a
	// "non-<type>" prefix on a group noun ("Nonland permanents you control are
	// artifacts ...", Encroaching Mycosynth). Lowering routes it onto the runtime
	// Selection.ExcludedTypes predicate.
	ExcludedTypes []types.Card
}

// StaticGroupReference describes WHERE a static declaration finds objects and
// carries the Selection that describes WHAT matches there.
type StaticGroupReference struct {
	Span          shared.Span
	Domain        StaticGroupDomain
	Selection     StaticSelection
	ExcludeSource bool
}

// StaticContinuousDeclaration is one layer-preserving characteristic change.
type StaticContinuousDeclaration struct {
	Layer          StaticContinuousLayer
	Operation      StaticContinuousOperation
	PowerDelta     CompiledSignedAmount
	ToughnessDelta CompiledSignedAmount
	DynamicAmount  CompiledAmount
	Keywords       []CompiledKeyword
	GrantedMana    *StaticGrantedManaAbility

	// GrantedAbility carries the parsed quoted ability a
	// StaticContinuousGrantAbility operation confers on its subject. The
	// lowering compiles its inner document and lowers the resulting ability
	// into the continuous effect's granted-ability set.
	GrantedAbility *parser.StaticGrantedAbilitySyntax

	// Set base power/toughness payload (StaticContinuousSetBasePowerToughness).
	SetPower     int
	SetToughness int

	// Color characteristic payload (StaticContinuousAddColors / SetColors).
	// SetColorless marks a SetColors operation that makes the affected object
	// colorless (its color set becomes empty); Colors is then empty.
	Colors       []color.Color
	SetColorless bool

	// Type characteristic payload. AddTypes/AddSubtypes are additive
	// (StaticContinuousAddTypes); SetTypes/SetSubtypes replace the affected
	// object's card types and creature types (StaticContinuousSetTypes,
	// StaticContinuousSetSubtypes).
	AddTypes    []types.Card
	AddSubtypes []types.Sub
	SetTypes    []types.Card
	SetSubtypes []types.Sub
	// AddEveryCreatureType adds every creature subtype at the type layer
	// (StaticContinuousAddTypes), the runtime expansion of "is/are every
	// creature type" (CR 702.73). It is mutually exclusive with the enumerated
	// AddTypes/AddSubtypes payload.
	AddEveryCreatureType bool
	// AddEveryBasicLandType adds all five basic land subtypes at LayerType
	// ("<group> is/are every basic land type", Dryad of the Ilysian Grove,
	// Prismatic Omen). The runtime expands it rather than enumerating subtypes
	// here.
	AddEveryBasicLandType bool
}

// StaticGrantedManaAbility is one closed activated mana ability granted in the
// ability layer by a static declaration.
type StaticGrantedManaAbility struct {
	TapCost  bool
	Amount   int
	AnyColor bool
	// Text is the granted ability's printed wording, carried verbatim so the
	// lowering reproduces it without re-deriving text from the typed fields.
	Text string
	// Sacrifice marks the "Sacrifice this artifact" additional cost.
	Sacrifice bool
	// AnyOneColor marks the "Add <Amount> mana of any one color" output (one
	// chosen color, Amount >= 2). It is mutually exclusive with AnyColor.
	AnyOneColor bool
	// Colorless marks the bare "{T}: Add {C}" ability that adds one colorless
	// mana. It is mutually exclusive with AnyColor and AnyOneColor.
	Colorless bool
}

// StaticBlockedObjectKind identifies the protected object an active "can't
// block" restriction shields. The empty value (None) is an unconditional block
// prohibition; the others scope the prohibition to blocking a specific group.
type StaticBlockedObjectKind uint8

// Static blocked-object scopes.
const (
	StaticBlockedObjectNone StaticBlockedObjectKind = iota
	// StaticBlockedObjectSource shields the source permanent itself ("can't block
	// it", "can't block this creature").
	StaticBlockedObjectSource
	// StaticBlockedObjectControlledCreatures shields the source controller's
	// creatures ("can't block creatures you control").
	StaticBlockedObjectControlledCreatures
)

// StaticRuleDeclaration is one prohibition, requirement, or permission.
type StaticRuleDeclaration struct {
	Domain  StaticRuleDomain
	Kind    StaticRuleKind
	Zone    StaticZone
	Blocker StaticBlockerRestriction
	// BlockedObject scopes an active "can't block" restriction to a protected
	// object ("can't block it", "can't block creatures you control"). It is None
	// for an unconditional block prohibition and unused for every other rule.
	BlockedObject StaticBlockedObjectKind
}

// StaticCostModifierKind identifies which semantic cost category is modified.
type StaticCostModifierKind uint8

// Static cost modifier kinds.
const (
	StaticCostModifierUnknown StaticCostModifierKind = iota
	StaticCostModifierAbility
	StaticCostModifierSpell
)

// StaticCostModifierDeclaration is a semantic cost change.
//
// MatchSpellColor constrains a spell cost modifier to spells of a single color.
// When MatchSpellColor is set, SpellColor names the required color; an empty
// SpellColor is the colorless sentinel, constraining the modifier to colorless
// spells. A color filter may combine with a SpellTypes card-type filter ("black
// creature spells"); SpellSubtypes is an alternative subtype filter.
type StaticCostModifierDeclaration struct {
	Kind                         StaticCostModifierKind
	AbilityKeyword               parser.KeywordKind
	SpellTypes                   []types.Card
	MatchSpellColor              bool
	SpellColor                   color.Color
	ChosenSubtypeFromEntryChoice bool
	GenericReduction             int
	GenericIncrease              int
	SetManaCost                  string
	ReplaceManaCost              bool
	FirstCycleEachTurn           bool

	// SpellColors constrains a spell cost modifier to spells carrying any one of
	// these colors ("... that's red or green ..."). It holds two or more real
	// colors and is mutually exclusive with MatchSpellColor and SpellTypes.
	SpellColors []color.Color

	// SpellSubtypes constrains a spell cost modifier to spells carrying any one
	// of these subtypes ("Aura and Equipment spells ..."). It may combine with a
	// color filter and is mutually exclusive with SpellTypes and SpellColors.
	SpellSubtypes []types.Sub

	// SourceZone constrains a spell cost modifier to spells being cast from a
	// single zone ("Spells you cast from your graveyard cost {N} less to cast.").
	// The empty kind applies no zone filter, so the modifier affects spells cast
	// from any zone. It combines with the card-type, color, and subtype filters.
	SourceZone parser.StaticDeclarationCastZoneKind

	// MinPower constrains a spell cost modifier to spells whose base printed
	// power is at least this threshold ("Creature spells you cast with power 4
	// or greater cost {2} less to cast.", Goreclaw). MatchMinPower marks the
	// threshold present so a zero threshold stays expressible. It combines with
	// the card-type, color, subtype, and zone filters.
	MinPower      int
	MatchMinPower bool

	// TargetsSource constrains a spell cost modifier to spells that target the
	// source permanent ("Spells your opponents cast that target this creature
	// cost {2} more to cast.", Boreal Elemental). Caster identifies which
	// players' spells the modifier affects.
	TargetsSource bool
	Caster        StaticSpellCasterKind
}

// StaticSpellCasterKind identifies which players' spells a cast-cost modifier
// affects.
type StaticSpellCasterKind uint8

// Static spell caster kinds recognized by Card Generation.
const (
	// StaticSpellCasterController is the default: the static ability
	// controller's own spells ("Spells you cast ...").
	StaticSpellCasterController StaticSpellCasterKind = iota
	// StaticSpellCasterOpponents is the controller's opponents' spells
	// ("Spells your opponents cast ...").
	StaticSpellCasterOpponents
	// StaticSpellCasterAny is every player's spells ("Spells that ...").
	StaticSpellCasterAny
)

// StaticPlayerRuleKind identifies a closed player-scoped static rule.
type StaticPlayerRuleKind uint8

// Static player rule kinds currently recognized by Card Generation.
const (
	StaticPlayerRuleUnknown StaticPlayerRuleKind = iota
	StaticPlayerRuleNoMaximumHandSize
	StaticPlayerRuleAttackTax
	StaticPlayerRuleAdditionalLandPlays
	StaticPlayerRulePlayLandsFromGraveyard
	StaticPlayerRulePlayLandsFromLibraryTop
	StaticPlayerRulePlayWithTopCardRevealed
	StaticPlayerRuleCastSpellsFromLibraryTop
	StaticPlayerRuleCastThisFromGraveyard
	StaticPlayerRuleLookAtTopCardAnyTime
	StaticPlayerRuleCastThisFromExile
	StaticPlayerRuleLifeForColoredMana
	StaticPlayerRuleLifeForCommanderTax
)

// StaticPlayerRuleDeclaration is one player-scoped static rule applied to the
// static ability's controller.
type StaticPlayerRuleDeclaration struct {
	Kind                StaticPlayerRuleKind
	AttackTaxGeneric    int
	AdditionalLandPlays int
	AffectsAllPlayers   bool

	// SpellTypes filters a StaticPlayerRuleCastSpellsFromLibraryTop permission by
	// card type (any one of the listed types); an empty SpellTypes permits casting
	// any spell. CastColorless additionally permits casting colorless spells, so a
	// spell qualifies when it matches SpellTypes or is colorless ("artifact spells
	// and colorless spells", Mystic Forge). AlsoPlayLands records the combined
	// "play lands and cast spells from the top of your library." wording, which
	// additionally grants the land-play permission. CastChosenCreatureType narrows
	// the permission to spells sharing the source permanent's entry-chosen creature
	// subtype ("creature spells of the chosen type", Realmwalker). All four are
	// unused for every other kind.
	SpellTypes             []types.Card
	CastColorless          bool
	AlsoPlayLands          bool
	CastChosenCreatureType bool

	// ManaColor carries the colored mana symbol of a
	// StaticPlayerRuleLifeForColoredMana rule ("For each {B} in a cost, ...",
	// K'rrik). It is empty for every other kind.
	ManaColor mana.Color
}

// StaticCardAbilityGrantDeclaration grants a keyword ability to cards in a
// non-battlefield group.
type StaticCardAbilityGrantDeclaration struct {
	Keyword CompiledKeyword
	Text    string
}

// StaticGraveyardKeywordGrantDeclaration grants a parameterless keyword to a
// filtered set of the controller's graveyard cards ("[During your turn,]
// <filter> cards in your graveyard have <keyword>.", Six, Wrenn and Six
// Emblem). Filter constrains the affected cards by card type; DuringControllerTurn
// scopes the grant to the controller's turn.
type StaticGraveyardKeywordGrantDeclaration struct {
	Keyword              CompiledKeyword
	Filter               parser.StaticDeclarationCardFilterKind
	DuringControllerTurn bool
}

// StaticOpponentActionRestrictionDeclaration is a continuous prohibition that
// stops the affected players from casting spells and/or activating abilities of
// permanents whose card type is in ActivateTypes. AffectsAllPlayers selects
// every player ("Players can't ...") rather than only the controller's opponents
// ("Your opponents can't ..."); DuringControllerTurn scopes the prohibition to
// the controller's turn.
type StaticOpponentActionRestrictionDeclaration struct {
	RestrictCastSpells   bool
	ActivateTypes        []types.Card
	AffectsAllPlayers    bool
	DuringControllerTurn bool

	// CastOnlyFromHand scopes the cast prohibition to every non-hand zone ("...
	// can't cast spells from anywhere other than their hands.", Drannith
	// Magistrate). The lowering expands it to the explicit non-hand cast zones.
	CastOnlyFromHand bool

	// CastFromZones scopes the cast prohibition to a set of source zones ("...
	// can't cast spells from graveyards or libraries."). When non-empty the cast
	// prohibition forbids casting only out of those zones rather than every zone.
	CastFromZones []parser.StaticDeclarationCastZoneKind
}

// StaticEnterBattlefieldRestrictionDeclaration forbids a filtered set of cards
// from entering the battlefield out of FromZones ("Creature cards in graveyards
// and libraries can't enter the battlefield."). The restriction is global. Filter
// selects which entering cards it affects.
type StaticEnterBattlefieldRestrictionDeclaration struct {
	Filter    parser.StaticDeclarationEnterFilterKind
	FromZones []parser.StaticDeclarationCastZoneKind
}

// StaticSpellUncounterableDeclaration makes a group of the controller's spells
// uncounterable ("[<type>] spells you control can't be countered."). SpellTypes
// is the disjunction of card types the affected spells must include; an empty
// SpellTypes affects every spell the controller casts.
type StaticSpellUncounterableDeclaration struct {
	SpellTypes []types.Card
}

// StaticEnteringTriggerMultiplierDeclaration makes a triggered ability of a
// permanent the controller controls trigger one additional time when an entering
// permanent caused it ("If an artifact or creature entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time.", Panharmonicon, Yarok, Ancient Greenwarden). EnteringTypes
// is the disjunction of card types the entering permanent must include; an empty
// EnteringTypes matches any entering permanent ("a permanent").
type StaticEnteringTriggerMultiplierDeclaration struct {
	EnteringTypes []types.Card
}

// StaticUntapStepDeclaration grants an extra untap to a group of the
// controller's permanents during each other player's untap step ("Untap all
// permanents you control during each other player's untap step."). Self scopes
// it to the source permanent itself; otherwise PermanentTypes filters the
// controller's permanents by card type (an empty PermanentTypes untaps every
// permanent the controller controls).
type StaticUntapStepDeclaration struct {
	Self           bool
	PermanentTypes []types.Card
}

// StaticCastAsThoughFlashDeclaration grants the controller a continuous timing
// permission to cast spells as though they had flash ("You may cast spells as
// though they had flash.", Vedalken Orrery). SpellTypes and SpellSubtypes
// optionally narrow the grant to spells of those card types ("sorcery spells")
// or subtypes ("Aura and Equipment spells"); empty filters permit every spell.
type StaticCastAsThoughFlashDeclaration struct {
	SpellTypes    []types.Card
	SpellSubtypes []types.Sub
}

// StaticDrawLimitDeclaration caps how many cards the affected players may draw
// each turn at Limit ("Each opponent can't draw more than one card each turn.",
// Narset, Parter of Veils). AffectsAllPlayers selects every player ("Each player
// can't ...", Spirit of the Labyrinth); AffectsController selects only the
// controller ("You can't ..."). With neither flag set the cap affects only the
// controller's opponents.
type StaticDrawLimitDeclaration struct {
	Limit             int
	AffectsAllPlayers bool
	AffectsController bool
}

// StaticCastLimitDeclaration caps how many spells the affected players may cast
// each turn at Limit ("Each player can't cast more than one spell each turn.",
// Rule of Law, Eidolon of Rhetoric). AffectsAllPlayers selects every player
// ("Each player"/"Players"); AffectsController selects only the controller
// ("You"). With neither flag set the cap affects only the controller's
// opponents.
type StaticCastLimitDeclaration struct {
	Limit             int
	AffectsAllPlayers bool
	AffectsController bool
}

// StaticDeclaration is source-spanned semantic data attached directly to a
// static ability. It is not Instruction content and never resolves.
type StaticDeclaration struct {
	Kind          StaticDeclarationKind
	Span          shared.Span
	OperationSpan shared.Span
	Group         StaticGroupReference
	Condition     *CompiledCondition

	// Exactly one variant payload matching Kind is non-nil.
	Continuous           *StaticContinuousDeclaration
	Rule                 *StaticRuleDeclaration
	Cost                 *StaticCostModifierDeclaration
	CardGrant            *StaticCardAbilityGrantDeclaration
	Player               *StaticPlayerRuleDeclaration
	OpponentRestriction  *StaticOpponentActionRestrictionDeclaration
	EnterRestriction     *StaticEnterBattlefieldRestrictionDeclaration
	SpellUncounterable   *StaticSpellUncounterableDeclaration
	EnteringMultiplier   *StaticEnteringTriggerMultiplierDeclaration
	ControlledMultiplier *StaticControlledTriggerMultiplierDeclaration
	Untap                *StaticUntapStepDeclaration
	CharacteristicPT     *StaticCharacteristicPowerToughnessDeclaration
	CastAsThoughFlash    *StaticCastAsThoughFlashDeclaration
	GraveyardGrant       *StaticGraveyardKeywordGrantDeclaration
	DrawLimit            *StaticDrawLimitDeclaration
	CastLimit            *StaticCastLimitDeclaration
}

// StaticCharacteristicPowerToughnessDeclaration carries the rules-derived count
// a characteristic-defining ability sets the source object's power and toughness
// equal to ("its power and toughness are each equal to the number of cards in
// your hand"). It applies only to the source object. SetsPower and SetsToughness
// record which characteristics the declaration sets; ToughnessOffset is the
// fixed integer added to the toughness count for the "that number plus N" form.
type StaticCharacteristicPowerToughnessDeclaration struct {
	Value           game.DynamicValueKind
	Subtype         types.Sub
	Color           color.Color
	SetsPower       bool
	SetsToughness   bool
	ToughnessOffset int
}

// CompiledStaticSemantics contains declarations recognized for a static
// ability, or the exact reason a declaration-shaped ability cannot be lowered.
type CompiledStaticSemantics struct {
	Declarations []StaticDeclaration
	Blocker      StaticDeclarationBlocker
}

// recognizeStaticDeclarations maps the typed static-declaration syntax the
// parser emitted for this ability onto closed semantic declarations. It consumes
// typed parser nodes and already-compiled semantic content only; it inspects no
// Oracle source text or tokens to derive meaning. Retained spans support exact
// source-consumption accounting and diagnostics.
func recognizeStaticDeclarations(compiled *CompiledAbility, syntax *parser.Ability) {
	if compiled.Kind != AbilityStatic {
		return
	}
	statics := syntax.StaticDeclarations
	if declarations, ok := recognizeTypedStaticRuleDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticEnchantedTypeChangeDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticLoseAbilitiesBecomeDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeMixedSourceStaticDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticPowerToughnessDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticCharacteristicPowerToughnessDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declarations, ok := recognizeStaticPowerToughnessRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticControlledGroupRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticBattlefieldBlockRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticGroupMustAttackDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticEntryChoiceSubtypeDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticChosenCreatureTypeTriggerMultiplier(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticEnteringTriggerMultiplier(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticControlledTriggerMultiplier(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declarations, ok := recognizeStaticComposedContinuousDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticQuotedAbilityGrantDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticSpellCostModifierDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCostModifierDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticAbilityCostSetDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCardAbilityGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticGraveyardCardKeywordGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticPermanentAbilityGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticControlGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticPlayerRuleDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticOpponentActionRestrictionDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticDrawLimitDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCastLimitDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticEnterBattlefieldRestrictionDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticSpellUncounterableDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCastAsThoughFlashDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticUntapStepDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if staticSyntaxIsHistoricCardGrant(*compiled, statics) {
		compiled.Static = &CompiledStaticSemantics{Blocker: StaticDeclarationBlockerHistoricCardSelection}
		return
	}
	if blocker := classifyStaticDeclarationBlocker(*compiled); blocker != StaticDeclarationBlockerNone {
		compiled.Static = &CompiledStaticSemantics{Blocker: blocker}
	}
}

func recognizeStaticChosenCreatureTypeTriggerMultiplier(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationChosenCreatureTypeTriggerMultiplier) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!chosenCreatureTypeTriggerMultiplierContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	return StaticDeclaration{
		Kind:          StaticDeclarationRule,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   node.Span,
			Domain: StaticGroupSource,
		},
		Rule: &StaticRuleDeclaration{
			Domain: StaticRuleDomainTrigger,
			Kind:   StaticRuleAdditionalTriggerForChosenCreatureType,
			Zone:   StaticZoneBattlefield,
		},
	}, true
}

// recognizeStaticEnteringTriggerMultiplier maps the parser-owned entering-trigger
// multiplier syntax ("If an artifact or creature entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time.", Panharmonicon, Yarok, Ancient Greenwarden) onto its closed
// semantic payload. The entering permanent's type filter travels in
// EnteringTypes; an empty filter matches any entering permanent.
func recognizeStaticEnteringTriggerMultiplier(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationEnteringTriggerMultiplier) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!enteringTriggerMultiplierContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	enteringTypes := make([]types.Card, 0, len(node.EnteringFilterTypes))
	for _, cardType := range node.EnteringFilterTypes {
		converted, ok := compilerCardType(cardType)
		if !ok {
			return StaticDeclaration{}, false
		}
		enteringTypes = append(enteringTypes, converted)
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationEnteringTriggerMultiplier,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		EnteringMultiplier: &StaticEnteringTriggerMultiplierDeclaration{
			EnteringTypes: enteringTypes,
		},
	}, true
}

// StaticControlledTriggerMultiplierDeclaration makes a triggered ability of a
// permanent the controller controls trigger one additional time when that
// permanent matches the source-permanent filter ("If a triggered ability of a
// legendary creature you control triggers, that ability triggers an additional
// time.", Annie Joins Up; Katara, the Fearless; Splinter, Radical Rat). Types
// and Supertypes are conjunctive; Subtypes is disjunctive. Unlike the
// chosen-type and entering-permanent doublers this family includes the doubler's
// own triggers ("a ... you control", not "another").
type StaticControlledTriggerMultiplierDeclaration struct {
	Types      []types.Card
	Supertypes []types.Super
	Subtypes   []types.Sub
}

// recognizeStaticControlledTriggerMultiplier maps the parser-owned
// controlled-trigger multiplier syntax onto its closed semantic payload. The
// source permanent's type, supertype, and subtype filter travels in Types,
// Supertypes, and Subtypes.
func recognizeStaticControlledTriggerMultiplier(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationControlledTriggerMultiplier) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!enteringTriggerMultiplierContent(ability.Content) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	cardTypes := make([]types.Card, 0, len(node.ControlledFilterTypes))
	for _, cardType := range node.ControlledFilterTypes {
		converted, ok := compilerCardType(cardType)
		if !ok {
			return StaticDeclaration{}, false
		}
		cardTypes = append(cardTypes, converted)
	}
	supertypes := make([]types.Super, 0, len(node.ControlledFilterSupertypes))
	for _, supertype := range node.ControlledFilterSupertypes {
		converted, ok := compilerSupertype(supertype)
		if !ok {
			return StaticDeclaration{}, false
		}
		supertypes = append(supertypes, converted)
	}
	subtypes := append([]types.Sub(nil), node.ControlledFilterSubtypes...)
	if len(cardTypes) == 0 && len(supertypes) == 0 && len(subtypes) == 0 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationControlledTriggerMultiplier,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		ControlledMultiplier: &StaticControlledTriggerMultiplierDeclaration{
			Types:      cardTypes,
			Supertypes: supertypes,
			Subtypes:   subtypes,
		},
	}, true
}

// enteringTriggerMultiplierContent reports whether the leftover content matches
// the entering-trigger multiplier shape: a single unsupported "if ... causes ...
// to trigger" condition and no other content the static declaration would
// otherwise own. The controlled-permanent multiplier shares this leftover shape.
func enteringTriggerMultiplierContent(content AbilityContent) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 0 {
		return false
	}
	condition := content.Conditions[0]
	return condition.Kind == ConditionIf &&
		condition.Predicate == ConditionPredicateUnsupported &&
		!condition.Intervening &&
		!condition.Resolving
}

func recognizeStaticEntryChoiceSubtypeDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationContinuousEntryChoiceSubtype) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Keywords) != 0 {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	group, ok := staticGroupForParserSubject(node.Subject)
	if !ok {
		return StaticDeclaration{}, false
	}
	switch node.Subject.Kind {
	case parser.StaticDeclarationSubjectSourceCreature:
		if !entryChoiceSubtypeContent(ability.Content) {
			return StaticDeclaration{}, false
		}
	case parser.StaticDeclarationSubjectGroup:
		if !entryChoiceSubtypeGroupContent(ability.Content) {
			return StaticDeclaration{}, false
		}
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerType,
			Operation: StaticContinuousAddSubtypeFromEntryChoice,
		},
	}, true
}

func chosenCreatureTypeTriggerMultiplierContent(content AbilityContent) bool {
	if len(content.Conditions) != 1 ||
		len(content.Effects) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.References) != 1 {
		return false
	}
	condition := content.Conditions[0]
	reference := content.References[0]
	return condition.Kind == ConditionIf &&
		condition.Predicate == ConditionPredicateUnsupported &&
		!condition.Intervening &&
		!condition.Resolving &&
		reference.Kind == ReferencePronoun &&
		reference.Pronoun == ReferencePronounIt &&
		reference.Binding == ReferenceBindingAmbiguous
}

func entryChoiceSubtypeContent(content AbilityContent) bool {
	if len(content.References) != 2 {
		return false
	}
	source := content.References[0]
	possessive := content.References[1]
	return source.Kind == ReferenceThisObject &&
		source.Binding == ReferenceBindingSource &&
		possessive.Kind == ReferencePronoun &&
		possessive.Pronoun == ReferencePronounIts &&
		possessive.Binding == ReferenceBindingSource
}

// entryChoiceSubtypeGroupContent reports whether a group "<group> is/are the
// chosen type in addition to its/their other types" declaration leaves only its
// trailing possessive pronoun as residual content. The group noun phrase is
// consumed by the static-declaration subject, so the body carries a single
// possessive pronoun ("its"/"their") and no other resolving content.
func entryChoiceSubtypeGroupContent(content AbilityContent) bool {
	if len(content.References) != 1 {
		return false
	}
	possessive := content.References[0]
	return possessive.Kind == ReferencePronoun &&
		(possessive.Pronoun == ReferencePronounIts || possessive.Pronoun == ReferencePronounTheir)
}

// staticSyntaxKindsAre reports whether the parser emitted exactly the given
// declaration kinds in order.
func staticSyntaxKindsAre(statics []parser.StaticDeclarationSyntax, kinds ...parser.StaticDeclarationKind) bool {
	actual := make([]parser.StaticDeclarationKind, len(statics))
	for i := range statics {
		actual[i] = statics[i].Kind
	}
	return slices.Equal(actual, kinds)
}

func classifyStaticDeclarationBlocker(ability CompiledAbility) StaticDeclarationBlocker {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 {
		return StaticDeclarationBlockerNone
	}
	if len(ability.Content.Effects) != 1 {
		return StaticDeclarationBlockerNone
	}
	effect := ability.Content.Effects[0]
	rule := staticRuleForEffect(effect.Kind) != StaticRuleUnknown
	if effect.Kind != EffectModifyPT && effect.Kind != EffectGrantKeyword && !rule {
		return StaticDeclarationBlockerNone
	}
	if effect.Duration != DurationNone {
		return StaticDeclarationBlockerDuration
	}
	if len(ability.Content.Conditions) > 1 || (rule && len(ability.Content.Conditions) != 0) {
		return StaticDeclarationBlockerCondition
	}
	if len(ability.Content.Conditions) == 1 &&
		ability.Content.Conditions[0].Predicate == ConditionPredicateUnsupported {
		return StaticDeclarationBlockerCondition
	}
	if rule {
		if len(ability.Content.References) != 1 ||
			ability.Content.References[0].Binding != ReferenceBindingSource {
			return StaticDeclarationBlockerGroup
		}
		return StaticDeclarationBlockerOperation
	}
	if effect.StaticSubject == StaticSubjectNone {
		if len(ability.Content.References) != 1 ||
			ability.Content.References[0].Binding != ReferenceBindingSource {
			return StaticDeclarationBlockerGroup
		}
	}
	if ability.AbilityWord != "" && !recognizedStaticAbilityWord(ability.AbilityWord) {
		return StaticDeclarationBlockerShell
	}
	return StaticDeclarationBlockerOperation
}

func recognizedStaticAbilityWord(word string) bool {
	switch word {
	case "",
		"Coven",
		"Delirium",
		"Domain",
		"Ferocious",
		"Hellbent",
		"Metalcraft",
		"Threshold":
		return true
	default:
		return false
	}
}

func recognizeTypedStaticRuleDeclarations(ability CompiledAbility, syntax *parser.Ability) ([]StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" ||
		len(syntax.Sentences) != 1 ||
		syntax.Sentences[0].StaticRule == nil ||
		len(syntax.Reminders) != 0 ||
		len(syntax.Quoted) != 0 {
		return nil, false
	}
	node := syntax.Sentences[0].StaticRule
	rule, zone, ok := semanticStaticRuleForSyntax(*node)
	if !ok {
		return nil, false
	}
	condition, ok := staticRuleGuardCondition(ability, *node, rule)
	if !ok {
		return nil, false
	}
	group, ok := staticRuleGroupDomain(node.Subject.Kind)
	if !ok {
		return nil, false
	}
	if len(ability.Content.Effects) != 1 ||
		staticRuleForEffect(ability.Content.Effects[0].Kind) != rule ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingSource {
		return nil, false
	}
	return []StaticDeclaration{staticRuleDeclaration(node.Span, node.Subject.Span, node.Operation.Span, rule, zone, group, staticBlockerRestrictionForSyntax(*node), condition)}, true
}

// staticRuleGuardCondition pairs a static rule's trailing guard clause (its
// Guarded flag, e.g. "unless you control seven or more lands.") with the single
// supported compiled condition the condition machinery produced for it. An
// unguarded rule must carry no conditions. A guarded rule is supported only for
// the land-gated can't-attack-or-block restriction (Topiary Stomper) and must
// carry exactly one supported condition; every other guarded rule fails closed
// so the broadening stays narrow and text-blind.
func staticRuleGuardCondition(ability CompiledAbility, node parser.StaticRuleSyntax, rule StaticRuleKind) (*CompiledCondition, bool) {
	if !node.Guarded {
		return nil, len(ability.Content.Conditions) == 0
	}
	if rule != StaticRuleCantAttackOrBlock {
		return nil, false
	}
	if len(ability.Content.Conditions) != 1 {
		return nil, false
	}
	condition := &ability.Content.Conditions[0]
	if condition.Predicate == ConditionPredicateUnsupported || condition.Resolving {
		return nil, false
	}
	return condition, true
}

// staticRuleGroupDomain maps a parsed static rule subject to the affected group
// domain. Source subjects affect the object itself; an Aura or Equipment subject
// ("enchanted creature"/"equipped creature") affects the object it is attached to.
func staticRuleGroupDomain(kind parser.StaticRuleSubjectKind) (StaticGroupDomain, bool) {
	switch kind {
	case parser.StaticRuleSubjectSourceCreature, parser.StaticRuleSubjectSourcePermanent, parser.StaticRuleSubjectSourceSpell:
		return StaticGroupSource, true
	case parser.StaticRuleSubjectAttachedObject:
		return StaticGroupAttachedObject, true
	case parser.StaticRuleSubjectControlledCreatures:
		return StaticGroupSourceControllerPermanents, true
	case parser.StaticRuleSubjectBattlefieldCreatures:
		return StaticGroupBattlefield, true
	default:
		return StaticGroupUnknown, false
	}
}

// isCreatureRuleSubject reports whether a static rule subject scopes a creature:
// either the source creature itself, the creature an Aura or Equipment is
// attached to, or the creatures the source's controller controls. Combat and
// untap rule operations apply to either.
func isCreatureRuleSubject(kind parser.StaticRuleSubjectKind) bool {
	switch kind {
	case parser.StaticRuleSubjectSourceCreature, parser.StaticRuleSubjectAttachedObject,
		parser.StaticRuleSubjectControlledCreatures, parser.StaticRuleSubjectBattlefieldCreatures,
		parser.StaticRuleSubjectOpponentControlledCreatures:
		return true
	default:
		return false
	}
}

func isUntapRuleSubject(kind parser.StaticRuleSubjectKind) bool {
	return isCreatureRuleSubject(kind) || kind == parser.StaticRuleSubjectSourcePermanent
}

func semanticStaticRuleForSyntax(rule parser.StaticRuleSyntax) (StaticRuleKind, StaticZone, bool) {
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		len(rule.Qualifiers) == 0 {
		switch rule.Operation.Voice {
		case parser.StaticRuleVoiceActive:
			return StaticRuleCantBlock, StaticZoneBattlefield, true
		case parser.StaticRuleVoicePassive:
			return StaticRuleCantBeBlocked, StaticZoneBattlefield, true
		default:
			return StaticRuleUnknown, StaticZoneBattlefield, false
		}
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierByMoreThanOne) {
		return StaticRuleCantBeBlockedByMoreThanOne, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticBlockerRestrictionForSyntax(rule).Kind != StaticBlockerRestrictionNone {
		return StaticRuleCantBeBlockedByCreaturesWith, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantAttack, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierDefenderYou) {
		return StaticRuleCantAttackYou, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttackOrBlock &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantAttackOrBlock, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlockAndBeBlocked &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantBlockAndCantBeBlocked, StaticZoneBattlefield, true
	}
	if isUntapRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationUntap &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleDoesntUntap, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintRequirement &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierEachCombat, parser.StaticRuleQualifierIfAble) {
		return StaticRuleMustAttack, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintRequirement &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierIfAble) {
		return StaticRuleMustBeBlocked, StaticZoneBattlefield, true
	}
	if rule.Subject.Kind == parser.StaticRuleSubjectSourceSpell &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationCounter &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantBeCountered, StaticZoneStack, true
	}
	return StaticRuleUnknown, StaticZoneBattlefield, false
}

// staticBlockerRestrictionForSyntax derives the closed blocker characteristic
// from a parsed passive block prohibition's qualifiers. A non-None result names
// a "can't be blocked by creatures with ..." restriction; an absent or
// unrecognized qualifier yields StaticBlockerRestrictionNone.
func staticBlockerRestrictionForSyntax(rule parser.StaticRuleSyntax) StaticBlockerRestriction {
	if len(rule.Qualifiers) != 1 {
		return StaticBlockerRestriction{}
	}
	qualifier := rule.Qualifiers[0]
	switch qualifier.Kind {
	case parser.StaticRuleQualifierBlockerFlying:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionFlying}
	case parser.StaticRuleQualifierBlockerPowerOrLess:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrLess, Amount: qualifier.Amount}
	case parser.StaticRuleQualifierBlockerPowerOrGreater:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrGreater, Amount: qualifier.Amount}
	case parser.StaticRuleQualifierBlockerColor:
		runtimeColor, ok := compilerColor(qualifier.Color)
		if !ok {
			return StaticBlockerRestriction{}
		}
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionColor, Color: runtimeColor}
	case parser.StaticRuleQualifierBlockerArtifact:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionArtifact}
	default:
		return StaticBlockerRestriction{}
	}
}

func staticRuleForEffect(kind EffectKind) StaticRuleKind {
	switch kind {
	case EffectCantBlock:
		return StaticRuleCantBlock
	case EffectCantBeBlocked:
		return StaticRuleCantBeBlocked
	case EffectCantAttack:
		return StaticRuleCantAttack
	case EffectMustAttack:
		return StaticRuleMustAttack
	case EffectMustBeBlocked:
		return StaticRuleMustBeBlocked
	case EffectCantBeCountered:
		return StaticRuleCantBeCountered
	case EffectCantBeBlockedByCreaturesWith:
		return StaticRuleCantBeBlockedByCreaturesWith
	case EffectCantBeBlockedByMoreThanOne:
		return StaticRuleCantBeBlockedByMoreThanOne
	case EffectCantAttackOrBlock:
		return StaticRuleCantAttackOrBlock
	case EffectCantBlockAndCantBeBlocked:
		return StaticRuleCantBlockAndCantBeBlocked
	case EffectDoesntUntap:
		return StaticRuleDoesntUntap
	default:
		return StaticRuleUnknown
	}
}

func staticRuleDeclaration(
	span, subjectSpan, operationSpan shared.Span,
	rule StaticRuleKind,
	zone StaticZone,
	group StaticGroupDomain,
	blocker StaticBlockerRestriction,
	condition *CompiledCondition,
) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationRule,
		Span:          span,
		OperationSpan: operationSpan,
		Group: StaticGroupReference{
			Span:   subjectSpan,
			Domain: group,
		},
		Condition: condition,
		Rule: &StaticRuleDeclaration{
			Domain:  staticRuleDomain(rule),
			Kind:    rule,
			Zone:    zone,
			Blocker: blocker,
		},
	}
}

func staticRuleDomain(rule StaticRuleKind) StaticRuleDomain {
	switch rule {
	case StaticRuleCantAttack, StaticRuleMustAttack, StaticRuleCantAttackYou:
		return StaticRuleDomainAttack
	case StaticRuleCantBlock, StaticRuleCantBeBlocked, StaticRuleMustBeBlocked, StaticRuleCantBeBlockedByMoreThanOne,
		StaticRuleCantBeBlockedByCreaturesWith, StaticRuleCantBlockAndCantBeBlocked:
		return StaticRuleDomainBlock
	case StaticRuleCantBeCountered:
		return StaticRuleDomainCountering
	case StaticRuleCantAttackOrBlock:
		return StaticRuleDomainAttackBlock
	case StaticRuleDoesntUntap:
		return StaticRuleDomainUntap
	default:
		return StaticRuleDomainUnknown
	}
}

func recognizeMixedSourceStaticDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant,
		parser.StaticDeclarationRule) {
		return nil, false
	}
	rule, _, ok := semanticStaticRuleForSyntax(statics[2].Rule)
	if !ok || rule != StaticRuleMustAttack {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate == ConditionPredicateUnsupported ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingSource ||
		len(ability.Content.Keywords) == 0 {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	condition := &ability.Content.Conditions[0]
	group := StaticGroupReference{Span: ability.Content.References[0].Span, Domain: StaticGroupSource}
	return []StaticDeclaration{
		staticPTDeclaration(ability.Span, group, condition, effect),
		staticKeywordGrantDeclaration(ability.Span, group, condition, ability.Content.Keywords),
		staticRuleDeclaration(ability.Span, group.Span, ability.Span, StaticRuleMustAttack, StaticZoneBattlefield, StaticGroupSource, StaticBlockerRestriction{}, condition),
	}, true
}

// recognizeStaticLoseAbilitiesBecomeDeclaration maps the polymorph syntax
// "<subject> loses all abilities [and is/has ...]" onto layer-faithful semantic
// declarations: a remove-all-abilities ability-layer declaration, plus optional
// set-color, set-type, set-subtype, and base power/toughness declarations. The
// affected object's existing colors, card types, and creature types are replaced
// (set), so the colors and types travel as set operations rather than additions.
func recognizeStaticLoseAbilitiesBecomeDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationLoseAbilitiesBecome) {
		return nil, false
	}
	node := &statics[0]
	if !node.LoseAllAbilities {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.AbilityWord != "" {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(node.Subject)
	if !ok {
		return nil, false
	}
	declarations := []StaticDeclaration{{
		Kind:          StaticDeclarationContinuous,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerAbility,
			Operation: StaticContinuousRemoveAllAbilities,
		},
	}}
	if node.BecomeColorless {
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:        StaticLayerColor,
				Operation:    StaticContinuousSetColors,
				SetColorless: true,
			},
		})
	} else if len(node.Colors) != 0 {
		colors, ok := staticRuntimeColors(node.Colors)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerColor,
				Operation: StaticContinuousSetColors,
				Colors:    colors,
			},
		})
	}
	if len(node.CardTypes) != 0 || len(node.Subtypes) != 0 {
		cardTypes, ok := staticCardTypesFromParser(node.CardTypes)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:       StaticLayerType,
				Operation:   StaticContinuousSetTypes,
				SetTypes:    cardTypes,
				SetSubtypes: slices.Clone(node.Subtypes),
			},
		})
	}
	if node.BasePTSet {
		declarations = append(declarations, staticBasePowerToughnessDeclaration(node.Span, node, group, nil))
	}
	if keywords := staticDeclarationGrantKeywords(ability.Content); len(keywords) != 0 {
		declarations = append(declarations, staticKeywordGrantDeclaration(node.Span, group, nil, keywords))
	}
	return declarations, true
}

// recognizeStaticEnchantedTypeChangeDeclaration maps the removal-Aura syntax
// "<attached subject> is [colorless] <types> [with '<mana ability>' | with base
// power and toughness N/N] [and loses all other abilities]" onto layer-faithful
// semantic declarations: an optional remove-all-abilities ability-layer
// declaration, an optional make-colorless color-layer declaration, a
// set-type/subtype type-layer declaration, an optional base-power/toughness-set
// declaration, and an optional granted mana ability. The remove-all-abilities
// declaration precedes the granted ability so the ability survives the loss
// within the ability layer.
func recognizeStaticEnchantedTypeChangeDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationEnchantedTypeChange) {
		return nil, false
	}
	node := &statics[0]
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.AbilityWord != "" {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(node.Subject)
	if !ok {
		return nil, false
	}
	var declarations []StaticDeclaration
	if node.LoseAllAbilities {
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerAbility,
				Operation: StaticContinuousRemoveAllAbilities,
			},
		})
	}
	if node.BecomeColorless {
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:        StaticLayerColor,
				Operation:    StaticContinuousSetColors,
				SetColorless: true,
			},
		})
	} else if len(node.Colors) != 0 {
		colors, ok := staticRuntimeColors(node.Colors)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerColor,
				Operation: StaticContinuousSetColors,
				Colors:    colors,
			},
		})
	}
	if len(node.CardTypes) != 0 || len(node.Subtypes) != 0 {
		cardTypes, ok := staticCardTypesFromParser(node.CardTypes)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:       StaticLayerType,
				Operation:   StaticContinuousSetTypes,
				SetTypes:    cardTypes,
				SetSubtypes: slices.Clone(node.Subtypes),
			},
		})
	}
	if node.BasePTSet {
		declarations = append(declarations, staticBasePowerToughnessDeclaration(node.Span, node, group, nil))
	}
	if granted := node.GrantedManaAbility; granted != nil {
		if !granted.TapCost || !staticGrantedManaAbilityValid(granted) {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerAbility,
				Operation: StaticContinuousGrantManaAbility,
				GrantedMana: &StaticGrantedManaAbility{
					TapCost:     granted.TapCost,
					Amount:      granted.Amount,
					AnyColor:    granted.AnyColor,
					Text:        granted.Text,
					Sacrifice:   granted.Sacrifice,
					AnyOneColor: granted.AnyOneColor,
					Colorless:   granted.Colorless,
				},
			},
		})
	}
	if len(declarations) == 0 {
		return nil, false
	}
	return declarations, true
}

func recognizeStaticPowerToughnessDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	plain := staticSyntaxKindsAre(statics, parser.StaticDeclarationContinuousPowerToughness)
	withKeywords := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant)
	if !plain && !withKeywords {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	if statics[0].Dynamic != (effect.Amount.DynamicKind != DynamicAmountNone) {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if len(keywords) == 0 {
		if !plain {
			return nil, false
		}
	} else if !withKeywords {
		return nil, false
	}
	declarations := []StaticDeclaration{staticPTDeclaration(ability.Span, group.Group, condition, effect)}
	if withKeywords {
		declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group.Group, condition, keywords))
	}
	return declarations, true
}

// recognizeStaticPowerToughnessRuleDeclarations maps a paragraph that composes a
// power/toughness modification (optionally with a keyword grant) and a single
// creature-scoped rule operation onto closed semantic declarations, e.g.
// "Enchanted creature gets +2/+2 and can't block." The resolving content carries
// only the power/toughness effect, so the rule operation derives from the typed
// parser node; the affected group derives from the resolving effect, keeping the
// mapping text-blind. Conditional compounds fail closed because static rule
// effects are recognized only without a condition.
func recognizeStaticPowerToughnessRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	plain := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationRule)
	withKeywords := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant,
		parser.StaticDeclarationRule)
	if !plain && !withKeywords {
		return nil, false
	}
	ruleNode := &statics[len(statics)-1]
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	if statics[0].Dynamic != (effect.Amount.DynamicKind != DynamicAmountNone) {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if (len(keywords) != 0) != withKeywords {
		return nil, false
	}
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	ruleGroup, ok := staticRuleGroupDomain(ruleNode.Rule.Subject.Kind)
	if !ok || ruleGroup != group.Group.Domain {
		return nil, false
	}
	declarations := []StaticDeclaration{staticPTDeclaration(ability.Span, group.Group, nil, effect)}
	if withKeywords {
		declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group.Group, nil, keywords))
	}
	declarations = append(declarations, staticRuleDeclaration(ability.Span, group.Group.Span, ruleNode.OperationSpan, rule, zone, group.Group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil))
	return declarations, true
}

// recognizeStaticKeywordGrantRuleDeclarations maps a paragraph that composes a
// keyword grant and a single creature-scoped rule operation, without any
// power/toughness change, onto closed semantic declarations, e.g. "Enchanted
// creature has trample and can't be blocked by more than one creature." The
// resolving content carries only the keyword-grant effect, so the rule operation
// derives from the typed parser node; the affected group derives from the
// resolving effect, keeping the mapping text-blind. Conditional compounds fail
// closed because static rule effects are recognized only without a condition.
// staticKeywordGrantRulePair recognizes a two-declaration group composed of one
// keyword grant and one static rule sharing an affected group, in either source
// order ("Equipped creature has shroud and can't be blocked." or "Equipped
// creature can't be blocked and has shroud."), returning the rule node. The
// keyword grant's payload is read from the compiled effect, so only the rule
// node position matters here.
func staticKeywordGrantRulePair(statics []parser.StaticDeclarationSyntax) (*parser.StaticDeclarationSyntax, bool) {
	if len(statics) != 2 {
		return nil, false
	}
	var ruleNode *parser.StaticDeclarationSyntax
	keywordGrants := 0
	for i := range statics {
		switch statics[i].Kind {
		case parser.StaticDeclarationKeywordGrant:
			keywordGrants++
		case parser.StaticDeclarationRule:
			ruleNode = &statics[i]
		default:
			return nil, false
		}
	}
	if keywordGrants != 1 || ruleNode == nil {
		return nil, false
	}
	return ruleNode, true
}

func recognizeStaticKeywordGrantRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	ruleNode, ok := staticKeywordGrantRulePair(statics)
	if !ok {
		return nil, false
	}
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectGrantKeyword ||
		ability.Content.Effects[0].Duration != DurationNone {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if len(keywords) == 0 {
		return nil, false
	}
	ruleGroup, ok := staticRuleGroupDomain(ruleNode.Rule.Subject.Kind)
	if !ok {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	group, ok := staticKeywordGrantRuleGroup(ability, effect, ruleNode, ruleGroup)
	if !ok {
		return nil, false
	}
	return []StaticDeclaration{
		staticKeywordGrantDeclaration(ability.Span, group, nil, keywords),
		staticRuleDeclaration(ability.Span, group.Span, ruleNode.OperationSpan, rule, zone, ruleGroup, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil),
	}, true
}

// recognizeStaticControlledGroupRuleDeclarations maps a standalone group-scoped
// static rule onto a closed semantic declaration, e.g. "Creatures you control
// can't be blocked." The rule has no resolving content effect, so the affected
// group derives entirely from the typed parser rule subject: the
// controlled-creatures subject yields a controller-permanents group restricted
// to creatures. Costs, triggers, conditions, or any resolving content fail
// closed because a continuous group rule carries none.
func recognizeStaticControlledGroupRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationRule) {
		return nil, false
	}
	ruleNode := &statics[0]
	if ruleNode.Rule.Subject.Kind != parser.StaticRuleSubjectControlledCreatures {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(ruleNode.Subject)
	if !ok || group.Domain != StaticGroupSourceControllerPermanents {
		return nil, false
	}
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 {
		return nil, false
	}
	declaration := staticRuleDeclaration(ability.Span, group.Span, ruleNode.OperationSpan, rule, zone, group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil)
	declaration.Group = group
	return []StaticDeclaration{declaration}, true
}

// recognizeStaticBattlefieldBlockRuleDeclarations maps a standalone
// battlefield-scoped "can't block" restriction onto a closed semantic
// declaration, e.g. "Creatures with power less than this creature's power can't
// block it." or "... can't block creatures you control." The battlefield-creatures
// subject yields an every-creature affected group narrowed by the typed parser
// rule subject (its source-relative power filter), and the protected object the
// restriction shields travels in the rule declaration's BlockedObject. Costs,
// triggers, conditions, or any resolving content fail closed because a continuous
// group rule carries none.
func recognizeStaticBattlefieldBlockRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationRule) {
		return nil, false
	}
	ruleNode := &statics[0]
	if ruleNode.Rule.Subject.Kind != parser.StaticRuleSubjectBattlefieldCreatures {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(ruleNode.Subject)
	if !ok || group.Domain != StaticGroupBattlefield {
		return nil, false
	}
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok || rule != StaticRuleCantBlock {
		return nil, false
	}
	blocked, ok := compileStaticBlockedObject(ruleNode.Rule.BlockedObject)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 {
		return nil, false
	}
	declaration := staticRuleDeclaration(ability.Span, group.Span, ruleNode.OperationSpan, rule, zone, group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil)
	declaration.Group = group
	declaration.Rule.BlockedObject = blocked
	return []StaticDeclaration{declaration}, true
}

// recognizeStaticGroupMustAttackDeclarations maps a standalone battlefield- or
// opponent-scoped forced-attack requirement onto a closed semantic declaration,
// e.g. "All creatures attack each combat if able." or "Creatures your opponents
// control attack each combat if able." The affected group derives entirely from
// the typed parser subject: the all-creatures subject yields an every-creature
// group and the opponent-controlled subject yields a battlefield group whose
// affected-permanent Selection scopes the controller to the opponent relation.
// The controller-scoped "Creatures you control" form is handled by
// recognizeStaticControlledGroupRuleDeclarations. Costs, triggers, conditions, or
// any resolving content fail closed because a continuous group rule carries none.
func recognizeStaticGroupMustAttackDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationRule) {
		return nil, false
	}
	ruleNode := &statics[0]
	switch ruleNode.Rule.Subject.Kind {
	case parser.StaticRuleSubjectBattlefieldCreatures, parser.StaticRuleSubjectOpponentControlledCreatures:
	default:
		return nil, false
	}
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok || rule != StaticRuleMustAttack {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(ruleNode.Subject)
	if !ok || group.Domain != StaticGroupBattlefield {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 {
		return nil, false
	}
	declaration := staticRuleDeclaration(ability.Span, group.Span, ruleNode.OperationSpan, rule, zone, group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil)
	declaration.Group = group
	return []StaticDeclaration{declaration}, true
}

// compileStaticBlockedObject maps the parser's typed "can't block" protected
// object onto its compiler scope. It fails closed for an unrepresentable scope.
func compileStaticBlockedObject(kind parser.StaticRuleBlockedObjectKind) (StaticBlockedObjectKind, bool) {
	switch kind {
	case parser.StaticRuleBlockedObjectNone:
		return StaticBlockedObjectNone, true
	case parser.StaticRuleBlockedObjectSource:
		return StaticBlockedObjectSource, true
	case parser.StaticRuleBlockedObjectControlledCreatures:
		return StaticBlockedObjectControlledCreatures, true
	default:
		return StaticBlockedObjectNone, false
	}
}

// can't be blocked and has shroud."); in that case it recovers the group from
// the rule node's subject, which the parser resolves independently of clause
// order. The recovered group is limited to the attached-object domain so a
// source-affecting grant keeps flowing through its reference-bound path.
func staticKeywordGrantRuleGroup(ability CompiledAbility, effect *CompiledEffect, ruleNode *parser.StaticDeclarationSyntax, ruleGroup StaticGroupDomain) (StaticGroupReference, bool) {
	if effect.StaticSubject != StaticSubjectNone {
		result, ok := staticDeclarationEffectGroup(ability, effect)
		if !ok || result.Group.Domain != ruleGroup {
			return StaticGroupReference{}, false
		}
		return result.Group, true
	}
	if ruleGroup != StaticGroupAttachedObject || len(ability.Content.References) != 0 {
		return StaticGroupReference{}, false
	}
	return StaticGroupReference{Span: ruleNode.Subject.Span, Domain: StaticGroupAttachedObject}, true
}

// shared affected group with one or more layer-preserving characteristic changes
// onto closed semantic declarations. It recognizes power/toughness modification,
// base power/toughness setting, keyword grants, and color/type characteristic
// additions, requiring at least one base-power/toughness or characteristic node so
// the simpler single-family recognizers keep ownership of their shapes. The group
// and payload derive from the typed parser nodes and already-resolved content
// only; no Oracle text is inspected.
func recognizeStaticComposedContinuousDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if len(statics) == 0 {
		return nil, false
	}
	ptNodes := 0
	keywordNodes := 0
	newNodes := 0
	for i := range statics {
		switch statics[i].Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			ptNodes++
		case parser.StaticDeclarationKeywordGrant:
			keywordNodes++
		case parser.StaticDeclarationContinuousBasePowerToughness,
			parser.StaticDeclarationContinuousCharacteristic:
			newNodes++
		case parser.StaticDeclarationRule:
		default:
			return nil, false
		}
	}
	if newNodes == 0 {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	subject := statics[0].Subject
	for i := range statics {
		if !staticSubjectsEquivalent(statics[i].Subject, subject) {
			return nil, false
		}
	}
	group, ok := staticGroupForParserSubject(subject)
	if !ok {
		return nil, false
	}
	// Cross-check the resolving content shape against the typed operations. The
	// "has base power and toughness" verb yields an empty keyword-grant effect
	// shell with no keywords, which is tolerated only when no keyword node is
	// present.
	modifyPT := 0
	for i := range ability.Content.Effects {
		switch ability.Content.Effects[i].Kind {
		case EffectModifyPT:
			modifyPT++
		case EffectGrantKeyword:
		default:
			return nil, false
		}
	}
	if modifyPT != ptNodes {
		return nil, false
	}
	if (keywordNodes > 0) != (len(ability.Content.Keywords) > 0) {
		return nil, false
	}
	if keywordNodes > 1 {
		return nil, false
	}
	var ptEffect *CompiledEffect
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].Kind == EffectModifyPT {
			ptEffect = &ability.Content.Effects[i]
		}
	}
	keywordsEmitted := false
	var declarations []StaticDeclaration
	for i := range statics {
		node := &statics[i]
		switch node.Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			if ptEffect == nil ||
				!ptEffect.PowerDelta.Known ||
				!ptEffect.ToughnessDelta.Known ||
				ptEffect.Duration != DurationNone {
				return nil, false
			}
			if node.Dynamic != (ptEffect.Amount.DynamicKind != DynamicAmountNone) {
				return nil, false
			}
			declarations = append(declarations, staticPTDeclaration(ability.Span, group, condition, ptEffect))
		case parser.StaticDeclarationKeywordGrant:
			if keywordsEmitted || len(ability.Content.Keywords) == 0 {
				return nil, false
			}
			keywordsEmitted = true
			declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group, condition, ability.Content.Keywords))
		case parser.StaticDeclarationContinuousBasePowerToughness:
			if !node.BasePTSet {
				return nil, false
			}
			declarations = append(declarations, staticBasePowerToughnessDeclaration(ability.Span, node, group, condition))
		case parser.StaticDeclarationContinuousCharacteristic:
			characteristic, ok := staticCharacteristicDeclarations(ability.Span, node, group, condition)
			if !ok {
				return nil, false
			}
			declarations = append(declarations, characteristic...)
		case parser.StaticDeclarationRule:
			rule, zone, ok := semanticStaticRuleForSyntax(node.Rule)
			if !ok {
				return nil, false
			}
			ruleGroup, ok := staticRuleGroupDomain(node.Rule.Subject.Kind)
			if !ok || ruleGroup != group.Domain {
				return nil, false
			}
			declarations = append(declarations, staticRuleDeclaration(ability.Span, group.Span, node.OperationSpan, rule, zone, ruleGroup, staticBlockerRestrictionForSyntax(node.Rule), condition))
		default:
			return nil, false
		}
	}
	if len(declarations) == 0 {
		return nil, false
	}
	return declarations, true
}

// recognizeStaticQuotedAbilityGrantDeclarations maps a static grant that confers
// a full quoted triggered or activated ability ("Equipped creature has '<quoted
// ability>'.") onto closed semantic declarations. The grant may be preceded by
// an optional power/toughness modification and an optional keyword grant sharing
// the same subject ("Equipped creature gets +1/+1 and has trample and '<quoted
// ability>'."). The affected group derives from the typed subject node, the
// power/toughness and keyword payloads derive from the resolving content, and
// the quoted ability is carried verbatim for the lowering to compile and lower
// into the continuous effect's granted-ability set. The resolving content's
// "has" verb yields an empty keyword-grant shell that is tolerated.
func recognizeStaticQuotedAbilityGrantDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if len(statics) == 0 {
		return nil, false
	}
	ptNodes := 0
	keywordNodes := 0
	grantNodes := 0
	for i := range statics {
		switch statics[i].Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			ptNodes++
		case parser.StaticDeclarationKeywordGrant:
			keywordNodes++
		case parser.StaticDeclarationContinuousQuotedAbilityGrant:
			grantNodes++
		default:
			return nil, false
		}
	}
	if grantNodes != 1 || ptNodes > 1 || keywordNodes > 1 {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		ability.AbilityWord != "" {
		return nil, false
	}
	subject := statics[0].Subject
	for i := range statics {
		if !staticSubjectsEquivalent(statics[i].Subject, subject) {
			return nil, false
		}
	}
	if !staticParserSubjectReferencesTolerated(ability.Content.References, subject) {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(subject)
	if !ok {
		return nil, false
	}
	modifyPT := 0
	for i := range ability.Content.Effects {
		switch ability.Content.Effects[i].Kind {
		case EffectModifyPT:
			modifyPT++
		case EffectGrantKeyword:
		default:
			return nil, false
		}
	}
	if modifyPT != ptNodes || (keywordNodes > 0) != (len(ability.Content.Keywords) > 0) {
		return nil, false
	}
	var ptEffect *CompiledEffect
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].Kind == EffectModifyPT {
			ptEffect = &ability.Content.Effects[i]
		}
	}
	return buildStaticQuotedAbilityGrantDeclarations(ability, statics, group, ptEffect)
}

// buildStaticQuotedAbilityGrantDeclarations emits the closed declarations for a
// recognized quoted-ability grant: the optional power/toughness modification,
// the optional keyword grant, and the quoted ability grant itself, all sharing
// the subject-derived group.
func buildStaticQuotedAbilityGrantDeclarations(
	ability CompiledAbility,
	statics []parser.StaticDeclarationSyntax,
	group StaticGroupReference,
	ptEffect *CompiledEffect,
) ([]StaticDeclaration, bool) {
	keywordsEmitted := false
	var declarations []StaticDeclaration
	for i := range statics {
		node := &statics[i]
		switch node.Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			if ptEffect == nil ||
				!ptEffect.PowerDelta.Known ||
				!ptEffect.ToughnessDelta.Known ||
				ptEffect.Duration != DurationNone ||
				node.Dynamic != (ptEffect.Amount.DynamicKind != DynamicAmountNone) {
				return nil, false
			}
			declarations = append(declarations, staticPTDeclaration(ability.Span, group, nil, ptEffect))
		case parser.StaticDeclarationKeywordGrant:
			if keywordsEmitted || len(ability.Content.Keywords) == 0 {
				return nil, false
			}
			keywordsEmitted = true
			declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group, nil, ability.Content.Keywords))
		case parser.StaticDeclarationContinuousQuotedAbilityGrant:
			if node.GrantedAbility == nil {
				return nil, false
			}
			declarations = append(declarations, staticQuotedAbilityGrantDeclaration(ability.Span, group, node))
		default:
			return nil, false
		}
	}
	return declarations, true
}

// staticQuotedAbilityGrantDeclaration builds the ability-layer grant declaration
// that confers a quoted triggered or activated ability on its subject. The
// parsed quoted ability is carried unchanged so the lowering compiles and lowers
// its inner document into the continuous effect's granted-ability set.
func staticQuotedAbilityGrantDeclaration(span shared.Span, group StaticGroupReference, node *parser.StaticDeclarationSyntax) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Continuous: &StaticContinuousDeclaration{
			Layer:          StaticLayerAbility,
			Operation:      StaticContinuousGrantAbility,
			GrantedAbility: node.GrantedAbility,
		},
	}
}

// staticSubjectsEquivalent reports whether two typed parser subjects name the
// same affected group. It compares only typed identity fields and ignores source
// spans so recognition stays position-blind.
func staticSubjectsEquivalent(a, b parser.StaticDeclarationSubject) bool {
	return a.Kind == b.Kind &&
		a.CardFilter == b.CardFilter &&
		a.Group.Kind == b.Group.Kind &&
		a.Group.Subtype == b.Group.Subtype &&
		a.Group.SubtypeKnown == b.Group.SubtypeKnown &&
		a.Group.Colorless == b.Group.Colorless &&
		a.Group.Multicolored == b.Group.Multicolored &&
		a.Group.ChosenColorFromEntry == b.Group.ChosenColorFromEntry &&
		slices.Equal(a.Group.Colors, b.Group.Colors)
}

// staticGroupForParserSubject maps a typed parser subject onto the affected group
// reference, failing closed for subjects whose runtime group is not representable.
func staticGroupForParserSubject(subject parser.StaticDeclarationSubject) (StaticGroupReference, bool) {
	switch subject.Kind {
	case parser.StaticDeclarationSubjectSourceCreature,
		parser.StaticDeclarationSubjectSourceNamed:
		return StaticGroupReference{Span: subject.Span, Domain: StaticGroupSource}, true
	case parser.StaticDeclarationSubjectGroup:
		kind := compileStaticSubjectKind(subject.Group.Kind)
		if kind == StaticSubjectNone {
			return StaticGroupReference{}, false
		}
		group, ok := staticGroupForSubject(kind, subject.Group.Span, subject.Group.Subtype, subject.Group.SubtypeKnown, staticColorFilter{
			Colors:       subject.Group.Colors,
			Colorless:    subject.Group.Colorless,
			Multicolored: subject.Group.Multicolored,
		}, parser.KeywordUnknown, parser.KeywordUnknown, subject.Group.ChosenColorFromEntry)
		if ok && subject.Group.CounterRequired {
			if subject.Group.CounterAny {
				group.Selection.MatchAnyCounter = true
			} else {
				group.Selection.MatchCounter = true
				group.Selection.RequiredCounter = subject.Group.CounterKind
			}
		}
		if ok {
			group.Selection.Power = subject.Group.Power
			group.Selection.MatchPower = subject.Group.MatchPower
			group.Selection.Toughness = subject.Group.Toughness
			group.Selection.MatchToughness = subject.Group.MatchToughness
			group.Selection.PowerOrToughness = subject.Group.PowerOrToughness
			group.Selection.PowerLessThanSource = subject.Group.PowerLessThanSource
			group.Selection.PowerGreaterThanSource = subject.Group.PowerGreaterThanSource
		}
		if ok && len(subject.Group.ExcludedTypes) > 0 {
			excluded, mapped := staticCardTypesFromParser(subject.Group.ExcludedTypes)
			if !mapped {
				return StaticGroupReference{}, false
			}
			group.Selection.ExcludedTypes = excluded
		}
		return group, ok
	default:
		return StaticGroupReference{}, false
	}
}

// staticBasePowerToughnessDeclaration builds a base power/toughness setting
// declaration from the typed parser payload.
func staticBasePowerToughnessDeclaration(span shared.Span, node *parser.StaticDeclarationSyntax, group StaticGroupReference, condition *CompiledCondition) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:        StaticLayerPowerToughnessSet,
			Operation:    StaticContinuousSetBasePowerToughness,
			SetPower:     node.BasePower,
			SetToughness: node.BaseToughness,
		},
	}
}

// recognizeStaticCharacteristicPowerToughnessDeclaration maps the parser's
// characteristic-defining power/toughness declaration ("<source>'s power and
// toughness are each equal to <count>") onto a closed semantic declaration. The
// declaration sets only the source object's power and toughness, so the ability
// shell must carry no resolving content; group subjects fail closed.
func recognizeStaticCharacteristicPowerToughnessDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCharacteristicDefiningPowerToughness) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 {
		return StaticDeclaration{}, false
	}
	node := &statics[0]
	if node.Subject.Kind != parser.StaticDeclarationSubjectSourceCreature &&
		node.Subject.Kind != parser.StaticDeclarationSubjectSourceNamed {
		return StaticDeclaration{}, false
	}
	value, ok := compileStaticDynamicValueKind(node.DynamicValue)
	if !ok {
		return StaticDeclaration{}, false
	}
	if !node.DynamicSetsPower && !node.DynamicSetsToughness {
		return StaticDeclaration{}, false
	}
	var countColor color.Color
	if node.DynamicValueColor != "" {
		countColor, ok = compilerColor(node.DynamicValueColor)
		if !ok {
			return StaticDeclaration{}, false
		}
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCharacteristicPowerToughness,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		Group:         StaticGroupReference{Span: node.Subject.Span, Domain: StaticGroupSource},
		CharacteristicPT: &StaticCharacteristicPowerToughnessDeclaration{
			Value:           value,
			Subtype:         node.DynamicValueSubtype,
			Color:           countColor,
			SetsPower:       node.DynamicSetsPower,
			SetsToughness:   node.DynamicSetsToughness,
			ToughnessOffset: node.DynamicToughnessOffset,
		},
	}, true
}

// compileStaticDynamicValueKind maps a parser characteristic-defining count kind
// onto its runtime dynamic-value kind. It fails closed for unrepresentable kinds.
func compileStaticDynamicValueKind(kind parser.StaticDeclarationDynamicValueKind) (game.DynamicValueKind, bool) {
	switch kind {
	case parser.StaticDeclarationDynamicValueControllerHandSize:
		return game.DynamicValueControllerHandSize, true
	case parser.StaticDeclarationDynamicValueControllerGraveyardSize:
		return game.DynamicValueControllerGraveyardSize, true
	case parser.StaticDeclarationDynamicValueControllerCreatureCount:
		return game.DynamicValueControllerCreatureCount, true
	case parser.StaticDeclarationDynamicValueControllerLandCount:
		return game.DynamicValueControllerLandCount, true
	case parser.StaticDeclarationDynamicValueControllerArtifactCount:
		return game.DynamicValueControllerArtifactCount, true
	case parser.StaticDeclarationDynamicValueAllBattlefieldCreatureCount:
		return game.DynamicValueAllBattlefieldCreatureCount, true
	case parser.StaticDeclarationDynamicValueAllGraveyardsSize:
		return game.DynamicValueAllGraveyardsSize, true
	case parser.StaticDeclarationDynamicValueCreatureCardsInAllGraveyards:
		return game.DynamicValueCreatureCardsInAllGraveyards, true
	case parser.StaticDeclarationDynamicValueCardTypesAmongAllGraveyards:
		return game.DynamicValueCardTypesAmongAllGraveyards, true
	case parser.StaticDeclarationDynamicValueControllerCreatureCardsInGraveyard:
		return game.DynamicValueControllerCreatureCardsInGraveyard, true
	case parser.StaticDeclarationDynamicValueControllerInstantOrSorceryCardsInGraveyard:
		return game.DynamicValueControllerInstantOrSorceryCardsInGraveyard, true
	case parser.StaticDeclarationDynamicValueControllerLandCardsInGraveyard:
		return game.DynamicValueControllerLandCardsInGraveyard, true
	case parser.StaticDeclarationDynamicValueControllerCardTypesInGraveyard:
		return game.DynamicValueControllerCardTypesInGraveyard, true
	case parser.StaticDeclarationDynamicValueControllerPermanentCardsInGraveyard:
		return game.DynamicValueControllerPermanentCardsInGraveyard, true
	case parser.StaticDeclarationDynamicValueControllerSubtypeCount:
		return game.DynamicValueControllerSubtypeCount, true
	case parser.StaticDeclarationDynamicValueControllerBasicLandTypeCount:
		return game.DynamicValueControllerBasicLandTypeCount, true
	case parser.StaticDeclarationDynamicValueControllerLifeTotal:
		return game.DynamicValueControllerLifeTotal, true
	case parser.StaticDeclarationDynamicValueAllPlayersHandSize:
		return game.DynamicValueAllPlayersHandSize, true
	case parser.StaticDeclarationDynamicValueControllerColorPermanentCount:
		return game.DynamicValueControllerColorPermanentCount, true
	default:
		return game.DynamicValueNone, false
	}
}

// staticCharacteristicDeclarations splits a "<group> is/are ... in addition"
// declaration into separate color and type layer declarations. Colors are set
// when no "in addition" tail is present and added otherwise; card types and
// subtypes are always additive. It fails closed for an unrepresentable color or
// card type.
func staticCharacteristicDeclarations(span shared.Span, node *parser.StaticDeclarationSyntax, group StaticGroupReference, condition *CompiledCondition) ([]StaticDeclaration, bool) {
	var declarations []StaticDeclaration
	if len(node.Colors) != 0 {
		colors, ok := staticRuntimeColors(node.Colors)
		if !ok {
			return nil, false
		}
		operation := StaticContinuousSetColors
		if node.ColorsAdd {
			operation = StaticContinuousAddColors
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerColor,
				Operation: operation,
				Colors:    colors,
			},
		})
	}
	if len(node.CardTypes) != 0 || len(node.Subtypes) != 0 {
		cardTypes, ok := staticCardTypesFromParser(node.CardTypes)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:       StaticLayerType,
				Operation:   StaticContinuousAddTypes,
				AddTypes:    cardTypes,
				AddSubtypes: slices.Clone(node.Subtypes),
			},
		})
	}
	if node.EveryCreatureType {
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:                StaticLayerType,
				Operation:            StaticContinuousAddTypes,
				AddEveryCreatureType: true,
			},
		})
	}
	if node.EveryBasicLandType {
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:                 StaticLayerType,
				Operation:             StaticContinuousAddTypes,
				AddEveryBasicLandType: true,
			},
		})
	}
	if len(declarations) == 0 {
		return nil, false
	}
	return declarations, true
}

func staticRuntimeColors(colors []parser.Color) ([]color.Color, bool) {
	result := make([]color.Color, 0, len(colors))
	for _, value := range colors {
		runtime, ok := compilerColor(value)
		if !ok {
			return nil, false
		}
		result = append(result, runtime)
	}
	return result, true
}

func staticCardTypesFromParser(cardTypes []parser.CardType) ([]types.Card, bool) {
	result := make([]types.Card, 0, len(cardTypes))
	for _, value := range cardTypes {
		mapped, ok := staticCardTypeFromParser(value)
		if !ok {
			return nil, false
		}
		result = append(result, mapped)
	}
	return result, true
}

func staticCardTypeFromParser(value parser.CardType) (types.Card, bool) {
	switch value {
	case parser.CardTypeArtifact:
		return types.Artifact, true
	case parser.CardTypeCreature:
		return types.Creature, true
	case parser.CardTypeLand:
		return types.Land, true
	case parser.CardTypeEnchantment:
		return types.Enchantment, true
	case parser.CardTypeInstant:
		return types.Instant, true
	case parser.CardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}

func staticDeclarationGrantKeywords(content AbilityContent) []CompiledKeyword {
	usesCyclingPredicate := false
	for i := range content.Effects {
		effect := &content.Effects[i]
		if effect.Selector.Keyword == parser.KeywordCycling ||
			effect.Amount.Selector().Keyword == parser.KeywordCycling {
			usesCyclingPredicate = true
			break
		}
	}
	if !usesCyclingPredicate {
		return content.Keywords
	}
	filtered := make([]CompiledKeyword, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keyword.Kind == parser.KeywordCycling && keyword.ParameterKind == parser.KeywordParameterNone {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func recognizeStaticKeywordGrantDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationKeywordGrant) {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectGrantKeyword ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Keywords) == 0 ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	// "As long as equipped/enchanted creature is <state>, it has <keyword>": the
	// pronoun "it" co-refers with the attached creature named by the gating
	// condition, so the grant applies to the attached object. The pronoun has no
	// antecedent reference of its own and binds Ambiguous, so the attached group
	// is taken from the condition's binding rather than from the effect group.
	if condition != nil && condition.ObjectBinding == ReferenceBindingSourceAttached {
		if !staticKeywordGrantBindsAttachedPronoun(ability, effect) {
			return nil, false
		}
		group := StaticGroupReference{Span: ability.Content.References[0].Span, Domain: StaticGroupAttachedObject}
		return []StaticDeclaration{staticKeywordGrantDeclaration(ability.Span, group, condition, ability.Content.Keywords)}, true
	}
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	if group.AffectedSource {
		if condition == nil {
			return nil, false
		}
	} else if condition != nil && !condition.SourceInGraveyard {
		// A group anthem ("creatures you control have ...") may carry a condition
		// only when the static ability functions from the graveyard, as on the
		// Incarnation cycle ("As long as this card is in your graveyard and you
		// control a <land>, ..."). Other conditioned group anthems fail closed.
		return nil, false
	}
	return []StaticDeclaration{staticKeywordGrantDeclaration(ability.Span, group.Group, condition, ability.Content.Keywords)}, true
}

// staticKeywordGrantBindsAttachedPronoun reports whether a conditional keyword
// grant is the attached-creature pronoun form "..., it has <keyword>": the
// effect recipient is a referenced object filled by exactly one "it"/"them"
// pronoun reference whose own antecedent is unresolved (Ambiguous). The pronoun
// co-refers with the attached creature named by the gating condition.
func staticKeywordGrantBindsAttachedPronoun(ability CompiledAbility, effect *CompiledEffect) bool {
	if effect.StaticSubject != StaticSubjectNone ||
		effect.Context != parser.EffectContextReferencedObject ||
		len(ability.Content.References) != 1 {
		return false
	}
	reference := ability.Content.References[0]
	return reference.Kind == ReferencePronoun &&
		(reference.Pronoun == ReferencePronounIt || reference.Pronoun == ReferencePronounThem) &&
		reference.Binding == ReferenceBindingAmbiguous
}

func staticDeclarationCondition(conditions []CompiledCondition) (*CompiledCondition, bool) {
	if len(conditions) == 0 {
		return nil, true
	}
	if len(conditions) != 1 || conditions[0].Predicate == ConditionPredicateUnsupported {
		return nil, false
	}
	return &conditions[0], true
}

type staticDeclarationEffectGroupResult struct {
	Group          StaticGroupReference
	AffectedSource bool
}

// staticParserSubjectReferencesTolerated reports whether the ability's free
// references are compatible with a typed parser subject's affected group. A
// static subject names its own affected group, so a free reference normally
// disqualifies it. The "with a/an <kind> counter on it/them" group filter
// (Rishkar's "Each creature you control with a counter on it has ...") names the
// affected creature itself with the pronoun "it"/"them"; that self-reference is
// tolerated rather than treated as a separate antecedent.
func staticParserSubjectReferencesTolerated(references []CompiledReference, subject parser.StaticDeclarationSubject) bool {
	if len(references) == 0 {
		return true
	}
	if subject.Kind != parser.StaticDeclarationSubjectGroup || !subject.Group.CounterRequired {
		return false
	}
	for i := range references {
		if references[i].Pronoun != ReferencePronounIt && references[i].Pronoun != ReferencePronounThem {
			return false
		}
	}
	return true
}

// staticSubjectGroupReferencesTolerated reports whether the ability's free
// references are compatible with a static-subject affected group. A static
// subject names its own affected group, so a free reference normally signals a
// referent-bound group and disqualifies it. Two exceptions tolerate a reference
// that names the affected creature itself with the pronoun "it"/"them" rather
// than a separate antecedent: the shared-creature-type bonus, whose amount reads
// "for each other creature ... that shares a creature type with it", and the
// counter-matters group filter, whose subject reads "creature you control with a
// +1/+1 counter on it/them".
func staticSubjectGroupReferencesTolerated(references []CompiledReference, effect *CompiledEffect) bool {
	if len(references) == 0 {
		return true
	}
	_, _, counterFilter := effect.StaticSubjectCounter()
	if effect.Amount.DynamicKind != DynamicAmountSharedCreatureTypeCount && !counterFilter {
		return false
	}
	for i := range references {
		if references[i].Pronoun != ReferencePronounIt && references[i].Pronoun != ReferencePronounThem {
			return false
		}
	}
	return true
}

func staticDeclarationEffectGroup(ability CompiledAbility, effect *CompiledEffect) (staticDeclarationEffectGroupResult, bool) {
	freeReferences := staticFreeReferences(ability)
	if effect.StaticSubject != StaticSubjectNone {
		if !staticSubjectGroupReferencesTolerated(freeReferences, effect) {
			return staticDeclarationEffectGroupResult{}, false
		}
		keyword, excludedKeyword := staticSubjectKeywordFilter(effect)
		group, ok := staticGroupForSubject(effect.StaticSubject, effect.StaticSubjectSpan, effect.StaticSubjectSub(), effect.StaticSubjectSubKnown(), staticColorFilter{
			Colors:       effect.StaticSubjectColorsAny(),
			Colorless:    effect.StaticSubjectColorless(),
			Multicolored: effect.StaticSubjectMulticolored(),
		}, keyword, excludedKeyword, effect.StaticSubjectChosenColorFromEntry())
		if ok {
			if kind, anyKind, present := effect.StaticSubjectCounter(); present {
				if anyKind {
					group.Selection.MatchAnyCounter = true
				} else {
					group.Selection.MatchCounter = true
					group.Selection.RequiredCounter = kind
				}
			}
		}
		return staticDeclarationEffectGroupResult{Group: group}, ok
	}
	if len(freeReferences) == 1 && freeReferences[0].Binding == ReferenceBindingSource {
		return staticDeclarationEffectGroupResult{
			Group: StaticGroupReference{
				Span:   freeReferences[0].Span,
				Domain: StaticGroupSource,
			},
			AffectedSource: true,
		}, true
	}
	return staticDeclarationEffectGroupResult{}, false
}

// staticFreeReferences returns the ability's references that are not consumed by
// a recognized condition clause. A condition that names the source ("as long as
// ~ has seven or more quest counters on it") contributes source references that
// belong to the gate rather than to the affected group, so the group derivation
// must not treat them as free referents that bind a separate antecedent.
func staticFreeReferences(ability CompiledAbility) []CompiledReference {
	references := ability.Content.References
	free := make([]CompiledReference, 0, len(references))
	for i := range references {
		consumed := false
		for j := range ability.Content.Conditions {
			condition := ability.Content.Conditions[j]
			if condition.Predicate == ConditionPredicateUnsupported {
				continue
			}
			if condition.Order.Contains(references[i].Order) {
				consumed = true
				break
			}
		}
		if !consumed {
			free = append(free, references[i])
		}
	}
	return free
}

func staticGroupForSubject(subject StaticSubjectKind, span shared.Span, subtype types.Sub, subtypeKnown bool, colors staticColorFilter, keyword, excludedKeyword parser.KeywordKind, chosenColorFromEntry bool) (StaticGroupReference, bool) {
	group := StaticGroupReference{Span: span}
	switch subject {
	case StaticSubjectAttachedObject:
		group.Domain = StaticGroupAttachedObject
	case StaticSubjectAllCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
	case StaticSubjectAllOtherCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.ExcludeSource = true
	case StaticSubjectAttackingCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.CombatState = StaticCombatStateAttacking
	case StaticSubjectOtherAttackingCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.CombatState = StaticCombatStateAttacking
		group.ExcludeSource = true
	case StaticSubjectBlockingCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.CombatState = StaticCombatStateBlocking
	case StaticSubjectControlledPermanents:
		group.Domain = StaticGroupSourceControllerPermanents
	case StaticSubjectOtherControlledPermanents:
		group.Domain = StaticGroupSourceControllerPermanents
		group.ExcludeSource = true
	case StaticSubjectControlledCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
	case StaticSubjectOtherControlledCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.ExcludeSource = true
	case StaticSubjectControlledWalls:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{types.Wall}
	case StaticSubjectControlledArtifacts:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Artifact}
	case StaticSubjectControlledSagas:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{types.Saga}
	case StaticSubjectControlledTokens:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.TokenOnly = true
	case StaticSubjectOpponentControlledCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.Controller = ControllerOpponent
	case StaticSubjectControlledCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{subtype}
	case StaticSubjectOtherControlledCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{subtype}
		group.ExcludeSource = true
	case StaticSubjectAllCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupBattlefield
		group.Selection.SubtypesAny = []types.Sub{subtype}
	case StaticSubjectOtherCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupBattlefield
		group.Selection.SubtypesAny = []types.Sub{subtype}
		group.ExcludeSource = true
	case StaticSubjectControlledAttackingCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.CombatState = StaticCombatStateAttacking
	case StaticSubjectControlledCreatureTokens:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.TokenOnly = true
	case StaticSubjectBattlefieldCreatureTokens:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.TokenOnly = true
	case StaticSubjectControlledLegendaryCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.Supertypes = []types.Super{types.Legendary}
	case StaticSubjectControlledNonlegendaryCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.ExcludedSupertypes = []types.Super{types.Legendary}
	case StaticSubjectControlledCommanderCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.Commander = true
	case StaticSubjectControlledCommanders:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.Commander = true
	case StaticSubjectControlledUntappedCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.TapState = StaticTapStateUntapped
	case StaticSubjectControlledModifiedCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.Modified = true
	case StaticSubjectOtherControlledTappedCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.TapState = StaticTapStateTapped
		group.ExcludeSource = true
	case StaticSubjectControlledArtifactCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Artifact, types.Creature}
	case StaticSubjectOtherControlledArtifactCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Artifact, types.Creature}
		group.ExcludeSource = true
	case StaticSubjectControlledNontokenCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.NonToken = true
	case StaticSubjectOtherControlledNontokenCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.NonToken = true
		group.ExcludeSource = true
	case StaticSubjectAllLands:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []types.Card{types.Land}
	case StaticSubjectControlledLands:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Land}
	case StaticSubjectControlledCreaturesChosenType:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.SubtypeFromEntryChoice = true
	case StaticSubjectOtherControlledCreaturesChosenType:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		group.Selection.SubtypeFromEntryChoice = true
		group.ExcludeSource = true
	default:
		return StaticGroupReference{}, false
	}
	if !applyStaticColorFilter(&group.Selection, colors) {
		return StaticGroupReference{}, false
	}
	if !applyStaticKeywordFilter(&group.Selection, keyword, excludedKeyword) {
		return StaticGroupReference{}, false
	}
	if chosenColorFromEntry {
		group.Selection.ColorFromEntryChoice = true
	}
	return group, true
}

// staticColorFilter is the closed color constraint an affected creature group
// may carry ("Other red creatures you control ..."). The zero value applies no
// color constraint.
type staticColorFilter struct {
	Colors       []parser.Color
	Colorless    bool
	Multicolored bool
}

// applyStaticColorFilter sets the Selection's color predicate from a typed color
// filter, failing closed for any color word that has no runtime representation.
func applyStaticColorFilter(selection *StaticSelection, colors staticColorFilter) bool {
	for _, value := range colors.Colors {
		runtime, ok := compilerColor(value)
		if !ok {
			return false
		}
		selection.ColorsAny = append(selection.ColorsAny, runtime)
	}
	selection.Colorless = colors.Colorless
	selection.Multicolored = colors.Multicolored
	return true
}

// applyStaticKeywordFilter records a single static-group keyword predicate. The
// keyword's runtime representability is validated later by the cardgen lowerer,
// which fails closed for keywords with no runtime Selection mapping.
func applyStaticKeywordFilter(selection *StaticSelection, keyword, excludedKeyword parser.KeywordKind) bool {
	selection.Keyword = keyword
	selection.ExcludedKeyword = excludedKeyword
	return true
}

// staticSubjectKeywordFilter splits an effect's optional static-subject keyword
// filter into the required and excluded keyword slots.
func staticSubjectKeywordFilter(effect *CompiledEffect) (required, excludedKeyword parser.KeywordKind) {
	keyword, excluded, ok := effect.StaticSubjectKeyword()
	if !ok {
		return parser.KeywordUnknown, parser.KeywordUnknown
	}
	if excluded {
		return parser.KeywordUnknown, keyword
	}
	return keyword, parser.KeywordUnknown
}

func staticPTDeclaration(span shared.Span, group StaticGroupReference, condition *CompiledCondition, effect *CompiledEffect) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: effect.VerbSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:          StaticLayerPowerToughnessModify,
			Operation:      StaticContinuousModifyPowerToughness,
			PowerDelta:     effect.PowerDelta,
			ToughnessDelta: effect.ToughnessDelta,
			DynamicAmount:  effect.Amount,
		},
	}
}

func staticKeywordGrantDeclaration(span shared.Span, group StaticGroupReference, condition *CompiledCondition, keywords []CompiledKeyword) StaticDeclaration {
	operationSpan := keywords[0].Span
	operationSpan.End = keywords[len(keywords)-1].Span.End
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: operationSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerAbility,
			Operation: StaticContinuousGrantKeywords,
			Keywords:  append([]CompiledKeyword(nil), keywords...),
		},
	}
}

// recognizeStaticSpellCostModifierDeclaration maps the typed spell cast-cost
// modifier syntax onto a closed semantic cost declaration. The affected group is
// the static ability's controller's spells; the optional spell-type filter is a
// closed set of card types matched as a disjunction at runtime.
func recognizeStaticSpellCostModifierDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCostModifier) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.CostModifier != parser.StaticDeclarationCostModifierSpellReduction &&
		node.CostModifier != parser.StaticDeclarationCostModifierSpellIncrease {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Conditions) != 0 {
		return StaticDeclaration{}, false
	}
	// A targets-source cost modifier ("Spells your opponents cast that target
	// this creature cost {N} more to cast.") intentionally carries the "that
	// target <source>" phrase, which the effect machinery records as a target
	// and a reference; the parser owns that wording and marks it via
	// SpellTargetsSource, so those payloads are expected here. Every other
	// cast-cost modifier must carry no targets or references.
	if !node.SpellTargetsSource &&
		(len(ability.Content.Targets) != 0 || len(ability.Content.References) != 0) {
		return StaticDeclaration{}, false
	}
	spellTypes, ok := staticSpellTypeCardTypes(node.SpellType)
	if !ok {
		return StaticDeclaration{}, false
	}
	spellColor, matchColor, ok := staticSpellColorMatch(node.SpellColor)
	if !ok {
		return StaticDeclaration{}, false
	}
	spellColors, ok := staticSpellColorDisjunctionMatch(node.SpellColors)
	if !ok {
		return StaticDeclaration{}, false
	}
	if len(node.SpellSubtypes) != 0 &&
		(len(spellTypes) != 0 || len(spellColors) != 0 || node.ChosenCreatureType) {
		return StaticDeclaration{}, false
	}
	if len(spellColors) != 0 && (matchColor || len(spellTypes) != 0 || node.ChosenCreatureType) {
		return StaticDeclaration{}, false
	}
	if node.ChosenCreatureType &&
		(node.CostModifier != parser.StaticDeclarationCostModifierSpellReduction ||
			node.SpellType != parser.StaticDeclarationSpellTypeCreature ||
			matchColor) {
		return StaticDeclaration{}, false
	}
	if node.CostReductionAmount <= 0 {
		return StaticDeclaration{}, false
	}
	if node.SpellCastZone != "" && node.ChosenCreatureType {
		return StaticDeclaration{}, false
	}
	if node.MatchSpellPowerAtLeast && node.SpellPowerAtLeast <= 0 {
		return StaticDeclaration{}, false
	}
	caster, ok := staticSpellCasterKind(node.SpellCaster)
	if !ok {
		return StaticDeclaration{}, false
	}
	cost := StaticCostModifierDeclaration{
		Kind:                         StaticCostModifierSpell,
		SpellTypes:                   spellTypes,
		MatchSpellColor:              matchColor,
		SpellColor:                   spellColor,
		SpellColors:                  spellColors,
		SpellSubtypes:                node.SpellSubtypes,
		ChosenSubtypeFromEntryChoice: node.ChosenCreatureType,
		SourceZone:                   node.SpellCastZone,
		MinPower:                     node.SpellPowerAtLeast,
		MatchMinPower:                node.MatchSpellPowerAtLeast,
		TargetsSource:                node.SpellTargetsSource,
		Caster:                       caster,
	}
	if node.CostModifier == parser.StaticDeclarationCostModifierSpellIncrease {
		cost.GenericIncrease = node.CostReductionAmount
	} else {
		cost.GenericReduction = node.CostReductionAmount
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCostModifier,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   ability.Span,
			Domain: StaticGroupControllerSpells,
		},
		Cost: &cost,
	}, true
}

// staticSpellCasterKind maps the parser's closed caster filter onto the
// compiler's caster kind. An unrecognized filter fails closed.
func staticSpellCasterKind(filter parser.StaticDeclarationSpellCasterKind) (StaticSpellCasterKind, bool) {
	switch filter {
	case parser.StaticDeclarationSpellCasterController:
		return StaticSpellCasterController, true
	case parser.StaticDeclarationSpellCasterOpponents:
		return StaticSpellCasterOpponents, true
	case parser.StaticDeclarationSpellCasterAny:
		return StaticSpellCasterAny, true
	default:
		return StaticSpellCasterController, false
	}
}

// staticSpellTypeCardTypes maps a closed spell-type filter onto the card types
// whose disjunction the runtime matches. An all-spells filter returns no types.
func staticSpellTypeCardTypes(filter parser.StaticDeclarationSpellTypeKind) ([]types.Card, bool) {
	switch filter {
	case parser.StaticDeclarationSpellTypeAll:
		return nil, true
	case parser.StaticDeclarationSpellTypeArtifact:
		return []types.Card{types.Artifact}, true
	case parser.StaticDeclarationSpellTypeCreature:
		return []types.Card{types.Creature}, true
	case parser.StaticDeclarationSpellTypeEnchantment:
		return []types.Card{types.Enchantment}, true
	case parser.StaticDeclarationSpellTypeInstant:
		return []types.Card{types.Instant}, true
	case parser.StaticDeclarationSpellTypeSorcery:
		return []types.Card{types.Sorcery}, true
	case parser.StaticDeclarationSpellTypeInstantOrSorcery:
		return []types.Card{types.Instant, types.Sorcery}, true
	default:
		return nil, false
	}
}

// staticSpellColorMatch maps a closed spell-color filter onto a runtime color
// match. It returns the matched color, whether a color filter is present, and
// false for an unrecognized filter. The colorless filter yields the empty-color
// sentinel with a true match flag.
func staticSpellColorMatch(filter parser.StaticDeclarationSpellColorKind) (spellColor color.Color, match, ok bool) {
	switch filter {
	case parser.StaticDeclarationSpellColorNone:
		return "", false, true
	case parser.StaticDeclarationSpellColorWhite:
		return color.White, true, true
	case parser.StaticDeclarationSpellColorBlue:
		return color.Blue, true, true
	case parser.StaticDeclarationSpellColorBlack:
		return color.Black, true, true
	case parser.StaticDeclarationSpellColorRed:
		return color.Red, true, true
	case parser.StaticDeclarationSpellColorGreen:
		return color.Green, true, true
	case parser.StaticDeclarationSpellColorColorless:
		return "", true, true
	default:
		return "", false, false
	}
}

// staticSpellColorDisjunctionMatch maps a closed color disjunction onto runtime
// colors. It returns the colors and false for an empty or malformed list: a
// disjunction carries two or more real colors (colorless is not admitted). An
// absent disjunction returns no colors with ok true.
func staticSpellColorDisjunctionMatch(filters []parser.StaticDeclarationSpellColorKind) ([]color.Color, bool) {
	if len(filters) == 0 {
		return nil, true
	}
	if len(filters) < 2 {
		return nil, false
	}
	colors := make([]color.Color, 0, len(filters))
	for _, filter := range filters {
		spellColor, match, ok := staticSpellColorMatch(filter)
		if !ok || !match || spellColor == "" {
			return nil, false
		}
		colors = append(colors, spellColor)
	}
	return colors, true
}

func recognizeStaticCostModifierDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCostModifier) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordCycling ||
		ability.Content.Keywords[0].ParameterKind != parser.KeywordParameterNone {
		return StaticDeclaration{}, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	cost := StaticCostModifierDeclaration{
		Kind:           StaticCostModifierAbility,
		AbilityKeyword: ability.Content.Keywords[0].Kind,
	}
	switch node.CostModifier {
	case parser.StaticDeclarationCostModifierAbilityReduction:
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.GenericReduction = node.CostReductionAmount
	case parser.StaticDeclarationCostModifierReplaceCost:
		if condition == nil ||
			condition.Predicate != ConditionPredicateControllerHandSizeAtLeast ||
			condition.Threshold != 7 {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = node.CostReplacement
	case parser.StaticDeclarationCostModifierReplaceFirstCost:
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = node.CostReplacement
		cost.FirstCycleEachTurn = true
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCostModifier,
		Span:          ability.Span,
		OperationSpan: ability.Span,
		Group: StaticGroupReference{
			Span:   ability.Span,
			Domain: StaticGroupControllerHandCards,
		},
		Condition: condition,
		Cost:      &cost,
	}, true
}

// recognizeStaticAbilityCostSetDeclaration maps the parser's ability-cost setting
// syntax ("Equipment you control have equip {N}.") onto a semantic cost-modifier
// declaration that replaces the Equip activation cost of the controller's
// Equipment. The optional Metalcraft-style count condition gates the static.
func recognizeStaticAbilityCostSetDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationAbilityCostSet) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordEquip {
		return StaticDeclaration{}, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	cost := StaticCostModifierDeclaration{
		Kind:            StaticCostModifierAbility,
		AbilityKeyword:  node.AbilityCostKeyword,
		ReplaceManaCost: true,
		SetManaCost:     node.CostReplacement,
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCostModifier,
		Span:          ability.Span,
		OperationSpan: ability.Span,
		Group: StaticGroupReference{
			Span:   ability.Span,
			Domain: StaticGroupControllerEquipment,
		},
		Condition: condition,
		Cost:      &cost,
	}, true
}

func recognizeStaticCardAbilityGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCardAbilityGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if !staticCardAbilityGrantGatingHolds(ability) {
		return StaticDeclaration{}, false
	}
	keyword := ability.Content.Keywords[0]
	group := StaticGroupReference{
		Span:   ability.Span,
		Domain: StaticGroupControllerHandCards,
	}
	var text string
	switch node.Subject.CardFilter {
	case parser.StaticDeclarationCardFilterLand:
		group.Selection.RequiredTypes = []types.Card{types.Land}
		text = "Each land card in your hand has cycling " + keyword.Parameter + "."
	case parser.StaticDeclarationCardFilterCreature:
		group.Selection.RequiredTypes = []types.Card{types.Creature}
		text = "Each creature card in your hand has cycling " + keyword.Parameter + "."
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCardAbilityGrant,
		Span:          ability.Span,
		OperationSpan: keyword.Span,
		Group:         group,
		CardGrant: &StaticCardAbilityGrantDeclaration{
			Keyword: keyword,
			Text:    text,
		},
	}, true
}

func recognizeStaticPermanentAbilityGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationPermanentAbilityGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	granted := node.GrantedManaAbility
	if granted == nil ||
		!granted.TapCost ||
		node.Subject.Kind != parser.StaticDeclarationSubjectGroup ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	if !staticGrantedManaAbilityValid(granted) {
		return StaticDeclaration{}, false
	}
	selection, ok := staticPermanentGrantSelection(node.Subject.Group)
	if !ok {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:      node.Subject.Span,
			Domain:    StaticGroupSourceControllerPermanents,
			Selection: selection,
		},
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerAbility,
			Operation: StaticContinuousGrantManaAbility,
			GrantedMana: &StaticGrantedManaAbility{
				TapCost:     granted.TapCost,
				Amount:      granted.Amount,
				AnyColor:    granted.AnyColor,
				Text:        granted.Text,
				Sacrifice:   granted.Sacrifice,
				AnyOneColor: granted.AnyOneColor,
			},
		},
	}, true
}

// staticGrantedManaAbilityValid reports whether the parsed granted mana ability
// is one of the two closed shapes the runtime can confer: the bare
// tap-for-one-mana-of-any-color ability, and the Treasure-style sacrifice
// ability that adds N mana (N >= 2) of one chosen color.
func staticGrantedManaAbilityValid(granted *parser.StaticGrantedManaAbilitySyntax) bool {
	switch {
	case granted.AnyColor:
		return granted.Amount == 1 && !granted.Sacrifice && !granted.AnyOneColor && !granted.Colorless
	case granted.AnyOneColor:
		return granted.Amount >= 2 && granted.Sacrifice
	case granted.Colorless:
		return granted.Amount == 1 && !granted.Sacrifice
	default:
		return false
	}
}

// staticPermanentGrantSelection maps the grant's typed group subject onto the
// affected-permanent Selection. The controller is implied by the
// source-controller permanent domain, so only the type and subtype filters are
// set here.
func staticPermanentGrantSelection(group parser.EffectStaticSubjectSyntax) (StaticSelection, bool) {
	switch group.Kind {
	case parser.EffectStaticSubjectControlledLands:
		return StaticSelection{RequiredTypes: []types.Card{types.Land}}, true
	case parser.EffectStaticSubjectControlledCreatures:
		return StaticSelection{RequiredTypes: []types.Card{types.Creature}}, true
	case parser.EffectStaticSubjectControlledArtifacts:
		selection := StaticSelection{RequiredTypes: []types.Card{types.Artifact}}
		if group.SubtypeKnown {
			selection.SubtypesAny = []types.Sub{group.Subtype}
		}
		return selection, true
	default:
		return StaticSelection{}, false
	}
}

func recognizeStaticControlGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationControlGrant) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.Subject.Kind != parser.StaticDeclarationSubjectGroup ||
		node.Subject.Group.Kind != parser.EffectStaticSubjectAttachedObject {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   node.Subject.Span,
			Domain: StaticGroupAttachedObject,
		},
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerControl,
			Operation: StaticContinuousChangeControl,
		},
	}, true
}

type staticPlayerRuleSpec struct {
	kind                    StaticPlayerRuleKind
	usesAttackTax           bool
	usesAdditionalLandPlays bool
	usesManaColor           bool
	allowsAllPlayers        bool
	matchesContent          func(AbilityContent) bool
}

var staticPlayerRuleSpecs = map[parser.StaticDeclarationPlayerRuleKind]staticPlayerRuleSpec{
	parser.StaticDeclarationPlayerRuleNoMaximumHandSize: {
		kind:           StaticPlayerRuleNoMaximumHandSize,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleAttackTax: {
		kind:           StaticPlayerRuleAttackTax,
		usesAttackTax:  true,
		matchesContent: attackTaxStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleAdditionalLandPlays: {
		kind:                    StaticPlayerRuleAdditionalLandPlays,
		usesAdditionalLandPlays: true,
		allowsAllPlayers:        true,
		matchesContent:          emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRulePlayLandsFromGraveyard: {
		kind:           StaticPlayerRulePlayLandsFromGraveyard,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRulePlayLandsFromLibraryTop: {
		kind:           StaticPlayerRulePlayLandsFromLibraryTop,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRulePlayWithTopCardRevealed: {
		kind:           StaticPlayerRulePlayWithTopCardRevealed,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleCastSpellsFromLibraryTop: {
		kind:           StaticPlayerRuleCastSpellsFromLibraryTop,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleCastThisFromGraveyard: {
		kind:           StaticPlayerRuleCastThisFromGraveyard,
		matchesContent: castThisFromGraveyardStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleCastThisFromExile: {
		kind:           StaticPlayerRuleCastThisFromExile,
		matchesContent: castThisFromGraveyardStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleLookAtTopCardAnyTime: {
		kind:           StaticPlayerRuleLookAtTopCardAnyTime,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleLifeForColoredMana: {
		kind:           StaticPlayerRuleLifeForColoredMana,
		usesManaColor:  true,
		matchesContent: emptyStaticPlayerRuleContent,
	},
	parser.StaticDeclarationPlayerRuleLifeForCommanderTax: {
		kind:           StaticPlayerRuleLifeForCommanderTax,
		matchesContent: lifeForCommanderTaxStaticPlayerRuleContent,
	},
}

// recognizeStaticPlayerRuleDeclaration maps parser-owned player-rule syntax to
// the closed semantic player-rule vocabulary.
func recognizeStaticPlayerRuleDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationPlayerRule) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	spec, ok := staticPlayerRuleSpecs[node.PlayerRule]
	if !ok ||
		!staticPlayerRuleSubjectAllowed(node.Subject.Kind, spec) ||
		spec.matchesContent == nil ||
		!spec.matchesContent(ability.Content) ||
		(spec.usesAttackTax && node.AttackTaxGeneric <= 0) ||
		(!spec.usesAttackTax && node.AttackTaxGeneric != 0) ||
		(spec.usesAdditionalLandPlays && node.AdditionalLandPlays <= 0) ||
		(!spec.usesAdditionalLandPlays && node.AdditionalLandPlays != 0) ||
		(spec.usesManaColor && !compilerManaColorValid(node.ManaColor)) ||
		(!spec.usesManaColor && node.ManaColor != "") {
		return StaticDeclaration{}, false
	}
	var spellTypes []types.Card
	if spec.kind == StaticPlayerRuleCastSpellsFromLibraryTop {
		spellTypes = make([]types.Card, 0, len(node.CastSpellTypes))
		for _, cardType := range node.CastSpellTypes {
			converted, ok := compilerCardType(cardType)
			if !ok {
				return StaticDeclaration{}, false
			}
			spellTypes = append(spellTypes, converted)
		}
	} else if len(node.CastSpellTypes) != 0 || node.CastColorless || node.AlsoPlayLands || node.CastChosenCreatureType {
		return StaticDeclaration{}, false
	}
	var condition *CompiledCondition
	if spec.kind == StaticPlayerRuleCastThisFromGraveyard || spec.kind == StaticPlayerRuleCastThisFromExile {
		compiledCondition, ok := staticDeclarationCondition(ability.Content.Conditions)
		if !ok {
			return StaticDeclaration{}, false
		}
		condition = compiledCondition
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationPlayerRule,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Condition:     condition,
		Player: &StaticPlayerRuleDeclaration{
			Kind:                   spec.kind,
			AttackTaxGeneric:       node.AttackTaxGeneric,
			AdditionalLandPlays:    node.AdditionalLandPlays,
			AffectsAllPlayers:      node.Subject.Kind == parser.StaticDeclarationSubjectEachPlayer,
			SpellTypes:             spellTypes,
			CastColorless:          node.CastColorless,
			AlsoPlayLands:          node.AlsoPlayLands,
			CastChosenCreatureType: node.CastChosenCreatureType,
			ManaColor:              node.ManaColor,
		},
	}, true
}

// staticPlayerRuleSubjectAllowed reports whether a player-rule subject scope is
// valid for the given spec. The controller scope is always accepted; the
// each-player scope is accepted only for specs that grant the rule to every
// player.
func staticPlayerRuleSubjectAllowed(subject parser.StaticDeclarationSubjectKind, spec staticPlayerRuleSpec) bool {
	switch subject {
	case parser.StaticDeclarationSubjectController:
		return true
	case parser.StaticDeclarationSubjectEachPlayer:
		return spec.allowsAllPlayers
	default:
		return false
	}
}

func emptyStaticPlayerRuleContent(content AbilityContent) bool {
	return len(content.Conditions) == 0 && len(content.References) == 0
}

// lifeForCommanderTaxStaticPlayerRuleContent accepts the life-for-commander-tax
// cost-substitution player rule ("Rather than pay {2} for each previous time
// you've cast this spell from the command zone this game, pay 2 life that many
// times."). The sentence carries a single "this spell" source self-reference and
// no condition clauses.
func lifeForCommanderTaxStaticPlayerRuleContent(content AbilityContent) bool {
	if len(content.Conditions) != 0 {
		return false
	}
	for i := range content.References {
		if content.References[i].Binding != ReferenceBindingSource {
			return false
		}
	}
	return true
}

// compilerManaColorValid reports whether c is one of the five real colors of
// mana, backing the StaticPlayerRuleLifeForColoredMana color requirement.
func compilerManaColorValid(c mana.Color) bool {
	switch c {
	case mana.W, mana.U, mana.B, mana.R, mana.G:
		return true
	default:
		return false
	}
}

// castThisFromGraveyardStaticPlayerRuleContent accepts the self-scoped
// graveyard-cast permission with an optional "as long as <condition>" gate (zero
// or one condition clause) and only source self-references ("this card").
func castThisFromGraveyardStaticPlayerRuleContent(content AbilityContent) bool {
	if len(content.Conditions) > 1 {
		return false
	}
	for i := range content.References {
		if content.References[i].Binding != ReferenceBindingSource {
			return false
		}
	}
	return true
}

func attackTaxStaticPlayerRuleContent(content AbilityContent) bool {
	if len(content.Conditions) != 1 || len(content.References) != 2 {
		return false
	}
	condition := content.Conditions[0]
	if condition.Kind != ConditionUnless ||
		condition.Predicate != ConditionPredicateUnsupported ||
		!condition.Negated {
		return false
	}
	return content.References[0].Pronoun == ReferencePronounTheir &&
		content.References[0].Binding == ReferenceBindingAmbiguous &&
		content.References[1].Pronoun == ReferencePronounThey &&
		content.References[1].Binding == ReferenceBindingAmbiguous
}

// recognizeStaticOpponentActionRestrictionDeclaration maps the parser-owned
// opponent action restriction syntax ("Your opponents can't cast spells [or
// activate abilities of <types>].", Grand Abolisher) onto its closed semantic
// payload. The legacy resolving-effect machinery also classifies the "cast"
// verb, so unlike the controller-scoped player rules this recognizer tolerates
// the leftover content effects, which the static-declaration lowering consumes.
func recognizeStaticOpponentActionRestrictionDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationOpponentActionRestriction) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	activateTypes := make([]types.Card, 0, len(node.RestrictActivateTypes))
	for _, cardType := range node.RestrictActivateTypes {
		converted, ok := compilerCardType(cardType)
		if !ok {
			return StaticDeclaration{}, false
		}
		activateTypes = append(activateTypes, converted)
	}
	if !node.RestrictCastSpells && len(activateTypes) == 0 {
		return StaticDeclaration{}, false
	}
	if node.RestrictCastOnlyFromHand && len(node.RestrictCastFromZones) != 0 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationOpponentActionRestriction,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		OpponentRestriction: &StaticOpponentActionRestrictionDeclaration{
			RestrictCastSpells:   node.RestrictCastSpells,
			ActivateTypes:        activateTypes,
			CastOnlyFromHand:     node.RestrictCastOnlyFromHand,
			CastFromZones:        append([]parser.StaticDeclarationCastZoneKind(nil), node.RestrictCastFromZones...),
			AffectsAllPlayers:    node.RestrictAffectsAllPlayers,
			DuringControllerTurn: node.RestrictDuringControllerTurn,
		},
	}, true
}

// recognizeStaticDrawLimitDeclaration maps the parser-owned draw-limit syntax
// ("Each opponent can't draw more than one card each turn.", Narset, Parter of
// Veils) onto its closed semantic payload. The continuous draw cap consumes no
// resolving content effects, so the ability must carry no cost, trigger, modes,
// targets, keywords, or ability word.
func recognizeStaticDrawLimitDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationDrawLimit) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.DrawLimit < 1 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationDrawLimit,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		DrawLimit: &StaticDrawLimitDeclaration{
			Limit:             node.DrawLimit,
			AffectsAllPlayers: node.DrawLimitAffectsAllPlayers,
			AffectsController: node.DrawLimitAffectsController,
		},
	}, true
}

// recognizeStaticCastLimitDeclaration maps the parser-owned cast-limit syntax
// ("Each player can't cast more than one spell each turn.", Rule of Law) onto
// its closed semantic payload. The continuous spell cap consumes no resolving
// content effects, so the ability must carry no cost, trigger, modes, targets,
// keywords, or ability word.
func recognizeStaticCastLimitDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCastLimit) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.CastLimit < 1 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCastLimit,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		CastLimit: &StaticCastLimitDeclaration{
			Limit:             node.CastLimit,
			AffectsAllPlayers: node.CastLimitAffectsAllPlayers,
			AffectsController: node.CastLimitAffectsController,
		},
	}, true
}

// recognizeStaticEnterBattlefieldRestrictionDeclaration maps the parser-owned
// entry-restriction syntax ("Creature cards in graveyards and libraries can't
// enter the battlefield.", Grafdigger's Cage) onto its closed semantic payload.
// The restriction is global; it carries the entering-card filter and the source
// zones cards cannot enter the battlefield out of.
func recognizeStaticEnterBattlefieldRestrictionDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationEnterBattlefieldRestriction) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.EnterRestrictFilter == "" || len(node.EnterRestrictFromZones) == 0 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationEnterBattlefieldRestriction,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		EnterRestriction: &StaticEnterBattlefieldRestrictionDeclaration{
			Filter:    node.EnterRestrictFilter,
			FromZones: append([]parser.StaticDeclarationCastZoneKind(nil), node.EnterRestrictFromZones...),
		},
	}, true
}

// recognizeStaticSpellUncounterableDeclaration maps the parser-owned
// group-uncounterable syntax ("[<type>] spells you control can't be countered.",
// Rhythm of the Wild) onto its closed semantic payload. The affected group is
// always the static ability's controller's spells.
func recognizeStaticSpellUncounterableDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationSpellUncounterable) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	spellTypes, ok := staticSpellTypeCardTypes(node.SpellType)
	if !ok {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationSpellUncounterable,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   node.Span,
			Domain: StaticGroupControllerSpells,
		},
		SpellUncounterable: &StaticSpellUncounterableDeclaration{
			SpellTypes: spellTypes,
		},
	}, true
}

// recognizeStaticCastAsThoughFlashDeclaration maps the parser-owned "You may
// cast [<filter>] spells as though they had flash." syntax onto its closed
// semantic payload. The permission is always scoped to the static ability's
// controller; the optional card-type and subtype filters are carried through.
func recognizeStaticCastAsThoughFlashDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCastAsThoughFlash) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	spellTypes, ok := staticSpellTypeCardTypes(node.FlashSpellType)
	if !ok {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCastAsThoughFlash,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   node.Span,
			Domain: StaticGroupControllerSpells,
		},
		CastAsThoughFlash: &StaticCastAsThoughFlashDeclaration{
			SpellTypes:    spellTypes,
			SpellSubtypes: node.FlashSpellSubtypes,
		},
	}, true
}

// recognizeStaticUntapStepDeclaration maps the parser-owned "Untap <group> you
// control during each other player's untap step." syntax onto its closed
// semantic payload. The affected group is always scoped to the static ability's
// controller (or the source permanent itself for the self form).
func recognizeStaticUntapStepDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationUntapDuringOtherUntapStep) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	payload, group, ok := staticUntapStepPayload(node.UntapGroup, node.Span)
	if !ok {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationUntapStep,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Untap:         &payload,
	}, true
}

// staticUntapStepPayload maps the closed parser untap-group filter onto the
// semantic payload and the affected group reference.
func staticUntapStepPayload(group parser.StaticUntapGroupKind, span shared.Span) (StaticUntapStepDeclaration, StaticGroupReference, bool) {
	switch group {
	case parser.StaticUntapGroupSelf:
		return StaticUntapStepDeclaration{Self: true},
			StaticGroupReference{Span: span, Domain: StaticGroupSource}, true
	case parser.StaticUntapGroupPermanents:
		return StaticUntapStepDeclaration{},
			StaticGroupReference{Span: span, Domain: StaticGroupSourceControllerPermanents}, true
	case parser.StaticUntapGroupCreatures:
		return StaticUntapStepDeclaration{PermanentTypes: []types.Card{types.Creature}},
			StaticGroupReference{
				Span:      span,
				Domain:    StaticGroupSourceControllerPermanents,
				Selection: StaticSelection{RequiredTypes: []types.Card{types.Creature}},
			}, true
	case parser.StaticUntapGroupArtifacts:
		return StaticUntapStepDeclaration{PermanentTypes: []types.Card{types.Artifact}},
			StaticGroupReference{
				Span:      span,
				Domain:    StaticGroupSourceControllerPermanents,
				Selection: StaticSelection{RequiredTypes: []types.Card{types.Artifact}},
			}, true
	case parser.StaticUntapGroupLands:
		return StaticUntapStepDeclaration{PermanentTypes: []types.Card{types.Land}},
			StaticGroupReference{
				Span:      span,
				Domain:    StaticGroupSourceControllerPermanents,
				Selection: StaticSelection{RequiredTypes: []types.Card{types.Land}},
			}, true
	default:
		return StaticUntapStepDeclaration{}, StaticGroupReference{}, false
	}
}

func staticSyntaxIsHistoricCardGrant(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) bool {
	return staticSyntaxKindsAre(statics, parser.StaticDeclarationCardAbilityGrant) &&
		statics[0].Subject.CardFilter == parser.StaticDeclarationCardFilterHistoric &&
		staticCardAbilityGrantGatingHolds(ability)
}

func staticCardAbilityGrantGatingHolds(ability CompiledAbility) bool {
	return ability.Cost == nil &&
		ability.Trigger == nil &&
		len(ability.Content.Modes) == 0 &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.References) == 0 &&
		len(ability.Content.Keywords) == 1 &&
		ability.Content.Keywords[0].Kind == parser.KeywordCycling &&
		ability.Content.Keywords[0].ParameterKind == parser.KeywordParameterManaCost
}

func staticRuleQualifiersAre(qualifiers []parser.StaticRuleQualifier, kinds ...parser.StaticRuleQualifierKind) bool {
	actual := make([]parser.StaticRuleQualifierKind, len(qualifiers))
	for i := range qualifiers {
		actual[i] = qualifiers[i].Kind
	}
	return slices.Equal(actual, kinds)
}
