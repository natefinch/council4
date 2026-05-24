package game

import "github.com/natefinch/council4/mtg/game/mana"

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
)

// TriggerCondition describes when a triggered ability fires.
type TriggerCondition struct {
	// Type is whether this is a When, Whenever, or At trigger.
	Type TriggerType

	// Pattern is the structured event pattern this ability listens for.
	Pattern TriggerPattern

	// Event is a description of the triggering event (e.g., "this creature
	// enters the battlefield", "a creature dies").
	// Deprecated: use Pattern for rules behavior.
	Event string

	// InterveningIf is the "if" condition that must be true both when the
	// event occurs and when the trigger resolves (CR 603.4). Empty if none.
	InterveningIf string
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

	MatchPermanentType bool
	PermanentType      CardType

	MatchFromZone bool
	FromZone      ZoneType
	MatchToZone   bool
	ToZone        ZoneType

	DamageRecipient DamageRecipientKind
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

// Effect describes a single game effect produced by an ability.
// TargetIndex indexes into the runtime targets chosen for the spell or ability;
// -1 means the effect applies to that spell or ability's controller.
type Effect struct {
	Type           EffectType
	Amount         int
	TargetIndex    int
	PowerDelta     int
	ToughnessDelta int
	ManaColor      mana.Color
	UntilEndOfTurn bool
	Selector       EffectSelector
	Token          *CardDef
	Description    string
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
}

// Mode represents one mode of a modal spell or ability ("Choose one —").
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

	// AdditionalCost describes non-mana costs (e.g., "Sacrifice a creature",
	// "Tap", "Pay 2 life"). Empty if none.
	AdditionalCost string

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
