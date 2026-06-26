package game

// The choice context is a transient hook that lets the rules layer prompt a
// player for an engine-mediated decision (specifically the CR 616.1 replacement
// selection) from code that is reached through free functions that do not carry
// the player agents. The rules layer sets it for the duration of an agent-driven
// turn and clears it afterward; it is held as an opaque any to avoid an import
// cycle between the game and rules packages, mirroring the static-source frame.

// SetChoiceContext stores the rules-owned choice context for the current turn.
func (g *Game) SetChoiceContext(ctx any) {
	g.choiceCtx = ctx
}

// ChoiceContext returns the current choice context and whether one is set.
func (g *Game) ChoiceContext() (any, bool) {
	if g.choiceCtx == nil {
		return nil, false
	}
	return g.choiceCtx, true
}

// ClearChoiceContext removes the current choice context.
func (g *Game) ClearChoiceContext() {
	g.choiceCtx = nil
}
