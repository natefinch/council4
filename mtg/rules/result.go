package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// GameResult is the structured output of a completed game. It folds in the final
// event stream and end-state so consumers (such as the report package) can work
// purely from a GameResult, never from a live *game.Game.
type GameResult struct {
	Winner           game.PlayerID
	HasWinner        bool
	EliminationOrder []game.PlayerID
	Losses           []LossLog
	TurnCount        int
	Turns            []TurnLog

	// Events is the game's full event stream, copied from the game at the end of
	// the run so reports can mine it without a live game.
	Events []game.Event
	// EndState is the final per-player state (life, elimination, remaining hand)
	// for end-of-game analysis such as cards stranded in hand.
	EndState EndState
	// Cards resolves every card instance that appeared in the game to its public
	// name and owner, so event and end-state consumers can attribute cards by
	// name and to the deck that owns them.
	Cards map[id.ID]CardInfo
}

// CardInfo is the public identity of a card instance used by reports.
type CardInfo struct {
	Name      string
	Owner     game.PlayerID
	ManaValue int
	Types     []types.Card
}

// EndState is the final state of every seat at the end of a game.
type EndState struct {
	Players [game.NumPlayers]PlayerEndState
}

// PlayerEndState is one seat's final state.
type PlayerEndState struct {
	Life           int
	Eliminated     bool
	Hand           []id.ID
	LibrarySize    int
	CommanderCasts int
}

// TurnLog records the decisions and outcomes from a single turn.
type TurnLog struct {
	TurnNumber     int
	ActivePlayer   game.PlayerID
	Entries        []TurnLogEntry
	Draws          []DrawLog
	Losses         []LossLog
	Actions        []ActionLog
	Choices        []game.ChoiceDecision
	Resolves       []ResolveLog
	CombatDamage   []CombatDamageLog
	CreatureDamage []CreatureDamageLog
	Deaths         []PermanentDeathLog
}

// TurnLogEntryKind identifies the kind of chronological turn log entry.
type TurnLogEntryKind int

// TurnLogEntry constants identify chronological turn log entry kinds.
const (
	// TurnLogEntryUnknown is the zero value for an unspecified turn log entry.
	TurnLogEntryUnknown TurnLogEntryKind = iota
	TurnLogEntryDraw
	TurnLogEntryLoss
	TurnLogEntryAction
	TurnLogEntryChoice
	TurnLogEntryResolve
	TurnLogEntryCombatDamage
	TurnLogEntryCreatureDamage
	TurnLogEntryDeath
)

// TurnLogEntry records one event in the order it happened during the turn.
type TurnLogEntry struct {
	Kind           TurnLogEntryKind
	Draw           DrawLog
	Loss           LossLog
	Action         ActionLog
	Choice         game.ChoiceDecision
	Resolve        ResolveLog
	CombatDamage   CombatDamageLog
	CreatureDamage CreatureDamageLog
	Death          PermanentDeathLog
}

// DrawLog records a player draw during a game.
type DrawLog struct {
	Player game.PlayerID
	CardID id.ID
	Failed bool
}

// LossReason describes why a player lost the game.
type LossReason string

// LossReason constants describe supported loss causes.
const (
	// LossReasonEmptyLibraryDraw means a player tried to draw from an empty library.
	LossReasonEmptyLibraryDraw    LossReason = "draw from empty library"
	LossReasonZeroLife            LossReason = "0 life"
	LossReasonPoisonCounters      LossReason = "10 poison counters"
	LossReasonCommanderDamage     LossReason = "21 commander damage"
	LossReasonStateBasedEliminate LossReason = "state-based elimination"
)

// LossLog records a player losing the game.
type LossLog struct {
	Player game.PlayerID
	Reason LossReason
}

// ActionLog records a player action that occurred during a game.
type ActionLog struct {
	Player              game.PlayerID
	Action              action.Action
	PermanentSources    map[id.ID]id.ID
	PermanentTokenNames map[id.ID]string
}

// ResolveLog records a stack object resolving.
type ResolveLog struct {
	StackObjectID id.ID
	SourceID      id.ID
	Controller    game.PlayerID
	Kind          game.StackObjectKind
	Result        string
}

// CombatDamageLog records combat damage dealt to a player.
type CombatDamageLog struct {
	Attacker        id.ID
	SourceID        id.ID
	Controller      game.PlayerID
	DefendingPlayer game.PlayerID
	Damage          int
}

// CreatureDamageLog records combat damage dealt to a creature.
type CreatureDamageLog struct {
	SourcePermanent   id.ID
	SourceID          id.ID
	Controller        game.PlayerID
	DamagedPermanent  id.ID
	DamagedSourceID   id.ID
	DamagedController game.PlayerID
	Damage            int
}

// PermanentDeathReason describes why a permanent died or left the battlefield.
type PermanentDeathReason string

// PermanentDeathReason constants describe state-based permanent deaths.
const (
	// PermanentDeathReasonLethalDamage means marked damage was lethal.
	PermanentDeathReasonLethalDamage  PermanentDeathReason = "lethal damage"
	PermanentDeathReasonZeroToughness PermanentDeathReason = "0 toughness"
	PermanentDeathReasonZeroLoyalty   PermanentDeathReason = "0 loyalty"
	PermanentDeathReasonZeroDefense   PermanentDeathReason = "0 defense"
	PermanentDeathReasonIllegalAura   PermanentDeathReason = "illegal aura"
	PermanentDeathReasonLegendaryRule PermanentDeathReason = "legendary rule"
	PermanentDeathReasonSagaComplete  PermanentDeathReason = "Saga final chapter complete"
)

// PermanentDeathLog records a permanent leaving the battlefield due to rules.
type PermanentDeathLog struct {
	Permanent  id.ID
	SourceID   id.ID
	TokenName  string
	Owner      game.PlayerID
	Controller game.PlayerID
	Reason     PermanentDeathReason
}

func (log *TurnLog) addDraw(draw DrawLog) {
	if log == nil {
		return
	}
	log.Draws = append(log.Draws, draw)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryDraw, Draw: draw})
}

func (log *TurnLog) addLoss(loss LossLog) {
	if log == nil {
		return
	}
	log.Losses = append(log.Losses, loss)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryLoss, Loss: loss})
}

func (log *TurnLog) addAction(actionLog *ActionLog) {
	if log == nil {
		return
	}
	log.Actions = append(log.Actions, *actionLog)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryAction, Action: *actionLog})
}

func (log *TurnLog) addChoice(choice game.ChoiceDecision) {
	if log == nil {
		return
	}
	log.Choices = append(log.Choices, choice)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryChoice, Choice: choice})
}

func (log *TurnLog) addResolve(resolve ResolveLog) {
	if log == nil {
		return
	}
	log.Resolves = append(log.Resolves, resolve)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryResolve, Resolve: resolve})
}

func (log *TurnLog) addCombatDamage(damage CombatDamageLog) {
	if log == nil {
		return
	}
	log.CombatDamage = append(log.CombatDamage, damage)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryCombatDamage, CombatDamage: damage})
}

func (log *TurnLog) addCreatureDamage(damage CreatureDamageLog) {
	if log == nil {
		return
	}
	log.CreatureDamage = append(log.CreatureDamage, damage)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryCreatureDamage, CreatureDamage: damage})
}

func (log *TurnLog) addDeath(death PermanentDeathLog) {
	if log == nil {
		return
	}
	log.Deaths = append(log.Deaths, death)
	log.Entries = append(log.Entries, TurnLogEntry{Kind: TurnLogEntryDeath, Death: death})
}

func (r *GameResult) addLosses(losses []LossLog) {
	r.Losses = append(r.Losses, losses...)
	for _, loss := range losses {
		r.EliminationOrder = append(r.EliminationOrder, loss.Player)
	}
}
