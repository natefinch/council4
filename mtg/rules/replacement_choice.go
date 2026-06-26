package rules

import "github.com/natefinch/council4/mtg/game"

// replacementChoiceContext carries the agents and log needed to prompt a player
// for a CR 616.1 replacement selection from within the free functions that apply
// replacement and prevention effects (e.g. replacementZoneChange,
// replacementDamageAmount). Those functions are reached through deep zone-change
// and damage call chains that do not carry the player agents, so the rules layer
// stashes this context on the game for the duration of an agent-driven turn (see
// runTurn) and retrieves it at the selection point.
type replacementChoiceContext struct {
	engine *Engine
	agents [game.NumPlayers]PlayerAgent
	log    *TurnLog
}

// setReplacementChoiceContext stashes the choice context on the game so the
// replacement-selection chokepoint can prompt the affected player. The caller is
// responsible for clearing it (g.ClearChoiceContext) when the turn ends.
func (e *Engine) setReplacementChoiceContext(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.SetChoiceContext(&replacementChoiceContext{engine: e, agents: agents, log: log})
}

func replacementChoiceContextFor(g *game.Game) (*replacementChoiceContext, bool) {
	value, ok := g.ChoiceContext()
	if !ok {
		return nil, false
	}
	ctx, ok := value.(*replacementChoiceContext)
	return ctx, ok
}

// chooseReplacement asks the chooser which applicable replacement or prevention
// effect to apply (CR 616.1), returning the chosen option index and whether the
// engine fell back to a deterministic default (no agent answered).
func (ctx *replacementChoiceContext) chooseReplacement(g *game.Game, chooser game.PlayerID, options []string) (int, bool) {
	reqOptions := make([]game.ChoiceOption, len(options))
	for i, label := range options {
		reqOptions[i] = game.ChoiceOption{Index: i, Label: label}
	}
	selected, usedFallback := ctx.engine.chooseChoiceWithFallback(g, ctx.agents, game.ChoiceRequest{
		Kind:             game.ChoiceReplacement,
		Player:           chooser,
		Prompt:           "Choose which replacement or prevention effect to apply.",
		Options:          reqOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, ctx.log)
	if len(selected) > 0 && selected[0] >= 0 && selected[0] < len(options) {
		return selected[0], usedFallback
	}
	return 0, true
}

// chooseReplacementDecision selects which of several applicable replacement or
// prevention effects to apply to an event (CR 616.1). When an agent-driven turn
// is in progress and more than one effect applies, it prompts the chooser (the
// affected object's controller or the affected player); otherwise it falls back
// to the first option deterministically. The decision is recorded for the turn
// log either way, including whether a deterministic fallback was used.
func chooseReplacementDecision(g *game.Game, chooser game.PlayerID, options []string) game.ReplacementDecision {
	chosen := 0
	usedFallback := true
	if ctx, ok := replacementChoiceContextFor(g); ok && len(options) > 1 {
		chosen, usedFallback = ctx.chooseReplacement(g, chooser, options)
	}
	decision := game.ReplacementDecision{
		Player:       chooser,
		Options:      append([]string(nil), options...),
		Selected:     []int{chosen},
		UsedFallback: usedFallback,
	}
	g.ReplacementDecisions = append(g.ReplacementDecisions, decision)
	return decision
}
