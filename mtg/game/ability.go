package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbilityKind classifies an ability by how it functions in the game (CR 113.3).
type AbilityKind int

const (
	// SpellAbility is an instruction on an instant or sorcery, executed when
	// the spell resolves (CR 113.3a).
	SpellAbility AbilityKind = iota

	// ActivatedAbility has the form "[Cost]: [Effect]" and can be activated
	// by paying the cost (CR 113.3b, 602).
	ActivatedAbility

	// TriggeredAbility begins with "When", "Whenever", or "At" and triggers
	// automatically when its condition is met (CR 113.3c, 603).
	TriggeredAbility

	// StaticAbility is a declarative statement that generates a continuous
	// effect while its source is in the appropriate zone (CR 113.3d, 604).
	StaticAbility
)

// Keyword represents an evergreen or commonly-used keyword ability (CR 702).
type Keyword int

// Keyword values enumerate supported keyword abilities.
const (
	KeywordNone Keyword = iota
	Deathtouch
	Defender
	DoubleStrike
	FirstStrike
	Flash
	Flying
	Haste
	Hexproof
	Indestructible
	Lifelink
	Menace
	Protection
	Reach
	Shroud
	Trample
	Vigilance
	Ward
	SplitSecond
	Equip
	Enchant
	Cycling
	Flashback
	Kicker
	Madness
	Morph
	Disguise
	Convoke
	Delve
	Suspend
	Storm
	Cascade
	Prowess
	Mutate
	Companion
	Ninjutsu
	Escape
	Foretell
	Craft
	Discover
	Eternalize
	Affinity
	Improvise
	Emerge
	Undying
	Persist
	Wither
	Infect
	Toxic
	Annihilator
	Exalted
)

// Reusable StaticAbilityBody templates for non-parameterized keyword abilities.
// Use these in CardFace.StaticAbilities slices or initializer-function appends.
// Treat these values as immutable.
var (
	// DeathtouchStaticBody is the reusable StaticAbilityBody for deathtouch.
	DeathtouchStaticBody = simpleKeywordStaticBody("Deathtouch", Deathtouch)

	// DefenderStaticBody is the reusable StaticAbilityBody for defender.
	DefenderStaticBody = simpleKeywordStaticBody("Defender", Defender)

	// DoubleStrikeStaticBody is the reusable StaticAbilityBody for double strike.
	DoubleStrikeStaticBody = simpleKeywordStaticBody("Double strike", DoubleStrike)

	// FirstStrikeStaticBody is the reusable StaticAbilityBody for first strike.
	FirstStrikeStaticBody = simpleKeywordStaticBody("First strike", FirstStrike)

	// FlashStaticBody is the reusable StaticAbilityBody for flash.
	FlashStaticBody = simpleKeywordStaticBody("Flash", Flash)

	// FlyingStaticBody is the reusable StaticAbilityBody for flying.
	FlyingStaticBody = simpleKeywordStaticBody("Flying", Flying)

	// HasteStaticBody is the reusable StaticAbilityBody for haste.
	HasteStaticBody = simpleKeywordStaticBody("Haste", Haste)

	// HexproofStaticBody is the reusable StaticAbilityBody for hexproof.
	HexproofStaticBody = simpleKeywordStaticBody("Hexproof", Hexproof)

	// IndestructibleStaticBody is the reusable StaticAbilityBody for indestructible.
	IndestructibleStaticBody = simpleKeywordStaticBody("Indestructible", Indestructible)

	// LifelinkStaticBody is the reusable StaticAbilityBody for lifelink.
	LifelinkStaticBody = simpleKeywordStaticBody("Lifelink", Lifelink)

	// MenaceStaticBody is the reusable StaticAbilityBody for menace.
	MenaceStaticBody = simpleKeywordStaticBody("Menace", Menace)

	// ReachStaticBody is the reusable StaticAbilityBody for reach.
	ReachStaticBody = simpleKeywordStaticBody("Reach", Reach)

	// ShroudStaticBody is the reusable StaticAbilityBody for shroud.
	ShroudStaticBody = simpleKeywordStaticBody("Shroud", Shroud)

	// TrampleStaticBody is the reusable StaticAbilityBody for trample.
	TrampleStaticBody = simpleKeywordStaticBody("Trample", Trample)

	// VigilanceStaticBody is the reusable StaticAbilityBody for vigilance.
	VigilanceStaticBody = simpleKeywordStaticBody("Vigilance", Vigilance)

	// SplitSecondStaticBody is the reusable StaticAbilityBody for split second.
	SplitSecondStaticBody = simpleKeywordStaticBody("Split second", SplitSecond)

	// ConvokeStaticBody is the reusable StaticAbilityBody for convoke.
	ConvokeStaticBody = simpleKeywordStaticBody("Convoke", Convoke)

	// DelveStaticBody is the reusable StaticAbilityBody for delve.
	DelveStaticBody = simpleKeywordStaticBody("Delve", Delve)

	// StormStaticBody is the reusable StaticAbilityBody for storm.
	StormStaticBody = simpleKeywordStaticBody("Storm", Storm)

	// CascadeStaticBody is the reusable StaticAbilityBody for cascade.
	CascadeStaticBody = simpleKeywordStaticBody("Cascade", Cascade)

	// ProwessStaticBody is the reusable StaticAbilityBody for prowess.
	ProwessStaticBody = simpleKeywordStaticBody("Prowess", Prowess)

	// ImproviseStaticBody is the reusable StaticAbilityBody for improvise.
	ImproviseStaticBody = simpleKeywordStaticBody("Improvise", Improvise)

	// UndyingStaticBody is the reusable StaticAbilityBody for undying.
	UndyingStaticBody = simpleKeywordStaticBody("Undying", Undying)

	// PersistStaticBody is the reusable StaticAbilityBody for persist.
	PersistStaticBody = simpleKeywordStaticBody("Persist", Persist)

	// WitherStaticBody is the reusable StaticAbilityBody for wither.
	WitherStaticBody = simpleKeywordStaticBody("Wither", Wither)

	// InfectStaticBody is the reusable StaticAbilityBody for infect.
	InfectStaticBody = simpleKeywordStaticBody("Infect", Infect)

	// ExaltedStaticBody is the reusable StaticAbilityBody for exalted.
	ExaltedStaticBody = simpleKeywordStaticBody("Exalted", Exalted)
)

// Reusable AbilityDef compatibility templates for rules and tests.
// New card definitions should use the StaticAbilityBody templates above.
// Treat these values as immutable.
var (
	// DeathtouchAbility is the reusable AbilityDef template for deathtouch.
	DeathtouchAbility = simpleKeywordAbility("Deathtouch", Deathtouch)

	// DefenderAbility is the reusable AbilityDef template for defender.
	DefenderAbility = simpleKeywordAbility("Defender", Defender)

	// DoubleStrikeAbility is the reusable AbilityDef template for double strike.
	DoubleStrikeAbility = simpleKeywordAbility("Double strike", DoubleStrike)

	// FirstStrikeAbility is the reusable AbilityDef template for first strike.
	FirstStrikeAbility = simpleKeywordAbility("First strike", FirstStrike)

	// FlashAbility is the reusable AbilityDef template for flash.
	FlashAbility = simpleKeywordAbility("Flash", Flash)

	// FlyingAbility is the reusable AbilityDef template for flying.
	FlyingAbility = simpleKeywordAbility("Flying", Flying)

	// HasteAbility is the reusable AbilityDef template for haste.
	HasteAbility = simpleKeywordAbility("Haste", Haste)

	// HexproofAbility is the reusable AbilityDef template for hexproof.
	HexproofAbility = simpleKeywordAbility("Hexproof", Hexproof)

	// IndestructibleAbility is the reusable AbilityDef template for indestructible.
	IndestructibleAbility = simpleKeywordAbility("Indestructible", Indestructible)

	// LifelinkAbility is the reusable AbilityDef template for lifelink.
	LifelinkAbility = simpleKeywordAbility("Lifelink", Lifelink)

	// MenaceAbility is the reusable AbilityDef template for menace.
	MenaceAbility = simpleKeywordAbility("Menace", Menace)

	// ReachAbility is the reusable AbilityDef template for reach.
	ReachAbility = simpleKeywordAbility("Reach", Reach)

	// ShroudAbility is the reusable AbilityDef template for shroud.
	ShroudAbility = simpleKeywordAbility("Shroud", Shroud)

	// TrampleAbility is the reusable AbilityDef template for trample.
	TrampleAbility = simpleKeywordAbility("Trample", Trample)

	// VigilanceAbility is the reusable AbilityDef template for vigilance.
	VigilanceAbility = simpleKeywordAbility("Vigilance", Vigilance)

	// SplitSecondAbility is the reusable AbilityDef template for split second.
	SplitSecondAbility = simpleKeywordAbility("Split second", SplitSecond)

	// ConvokeAbility is the reusable AbilityDef template for convoke.
	ConvokeAbility = simpleKeywordAbility("Convoke", Convoke)

	// DelveAbility is the reusable AbilityDef template for delve.
	DelveAbility = simpleKeywordAbility("Delve", Delve)

	// StormAbility is the reusable AbilityDef template for storm.
	StormAbility = simpleKeywordAbility("Storm", Storm)

	// CascadeAbility is the reusable AbilityDef template for cascade.
	CascadeAbility = simpleKeywordAbility("Cascade", Cascade)

	// ProwessAbility is the reusable AbilityDef template for prowess.
	ProwessAbility = simpleKeywordAbility("Prowess", Prowess)

	// ImproviseAbility is the reusable AbilityDef template for improvise.
	ImproviseAbility = simpleKeywordAbility("Improvise", Improvise)

	// UndyingAbility is the reusable AbilityDef template for undying.
	UndyingAbility = simpleKeywordAbility("Undying", Undying)

	// PersistAbility is the reusable AbilityDef template for persist.
	PersistAbility = simpleKeywordAbility("Persist", Persist)

	// WitherAbility is the reusable AbilityDef template for wither.
	WitherAbility = simpleKeywordAbility("Wither", Wither)

	// InfectAbility is the reusable AbilityDef template for infect.
	InfectAbility = simpleKeywordAbility("Infect", Infect)

	// ExaltedAbility is the reusable AbilityDef template for exalted.
	ExaltedAbility = simpleKeywordAbility("Exalted", Exalted)
)

func simpleKeywordAbility(text string, keyword Keyword) AbilityDef {
	return AbilityDef{Body: simpleKeywordStaticBody(text, keyword)}
}

func simpleKeywordStaticBody(text string, keyword Keyword) StaticAbilityBody {
	return StaticAbilityBody{Text: text, KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: keyword}}}
}

// TriggerType classifies what kind of event triggers a triggered ability.
type TriggerType int

// Trigger type values identify supported trigger wordings.
const (
	TriggerWhen     TriggerType = iota // "When [event]" — fires once
	TriggerWhenever                    // "Whenever [event]" — fires each time
	TriggerAt                          // "At the beginning of [step]"
	TriggerState                       // State trigger checked whenever a player would get priority
)

// TriggerCondition describes when a triggered ability fires.
type TriggerCondition struct {
	// Type is whether this is a When, Whenever, or At trigger.
	Type TriggerType

	// Pattern is the structured event pattern this ability listens for.
	Pattern TriggerPattern

	// InterveningIf is the "if" condition that must be true both when the
	// event occurs and when the trigger resolves (CR 603.4). Empty if none.
	InterveningIf string

	// InterveningCondition is the structured form of InterveningIf. The rules
	// layer evaluates it with the trigger controller and triggering event bound.
	InterveningCondition opt.V[Condition]

	// InterveningIfControllerLifeAtLeast is a structured initial intervening-if
	// condition for life-threshold triggers.
	InterveningIfControllerLifeAtLeast int

	// InterveningIfEventPermanentHadCounters is true for intervening-if clauses
	// such as "if it had counters on it" on zone-change triggers. mtg/rules
	// checks the event permanent's current object or last-known information.
	InterveningIfEventPermanentHadCounters bool

	// State describes a state trigger. State triggers latch while true and only
	// trigger again after becoming false, then true again (CR 603.8).
	State opt.V[StateTriggerCondition]
}

// StateTriggerCondition describes a simple state trigger condition. Empty
// fields mean no state condition is active.
type StateTriggerCondition struct {
	MatchControllerLifeLessOrEqual bool
	ControllerLifeLessOrEqual      int
}

// TriggerControllerFilter constrains a trigger by the controller recorded on an event.
type TriggerControllerFilter int

// Trigger controller filters match events by controller.
const (
	TriggerControllerAny TriggerControllerFilter = iota
	TriggerControllerYou
	TriggerControllerOpponent
)

// TriggerSourceFilter constrains a trigger by the source of the event.
type TriggerSourceFilter int

// Trigger source filters match events by source.
const (
	TriggerSourceAny TriggerSourceFilter = iota
	TriggerSourceSelf
	TriggerSourceAttachedPermanent
)

// TriggerSubjectObject identifies which permanent on an event is the trigger
// subject for source/controller matching. Event-specific object fields that are
// not the subject, such as EventBlockerDeclared.PermanentID for the blocker,
// continue to feed general permanent filters.
type TriggerSubjectObject int

// Trigger subject object values identify event permanent roles.
const (
	TriggerSubjectDefault TriggerSubjectObject = iota
	TriggerSubjectPermanent
	TriggerSubjectBlockedAttacker
)

// TriggerPlayerFilter constrains a trigger by the affected player recorded on an event.
type TriggerPlayerFilter int

// Trigger player filters match events by affected player.
const (
	TriggerPlayerAny TriggerPlayerFilter = iota
	TriggerPlayerYou
	TriggerPlayerOpponent
)

// TriggerPattern matches a GameEvent for triggered-ability detection.
// Zero-valued filters are wildcards except Event, which must be set.
type TriggerPattern struct {
	Event EventKind

	Controller  TriggerControllerFilter
	Source      TriggerSourceFilter
	ExcludeSelf bool
	Player      TriggerPlayerFilter

	Subject TriggerSubjectObject

	RequirePermanentTypes []types.Card
	ExcludePermanentTypes []types.Card
	RequireNonToken       bool

	// RequireCardTypes and ExcludeCardTypes filter spell-cast events by the
	// spell's types as chosen/cast on the stack (CR 601.2, CR 603.2).
	RequireCardTypes []types.Card
	ExcludeCardTypes []types.Card

	MatchFromZone bool
	FromZone      ZoneType
	MatchToZone   bool
	ToZone        ZoneType

	MatchStackObjectKind bool
	StackObjectKind      StackObjectKind

	DamageRecipient            DamageRecipientKind
	DamageRecipientCombatState CombatStateFilter

	SpellTargetsSource bool
	SpellTargetAllow   TargetAllow
	SpellTargetPattern opt.V[TargetPredicate]

	// OneOrMore coalesces matching events consumed in the same trigger detection
	// pass into one trigger. The first matching event is retained as TriggerEvent.
	OneOrMore bool

	// Step filters EventBeginningOfStep triggers such as "At the beginning of
	// your upkeep" (CR 603.6c).
	Step Step
}

// TimingRestriction constrains when an activated ability can be used.
type TimingRestriction int

const (
	// NoTimingRestriction means the ability can be activated at instant speed.
	NoTimingRestriction TimingRestriction = iota

	// SorceryOnly means "activate only as a sorcery" (CR 113.6e).
	SorceryOnly

	// OncePerTurn means "activate only once each turn.".
	OncePerTurn

	// SorceryOncePerTurn combines both restrictions.
	SorceryOncePerTurn

	// DuringCombat means "activate only during combat.".
	DuringCombat

	// DuringUpkeep means "activate only during your upkeep.".
	DuringUpkeep
)

// EffectType classifies a legacy Effect. Card Implementations use typed
// Primitive variants instead.
type EffectType int

// Effect type values enumerate supported effect categories.
const (
	EffectUnknown EffectType = iota
	EffectDamage
	EffectDestroy
	EffectExile
	EffectBounce
	EffectCounter
	EffectDraw
	EffectDiscard
	EffectMill
	EffectSearch
	EffectCreateToken
	EffectGainLife
	EffectLoseLife
	EffectAddMana
	EffectModifyPT
	EffectAddCounter
	EffectRemoveCounter
	EffectPutOnBattlefield
	EffectSacrifice
	EffectTap
	EffectUntap
	EffectGainControl
	EffectCopy
	EffectScry
	EffectSurveil
	EffectFight
	EffectTransform
	EffectAttach
	EffectReplace
	EffectPrevent
	EffectCreateDelayedTrigger
	EffectRegenerate
	EffectSkipStep
	EffectPhaseOut
	EffectCreateEmblem
	EffectApplyContinuous
	EffectMoveCounters
	EffectChoose
	EffectPay
	EffectApplyRule
	EffectProliferate
	EffectGoad
	EffectReveal
	EffectInvestigate
	EffectShufflePermanentIntoLibrary
	EffectDiscover
	EffectStartEngines
	EffectSetClassLevel
	EffectMonstrosity
)

// EffectResultAmountKind identifies which numeric result an effect records for
// later linked "that much" or X instructions.
type EffectResultAmountKind int

// Effect result amount values select stored numeric results.
const (
	EffectResultAmountDefault EffectResultAmountKind = iota
	EffectResultAmountExcessDamage
)

// EffectSelector identifies a set of permanents affected by a mass effect.
type EffectSelector string

// Effect selector values identify mass-effect recipient groups.
const (
	EffectSelectorNone                                  EffectSelector = ""
	EffectSelectorAllCreatures                          EffectSelector = "all creatures"
	EffectSelectorAllCreaturesExceptTarget              EffectSelector = "all creatures except target"
	EffectSelectorAllArtifacts                          EffectSelector = "all artifacts"
	EffectSelectorAllEnchantments                       EffectSelector = "all enchantments"
	EffectSelectorAllNonlandPermanents                  EffectSelector = "all nonland permanents"
	EffectSelectorAllPermanents                         EffectSelector = "all permanents"
	EffectSelectorCreaturesYouControl                   EffectSelector = "creatures you control"
	EffectSelectorOtherCreaturesYouControl              EffectSelector = "other creatures you control"
	EffectSelectorEquippedCreature                      EffectSelector = "equipped creature"
	EffectSelectorOtherCreaturesDefendingPlayerControls EffectSelector = "other creatures defending player controls"
)

// PlayerSelector identifies a set of players affected by a mass effect.
type PlayerSelector string

// Player selector values identify groups of affected players.
const (
	PlayerSelectorNone      PlayerSelector = ""
	PlayerSelectorOpponents PlayerSelector = "opponents"
)

// CounterSourceKind identifies where an effect reads counters from.
type CounterSourceKind int

// Counter source values identify where counter-moving effects read counters.
const (
	CounterSourceNone CounterSourceKind = iota

	// CounterSourceTarget reads counters from another chosen target.
	CounterSourceTarget

	// CounterSourceEventPermanent reads counters from the event permanent that
	// caused a triggered ability to trigger. If that permanent has left the
	// battlefield, mtg/rules reads its last-known information.
	CounterSourceEventPermanent
)

// CounterSourceSpec describes the source object for counter-moving effects.
type CounterSourceSpec struct {
	Kind        CounterSourceKind
	TargetIndex int
}

// TokenCopySource identifies what object/card supplies copiable values for a
// token-copy effect.
type TokenCopySource int

// Token copy source values identify what supplies copiable values.
const (
	TokenCopySourceNone TokenCopySource = iota
	TokenCopySourceObject
	TokenCopySourceSourceCard
)

// TokenCopySpec describes a token that starts as a copy of another object/card,
// then applies explicit copy-modifying exceptions such as Eternalize's color,
// type, power/toughness, and mana-cost overrides.
type TokenCopySpec struct {
	Source TokenCopySource
	Object ObjectReference

	SetName       string
	SetColors     []color.Color
	SetTypes      []types.Card
	SetSubtypes   []types.Sub
	SetPower      opt.V[PT]
	SetToughness  opt.V[PT]
	NoManaCost    bool
	NoPrintedText bool
}

// EternalizeAbility builds the keyword's full activated ability: a sorcery-speed
// graveyard activation that exiles the source card and creates the standard
// 4/4 black Zombie copy token with no mana cost.
func EternalizeAbility(manaCost cost.Mana, creatureSubtypes ...types.Sub) AbilityDef {
	return AbilityDef{Body: EternalizeActivatedBody(manaCost, creatureSubtypes...)}
}

// EternalizeActivatedBody builds the ActivatedAbilityBody for the Eternalize
// keyword. Use this in CardFace.ActivatedAbilities when the card source uses
// categorized fields instead of the legacy Abilities slice.
func EternalizeActivatedBody(manaCost cost.Mana, creatureSubtypes ...types.Sub) ActivatedAbilityBody {
	tokenSubtypes := make([]types.Sub, 0, len(creatureSubtypes)+1)
	tokenSubtypes = append(tokenSubtypes, types.Zombie)
	tokenSubtypes = append(tokenSubtypes, creatureSubtypes...)
	return ActivatedAbilityBody{
		Text:           "Eternalize " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: ZoneGraveyard,
		Timing:         SorceryOnly,
		AdditionalCosts: []AdditionalCost{{
			Kind: AdditionalCostExileSource,
			Text: "Exile this card from your graveyard",
		}},
		Content: PlainAbilityContent{Sequence: []Instruction{{
			Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenCopyOf(TokenCopySpec{
					Source:       TokenCopySourceSourceCard,
					SetColors:    []color.Color{color.Black},
					SetSubtypes:  tokenSubtypes,
					SetPower:     opt.Val(PT{Value: 4}),
					SetToughness: opt.Val(PT{Value: 4}),
					NoManaCost:   true,
				}),
			},
		}}},
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Eternalize}},
	}
}

// SearchSpec describes a deterministic library-search slice. The rules
// implementation supports library -> hand and library -> battlefield templates
// with common type, supertype, and subtype filters.
type SearchSpec struct {
	SourceZone  ZoneType
	Destination ZoneType

	CardType  opt.V[types.Card]
	Supertype opt.V[types.Super]

	SubtypesAny []types.Sub

	Reveal       bool
	Shuffle      bool
	EntersTapped bool
}

// EffectCondition describes a simple condition that must be true when an
// effect resolves. It is data only; mtg/rules owns evaluation.
type EffectCondition struct {
	// Text preserves the printed condition for logs, diagnostics, and review.
	Text string

	// TargetIndex identifies the target whose current characteristics are tested.
	TargetIndex int

	PermanentType opt.V[types.Card]

	// Negate inverts the permanent-type match, e.g. "it isn't a creature".
	Negate bool

	// Condition is an additional shared condition evaluated with the resolving
	// stack object bound.
	Condition opt.V[Condition]
}

// TargetIndex sentinel values identify objects or players that are not chosen
// targets. Non-negative TargetIndex values index into the runtime targets chosen
// for the spell or ability.
const (
	// TargetIndexController means the effect applies to that spell or ability's controller.
	TargetIndexController = -1

	// TargetIndexSourcePermanent means the effect refers to the source permanent.
	TargetIndexSourcePermanent = -2
)

// Effect describes a legacy game effect produced by an ability. Card
// Implementations use Instruction and typed Primitive variants instead.
type Effect struct {
	Type          EffectType
	Amount        int
	DynamicAmount opt.V[DynamicAmount]
	TargetIndex   int
	// RelatedTargetIndex identifies a second chosen target used with the
	// primary TargetIndex by paired-object effects such as fight. If unset,
	// those effects use their historical default target ordering.
	RelatedTargetIndex opt.V[int]
	Object             opt.V[ObjectReference]
	DamageSource       opt.V[ObjectReference]
	Recipient          opt.V[PlayerReference]
	Condition          opt.V[EffectCondition]
	CardCondition      opt.V[CardCondition]

	// Optional asks the effect's controller whether to apply this single
	// resolution instruction. LinkID can be used with ResultCondition on later
	// effects to model "if you do" / "if you don't" branches as instructions are
	// followed in order (CR 608.2c).
	Optional        bool
	ResultCondition opt.V[EffectResultCondition]

	PowerDelta            int
	ToughnessDelta        int
	PowerDeltaDynamic     opt.V[DynamicAmount]
	ToughnessDeltaDynamic opt.V[DynamicAmount]
	ResultAmount          EffectResultAmountKind
	CounterKind           counter.Kind
	CounterSource         CounterSourceSpec
	ManaColor             mana.Color
	// Choice asks for a value while resolving this instruction and stores it
	// under LinkID for later instructions to consume (CR 608.2c, CR 609.3).
	Choice opt.V[ResolutionChoice]
	// ChoiceLinkID consumes a value produced by a prior resolution choice, such
	// as a chosen color for mana or a chosen player for player effects.
	ChoiceLinkID string
	// Payment asks the controller whether to pay a cost during resolution; the
	// payment result is recorded through LinkID/ResultCondition (CR 608.2c,
	// CR 117.12).
	Payment           opt.V[ResolutionPayment]
	UntilEndOfTurn    bool
	Duration          EffectDuration
	Step              Step
	Selector          EffectSelector
	PlayerSelector    PlayerSelector
	Token             opt.V[*CardDef]
	TokenCopy         opt.V[TokenCopySpec]
	ContinuousEffects []ContinuousEffect
	DelayedTrigger    opt.V[DelayedTriggerDef]
	EmblemAbilities   []AbilityDef
	Replacement       opt.V[ReplacementEffect]
	RuleEffects       []RuleEffect
	Search            opt.V[SearchSpec]
	Card              opt.V[CardReference]
	LinkID            string
	Description       string
}

// TargetSpec describes the targeting requirements of an ability.
type TargetSpec struct {
	// MinTargets is the minimum number of targets (0 for "up to").
	MinTargets int

	// MaxTargets is the maximum number of targets.
	MaxTargets int

	// Constraint describes what can be targeted (e.g., "creature",
	// "creature or planeswalker", "player").
	Constraint string

	// Allow and Predicate provide structured target legality for generated card
	// definitions. Constraint remains for display and as a legacy fallback.
	Allow     TargetAllow
	Predicate TargetPredicate

	// Chooser identifies who chooses this target slot during announcement. The
	// default controller chooser preserves normal targeting. For non-controller
	// choosers, structured "you" predicates are evaluated relative to the
	// choosing player.
	Chooser TargetChooser
}

// TargetChooser identifies who chooses a target slot during announcement.
type TargetChooser int

// Target chooser values identify who chooses a target slot.
const (
	TargetChooserController TargetChooser = iota
	// TargetChooserOpponent means the ability controller chooses an opponent,
	// then that opponent chooses this target slot.
	TargetChooserOpponent
)

// Mode represents one mode of a modal spell or ability ("Choose one —",
// "Choose two —", etc.; CR 700.2).
type Mode struct {
	// Text is the oracle text of this mode.
	Text string

	// LegacyEffects are the legacy effects this mode produces.
	// New Card Implementations use Sequence.
	LegacyEffects []Effect

	// Sequence is the typed instruction sequence this mode produces.
	Sequence []Instruction

	// Targets are the targeting requirements of this mode.
	Targets []TargetSpec
}

// AbilityDef is the static definition of a single ability on a card.
// It describes the ability as printed, not as modified by continuous
// effects during gameplay.
type AbilityDef struct {
	// Kind classifies this ability (spell, activated, triggered, or static).
	//
	// Deprecated: use Body for new card definitions.
	Kind AbilityKind

	// Text is the full oracle text of this ability paragraph.
	Text string

	// Body is the sealed ability-body variant for new card definitions. Legacy
	// fields remain populated while the rules engine migrates incrementally.
	Body AbilityBody

	// Condition restricts when a static ability functions. It is currently used
	// only for static abilities; activation restrictions belong in
	// ActivationCondition.
	Condition opt.V[Condition]

	// KeywordAbilities lists sealed keyword variants this ability provides.
	KeywordAbilities []KeywordAbility

	// ManaCost is the mana component of an activated ability's cost.
	// Nil for non-activated abilities.
	ManaCost opt.V[cost.Mana]

	// AdditionalCosts describes typed non-mana costs. mtg/rules owns choosing
	// and applying these costs.
	AdditionalCosts []AdditionalCost

	// AlternativeCosts are optional costs that replace the normal mana cost
	// when selected. Required additional costs still apply.
	AlternativeCosts []AlternativeCost

	// Trigger defines when a triggered ability fires. Nil for non-triggered.
	Trigger opt.V[TriggerCondition]

	// MaxTriggersPerTurn limits how many times this ability can trigger each
	// turn. Zero means no limit.
	MaxTriggersPerTurn int

	// Optional is true for "you may" abilities. Triggered abilities still go on
	// the stack; the controller chooses whether to apply their effects on
	// resolution.
	Optional bool

	// Effects lists the effects this ability produces.
	Effects []Effect

	// Targets lists targeting requirements. Empty for untargeted abilities.
	Targets []TargetSpec

	// Modes lists the modes for modal spells/abilities. Empty for non-modal.
	Modes []Mode

	// MinModes and MaxModes constrain modal choices locked in while casting or
	// activating a modal object (CR 601.2d, CR 700.2). For legacy choose-one
	// modal abilities, leave both zero and the rules layer treats the ability as
	// choosing exactly one mode.
	MinModes int
	MaxModes int

	// AllowDuplicateModes permits the same mode to be chosen more than once when
	// the modal text explicitly allows it (CR 700.2d).
	AllowDuplicateModes bool

	// ZoneOfFunction is the zone where this ability functions.
	// Defaults to Battlefield for permanents (CR 113.6).
	ZoneOfFunction ZoneType

	// Timing restricts when an activated ability can be used.
	Timing TimingRestriction

	// ActivationCondition restricts when an activated ability can be activated,
	// e.g. "Activate only if you control a Mountain".
	ActivationCondition opt.V[Condition]

	// IsLoyaltyAbility is true for planeswalker loyalty abilities (CR 606).
	IsLoyaltyAbility bool

	// LoyaltyCost is the loyalty cost (+N, 0, or -N) for loyalty abilities.
	LoyaltyCost int

	// IsManaAbility is true if this is a mana ability (CR 605.1):
	// produces mana, no targets, not a loyalty ability.
	IsManaAbility bool
}
