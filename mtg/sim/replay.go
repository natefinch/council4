package sim

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// ReplayRecord captures everything needed to reconstruct one game deterministically
// from its seed: the per-game seed and, per seat, the agent decisions made during
// the game. Decisions are stored as indices and option selections, so a record is
// plain data that round-trips through JSON.
//
// Replay re-runs the same seed (so the engine RNG reproduces every shuffle and
// coin flip) and scripts each seat's agent to repeat its recorded decisions, which
// reproduces the original GameResult exactly.
type ReplayRecord struct {
	Seed     uint64                      `json:"seed"`
	Seats    [game.NumPlayers]SeatScript `json:"seats"`
	Programs []game.ProgramReference     `json:"programs,omitempty"`
}

// SeatScript is one seat's recorded decisions, in the order the engine asked for
// them. Actions[k] is the index, within that call's legal-action list, of the
// action the seat took on its k-th priority decision. Choices[k] is the option
// selection the seat returned on its k-th engine-mediated choice. ChoiceCapable
// records whether the seat's agent answered choices itself; a seat that left
// choices to the engine's fallback must replay the same way, so its scripted
// agent must not answer choices either.
type SeatScript struct {
	Actions       []int   `json:"actions"`
	Choices       [][]int `json:"choices"`
	ChoiceCapable bool    `json:"choiceCapable"`
}

// RecordGame plays game index from cfg with recording agents and returns both its
// result and a ReplayRecord that reproduces it. The result is identical to
// RunOne(cfg, index); the record additionally lets the game be replayed later
// without the original agents.
func RecordGame(cfg Config, index int) (rules.GameResult, ReplayRecord) {
	seed := GameSeed(cfg.Seed, index)
	return recordedRunWithPrograms(cfg.Configs, seed, newAgents(cfg)(seed), cfg.Programs)
}

func recordedRun(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent) (rules.GameResult, ReplayRecord) {
	return recordedRunWithPrograms(configs, seed, base, nil)
}

func recordedRunWithPrograms(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent, programs *cardprogram.Catalog) (rules.GameResult, ReplayRecord) {
	var recorders [game.NumPlayers]seatRecorder
	var agents [game.NumPlayers]rules.PlayerAgent
	var choiceCapable [game.NumPlayers]bool
	for i := range base {
		_, choiceCapable[i] = base[i].(rules.ChoiceAgent)
		agents[i] = wrapForRecording(base[i], &recorders[i])
	}
	engine := engineFor(seed, programs)
	g := engine.NewGame(configs)
	result := *engine.RunGame(g, agents)

	record := ReplayRecord{Seed: seed, Programs: configProgramReferences(configs)}
	for i := range recorders {
		record.Seats[i] = SeatScript{
			Actions:       recorders[i].actions,
			Choices:       recorders[i].choices,
			ChoiceCapable: choiceCapable[i],
		}
	}
	return result, record
}

// Replay reconstructs the game described by record over configs and returns its
// result, which equals the originally recorded GameResult.
func Replay(configs [game.NumPlayers]game.PlayerConfig, record ReplayRecord, catalogs ...cardprogram.Catalog) rules.GameResult {
	if len(catalogs) > 1 || !slices.Equal(record.Programs, configProgramReferences(configs)) {
		panic(rules.UnsupportedError{Reason: "replay program bindings do not match the recorded definitions"})
	}
	var catalog *cardprogram.Catalog
	if len(catalogs) == 1 {
		catalog = &catalogs[0]
	}
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range record.Seats {
		actionAgent := &scriptedActionAgent{actions: record.Seats[i].Actions}
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
	g := engine.NewGame(configs)
	return *engine.RunGame(g, agents)
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
	actions []int
	choices [][]int
}

func wrapForRecording(base rules.PlayerAgent, rec *seatRecorder) rules.PlayerAgent {
	recorder := &recordingAgent{base: base, rec: rec}
	if choiceBase, ok := base.(rules.ChoiceAgent); ok {
		return &recordingChoiceAgent{recordingAgent: recorder, choiceBase: choiceBase}
	}
	return recorder
}

// recordingAgent wraps a base agent and records, for each priority decision, the
// index of the chosen action within the legal list it was offered.
type recordingAgent struct {
	base rules.PlayerAgent
	rec  *seatRecorder
}

func (a *recordingAgent) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	chosen := a.base.ChooseAction(obs, legal)
	a.rec.actions = append(a.rec.actions, indexOfAction(legal, chosen))
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
	actions []int
	cursor  int
}

func (a *scriptedActionAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	index := -1
	if a.cursor < len(a.actions) {
		index = a.actions[a.cursor]
		a.cursor++
	}
	if index >= 0 && index < len(legal) {
		return legal[index]
	}
	return action.Pass()
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
// targets and payloads are matched. It returns -1 when chosen is not a legal
// action, which a well-behaved agent never returns.
func indexOfAction(legal []action.Action, chosen action.Action) int {
	for i := range legal {
		if reflect.DeepEqual(legal[i], chosen) {
			return i
		}
	}
	return -1
}
