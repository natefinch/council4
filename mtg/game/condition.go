package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
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

	// Aggregates compares player- or board-derived quantities (see
	// AggregateKind) against thresholds using typed comparators. The entries are
	// ANDed; an empty slice disables the predicate. It unifies the controller
	// life-total, hand-size, library-size, graveyard-count, basic-land-type,
	// creature-power-diversity, opponent-count, attacker-count, gained-life, and
	// resolving-spell {X} comparisons that were previously modeled as separate
	// AtLeast/AtMost/Exactly fields.
	Aggregates []AggregateComparison

	// AnyPlayerLifeAtMost checks every non-eliminated player. Zero disables it.
	AnyPlayerLifeAtMost int

	// AnyOpponentPoisonAtLeast requires at least one non-eliminated opponent to
	// have at least this many poison counters. Zero disables the predicate.
	AnyOpponentPoisonAtLeast int

	// ControllerHandEmpty models the live hand-empty game-state predicate used by
	// the hellbent ability word. The controller-relative count quantities for
	// ability words such as threshold, delirium, domain, and coven are modeled by
	// Aggregates.
	ControllerHandEmpty bool

	// AllPlayersHandEmpty is satisfied when every non-eliminated player has no
	// cards in hand ("if each player has no cards in hand", Howltooth Hollow). It
	// is the all-players counterpart of ControllerHandEmpty.
	AllPlayersHandEmpty bool

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
	// SourceCountersAtLeast gates on a named counter kind on the source.
	SourceCounterKind      counter.Kind
	SourceCounterKindKnown bool
	SourceCountersAtLeast  int
	// SourceAttachedCombatCounterpartSubtypes requires the source's attached
	// permanent to be blocking or blocked by a creature with either subtype.
	SourceAttachedCombatCounterpartSubtypes [2]types.Sub
	SourceNotMonstrous                      bool
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

	// GiftPromised is satisfied when the resolving spell's Gift keyword action
	// promised a gift to an opponent as it was cast ("if the gift was promised,
	// ..."; CR 702.171). It is evaluated against the resolving stack object's
	// captured gift-promised state. Unlike the kicker gate, a copy of a promised
	// spell is itself promised to the same opponent (CR 707.10), so this holds
	// for copies too. When Condition.Negate is set it matches the "if the gift
	// wasn't promised" penalty clause.
	GiftPromised bool

	// EventPermanentWasKicked is satisfied when the permanent named by the
	// triggering or entering event was kicked ("If this creature was kicked, it
	// enters with N +1/+1 counters on it." — the kicker enters-with-counters
	// cycle). It is evaluated against the event's captured kicker-paid state,
	// which the entering-permanent event preserves from the spell that became the
	// permanent, and is false when no such event is in context.
	EventPermanentWasKicked bool

	// EventPermanentWasCastFromControllerHand is satisfied when the entering
	// permanent was cast by the condition controller from that player's hand
	// ("enters with a divinity counter on it if you cast it from your hand" —
	// the original Myojin cycle). It is evaluated against the entering event's
	// captured cast controller and source zone.
	EventPermanentWasCastFromControllerHand bool

	// ControllerGraveyardCardOfTypeCountAtLeast requires the context controller's
	// graveyard to hold at least this many cards of ControllerGraveyardCountCardType
	// ("if twenty or more creature cards are in your graveyard", Mortal Combat).
	// Zero disables the predicate.
	ControllerGraveyardCardOfTypeCountAtLeast int
	// ControllerGraveyardCountCardType is the card type counted by
	// ControllerGraveyardCardOfTypeCountAtLeast.
	ControllerGraveyardCountCardType types.Card
	// ControllerGraveyardInstantOrSorceryCountAtLeast requires the context
	// controller's graveyard to hold at least this many cards that are instants
	// and/or sorceries ("Spell mastery — If there are two or more instant and/or
	// sorcery cards in your graveyard, ...", Fiery Impulse). Zero disables the
	// predicate.
	ControllerGraveyardInstantOrSorceryCountAtLeast int

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

	// ControllerControlsGreatestToughnessCreature is satisfied when the context
	// controller controls a creature whose toughness is greater than or equal to
	// every creature's toughness on the battlefield ("if you control the creature
	// with the greatest toughness or tied for the greatest toughness"; Abzan
	// Beastmaster). It holds when the controller has the sole highest-toughness
	// creature or is tied for highest, and is false when no creatures exist.
	ControllerControlsGreatestToughnessCreature bool

	// EventPermanentPowerGreaterThanEachOtherCreature is satisfied when the
	// permanent named by the triggering zone-change event has power strictly
	// greater than every other creature's power on the battlefield ("if its power
	// is greater than each other creature's power"; Selvala, Heart of the Wilds).
	// It reads the entering creature's power and compares it against every other
	// creature; a tie, the absence of the event permanent, or an event permanent
	// that is not a creature fails closed.
	EventPermanentPowerGreaterThanEachOtherCreature bool

	// ControllerIsMonarch is satisfied when the context controller is the
	// monarch (CR 720), as in "At the beginning of your end step, if you're the
	// monarch, ...". It reads the controller's live IsMonarch designation flag.
	ControllerIsMonarch bool

	// ControllerWasMonarchAtTurnStart is satisfied when the context controller was
	// the monarch (CR 720) as the current turn began, as in "if you were the
	// monarch as the turn began" (Knights of the Black Rose). It reads the monarch
	// snapshot taken when the turn advanced (Turn.MonarchAtTurnStart), not the live
	// designation.
	ControllerWasMonarchAtTurnStart bool

	// AnOpponentIsMonarch is satisfied when any of the context controller's
	// opponents is the monarch (CR 720), as in "At the beginning of your upkeep,
	// if an opponent is the monarch, ..." (Queen Marchesa). It reads the live
	// IsMonarch designation flag of each opponent.
	AnOpponentIsMonarch bool

	// NoMonarch is satisfied when no player currently holds the monarch
	// designation ("if there is no monarch, you become the monarch." — Crown of
	// Gondor, Archivist of Gondor). It reads the live IsMonarch designation flag
	// of every player.
	NoMonarch bool

	// EventDefendingPlayerIsMonarch is satisfied when the defending player of the
	// triggering attack event ("that player") currently holds the monarch (CR
	// 720), as in "Whenever a creature attacks one of your opponents, if that
	// player is the monarch, ..." (M'Baku, Jabari Chieftain). It reads the
	// triggering event's defending player (Event.Player) and the live monarch
	// designation, so it is meaningful only in a trigger context whose bound
	// event is an attacker-declared event. It is false when no event is bound or
	// the defending player is not the living monarch.
	EventDefendingPlayerIsMonarch bool

	// ControllerHasInitiative is satisfied when the context controller has the
	// initiative (CR 720), as in "At the beginning of your end step, if you have
	// the initiative, ...". It reads the controller's live HasInitiative flag.
	ControllerHasInitiative bool

	// ControllerHasCityBlessing is satisfied when the context controller has the
	// city's blessing (CR 702.131 ascend), as in "if you have the city's
	// blessing, ...". It reads the controller's live HasCityBlessing flag.
	ControllerHasCityBlessing bool
	// SourceControllerTurn is satisfied while it is the context controller's turn,
	// i.e. the controller is the active player ("During your turn, this creature
	// has first strike"; Fresh-Faced Recruit, Embereth Skyblazer). It gates a
	// conditional self-static so the granted keyword or power/toughness bonus
	// applies only on the controller's own turns.
	SourceControllerTurn bool

	// SpellColorManaSpent gates the Adamant ability word "If at least three
	// <color> mana was spent to cast this spell, ..." (CR 702.132). It is
	// satisfied when at least SpellColorManaSpent.Count mana of
	// SpellColorManaSpent.Color was spent to cast the resolving spell. It reads
	// the resolving stack object's per-color mana-spend record and is false for
	// copies and for permanents that did not enter from a cast spell. Its zero
	// value (Count == 0) disables the predicate.
	SpellColorManaSpent ColorManaSpendThreshold

	// SpellSameColorManaSpentAtLeast gates the Adamant ability word "If at least
	// three mana of the same color was spent to cast this spell, ..." (Henge
	// Walker). It is satisfied when some single color contributed at least this
	// many mana to the resolving spell's cost. Zero disables the predicate.
	SpellSameColorManaSpentAtLeast int

	// LandEnteredThisTurnOrControlsBasicLand is satisfied when the condition
	// source land entered the battlefield this turn or its controller controls a
	// basic land ("Activate only if this land entered this turn or if you control
	// a basic land."; the Mercadian Masques tap-for-two-colors land cycle). It is
	// the disjunctive activation gate those lands print to bar second-turn
	// fixing, holding on either half. It reads the source's enter history and the
	// controller's basic-land board state.
	LandEnteredThisTurnOrControlsBasicLand bool

	// SourceAbilityResolutionOrdinalThisTurn is satisfied when the resolving
	// triggered ability has resolved exactly this many times during the current
	// turn, counting the current resolution ("if this is the second time this
	// ability has resolved this turn"; Prowl, Pursuit Vehicle). It reads the
	// resolving stack object's (source, ability) resolution tally from
	// Game.ResolvedTriggeredAbilitiesThisTurn, which the ability increments as it
	// begins resolving, and is meaningful only while a triggered ability is
	// resolving. Zero disables the predicate.
	SourceAbilityResolutionOrdinalThisTurn int
}

// ColorManaSpendThreshold names a single color and the minimum number of mana of
// that color that must have been spent to cast the resolving spell for the
// Adamant predicate to hold. A zero Count disables the predicate.
type ColorManaSpendThreshold struct {
	Color color.Color
	Count int
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
		len(c.Aggregates) == 0 &&
		c.AnyOpponentPoisonAtLeast == 0 &&
		c.AnyPlayerLifeAtMost == 0 &&
		!c.ControllerHandEmpty &&
		!c.AllPlayersHandEmpty &&
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
		c.SourceCountersAtLeast == 0 &&
		c.SourceAttachedCombatCounterpartSubtypes == [2]types.Sub{} &&
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
		!c.GiftPromised &&
		!c.EventPermanentWasKicked &&
		!c.EventPermanentWasCastFromControllerHand &&
		c.ControllerGraveyardCardOfTypeCountAtLeast == 0 &&
		c.ControllerGraveyardInstantOrSorceryCountAtLeast == 0 &&
		len(c.ControllerControlsNamed) == 0 &&
		!c.FirstCombatPhaseOfTurn &&
		!c.ControllerControlsGreatestPowerCreature &&
		!c.ControllerControlsGreatestToughnessCreature &&
		!c.EventPermanentPowerGreaterThanEachOtherCreature &&
		!c.ControllerIsMonarch &&
		!c.ControllerWasMonarchAtTurnStart &&
		!c.AnOpponentIsMonarch &&
		!c.NoMonarch &&
		!c.EventDefendingPlayerIsMonarch &&
		!c.ControllerHasInitiative &&
		!c.ControllerHasCityBlessing &&
		!c.SourceControllerTurn &&
		c.SpellColorManaSpent.Count == 0 &&
		c.SpellSameColorManaSpentAtLeast == 0 &&
		!c.LandEnteredThisTurnOrControlsBasicLand &&
		c.SourceAbilityResolutionOrdinalThisTurn == 0
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
