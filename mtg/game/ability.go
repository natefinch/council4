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

	// Event is a description of the triggering event (e.g., "this creature
	// enters the battlefield", "a creature dies").
	Event string

	// InterveningIf is the "if" condition that must be true both when the
	// event occurs and when the trigger resolves (CR 603.4). Empty if none.
	InterveningIf string
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

// Effect describes a single game effect produced by an ability.
// This is a placeholder structure — the rules engine will expand this
// with concrete implementations.
type Effect struct {
	Type        EffectType
	Description string
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

	// ManaCost is the mana component of an activated ability's cost.
	// Nil for non-activated abilities.
	ManaCost *mana.Cost

	// AdditionalCost describes non-mana costs (e.g., "Sacrifice a creature",
	// "Tap", "Pay 2 life"). Empty if none.
	AdditionalCost string

	// Trigger defines when a triggered ability fires. Nil for non-triggered.
	Trigger *TriggerCondition

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
