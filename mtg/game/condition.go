package game

import (
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Condition is a reusable rules predicate evaluated by mtg/rules in an explicit
// context such as a static ability, activation restriction, trigger, effect, or
// replacement event.
type Condition struct {
	// Text preserves the printed condition for diagnostics and generated-card
	// review.
	Text string

	// Negate inverts the whole condition, e.g. "unless you control...".
	Negate bool

	// ControlsMatching requires the context controller to control matching
	// permanents. When present, the context controller must control at least
	// MinCount objects matching the Selection (MinCount defaults to 1),
	// optionally constrained by TotalPower. It is ignored when absent.
	ControlsMatching opt.V[SelectionCount]

	// ControllerLifeAtLeast requires the context controller's current life total
	// to meet the threshold. AnyPlayerLifeAtMost checks every non-eliminated
	// player. Zero values disable these predicates.
	ControllerLifeAtLeast     int
	ControllerHandSizeAtLeast int
	AnyPlayerLifeAtMost       int

	// ControllerLifeAtMost requires the context controller's current life total
	// to be at most the threshold ("you have N or less life"). It uses opt.V so
	// a zero threshold ("0 or less life") is distinguishable from absence.
	ControllerLifeAtMost opt.V[int]

	// ControllerLifeAtLeastAboveStarting requires the context controller's
	// current life total to be at least this many points above their starting
	// life total ("you have at least N life more than your starting life
	// total"). Zero disables the predicate.
	ControllerLifeAtLeastAboveStarting int

	// ControllerHandSizeExactly requires the context controller to hold exactly
	// this many cards in hand. Negative disables it; zero is expressed via
	// ControllerHandEmpty, so a present exact-zero predicate is not modeled here.
	ControllerHandSizeExactly opt.V[int]

	// AnyOpponentPoisonAtLeast requires at least one non-eliminated opponent to
	// have at least this many poison counters. Zero disables the predicate.
	AnyOpponentPoisonAtLeast int

	// OpponentCountAtLeast requires this many non-eliminated opponents.
	OpponentCountAtLeast int

	// ControllerHandEmpty and the controller-relative thresholds model
	// live game-state predicates used by ability words such as threshold,
	// delirium, domain, hellbent, and coven.
	ControllerHandEmpty                     bool
	ControllerGraveyardCardCountAtLeast     int
	ControllerGraveyardCardTypeCountAtLeast int
	ControllerBasicLandTypeCountAtLeast     int
	ControllerCreaturePowerDiversityAtLeast int

	// ControllerCreatedTokenThisTurn requires the context controller to have
	// created at least one token during the current turn ("Activate only if you
	// created a token this turn").
	ControllerCreatedTokenThisTurn bool

	// AnyOpponentControls checks each opponent independently. OpponentsControl
	// counts matching permanents controlled by all opponents collectively.
	AnyOpponentControls opt.V[SelectionCount]
	OpponentsControl    opt.V[SelectionCount]

	// ControlComparison compares the number of permanents matching a Selection
	// controlled by two player scopes ("an opponent controls more lands than
	// you"). It is ignored when not present.
	ControlComparison opt.V[ControlCountComparison]

	// Object tests a referenced object in the current condition context, such as
	// a triggering event permanent. It may use last-known information.
	// ObjectMatches, when present, applies the shared Selection semantics to that
	// object. An empty ObjectMatches Selection is a wildcard existence check.
	Object                                                       opt.V[ObjectReference]
	ObjectMatches                                                opt.V[Selection]
	Types                                                        []types.Card
	EventPermanentNameUniqueAmongControlledAndGraveyardCreatures bool
	SourceClassLevelAtLeast                                      int
	SourceClassLevelLessThan                                     int
	// SourceLevelCountersAtLeast and SourceLevelCountersLessThan gate an ability
	// by the number of level counters on the condition source (CR 711.2),
	// modeling a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band. AtLeast applies
	// the band's lower bound; LessThan applies a non-final band's exclusive upper
	// bound (hi+1) and is zero for the open-ended final band. Zero disables each
	// predicate.
	SourceLevelCountersAtLeast  int
	SourceLevelCountersLessThan int
	SourceNotMonstrous          bool
	// SourceSaddled requires the condition source Mount to be saddled
	// (CR 702.166), as in "if this creature is saddled". Negate models the
	// "isn't saddled" wording.
	SourceSaddled         bool
	SourceTributeNotPaid  bool
	ControllerHasMaxSpeed bool
	TargetEnteredThisTurn opt.V[int]
	CastFromZone          opt.V[zone.Type]

	// CastDuringControllerMainPhase is satisfied when the resolving spell was
	// cast during its controller's main phase ("Addendum — If you cast this
	// spell during your main phase, ..."). It is evaluated against the resolving
	// stack object's captured cast timing and is false for copies.
	CastDuringControllerMainPhase bool

	// EventHistory is satisfied when the selected turn's event history contains
	// at least one event matching the stored pattern. When Condition.Negate is
	// true the predicate is inverted (e.g. "if no spells were cast last turn").
	EventHistory opt.V[EventHistoryCondition]

	// ControllerControlsCommander requires the context controller to control
	// their commander on the battlefield ("if you control your commander" / "as
	// long as you control your commander"). It gates the Lieutenant ability word.
	ControllerControlsCommander bool

	// SpellWasKicked is satisfied when the resolving spell was kicked ("if this
	// spell was kicked, ... instead"). It is evaluated against the resolving
	// stack object's captured kicker-paid state and is false for copies.
	SpellWasKicked bool

	// AttackersAttackingControllerAtLeast requires at least this many of the
	// attackers declared this combat to be attacking the context controller
	// directly or one of the controller's planeswalkers ("if two or more of
	// those creatures are attacking you and/or planeswalkers you control";
	// Mangara, the Diplomat). It is evaluated against live combat state and is
	// zero (disabled) elsewhere.
	AttackersAttackingControllerAtLeast int

	// ControllerLibrarySizeAtLeast requires the context controller's library to
	// hold at least this many cards ("if you have N or more cards in your
	// library", Battle of Wits). Zero disables the predicate.
	ControllerLibrarySizeAtLeast int

	// ControllerLifeExactly requires the context controller's current life total
	// to equal this value ("if you have exactly N life", Near-Death Experience).
	// It uses opt.V so an exact-zero threshold is distinguishable from absence.
	ControllerLifeExactly opt.V[int]

	// ControllerGainedLifeThisTurnAtLeast requires the context controller to have
	// gained at least this much total life so far this turn ("if you gained 3 or
	// more life this turn"; Angelic Accord). It is zero (disabled) otherwise.
	ControllerGainedLifeThisTurnAtLeast int
	// SpellXAtLeast requires the resolving spell's chosen value of {X} to be at
	// least this value ("if X is N or more", the Finale cycle). It is evaluated
	// against the resolving stack object's captured X value. Zero disables the
	// predicate.
	SpellXAtLeast int
	// ControllerGraveyardCardOfTypeCountAtLeast requires the context controller's
	// graveyard to hold at least this many cards of ControllerGraveyardCountCardType
	// ("if twenty or more creature cards are in your graveyard", Mortal Combat).
	// Zero disables the predicate.
	ControllerGraveyardCardOfTypeCountAtLeast int
	// ControllerGraveyardCountCardType is the card type counted by
	// ControllerGraveyardCardOfTypeCountAtLeast.
	ControllerGraveyardCountCardType types.Card

	// ControllerControlsNamed requires the context controller to control at
	// least one permanent matching each listed card name ("If you control an
	// Urza's Mine and an Urza's Tower, ..."; the Urza tron lands). Names are
	// compared case-insensitively with hyphens and spaces treated alike, so the
	// printed Oracle spelling ("Urza's Power-Plant") matches the canonical card
	// name ("Urza's Power Plant"). An empty slice disables the predicate.
	ControllerControlsNamed []string

	// FirstCombatPhaseOfTurn is satisfied while the current turn is still in its
	// first combat phase ("if it's the first combat phase of the turn"; Raiyuu,
	// Storm's Edge, Karlach, Fury of Avernus). It is evaluated against
	// TurnState.CombatPhasesThisTurn, holding only while that count is 1, so the
	// extra-combat insertion it gates fires once per turn rather than looping.
	FirstCombatPhaseOfTurn bool

	// ControllerControlsGreatestPowerCreature is satisfied when the context
	// controller controls a creature whose power is greater than or equal to
	// every creature's power on the battlefield ("if you control the creature
	// with the greatest power or tied for the greatest power"; Summon: Fenrir
	// chapter III). It holds when the controller has the sole highest-power
	// creature or is tied for highest, and is false when no creatures exist.
	ControllerControlsGreatestPowerCreature bool
}

// ControlPlayerScope selects which players' battlefields a control-count
// comparison counts.
type ControlPlayerScope uint8

// Control player scope values.
const (
	// ControlPlayerController counts permanents controlled by the condition's
	// controller ("you").
	ControlPlayerController ControlPlayerScope = iota
	// ControlPlayerAnyOpponent quantifies existentially over opponents: the
	// comparison holds when at least one opponent satisfies it.
	ControlPlayerAnyOpponent
	// ControlPlayerEachOpponent quantifies universally over opponents: the
	// comparison holds when every opponent satisfies it.
	ControlPlayerEachOpponent
	// ControlPlayerTriggeringPlayer counts permanents controlled by the player
	// tied to the triggering event ("that player"), resolved from the event's
	// controller. It compares a single specific player rather than quantifying
	// over opponents.
	ControlPlayerTriggeringPlayer
)

// ControlCountComparison compares the number of permanents matching Selection
// controlled by two player scopes ("an opponent controls more lands than you").
// It is satisfied when Left's count compares to Right's count under Op,
// quantified by whichever side is an opponent scope (existential for
// AnyOpponent, universal for EachOpponent). Exactly one side is the controller.
type ControlCountComparison struct {
	Selection Selection
	Left      ControlPlayerScope
	Right     ControlPlayerScope
	Op        compare.Op
}

// Empty reports whether the condition contains no active predicate.
func (c *Condition) Empty() bool {
	return !c.ControlsMatching.Exists &&
		c.ControllerLifeAtLeast == 0 &&
		!c.ControllerLifeAtMost.Exists &&
		c.ControllerLifeAtLeastAboveStarting == 0 &&
		c.ControllerHandSizeAtLeast == 0 &&
		!c.ControllerHandSizeExactly.Exists &&
		c.AnyOpponentPoisonAtLeast == 0 &&
		c.AnyPlayerLifeAtMost == 0 &&
		c.OpponentCountAtLeast == 0 &&
		!c.ControllerHandEmpty &&
		c.ControllerGraveyardCardCountAtLeast == 0 &&
		c.ControllerGraveyardCardTypeCountAtLeast == 0 &&
		c.ControllerBasicLandTypeCountAtLeast == 0 &&
		c.ControllerCreaturePowerDiversityAtLeast == 0 &&
		!c.ControllerCreatedTokenThisTurn &&
		!c.AnyOpponentControls.Exists &&
		!c.OpponentsControl.Exists &&
		!c.ControlComparison.Exists &&
		!c.Object.Exists &&
		!c.ObjectMatches.Exists &&
		len(c.Types) == 0 &&
		!c.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures &&
		c.SourceClassLevelAtLeast == 0 &&
		c.SourceClassLevelLessThan == 0 &&
		c.SourceLevelCountersAtLeast == 0 &&
		c.SourceLevelCountersLessThan == 0 &&
		!c.SourceNotMonstrous &&
		!c.SourceSaddled &&
		!c.SourceTributeNotPaid &&
		!c.ControllerHasMaxSpeed &&
		!c.TargetEnteredThisTurn.Exists &&
		!c.CastFromZone.Exists &&
		!c.CastDuringControllerMainPhase &&
		!c.EventHistory.Exists &&
		!c.ControllerControlsCommander &&
		!c.SpellWasKicked &&
		c.AttackersAttackingControllerAtLeast == 0 &&
		c.ControllerLibrarySizeAtLeast == 0 &&
		!c.ControllerLifeExactly.Exists &&
		c.ControllerGainedLifeThisTurnAtLeast == 0 &&
		c.SpellXAtLeast == 0 &&
		c.ControllerGraveyardCardOfTypeCountAtLeast == 0 &&
		len(c.ControllerControlsNamed) == 0 &&
		!c.FirstCombatPhaseOfTurn &&
		!c.ControllerControlsGreatestPowerCreature
}

// EventHistoryWindow selects which turn's event log an EventHistoryCondition
// searches.
type EventHistoryWindow uint8

// Event history window values.
const (
	// EventHistoryCurrentTurn checks events emitted during the current turn.
	EventHistoryCurrentTurn EventHistoryWindow = iota
	// EventHistoryPreviousTurn checks events emitted during the immediately
	// preceding turn.
	EventHistoryPreviousTurn
)

// EventHistoryCondition checks that the chosen turn's event log contains at
// least one event matching Pattern. Negate on the enclosing Condition inverts
// the result (e.g. "if no spells were cast last turn").
type EventHistoryCondition struct {
	Pattern TriggerPattern
	Window  EventHistoryWindow
	// MinCount is the minimum number of events in Window that must match Pattern
	// for the condition to hold. A zero value requires a single matching event.
	MinCount int
}
