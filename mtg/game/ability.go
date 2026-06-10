package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Keyword represents an evergreen or commonly-used keyword ability (CR 702).
type Keyword int

// Keyword values enumerate supported keyword abilities.
const (
	KeywordNone Keyword = iota
	Devoid
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
	ReadAhead
)

// Reusable StaticAbilityBody templates for non-parameterized keyword abilities.
// Use these in CardFace.StaticAbilities slices or initializer-function appends.
// Treat these values as immutable.
var (
	// DevoidStaticBody is the reusable StaticAbilityBody for devoid.
	DevoidStaticBody = simpleKeywordStaticBody("Devoid", Devoid)

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

	// ReadAheadStaticBody is the reusable StaticAbilityBody for read ahead.
	ReadAheadStaticBody = simpleKeywordStaticBody("Read ahead", ReadAhead)
)

func simpleKeywordStaticBody(text string, keyword Keyword) StaticAbility {
	return StaticAbility{Text: text, KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: keyword}}}
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

	// InterveningIfEventPermanentHadNoCounterKind identifies a counter kind that
	// must be absent from the event permanent's current object or last-known
	// information.
	InterveningIfEventPermanentHadNoCounterKind opt.V[counter.Kind]

	// InterveningIfEventPermanentWasKicked is true for "if it was kicked" on
	// enter triggers. The entering permanent event preserves the spell's kicker
	// choice for both trigger-time and resolution-time checks.
	InterveningIfEventPermanentWasKicked bool

	// InterveningIfEventPermanentWasCast is true for "if it was cast" and "if
	// you cast it" on enter triggers.
	InterveningIfEventPermanentWasCast bool

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

// TriggerPattern matches a Event for triggered-ability detection.
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

	// SubjectSelection is the Selection-based form of the event subject
	// permanent filters (RequirePermanentTypes/ExcludePermanentTypes and
	// RequireNonToken). It is wildcard by default; the rules matcher adapts the
	// legacy fields when it is empty, and the two forms must not both be set.
	SubjectSelection Selection

	// RequireCardTypes and ExcludeCardTypes filter spell-cast events by the
	// spell's types as chosen/cast on the stack (CR 601.2, CR 603.2).
	RequireCardTypes []types.Card
	ExcludeCardTypes []types.Card

	// CardSelection is the Selection-based form of the cast-spell card filters
	// (RequireCardTypes/ExcludeCardTypes). It is wildcard by default; the rules
	// matcher adapts the legacy fields when it is empty, and the two forms must
	// not both be set.
	CardSelection Selection

	MatchFromZone bool
	FromZone      zone.Type
	MatchToZone   bool
	ToZone        zone.Type

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

// EffectResultAmountKind identifies which numeric result an effect records for
// later linked "that much" or X instructions.
type EffectResultAmountKind int

// Effect result amount values select stored numeric results.
const (
	EffectResultAmountDefault EffectResultAmountKind = iota
	EffectResultAmountExcessDamage
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
	Kind   CounterSourceKind
	Object ObjectReference
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

// EternalizeActivatedBody builds the ActivatedAbilityBody for the Eternalize
// keyword. Use this in CardFace.ActivatedAbilities with categorized fields.
func EternalizeActivatedBody(manaCost cost.Mana, creatureSubtypes ...types.Sub) ActivatedAbility {
	tokenSubtypes := make([]types.Sub, 0, len(creatureSubtypes)+1)
	tokenSubtypes = append(tokenSubtypes, types.Zombie)
	tokenSubtypes = append(tokenSubtypes, creatureSubtypes...)
	return ActivatedAbility{
		Text:           "Eternalize " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: zone.Graveyard,
		Timing:         SorceryOnly,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalExileSource,
			Text: "Exile this card from your graveyard",
		}},
		Content: Mode{Sequence: []Instruction{{
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
		}}}.Ability(),

		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Eternalize}},
	}
}

// SearchSpec describes a deterministic library-search slice. The rules
// implementation supports library -> hand and library -> battlefield templates
// with common type, supertype, and subtype filters.
type SearchSpec struct {
	SourceZone  zone.Type
	Destination zone.Type

	CardType  opt.V[types.Card]
	Supertype opt.V[types.Super]

	SubtypesAny []types.Sub

	Reveal       bool
	EntersTapped bool
}

// EffectCondition describes a simple condition that must be true when an
// effect resolves. It is data only; mtg/rules owns evaluation.
type EffectCondition struct {
	// Text preserves the printed condition for logs, diagnostics, and review.
	Text string

	// Object identifies the object whose current characteristics are tested.
	Object ObjectReference

	PermanentType opt.V[types.Card]

	// Negate inverts the permanent-type match, e.g. "it isn't a creature".
	Negate bool

	// Condition is an additional shared condition evaluated with the resolving
	// stack object bound.
	Condition opt.V[Condition]
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

	// Selection is the shared Selection-based form of the structured target
	// predicate. When present it supersedes Predicate; the two must not both be
	// specified. Predicate remains for existing cards and is adapted to a
	// Selection by the rules matcher when Selection is absent.
	Selection opt.V[Selection]

	// TargetZone restricts card targets to one zone. It is meaningful only when
	// Allow includes TargetAllowCard.
	TargetZone zone.Type

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

	// Targets are the targeting requirements of this mode.
	Targets []TargetSpec

	// Sequence is the typed instruction sequence this mode produces.
	Sequence []Instruction
}

// Ability creates ordinary non-modal ability content from this mode.
func (m Mode) Ability() AbilityContent {
	return AbilityContent{
		Modes:               []Mode{m},
		MinModes:            1,
		MaxModes:            1,
		AllowDuplicateModes: false,
	}
}
