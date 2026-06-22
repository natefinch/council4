package game

// Phase represents one of the five phases of a Magic turn (CR 500).
type Phase int

const (
	// PhaseBeginning contains the untap, upkeep, and draw steps.
	PhaseBeginning Phase = iota

	// PhasePrecombatMain is the first main phase where sorcery-speed
	// actions can be taken.
	PhasePrecombatMain

	// PhaseCombat contains the combat steps (declare attackers, blockers, etc.).
	PhaseCombat

	// PhasePostcombatMain is the second main phase.
	PhasePostcombatMain

	// PhaseEnding contains the end step and cleanup step.
	PhaseEnding
)

// String returns the phase name.
func (p Phase) String() string {
	switch p {
	case PhaseBeginning:
		return "Beginning"
	case PhasePrecombatMain:
		return "Precombat Main"
	case PhaseCombat:
		return "Combat"
	case PhasePostcombatMain:
		return "Postcombat Main"
	case PhaseEnding:
		return "Ending"
	default:
		return "Unknown"
	}
}

// Step identifies a turn step or a synthetic main-phase trigger boundary
// (CR 500–514).
type Step int

// Step values identify turn steps. Main-phase trigger markers are appended to
// preserve the numeric values of the existing steps.
const (
	// StepNone indicates no specific step (used during main phases which
	// have no sub-steps).
	StepNone Step = iota

	StepUntap  // No player gets priority (CR 502)
	StepUpkeep // "At the beginning of upkeep" triggers
	StepDraw   // Active player draws one card

	StepBeginningOfCombat // "Beginning of combat" triggers
	StepDeclareAttackers  // Active player declares attackers
	StepDeclareBlockers   // Defending players declare blockers
	StepFirstStrikeDamage // Only exists if a creature has first/double strike
	StepCombatDamage      // All combat damage is dealt
	StepEndOfCombat       // "End of combat" triggers

	StepEnd     // "At the beginning of the end step" triggers
	StepCleanup // Discard to hand size, remove damage, "until end of turn" expires

	// StepPrecombatMain marks the precombat-main boundary for trigger events.
	StepPrecombatMain
	// StepPostcombatMain marks the postcombat-main boundary for trigger events.
	StepPostcombatMain
)

// String returns the step name.
func (s Step) String() string {
	switch s {
	case StepNone:
		return "Main"
	case StepUntap:
		return "Untap"
	case StepUpkeep:
		return "Upkeep"
	case StepDraw:
		return "Draw"
	case StepBeginningOfCombat:
		return "Beginning of Combat"
	case StepDeclareAttackers:
		return "Declare Attackers"
	case StepDeclareBlockers:
		return "Declare Blockers"
	case StepFirstStrikeDamage:
		return "First Strike Damage"
	case StepCombatDamage:
		return "Combat Damage"
	case StepEndOfCombat:
		return "End of Combat"
	case StepEnd:
		return "End Step"
	case StepCleanup:
		return "Cleanup"
	case StepPrecombatMain:
		return "Precombat Main"
	case StepPostcombatMain:
		return "Postcombat Main"
	default:
		return "Unknown"
	}
}

// TurnState tracks the current position within the turn structure
// and turn-level game state.
type TurnState struct {
	// TurnNumber is the current turn number (1-indexed).
	TurnNumber int

	// ActivePlayer is the player whose turn it is.
	ActivePlayer PlayerID

	// Phase is the current phase of the turn.
	Phase Phase

	// Step is the current step within the phase. Main-phase boundary markers are
	// transient; StepNone is used while players have priority in main phases.
	Step Step

	// PriorityPlayer is the player who currently has priority (the right
	// to take an action). Priority passes clockwise after each action.
	PriorityPlayer PlayerID

	// LandsPlayedThisTurn counts how many lands the active player has
	// played this turn. Normally limited to 1 (CR 305.2).
	LandsPlayedThisTurn int

	// LandsAllowedThisTurn is the maximum number of lands the active player
	// may play this turn. Defaults to 1 but can be modified by effects
	// like Exploration or Oracle of Mul Daya.
	LandsAllowedThisTurn int

	// ExtraTurns is a queue of players who will take extra turns after
	// the current turn ends. Processed LIFO (most recently added first).
	ExtraTurns []PlayerID

	// ExtraPhases is a FIFO queue of additional phases inserted into the
	// current turn by extra-phase effects ("After this main phase, there is
	// an additional combat phase followed by an additional main phase." —
	// Aggravated Assault, Aurelia, World at War). The turn loop drains the
	// queue after the postcombat main phase, running each queued phase in
	// order; a queued phase that re-activates the source re-queues more
	// phases, so the combo loop continues until the queue empties.
	ExtraPhases []Phase
}

// CanPlayLand reports whether the active player can still play a land
// this turn (has not exceeded their land-per-turn limit).
func (ts *TurnState) CanPlayLand() bool {
	return ts.LandsPlayedThisTurn < ts.LandsAllowedThisTurn
}

// IsMainPhase reports whether the current phase is a main phase
// (precombat or postcombat).
func (ts *TurnState) IsMainPhase() bool {
	return ts.Phase == PhasePrecombatMain || ts.Phase == PhasePostcombatMain
}

// TurnOrder manages the player turn rotation in a 4-player game,
// handling eliminated players.
type TurnOrder struct {
	// Order is the seating order of players (clockwise).
	Order [NumPlayers]PlayerID

	// Eliminated tracks which players have been eliminated from the game.
	Eliminated map[PlayerID]bool
}

// NewTurnOrder creates a TurnOrder with the default seating arrangement
// (Player1 through Player4 in order).
func NewTurnOrder() TurnOrder {
	return TurnOrder{
		Order:      [NumPlayers]PlayerID{Player1, Player2, Player3, Player4},
		Eliminated: make(map[PlayerID]bool),
	}
}

// NextActivePlayer returns the next non-eliminated player after the
// given player in turn order.
func (to *TurnOrder) NextActivePlayer(current PlayerID) PlayerID {
	idx := -1
	for i, p := range to.Order {
		if p == current {
			idx = i
			break
		}
	}
	for i := 1; i <= NumPlayers; i++ {
		next := to.Order[(idx+i)%NumPlayers]
		if !to.Eliminated[next] {
			return next
		}
	}
	return current // all eliminated except current (shouldn't happen)
}

// NextPriority returns the next non-eliminated player who should
// receive priority after the given player.
func (to *TurnOrder) NextPriority(current PlayerID) PlayerID {
	return to.NextActivePlayer(current)
}

// IsEliminated reports whether the given player has been eliminated.
func (to *TurnOrder) IsEliminated(p PlayerID) bool {
	return to.Eliminated[p]
}

// Eliminate marks a player as eliminated from the game.
func (to *TurnOrder) Eliminate(p PlayerID) {
	if to.Eliminated == nil {
		to.Eliminated = make(map[PlayerID]bool)
	}
	to.Eliminated[p] = true
}

// ActivePlayerCount returns the number of non-eliminated players.
func (to *TurnOrder) ActivePlayerCount() int {
	count := NumPlayers
	for _, e := range to.Eliminated {
		if e {
			count--
		}
	}
	return count
}
