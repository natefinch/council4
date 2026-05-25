package game

// EffectResultCondition gates an effect on a previous instruction from the same
// stack-object resolution. This models "if you do" / "if you don't" branches
// while instructions are followed in printed order (CR 608.2c).
type EffectResultCondition struct {
	LinkID string

	Accepted  TriState
	Succeeded TriState
}

// EffectResolutionResult records whether a linked effect was accepted and
// actually did anything rules-relevant. "Succeeded" is distinct from "Accepted"
// because impossible instructions do only as much as possible (CR 101.3).
type EffectResolutionResult struct {
	Accepted  bool
	Succeeded bool
	Amount    int
}
