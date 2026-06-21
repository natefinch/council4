package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// TriggerEvent identifies a representable rules event without depending on
// runtime game values.
type TriggerEvent uint8

// Trigger event families recognized by the semantic compiler.
const (
	TriggerEventUnknown TriggerEvent = iota
	TriggerEventSpellCast
	TriggerEventPermanentEnteredBattlefield
	TriggerEventPermanentDied
	TriggerEventZoneChanged
	TriggerEventCountersAdded
	TriggerEventDamageDealt
	TriggerEventCardDrawn
	TriggerEventAttackerDeclared
	TriggerEventBlockerDeclared
	TriggerEventCardDiscarded
	TriggerEventCycled
	TriggerEventBeginningOfStep
	TriggerEventLifeGained
	TriggerEventLifeLost
	TriggerEventPermanentTapped
	TriggerEventPermanentUntapped
	TriggerEventPermanentTurnedFaceUp
	TriggerEventPermanentSacrificed
	TriggerEventScry
	TriggerEventSurveil
	TriggerEventAbilityActivated
	TriggerEventObjectBecameTarget
	TriggerEventPermanentMutated
	TriggerEventAttackerBecameBlocked
	TriggerEventTokenCreated
	TriggerEventLibrarySearched
)

// TriggerSourceRelation identifies the event object's relationship to the
// ability source.
type TriggerSourceRelation uint8

// Trigger source relations.
const (
	TriggerSourceAny TriggerSourceRelation = iota
	TriggerSourceSelf
	TriggerSourceAttachedPermanent
)

// TriggerSubject identifies the event permanent role used for source and
// controller matching.
type TriggerSubject uint8

// Trigger subject roles.
const (
	TriggerSubjectDefault TriggerSubject = iota
	TriggerSubjectPermanent
	TriggerSubjectBlockedAttacker
	TriggerSubjectDamageSource
)

// TriggerPlayerRelation identifies an affected player's relationship to the
// ability controller.
type TriggerPlayerRelation uint8

// Trigger player relations.
const (
	TriggerPlayerAny TriggerPlayerRelation = iota
	TriggerPlayerYou
	TriggerPlayerOpponent
)

// TriggerZone identifies a zone used by a trigger pattern.
type TriggerZone uint8

// Trigger zones.
const (
	TriggerZoneNone TriggerZone = iota
	TriggerZoneGraveyard
	TriggerZoneBattlefield
	TriggerZoneHand
	TriggerZoneExile
	TriggerZoneLibrary
	TriggerZoneStack
	TriggerZoneCommand
)

// TriggerStep identifies a phase or step boundary used by a trigger pattern.
type TriggerStep uint8

// Trigger steps.
const (
	TriggerStepNone TriggerStep = iota
	TriggerStepUpkeep
	TriggerStepDraw
	TriggerStepBeginningOfCombat
	TriggerStepEndOfCombat
	TriggerStepEnd
	TriggerStepPrecombatMain
	TriggerStepPostcombatMain
)

// TriggerCombatQualifier identifies a combat-specific event restriction.
type TriggerCombatQualifier uint8

// Trigger combat qualifiers.
const (
	TriggerCombatAny TriggerCombatQualifier = iota
	TriggerCombatDamage
	TriggerNonCombatDamage
)

// TriggerAttackRecipient identifies what an attacker was declared against.
type TriggerAttackRecipient uint8

// Trigger attack recipient values are flags so exact recipient unions remain
// representable.
const (
	TriggerAttackRecipientAny    TriggerAttackRecipient = 0
	TriggerAttackRecipientPlayer TriggerAttackRecipient = 1 << (iota - 1)
	TriggerAttackRecipientPlaneswalker
	TriggerAttackRecipientBattle
)

// TriggerDamageRecipient identifies what received damage. Values are flags so a
// pattern can match either kind.
type TriggerDamageRecipient uint8

// Trigger damage recipient kinds.
const (
	TriggerDamageRecipientAny TriggerDamageRecipient = iota
	TriggerDamageRecipientPlayer
	TriggerDamageRecipientPermanent
)

// TriggerStackObject identifies a stack object involved in an event.
type TriggerStackObject uint8

// Trigger stack object kinds.
const (
	TriggerStackObjectAny TriggerStackObject = iota
	TriggerStackObjectSpell
)

// TriggerCounter identifies a counter kind used by a trigger pattern.
type TriggerCounter uint8

// Trigger counter kinds.
const (
	TriggerCounterAny TriggerCounter = iota
	TriggerCounterPlusOnePlusOne
	TriggerCounterMinusOneMinusOne
)

// TriggerCardType identifies a card type used by a semantic trigger Selection.
type TriggerCardType uint8

// Trigger card types.
const (
	TriggerCardTypeUnknown TriggerCardType = iota
	TriggerCardTypeArtifact
	TriggerCardTypeBattle
	TriggerCardTypeCreature
	TriggerCardTypeEnchantment
	TriggerCardTypeInstant
	TriggerCardTypeLand
	TriggerCardTypePlaneswalker
	TriggerCardTypeSorcery
)

// TriggerColor identifies a color used by a semantic trigger Selection.
type TriggerColor uint8

// Trigger colors.
const (
	TriggerColorUnknown TriggerColor = iota
	TriggerColorWhite
	TriggerColorBlue
	TriggerColorBlack
	TriggerColorRed
	TriggerColorGreen
)

// TriggerSubtype identifies a typed subtype used by a semantic trigger Selection.
type TriggerSubtype = types.Sub

// Trigger subtypes.
const (
	TriggerSubtypeUnknown TriggerSubtype = ""
	TriggerSubtypeSpirit  TriggerSubtype = types.Spirit
	TriggerSubtypeArcane  TriggerSubtype = types.Arcane
)

// TriggerSupertype identifies a supertype used by a semantic trigger Selection.
type TriggerSupertype uint8

// Trigger supertypes.
const (
	TriggerSupertypeUnknown TriggerSupertype = iota
	TriggerSupertypeLegendary
	TriggerSupertypeSnow
)

// TriggerKeyword identifies a keyword used by a semantic trigger Selection.
type TriggerKeyword uint8

// Trigger keywords.
const (
	TriggerKeywordUnknown TriggerKeyword = iota
	TriggerKeywordDefender
	TriggerKeywordFlash
	TriggerKeywordFlying
	TriggerKeywordHaste
	TriggerKeywordShadow
)

// TriggerTriState is a closed semantic true/false filter.
type TriggerTriState uint8

// Trigger tri-state values.
const (
	TriggerTriAny TriggerTriState = iota
	TriggerTriTrue
	TriggerTriFalse
)

// TriggerCombatState identifies a permanent's combat involvement.
type TriggerCombatState uint8

// Trigger combat-state values.
const (
	TriggerCombatStateAny TriggerCombatState = iota
	TriggerCombatStateAttacking
	TriggerCombatStateBlocking
)

// TriggerComparison identifies an integer-comparison relation.
type TriggerComparison uint8

// Trigger comparison relations.
const (
	TriggerComparisonUnknown TriggerComparison = iota
	TriggerComparisonEqual
	TriggerComparisonAtMost
	TriggerComparisonAtLeast
)

// TriggerNumberFilter is a closed semantic integer predicate.
type TriggerNumberFilter struct {
	Comparison TriggerComparison
	Value      int
}

// TriggerSelection is the closed semantic Selection vocabulary currently used
// by representable event subjects and cast spells. Its zero value is a
// wildcard.
type TriggerSelection struct {
	RequiredTypes    []TriggerCardType
	RequiredTypesAny []TriggerCardType
	ExcludedTypes    []TriggerCardType
	Supertypes       []TriggerSupertype
	SubtypesAny      []TriggerSubtype
	ColorsAny        []TriggerColor
	ExcludedColors   []TriggerColor
	Colorless        bool
	Multicolored     bool
	Tapped           TriggerTriState
	CombatState      TriggerCombatState
	Keyword          TriggerKeyword
	ExcludedKeyword  TriggerKeyword
	NonToken         bool
	TokenOnly        bool
	ManaValueAtLeast int
	MatchManaValue   bool
	ManaValue        TriggerNumberFilter
	Power            TriggerNumberFilter
	Toughness        TriggerNumberFilter
	Controller       ControllerKind
	// SubtypeFromEntryChoice requires the matched object to share the creature
	// subtype the predicate's source permanent chose as it entered ("of the
	// chosen type"). It lowers to Selection.SubtypeFromSourceEntryChoice.
	SubtypeFromEntryChoice bool
}

// TriggerPattern is a source-spanned semantic description of a representable
// event trigger. Raw trigger event text is deliberately not part of this
// lowering interface.
type TriggerPattern struct {
	Span shared.Span
	Kind TriggerKind

	Event      TriggerEvent
	Source     TriggerSourceRelation
	Subject    TriggerSubject
	Controller ControllerKind
	// UnionEvent names a second event family joined to Event under the pattern's
	// shared subject and player filters, expressing "create or sacrifice a
	// token". It is TriggerEventUnknown for single-event patterns.
	UnionEvent TriggerEvent
	// CauseController identifies the controller of the spell or ability that
	// caused an event, independently from the event subject's controller.
	CauseController ControllerKind
	Player          TriggerPlayerRelation

	SubjectSelection         TriggerSelection
	RelatedSubjectSelection  TriggerSelection
	CardSelection            TriggerSelection
	DamageRecipientSelection TriggerSelection
	DamageSourceSelection    TriggerSelection
	AttackRecipientSelection TriggerSelection

	// SubjectSelectionOrSelf widens a SubjectSelection-filtered event subject to
	// also match the ability's own source, expressing "this permanent or another
	// <Selection> you control" zone-change triggers.
	SubjectSelectionOrSelf bool

	MatchFromZone bool
	FromZone      TriggerZone
	MatchToZone   bool
	ToZone        TriggerZone
	ExcludeToZone bool

	MatchFaceDown bool
	FaceDown      bool

	Step                              TriggerStep
	StepPlayerSourceAttachedSelection TriggerSelection
	CombatQualifier                   TriggerCombatQualifier
	DamageRecipient                   TriggerDamageRecipient
	DamageRecipientIsSource           bool
	DamageSourceIsStackObject         bool
	AttackRecipient                   TriggerAttackRecipient
	StackObject                       TriggerStackObject
	Counter                           TriggerCounter

	// AttackAlone restricts an attacker-declared pattern to a creature that
	// attacks alone (the only attacking creature this combat).
	AttackAlone bool
	// AttackerCountAtLeast restricts a controller-scoped attacker-declared
	// pattern to combats with at least this many attacking creatures. Zero
	// imposes no minimum.
	AttackerCountAtLeast int

	ExcludeSelf                bool
	OneOrMore                  bool
	OneOrMorePerAttackTarget   bool
	RequireKickerPaid          bool
	RequireHistoric            bool
	ExcludeManaAbility         bool
	PlayerEventOrdinalThisTurn int
	// MatchSpellCopy widens a spell-cast pattern to also match spell-copy
	// events ("Whenever you cast or copy ...", magecraft).
	MatchSpellCopy bool

	// TappedForMana restricts a permanent-tapped pattern to taps that paid a
	// mana ability's cost ("is tapped for mana").
	TappedForMana bool

	// NextOccurrence marks a one-shot "next" phase/step relation ("your next
	// upkeep") rather than a recurring trigger. Such a pattern is representable
	// only as a delayed triggered ability created when a spell resolves
	// (CR 603.7), so direct trigger lowering rejects it.
	NextOccurrence bool

	InterveningCondition *CompiledCondition
}
