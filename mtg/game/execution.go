package game

// ExecutionDiagnostics makes policy substitution and bounded work observable.
// It contains no private card identities and is safe in public observations.
type ExecutionDiagnostics struct {
	Decisions         int
	ActionFallbacks   int
	ChoiceFallbacks   int
	InvalidActions    int
	InvalidChoices    int
	SearchBudgetStops int
}
