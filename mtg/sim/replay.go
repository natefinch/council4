package sim

import (
	"reflect"

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
	Seed  uint64                      `json:"seed"`
	Seats [game.NumPlayers]SeatScript `json:"seats"`
}

// SeatScript is one seat's recorded decisions, in the order the engine asked for
// them. Actions[k] is the index, within that call's legal-action list, of the
// action the seat took on its k-th priority decision. Choices[k] is the option
// selection the seat returned on its k-th engine-mediated choice.
type SeatScript struct {
	Actions []int   `json:"actions"`
	Choices [][]int `json:"choices"`
}

// RecordGame plays game index from cfg with recording agents and returns both its
// result and a ReplayRecord that reproduces it. The result is identical to
// RunOne(cfg, index); the record additionally lets the game be replayed later
// without the original agents.
func RecordGame(cfg Config, index int) (rules.GameResult, ReplayRecord) {
	seed := GameSeed(cfg.Seed, index)
	return recordedRun(cfg.Configs, seed, newAgents(cfg)(seed))
}

func recordedRun(configs [game.NumPlayers]game.PlayerConfig, seed uint64, base [game.NumPlayers]rules.PlayerAgent) (rules.GameResult, ReplayRecord) {
	var recorders [game.NumPlayers]seatRecorder
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range base {
		agents[i] = wrapForRecording(base[i], &recorders[i])
	}
	engine := rules.NewEngine(NewRand(seed))
	g := engine.NewGame(configs)
	result := *engine.RunGame(g, agents)

	record := ReplayRecord{Seed: seed}
	for i := range recorders {
		record.Seats[i] = SeatScript{Actions: recorders[i].actions, Choices: recorders[i].choices}
	}
	return result, record
}

// Replay reconstructs the game described by record over configs and returns its
// result, which equals the originally recorded GameResult.
func Replay(configs [game.NumPlayers]game.PlayerConfig, record ReplayRecord) rules.GameResult {
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range record.Seats {
		agents[i] = &scriptedAgent{script: record.Seats[i]}
	}
	engine := rules.NewEngine(NewRand(record.Seed))
	g := engine.NewGame(configs)
	return *engine.RunGame(g, agents)
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

// scriptedAgent replays one seat's recorded decisions. It always satisfies
// ChoiceAgent: when its recorded choices are exhausted (or the seat recorded
// none) it returns nil, which the engine treats as an invalid selection and
// resolves with the same deterministic fallback the recording pass used.
type scriptedAgent struct {
	script       SeatScript
	actionCursor int
	choiceCursor int
}

func (a *scriptedAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	index := -1
	if a.actionCursor < len(a.script.Actions) {
		index = a.script.Actions[a.actionCursor]
		a.actionCursor++
	}
	if index >= 0 && index < len(legal) {
		return legal[index]
	}
	return action.Pass()
}

func (a *scriptedAgent) ChooseChoice(_ rules.PlayerObservation, _ game.ChoiceRequest) []int {
	if a.choiceCursor < len(a.script.Choices) {
		selected := a.script.Choices[a.choiceCursor]
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
