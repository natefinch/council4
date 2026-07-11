package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
	TriggerEventAttackerBecameUnblocked
	TriggerEventClassBecameLevel
	// TriggerEventDoorUnlocked is the self-source door-unlock trigger of a Room
	// enchantment half ("When you unlock this door", CR 715). The runtime models
	// the cast-door unlock that happens as the Room enters; lowering maps it onto
	// the permanent-entered-battlefield event for that half.
	TriggerEventDoorUnlocked
	// TriggerEventCrimeCommitted is the acting-player "commit a crime" event
	// (CR 700.15): a player puts a spell or ability on the stack that targets an
	// opponent, an object an opponent controls, or a card in an opponent's
	// graveyard.
	TriggerEventCrimeCommitted
	// TriggerEventBecameMonarch is the "become the monarch" event (CR 720.2): a
	// player who was not already the monarch takes the crown, whether by a
	// designation effect or by dealing combat damage to the monarch.
	TriggerEventBecameMonarch
	// TriggerEventCardPlayedFromExile is the "plays a card exiled with <this
	// permanent>" event (Prowl, Stoic Strategist): a player casts or plays as a
	// land a card that a linked-exile ability of the source placed into exile.
	TriggerEventCardPlayedFromExile
	// TriggerEventLandPlayed is the "plays a land" event (Burgeoning, Dirtcowl
	// Wurm, Horn of Greed): a player plays a land as the land-play special
	// action (CR 305). It is distinct from TriggerEventCardPlayedFromExile, which
	// fires only for cards a linked-exile source put into exile.
	TriggerEventLandPlayed
)

// TriggerCastTurn restricts a spell-cast pattern by whose turn the spell was
// cast on, relative to the ability's controller.
type TriggerCastTurn uint8

// Spell-cast turn relations.
const (
	TriggerCastTurnAny TriggerCastTurn = iota
	TriggerCastTurnYours
	TriggerCastTurnNotYours
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
	// TriggerPlayerMonarch matches an event whose player is the current monarch
	// ("At the beginning of the monarch's end step, ...", Archivist of Gondor).
	TriggerPlayerMonarch
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
	TriggerCounterLore
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

// TriggerSelection is the closed semantic Selection vocabulary currently used
// by representable event subjects and cast spells. Its zero value is a
// wildcard.
type TriggerSelection struct {
	RequiredTypes    []types.Card
	RequiredTypesAny []types.Card
	ExcludedTypes    []types.Card
	Supertypes       []types.Super
	SubtypesAny      []types.Sub
	ColorsAny        []color.Color
	ExcludedColors   []color.Color
	Colorless        bool
	Multicolored     bool
	Tapped           TriggerTriState
	CombatState      TriggerCombatState
	Keyword          parser.KeywordKind
	ExcludedKeyword  parser.KeywordKind
	NonToken         bool
	TokenOnly        bool
	ManaValueAtLeast int
	ManaValueAtMost  int
	MatchManaValue   bool
	ManaValue        compare.Int
	Power            compare.Int
	Toughness        compare.Int
	Controller       ControllerKind
	// MatchAnyCounter records a kind-agnostic "with a counter on it" subject
	// qualifier. It lowers to the matching CompiledSelector counter dimension.
	MatchAnyCounter bool
	// MatchCounter and RequiredCounter record a kind-specific "with a <kind>
	// counter on it" subject qualifier. The matched permanent must carry a
	// counter of RequiredCounter's kind; a dying subject is matched against its
	// last-known counters. They lower to Selection.MatchCounter and
	// Selection.RequiredCounter.
	MatchCounter    bool
	RequiredCounter counter.Kind
	// SubtypeFromEntryChoice requires the matched object to share the creature
	// subtype the predicate's source permanent chose as it entered ("of the
	// chosen type"). It lowers to Selection.SubtypeFromSourceEntryChoice.
	SubtypeFromEntryChoice bool
	// ColorFromEntryChoice requires the matched object to share the color the
	// predicate's source permanent chose as it entered ("of the chosen color",
	// Prism Ring). It lowers to Selection.ColorChoice = ColorChoiceSourceEntry.
	ColorFromEntryChoice bool
	// Modified requires the matched permanent to be modified (a counter, Aura, or
	// Equipment, CR 701.50). It lowers to Selection.MatchModified.
	Modified bool
	// Commander requires the matched permanent to be a commander ("your commander
	// deals combat damage to a player", Archivist of Gondor). It lowers to
	// Selection.MatchCommander.
	Commander bool
	// Goaded requires the matched permanent to be goaded right now (CR 701.38,
	// "Whenever a goaded creature attacks", Vengeful Ancestor). It lowers to
	// Selection.MatchGoaded.
	Goaded bool
	// AnyOf is a disjunction of alternative selections; the subject matches when
	// it satisfies at least one alternative ("creature or Vehicle"). The other
	// fields remain conjunctive requirements shared by every alternative. It
	// lowers to Selection.AnyOf.
	AnyOf []TriggerSelection
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

	// DamageSourceSelectionOrSelf widens a combat-damage source filter to also
	// match the ability's own source, expressing "this creature or another
	// <Selection> you control deals combat damage" and "this creature or
	// equipped creature deals combat damage".
	DamageSourceSelectionOrSelf bool

	// DyingDamagedBySource restricts a permanent-died pattern to a permanent that
	// was dealt damage by the ability's own source earlier this turn ("Whenever a
	// creature dealt damage by this creature this turn dies", CR 603.2). It is
	// only valid with Event == TriggerEventPermanentDied.
	DyingDamagedBySource bool

	MatchFromZone bool
	FromZone      TriggerZone
	MatchToZone   bool
	ToZone        TriggerZone
	ExcludeToZone bool

	ExcludeFromZone bool

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
	// AttackWhileSaddled restricts an attacker-declared pattern to combats where
	// the attacking source is saddled ("attacks while saddled", CR 702.166).
	AttackWhileSaddled bool
	// AttacksDifferentPlayerThanAnother restricts an attacker-declared pattern to
	// combats where the source and at least one other attacking creature attack
	// different players ("this creature and another creature attack different
	// players", Canal Courier).
	AttacksDifferentPlayerThanAnother bool
	// AttackerCountAtLeast restricts a controller-scoped attacker-declared
	// pattern to combats with at least this many attacking creatures. Zero
	// imposes no minimum.
	AttackerCountAtLeast int
	// AttacksAlongsideSelection restricts a self-source attacker-declared pattern
	// to combats where at least AttacksAlongsideCount other attacking creatures
	// match this selection ("Whenever this creature and at least one Human
	// attack", Goldbug). It is only set with a positive AttacksAlongsideCount.
	AttacksAlongsideSelection TriggerSelection
	// AttacksAlongsideCount is the minimum number of other attacking creatures
	// matching AttacksAlongsideSelection. Zero imposes no such restriction.
	AttacksAlongsideCount int

	ExcludeSelf                bool
	OneOrMore                  bool
	OneOrMorePerAttackTarget   bool
	RequireKickerPaid          bool
	RequireHistoric            bool
	ExcludeManaAbility         bool
	PlayerEventOrdinalThisTurn int
	// PlaysExiledWithSource marks a player-event pattern whose card object is "a
	// card exiled with <this permanent>" (Prowl, Stoic Strategist): the event
	// fires when any player plays a card the source's linked-exile ability put
	// into exile. It is only valid with Event == TriggerEventCardPlayedFromExile.
	PlaysExiledWithSource bool
	// MatchSpellCopy widens a spell-cast pattern to also match spell-copy
	// events ("Whenever you cast or copy ...", magecraft).
	MatchSpellCopy bool

	// SelfWasCast restricts a spell-cast pattern to the casting of the ability's
	// own source spell ("When you cast this spell", CR 601.3i). It is only valid
	// with Event == TriggerEventSpellCast and fires once as the source is put on
	// the stack.
	SelfWasCast bool

	// SpellTargetsSource restricts a spell-cast pattern to spells that target
	// the source permanent ("Whenever you cast a spell that targets this
	// creature", the Heroic ability word).
	SpellTargetsSource bool

	// SpellTargetSelection restricts a spell-cast pattern to spells that target a
	// permanent matching this selection ("Whenever you cast a spell that targets
	// a creature you control"). It is nil when no such relation applies.
	SpellTargetSelection *TriggerSelection

	// CastDuringTurn restricts a spell-cast pattern by whose turn the spell was
	// cast on, relative to the ability's controller ("Whenever you cast a spell
	// during your turn" / "during an opponent's turn").
	CastDuringTurn TriggerCastTurn
	// TappedForMana restricts a permanent-tapped pattern to taps that paid a
	// mana ability's cost ("is tapped for mana").
	TappedForMana bool

	// TappedForManaColor narrows a TappedForMana pattern to taps that produced a
	// specific type of mana ("tap a permanent for {C}"). It is empty for the
	// unrestricted "for mana" wording.
	TappedForManaColor mana.Color

	// NextOccurrence marks a one-shot "next" phase/step relation ("your next
	// upkeep") rather than a recurring trigger. Such a pattern is representable
	// only as a delayed triggered ability created when a spell resolves
	// (CR 603.7), so direct trigger lowering rejects it.
	NextOccurrence bool

	// ExcludeFirstDrawInDrawStep narrows a card-draw pattern to draws other than
	// the first card a player draws during each of their draw steps ("except the
	// first one they draw in each of their draw steps", Orcish Bowmasters). It is
	// only meaningful for the card-draw event.
	ExcludeFirstDrawInDrawStep bool

	// ClassBecameLevel restricts a class-level-gained pattern to the level the
	// Class became ("When this Class becomes level N"). Zero imposes no
	// restriction.
	ClassBecameLevel int

	InterveningCondition *CompiledCondition

	// StateCondition holds the board-state predicate of a state trigger
	// (Kind == TriggerState). The runtime fires the trigger whenever this
	// condition holds while it is not already on the stack (CR 603.8). It is nil
	// for every event-based pattern.
	StateCondition *CompiledCondition
}
