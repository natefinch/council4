package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
	// Non-combat keywords
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

// TriggerType classifies what kind of event triggers a triggered ability.
type TriggerType int

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

	// InterveningIfControllerLifeAtLeast is a structured initial intervening-if
	// condition for life-threshold triggers.
	InterveningIfControllerLifeAtLeast int

	// InterveningIfEventPermanentHadCounters is true for intervening-if clauses
	// such as "if it had counters on it" on zone-change triggers. mtg/rules
	// checks the event permanent's current object or last-known information.
	InterveningIfEventPermanentHadCounters bool

	// State describes a state trigger. State triggers latch while true and only
	// trigger again after becoming false, then true again (CR 603.8).
	State *StateTriggerCondition
}

// StateTriggerCondition describes a simple state trigger condition. Empty
// fields mean no state condition is active.
type StateTriggerCondition struct {
	MatchControllerLifeLessOrEqual bool
	ControllerLifeLessOrEqual      int
}

// TriggerControllerFilter constrains a trigger by the controller recorded on an event.
type TriggerControllerFilter int

const (
	TriggerControllerAny TriggerControllerFilter = iota
	TriggerControllerYou
	TriggerControllerOpponent
)

// TriggerSourceFilter constrains a trigger by the source of the event.
type TriggerSourceFilter int

const (
	TriggerSourceAny TriggerSourceFilter = iota
	TriggerSourceSelf
)

// TriggerPlayerFilter constrains a trigger by the affected player recorded on an event.
type TriggerPlayerFilter int

const (
	TriggerPlayerAny TriggerPlayerFilter = iota
	TriggerPlayerYou
	TriggerPlayerOpponent
)

// TriggerPattern matches a GameEvent for triggered-ability detection.
// Zero-valued filters are wildcards except Event, which must be set.
type TriggerPattern struct {
	Event EventKind

	Controller TriggerControllerFilter
	Source     TriggerSourceFilter
	Player     TriggerPlayerFilter

	MatchPermanentType    bool
	PermanentType         CardType
	RequirePermanentTypes []CardType
	ExcludePermanentTypes []CardType

	// RequireCardTypes and ExcludeCardTypes filter spell-cast events by the
	// spell's types as chosen/cast on the stack (CR 601.2, CR 603.2).
	RequireCardTypes []CardType
	ExcludeCardTypes []CardType

	MatchFromZone bool
	FromZone      ZoneType
	MatchToZone   bool
	ToZone        ZoneType

	DamageRecipient DamageRecipientKind

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

	// OncePerTurn means "activate only once each turn."
	OncePerTurn

	// SorceryOncePerTurn combines both restrictions.
	SorceryOncePerTurn

	// DuringCombat means "activate only during combat."
	DuringCombat

	// DuringUpkeep means "activate only during your upkeep."
	DuringUpkeep
)

// EffectType classifies the broad category of an effect for future rules
// engine processing. This is a placeholder — the rules engine will define
// richer effect representations.
type EffectType int

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
)

// EffectSelector identifies a set of permanents affected by a mass effect.
type EffectSelector string

const (
	EffectSelectorNone                     EffectSelector = ""
	EffectSelectorAllCreatures             EffectSelector = "all creatures"
	EffectSelectorAllArtifacts             EffectSelector = "all artifacts"
	EffectSelectorAllEnchantments          EffectSelector = "all enchantments"
	EffectSelectorAllNonlandPermanents     EffectSelector = "all nonland permanents"
	EffectSelectorAllPermanents            EffectSelector = "all permanents"
	EffectSelectorCreaturesYouControl      EffectSelector = "creatures you control"
	EffectSelectorOtherCreaturesYouControl EffectSelector = "other creatures you control"
)

// CounterSourceKind identifies where an effect reads counters from.
type CounterSourceKind int

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

// EffectCondition describes a simple condition that must be true when an
// effect resolves. It is data only; mtg/rules owns evaluation.
type EffectCondition struct {
	// Text preserves the printed condition for logs, diagnostics, and review.
	Text string

	// TargetIndex identifies the target whose current characteristics are tested.
	TargetIndex int

	MatchPermanentType bool
	PermanentType      CardType

	// Negate inverts the permanent-type match, e.g. "it isn't a creature".
	Negate bool
}

// Effect describes a single game effect produced by an ability.
// TargetIndex indexes into the runtime targets chosen for the spell or ability;
// -1 means the effect applies to that spell or ability's controller.
type Effect struct {
	Type          EffectType
	Amount        int
	DynamicAmount *DynamicAmount
	TargetIndex   int
	Condition     *EffectCondition

	// Optional asks the effect's controller whether to apply this single
	// resolution instruction. LinkID can be used with ResultCondition on later
	// effects to model "if you do" / "if you don't" branches as instructions are
	// followed in order (CR 608.2c).
	Optional        bool
	ResultCondition *EffectResultCondition

	PowerDelta     int
	ToughnessDelta int
	CounterKind    counter.Kind
	CounterSource  CounterSourceSpec
	ManaColor      mana.Color
	// Choice asks for a value while resolving this instruction and stores it
	// under LinkID for later instructions to consume (CR 608.2c, CR 609.3).
	Choice *ResolutionChoice
	// ChoiceLinkID consumes a value produced by a prior resolution choice, such
	// as a chosen color for mana or a chosen player for player effects.
	ChoiceLinkID string
	// Payment asks the controller whether to pay a cost during resolution; the
	// payment result is recorded through LinkID/ResultCondition (CR 608.2c,
	// CR 117.12).
	Payment           *ResolutionPayment
	UntilEndOfTurn    bool
	Duration          EffectDuration
	Step              Step
	Selector          EffectSelector
	Token             *CardDef
	ContinuousEffects []ContinuousEffect
	DelayedTrigger    *DelayedTriggerDef
	EmblemAbilities   []AbilityDef
	Replacement       *ReplacementEffect
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
}

// Mode represents one mode of a modal spell or ability ("Choose one —",
// "Choose two —", etc.; CR 700.2).
type Mode struct {
	// Text is the oracle text of this mode.
	Text string

	// Effects are the effects this mode produces.
	Effects []Effect

	// Targets are the targeting requirements of this mode.
	Targets []TargetSpec
}

// AbilityDef is the static definition of a single ability on a card.
// It describes the ability as printed, not as modified by continuous
// effects during gameplay.
type AbilityDef struct {
	// Kind classifies this ability (spell, activated, triggered, or static).
	Kind AbilityKind

	// Text is the full oracle text of this ability paragraph.
	Text string

	// Keywords lists keyword abilities this provides (e.g., Flying, Haste).
	// A single ability line can grant multiple keywords.
	Keywords []Keyword

	// ProtectionFromColors parameterizes Protection for the initial protection
	// slice. Empty means this ability does not currently grant rules-relevant
	// protection, even if Keywords includes Protection.
	ProtectionFromColors []mana.Color

	// ManaCost is the mana component of an activated ability's cost.
	// Nil for non-activated abilities.
	ManaCost *mana.Cost

	// AdditionalCosts describes typed non-mana costs. mtg/rules owns choosing
	// and applying these costs.
	AdditionalCosts []AdditionalCost

	// AlternativeCosts are optional costs that replace the normal mana cost
	// when selected. Required additional costs still apply.
	AlternativeCosts []AlternativeCost

	// KickerCost is an optional additional mana cost for Kicker.
	KickerCost *mana.Cost

	// KickerEffects are additional effects applied if the spell was kicked.
	KickerEffects []Effect

	// Trigger defines when a triggered ability fires. Nil for non-triggered.
	Trigger *TriggerCondition

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

	// IsLoyaltyAbility is true for planeswalker loyalty abilities (CR 606).
	IsLoyaltyAbility bool

	// LoyaltyCost is the loyalty cost (+N, 0, or -N) for loyalty abilities.
	LoyaltyCost int

	// IsManaAbility is true if this is a mana ability (CR 605.1):
	// produces mana, no targets, not a loyalty ability.
	IsManaAbility bool
}
