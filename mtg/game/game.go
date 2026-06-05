package game

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game/id"
)

// DayNightState represents the day/night cycle (CR 728).
type DayNightState int

// Day/night values identify the current daybound or nightbound state.
const (
	Day   DayNightState = iota // It is day.
	Night                      // It is night.
)

// Emblem represents an emblem in the command zone. Emblems are created
// by planeswalker abilities and exist for the rest of the game (CR 114).
type Emblem struct {
	// Owner is the player who owns this emblem.
	Owner PlayerID

	// Abilities lists the abilities this emblem provides.
	Abilities []AbilityDef
}

// SuspendedCard tracks a card exiled with suspend.
type SuspendedCard struct {
	Owner        PlayerID
	Controller   PlayerID
	TimeCounters int
}

// PlayerConfig holds the configuration for a single player when setting
// up a new game — their name, deck list, and commander.
type PlayerConfig struct {
	// Name is the player's display name.
	Name string

	// Commander is the card definition for the player's commander.
	Commander *CardDef

	// Deck is the list of card definitions for the player's 99-card deck
	// (not including the commander).
	Deck []*CardDef

	// PowerBracket is optional deck metadata for future simulations/reports.
	PowerBracket string

	// PowerLevel is optional numeric deck metadata for future simulations/reports.
	PowerLevel int
}

// Game is the top-level state of a 4-player Commander game. It ties
// together all players, the shared battlefield, the stack, turn state,
// and global game properties.
type Game struct {
	// Players are the four players in the game, indexed by PlayerID.
	Players [NumPlayers]*Player

	// Battlefield holds all permanents currently on the battlefield.
	// This is a shared zone — permanents track their own Controller/Owner.
	Battlefield []*Permanent

	// CardInstances maps CardInstance IDs to their CardInstance structs.
	// This is the registry of all card instances in the game (deck cards
	// and any cards created during play).
	CardInstances map[id.ID]*CardInstance

	// CommanderIDs records the original commander card instances for the game.
	CommanderIDs map[id.ID]bool

	// Stack is the game stack where spells and abilities wait to resolve.
	Stack Stack

	// ContinuousEffects are runtime continuous effects applied through the
	// layer system when rules need effective permanent values.
	ContinuousEffects []ContinuousEffect

	// DelayedTriggers are delayed triggered abilities waiting for a future
	// timing condition such as the next end step.
	DelayedTriggers []DelayedTrigger

	// PreventionShields are runtime damage-prevention replacement effects.
	PreventionShields []PreventionShield

	// ReplacementDecisions records deterministic fallback ordering decisions.
	ReplacementDecisions []ReplacementDecision

	// ReplacementEffects are runtime effects that modify future events before
	// they happen (CR 614). mtg/rules owns matching, ordering, and expiry.
	ReplacementEffects []ReplacementEffect

	// SkippedSteps records upcoming turn steps a player should skip.
	SkippedSteps map[PlayerID]map[Step]int

	// CostModifiers are runtime generic cost increases/reductions/taxes.
	CostModifiers []CostModifier

	// AttackTaxes are Ghostly Prison-style costs to attack a player.
	AttackTaxes []AttackTax

	// RuleEffects are runtime permission/prohibition/cost effects.
	RuleEffects []RuleEffect

	// SuspendedCards tracks cards exiled with suspend and their time counters.
	SuspendedCards map[id.ID]SuspendedCard

	// LastKnownInformation stores snapshots for objects that have moved zones.
	LastKnownInformation map[id.ID]ObjectSnapshot

	// LinkedObjects stores objects associated with linked ability pairs.
	LinkedObjects map[LinkedObjectKey][]LinkedObjectRef

	// Turn tracks the current turn, phase, step, and priority.
	Turn TurnState

	// TurnOrder manages player rotation and elimination.
	TurnOrder TurnOrder

	// FailedDraws tracks players who attempted to draw from an empty library.
	// State-based actions eliminate those players the next time they are checked.
	FailedDraws map[PlayerID]bool

	// Combat holds the current combat state. Nil outside of the combat phase.
	Combat *CombatState

	// Emblems lists all emblems in the command zone.
	Emblems []Emblem

	// DayNight tracks whether it is currently day or night. Nil if the
	// day/night cycle has not been established.
	DayNight *DayNightState

	// Events records rules-relevant facts emitted by the rules engine as
	// state-changing helpers mutate this game. It is distinct from GameResult
	// logs, which are report-oriented summaries produced by rules.Engine.
	Events []GameEvent

	// EventTurnStarts records the Events index where each turn's partition
	// starts. Turn N uses index N-1.
	EventTurnStarts []int

	// TriggerEventCursor is the index of the next event the rules engine should
	// inspect for triggered ability detection.
	TriggerEventCursor int

	// StateTriggerLatches prevents state triggers from repeatedly triggering
	// while their condition remains true (CR 603.8).
	StateTriggerLatches map[StateTriggerKey]bool

	// ActivatedAbilitiesThisTurn records once-per-turn activated abilities used
	// during the current turn.
	ActivatedAbilitiesThisTurn map[ActivatedAbilityUse]bool

	// TriggeredAbilitiesThisTurn records triggered ability trigger counts during
	// the current turn for abilities with MaxTriggersPerTurn.
	TriggeredAbilitiesThisTurn map[TriggeredAbilityUse]int

	// IDGen generates unique IDs for game objects.
	IDGen id.Generator
}

// ActivatedAbilityUse identifies one activated ability on one source object.
type ActivatedAbilityUse struct {
	SourceID     id.ID
	AbilityIndex int
}

// TriggeredAbilityUse identifies one triggered ability on one source object.
type TriggeredAbilityUse struct {
	SourceID     id.ID
	AbilityIndex int
}

// NewGame creates and initializes a new 4-player Commander game from the given
// player configurations. Use NewGameWithRand when tests or simulations need
// reproducible library shuffles.
//
// It:
//   - Creates players with 40 life
//   - Creates CardInstances for all cards in each player's deck
//   - Places commanders in command zones
//   - Adds deck cards to libraries
//   - Shuffles libraries
//   - Sets turn 1 with Player1 as the active player
//
//nolint:gocritic // The fixed-size player array is the public setup API.
func NewGame(configs [NumPlayers]PlayerConfig) *Game {
	return NewGameWithRand(configs, rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))
}

// NewGameWithRand creates a game using rng for all setup randomness. The same
// rng is consumed sequentially across players' library shuffles.
//
//nolint:gocritic // The fixed-size player array is the public setup API.
func NewGameWithRand(configs [NumPlayers]PlayerConfig, rng *rand.Rand) *Game {
	if rng == nil {
		panic("nil rng")
	}
	g := &Game{
		CardInstances:              make(map[id.ID]*CardInstance),
		CommanderIDs:               make(map[id.ID]bool),
		SuspendedCards:             make(map[id.ID]SuspendedCard),
		LastKnownInformation:       make(map[id.ID]ObjectSnapshot),
		LinkedObjects:              make(map[LinkedObjectKey][]LinkedObjectRef),
		SkippedSteps:               make(map[PlayerID]map[Step]int),
		TurnOrder:                  NewTurnOrder(),
		FailedDraws:                make(map[PlayerID]bool),
		StateTriggerLatches:        make(map[StateTriggerKey]bool),
		ActivatedAbilitiesThisTurn: make(map[ActivatedAbilityUse]bool),
		TriggeredAbilitiesThisTurn: make(map[TriggeredAbilityUse]int),
		EventTurnStarts:            []int{0},
		Turn: TurnState{
			TurnNumber:           1,
			ActivePlayer:         Player1,
			Phase:                PhaseBeginning,
			Step:                 StepUntap,
			PriorityPlayer:       Player1,
			LandsAllowedThisTurn: 1,
		},
	}

	for i, cfg := range configs {
		pid := PlayerID(i)
		p := NewPlayer(pid, cfg.Name)
		p.PowerBracket = cfg.PowerBracket
		p.PowerLevel = cfg.PowerLevel

		// Create commander CardInstance and place in command zone.
		if cfg.Commander != nil {
			ci := &CardInstance{
				ID:    g.IDGen.Next(),
				Def:   cfg.Commander.WithAbilityBodies(),
				Owner: pid,
			}
			g.CardInstances[ci.ID] = ci
			g.CommanderIDs[ci.ID] = true
			p.CommanderInstanceID = ci.ID
			p.CommandZone.Add(ci.ID)
		}

		// Create CardInstances for the deck and add to library.
		for _, def := range cfg.Deck {
			ci := &CardInstance{
				ID:    g.IDGen.Next(),
				Def:   def.WithAbilityBodies(),
				Owner: pid,
			}
			g.CardInstances[ci.ID] = ci
			p.Library.AddToBottom(ci.ID)
		}

		// Shuffle the library.
		p.Library.Shuffle(rng)

		g.Players[i] = p
	}

	return g
}

// ActivePlayer returns the player whose turn it is.
func (g *Game) ActivePlayer() *Player {
	return g.Players[g.Turn.ActivePlayer]
}

// PriorityHolder returns the player who currently has priority.
func (g *Game) PriorityHolder() *Player {
	return g.Players[g.Turn.PriorityPlayer]
}

// GetCardInstance looks up a CardInstance by its ID.
func (g *Game) GetCardInstance(cardID id.ID) (*CardInstance, bool) {
	card, ok := g.CardInstances[cardID]
	return card, ok
}

// PermanentByID finds a permanent on the battlefield by its ObjectID.
func (g *Game) PermanentByID(objID id.ID) (*Permanent, bool) {
	for _, p := range g.Battlefield {
		if p.ObjectID == objID {
			return p, true
		}
	}
	return nil, false
}

// PermanentsControlledBy returns all permanents controlled by the given player.
func (g *Game) PermanentsControlledBy(pid PlayerID) []*Permanent {
	var result []*Permanent
	for _, p := range g.Battlefield {
		if p.Controller == pid {
			result = append(result, p)
		}
	}
	return result
}

// AlivePlayers returns the players who are still in the game.
func (g *Game) AlivePlayers() []*Player {
	var alive []*Player
	for _, p := range g.Players {
		if p.IsAlive() {
			alive = append(alive, p)
		}
	}
	return alive
}

// IsGameOver reports whether the game has ended (one or fewer players remain).
func (g *Game) IsGameOver() bool {
	return g.TurnOrder.ActivePlayerCount() <= 1
}

// Winner returns the last remaining player when the game is over.
func (g *Game) Winner() (*Player, bool) {
	alive := g.AlivePlayers()
	if len(alive) == 1 {
		return alive[0], true
	}
	return nil, false
}
