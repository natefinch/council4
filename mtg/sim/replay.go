package sim

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// ReplayRecord captures everything needed to reconstruct one game deterministically
// from its seed: the per-game seed and, per seat, the agent decisions made during
// the game. Decisions are stored as sampled indices, exact typed announcements
// and option/numeric selections, all plain data that round-trips through JSON.
//
// Replay re-runs the same seed (so the engine RNG reproduces every shuffle and
// coin flip) and scripts each seat's agent to repeat its recorded decisions, which
// reproduces the original GameResult exactly.
type ReplayRecord struct {
	Seed           uint64                      `json:"seed"`
	Seats          [game.NumPlayers]SeatScript `json:"seats"`
	Programs       []game.ProgramReference     `json:"programs,omitempty"`
	Bindings       []cardprogram.Binding       `json:"bindings,omitempty"`
	TurnLimit      int                         `json:"turn_limit,omitempty"`
	DecisionLimit  int                         `json:"decision_limit,omitempty"`
	SourceRevision string                      `json:"production_source_revision"`
	RulesRevision  string                      `json:"shared_rules_revision"`
}

// SeatScript is one seat's recorded decisions, in the order the engine asked for
// them. Actions[k] is the index, within that call's legal-action list, of the
// action the seat took on its k-th priority decision. Choices[k] is the option
// selection the seat returned on its k-th engine-mediated choice. ChoiceCapable
// records whether the seat's agent answered choices itself; a seat that left
// choices to the engine's fallback must replay the same way, so its scripted
// agent must not answer choices either. Announcements retains full proposals at
// action positions whose values were not among the sampled candidates.
type SeatScript struct {
	Actions       []int                        `json:"actions"`
	Choices       [][]int                      `json:"choices"`
	ChoiceCapable bool                         `json:"choiceCapable"`
	Announcements map[int]RecordedAnnouncement `json:"announcements,omitempty"`
}

// RecordedAnnouncement preserves a complete proposal outside the sampled list.
type RecordedAnnouncement struct {
	Cast       *action.CastSpellAction        `json:"cast,omitempty"`
	Activation *action.ActivateAbilityAction  `json:"activation,omitempty"`
	Attack     *action.DeclareAttackersAction `json:"attack,omitempty"`
	Block      *action.DeclareBlockersAction  `json:"block,omitempty"`
}

// RecordGame plays game index from cfg with recording agents and returns both its
// result and a ReplayRecord that reproduces it. The result is identical to
// RunOne(cfg, index); the record additionally lets the game be replayed later
// without the original agents.
func RecordGame(cfg Config, index int) (rules.GameResult, ReplayRecord) {
	seed := GameSeed(cfg.Seed, index)
	return recordedRunWithLimits(cfg.Configs, seed, newAgents(cfg)(seed), cfg.Programs, cfg.TurnLimit, cfg.DecisionLimit)
}

func recordedRun(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent) (rules.GameResult, ReplayRecord) {
	return recordedRunWithPrograms(configs, seed, base, nil)
}

func recordedRunWithPrograms(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent, programs *cardprogram.Catalog) (rules.GameResult, ReplayRecord) {
	return recordedRunWithLimits(configs, seed, base, programs, 0, 0)
}

func recordedRunWithLimits(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent, programs *cardprogram.Catalog, turnLimit, decisionLimit int) (rules.GameResult, ReplayRecord) {
	var recorders [game.NumPlayers]seatRecorder
	var agents [game.NumPlayers]rules.PlayerAgent
	var choiceCapable [game.NumPlayers]bool
	for i := range base {
		_, choiceCapable[i] = base[i].(rules.ChoiceAgent)
		agents[i] = wrapForRecording(base[i], &recorders[i])
	}
	engine := engineFor(seed, programs)
	engine.DecisionLimit = decisionLimit
	g := engine.NewGame(configs)
	result := *engine.RunGameWithTurnLimit(g, agents, turnLimit)

	record := ReplayRecord{Seed: seed, Programs: configProgramReferences(configs), TurnLimit: turnLimit, DecisionLimit: decisionLimit}
	record.SourceRevision = authored.SourceRevision
	record.RulesRevision = cardprogram.SharedRulesRevision
	bindings, err := engine.ProgramBindings(record.Programs)
	if err != nil {
		panic(rules.UnsupportedError{Reason: err.Error()})
	}
	record.Bindings = bindings
	for i := range recorders {
		record.Seats[i] = SeatScript{
			Actions:       recorders[i].actions,
			Choices:       recorders[i].choices,
			ChoiceCapable: choiceCapable[i],
			Announcements: recorders[i].announcements,
		}
	}
	return result, record
}

// Replay reconstructs the game described by record over configs and returns its
// result, which equals the originally recorded GameResult.
func Replay(configs [game.NumPlayers]game.PlayerConfig, record ReplayRecord, catalogs ...cardprogram.Catalog) rules.GameResult {
	if len(catalogs) > 1 || !slices.Equal(record.Programs, configProgramReferences(configs)) ||
		record.SourceRevision != authored.SourceRevision || record.RulesRevision != cardprogram.SharedRulesRevision {
		panic(rules.UnsupportedError{Reason: "replay program bindings do not match the recorded definitions"})
	}
	var catalog *cardprogram.Catalog
	if len(catalogs) == 1 {
		catalog = &catalogs[0]
	}
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range record.Seats {
		actionAgent := &scriptedActionAgent{actions: record.Seats[i].Actions, announcements: record.Seats[i].Announcements}
		if record.Seats[i].ChoiceCapable {
			// The seat answered choices itself, so replay answers them from the
			// recording. A correctly reproduced game asks the same number of
			// choices, so the script never runs short mid-game.
			agents[i] = &scriptedChoiceAgent{scriptedActionAgent: actionAgent, choices: record.Seats[i].Choices}
		} else {
			// The seat left choices to the engine's deterministic fallback; an
			// action-only agent is not a ChoiceAgent, so the engine falls back
			// identically on replay.
			agents[i] = actionAgent
		}
	}
	engine := engineFor(record.Seed, catalog)
	bindings, err := engine.ProgramBindings(record.Programs)
	if err != nil || !slices.Equal(record.Bindings, bindings) {
		panic(rules.UnsupportedError{Reason: "replay executable source/shared-rules bindings do not match"})
	}
	engine.DecisionLimit = record.DecisionLimit
	g := engine.NewGame(configs)
	return *engine.RunGameWithTurnLimit(g, agents, record.TurnLimit)
}

func configProgramReferences(configs [game.NumPlayers]game.PlayerConfig) []game.ProgramReference {
	var references []game.ProgramReference
	seen := make(map[game.ProgramReference]bool)
	for _, config := range configs {
		for _, def := range append(slices.Clone(config.CommanderDefs()), config.Deck...) {
			if def == nil {
				continue
			}
			for _, reference := range def.ProgramReferences() {
				if !seen[reference] {
					references = append(references, reference)
					seen[reference] = true
				}
			}
		}
	}
	slices.SortFunc(references, func(a, b game.ProgramReference) int {
		if a.ID < b.ID || a.ID == b.ID && a.Version < b.Version {
			return -1
		}
		if a == b {
			return 0
		}
		return 1
	})
	return references
}

// seatRecorder accumulates one seat's decisions during a recorded game.
type seatRecorder struct {
	actions       []int
	choices       [][]int
	announcements map[int]RecordedAnnouncement
}

func wrapForRecording(base rules.PlayerAgent, rec *seatRecorder) rules.PlayerAgent {
	recorder := &recordingAgent{base: base, rec: rec}
	if choiceBase, ok := base.(rules.ChoiceAgent); ok {
		return &recordingChoiceAgent{recordingAgent: recorder, choiceBase: choiceBase}
	}
	return recorder
}

// recordingAgent wraps a base agent and records, for each priority decision, the
// sampled index or exact typed proposal chosen by the base agent.
type recordingAgent struct {
	base rules.PlayerAgent
	rec  *seatRecorder
}

func (a *recordingAgent) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	chosen := a.base.ChooseAction(obs, legal)
	index := indexOfAction(legal, chosen)
	position := len(a.rec.actions)
	a.rec.actions = append(a.rec.actions, index)
	if index < 0 && chosen.Validate() == nil {
		var announcement RecordedAnnouncement
		if cast, ok := chosen.CastSpellPayload(); ok {
			announcement.Cast = &cast
		}
		if activation, ok := chosen.ActivateAbilityPayload(); ok {
			announcement.Activation = &activation
		}
		if attack, ok := chosen.DeclareAttackersPayload(); ok {
			announcement.Attack = &attack
		}
		if block, ok := chosen.DeclareBlockersPayload(); ok {
			announcement.Block = &block
		}
		if announcement.Cast != nil || announcement.Activation != nil || announcement.Attack != nil || announcement.Block != nil {
			if a.rec.announcements == nil {
				a.rec.announcements = make(map[int]RecordedAnnouncement)
			}
			a.rec.announcements[position] = announcement
		}
	}
	return chosen
}

// recordingChoiceAgent additionally records every engine-mediated choice when the
// base agent answers choices. A seat whose base is not a ChoiceAgent records no
// choices and lets the engine's deterministic fallback handle them, on both the
// recording and replay passes.
type recordingChoiceAgent struct {
	*recordingAgent

	choiceBase rules.ChoiceAgent
}

func (a *recordingChoiceAgent) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	selected := a.choiceBase.ChooseChoice(obs, request)
	a.rec.choices = append(a.rec.choices, append([]int(nil), selected...))
	return selected
}

// scriptedActionAgent replays one seat's recorded priority decisions. It is not
// a ChoiceAgent, so the engine answers any choices with its deterministic
// fallback — matching a recording seat whose base did not answer choices.
type scriptedActionAgent struct {
	actions       []int
	cursor        int
	announcements map[int]RecordedAnnouncement
}

func (a *scriptedActionAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	index := -1
	if a.cursor < len(a.actions) {
		index = a.actions[a.cursor]
		announcement, announced := a.announcements[a.cursor]
		a.cursor++
		if announced {
			count := 0
			for _, present := range []bool{announcement.Cast != nil, announcement.Activation != nil, announcement.Attack != nil, announcement.Block != nil} {
				if present {
					count++
				}
			}
			if count != 1 {
				return action.Action{Kind: action.Kind(-1)}
			}
			if announcement.Cast != nil && announcement.Activation == nil {
				return action.AnnounceSpell(*announcement.Cast)
			}
			if announcement.Activation != nil && announcement.Cast == nil {
				return action.AnnounceAbility(*announcement.Activation)
			}
			if announcement.Attack != nil {
				if announcement.Attack.Staged {
					return action.ChooseAttackers()
				}
				return action.DeclareAttackers(announcement.Attack.Attackers)
			}
			if announcement.Block != nil {
				if announcement.Block.Staged {
					return action.ChooseBlockers()
				}
				return action.DeclareBlockers(announcement.Block.Blockers)
			}
			return action.Action{Kind: action.Kind(-1)}
		}
	}
	if index >= 0 && index < len(legal) {
		return legal[index]
	}
	// An invalid recorded answer must take the engine's fallback path again,
	// preserving invalid/fallback counters rather than returning a valid pass.
	return action.Action{Kind: action.Kind(-1)}
}

// scriptedChoiceAgent replays a seat whose base agent answered choices. It
// returns the recorded selections in order; a correctly reproduced game asks the
// same choices in the same order, so the script never runs short mid-game.
type scriptedChoiceAgent struct {
	*scriptedActionAgent

	choices      [][]int
	choiceCursor int
}

func (a *scriptedChoiceAgent) ChooseChoice(_ rules.PlayerObservation, _ game.ChoiceRequest) []int {
	if a.choiceCursor < len(a.choices) {
		selected := a.choices[a.choiceCursor]
		a.choiceCursor++
		return selected
	}
	return nil
}

// indexOfAction returns the index of chosen within legal, comparing by value so
// targets and payloads are matched. It returns -1 for proposals outside the
// sampled list; syntactically valid declarations are recorded separately.
func indexOfAction(legal []action.Action, chosen action.Action) int {
	for i := range legal {
		if reflect.DeepEqual(legal[i], chosen) {
			return i
		}
	}
	return -1
}
