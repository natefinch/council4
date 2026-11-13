package game

import "slices"

// OpenClashRevealScope begins a resolving effect that may clash, and returns the
// function that ends it. Cards revealed by a clash are shown to every player
// and stay revealed while the rest of that effect uses the clash's result (CR
// 701.20a, 701.30); ending the scope discards the clashes recorded within it.
func (g *Game) OpenClashRevealScope() (closeScope func()) {
	depth := len(g.ClashReveals)
	return func() { g.ClashReveals = g.ClashReveals[:depth] }
}

// RecordClashReveals keeps a completed clash's public reveals and placements
// for the rest of the open clash-reveal scope.
func (g *Game) RecordClashReveals(decisions []ClashDecision) {
	for _, decision := range decisions {
		g.ClashReveals = append(g.ClashReveals, cloneClashDecision(decision))
	}
}

// PublicClashReveals returns a copy of every card revealed by a clash in the
// currently resolving effect, in the order the clashes happened, each with its
// public top/bottom placement.
func (g *Game) PublicClashReveals() []ClashDecision {
	return cloneSliceFunc(g.ClashReveals, cloneClashDecision)
}

func cloneClashDecision(decision ClashDecision) ClashDecision {
	decision.Card.Types = slices.Clone(decision.Card.Types)
	decision.Card.Colors = slices.Clone(decision.Card.Colors)
	return decision
}
