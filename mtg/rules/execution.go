package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// Termination separates Magic terminal states from incomplete harness runs.
type Termination string

// Termination values classify rules results separately from harness aborts.
const (
	TerminationUnknown     Termination = "unknown"
	TerminationWin         Termination = "rules-win"
	TerminationLoss        Termination = "rules-loss"
	TerminationDraw        Termination = "rules-draw"
	TerminationUnsupported Termination = "unsupported-execution"
	TerminationInvalid     Termination = "invalid-execution"
	TerminationTurnBudget  Termination = "turn-budget"
	TerminationWorkBudget  Termination = "work-budget"
)

// DefaultDecisionLimit bounds live action/choice requests per game.
const DefaultDecisionLimit = 100000

// WorkBudgetError carries bounded progress even when a nonterminating policy
// cannot return from a turn. It is not evidence of a mandatory rules loop.
type WorkBudgetError struct {
	Execution game.ExecutionDiagnostics
	Limit     int
}

func (e WorkBudgetError) Error() string {
	return fmt.Sprintf("decision work budget exhausted at %d/%d decisions", e.Execution.Decisions, e.Limit)
}

func (e *Engine) countDecision(g *game.Game) {
	limit := e.DecisionLimit
	if limit <= 0 {
		limit = DefaultDecisionLimit
	}
	if g.Execution.Decisions >= limit {
		panic(WorkBudgetError{Execution: g.Execution, Limit: limit})
	}
	g.Execution.Decisions++
}

// Limitation is an inspectable restriction, not a search budget. These remain
// conservative even when a particular run never hits a restricted choice.
type Limitation struct {
	ID      string
	Detail  string
	Blocker string
}

// ActionGenerationLimitations inventories production restrictions separately
// from Searcher.Budget/Lookahead and the simulation turn/work limits.
func ActionGenerationLimitations() []Limitation {
	return []Limitation{
		{ID: "combat-enumeration", Detail: "Attack and block declarations offer representative assignments, not every legal subset, defender assignment, blocker assignment or order.", Blocker: "https://github.com/natefinch/council4/issues/3126"},
		{ID: "evaluation-knowledge", Detail: "Heuristic scoring is incomplete; authored bodies execute in Simulator but their consequences and static/replacement/policy value are not completely modeled. Unknown atoms are explicit."},
	}
}

// Exact reports fidelity, not whether the rules happen to produce a winner.
func (r GameResult) Exact() bool {
	return r.Completed() && len(r.Limitations) == 0 &&
		r.Execution.ActionFallbacks == 0 && r.Execution.ChoiceFallbacks == 0
}
