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
	// TurnLimitReached reports that a goldfish run completed its requested turns.
	TurnLimitReached bool

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

	// LandsPlayed is the number of lands the active player played during this
	// turn. Zero on a turn the active player could have played a land but did
	// not is a missed land drop.
	LandsPlayed int
	// ManaAvailable is the active player's total mana available for the turn,
	// measured at the end of their first precombat main phase (after their land
	// drop). Each mana source the player controls and could tap for mana this
	// turn counts once — lands, mana rocks, and non-summoning-sick mana dorks —
	// approximating open mana the way the engine's own heuristic does (one mana
	// per source). Rituals are excluded because they are spells, not permanents.
	// A source that entered tapped this turn (a tapland) is still counted even
	// though it cannot tap until next turn, so the figure can overstate a
	// tapland turn by one.
	ManaAvailable int
	// ManaColors lists the distinct colors those sources can produce, as
	// single-letter codes (W, U, B, R, G).
	ManaColors []string
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
	// LossReasonGameLossEffect means an effect instructed the player to lose the
	// game (CR 104.3a), such as an unpaid Pact upkeep cost.
	LossReasonGameLossEffect LossReason = "game-loss effect"
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

	// ManaAbility reports that this action activated a mana ability (one that
	// produces mana and resolves without using the stack). It is set only for
	// ActionActivateAbility actions.
	ManaAbility bool

	// ManaTaps lists the permanents tapped for mana while applying this action,
	// in tap order, so a report can show how a spell or ability was paid for.
	// It includes lands and other sources tapped during cost payment.
	ManaTaps []ManaTap
}

// ManaTap records one permanent tapped for mana while paying for an action.
type ManaTap struct {
	// Source is the display name of the tapped permanent.
	Source string
	// Colors lists the mana colors the tap produced, in production order, as
	// single-letter codes (W, U, B, R, G) or the colorless symbol. It may be
	// empty when the produced color was not recorded.
	Colors []string
}

// ResolveLog records a stack object resolving.
type ResolveLog struct {
	StackObjectID id.ID
	SourceID      id.ID
	Controller    game.PlayerID
	Kind          game.StackObjectKind
	Result        string

	// SourceName is the display name of the spell, ability source, or token
	// that resolved, so a report can name an ability's source even though its
	// SourceID is a permanent object ID rather than a card instance ID.
	SourceName string
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
