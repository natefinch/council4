package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// replacementChoiceContext carries the agents and log needed to prompt a player
// for a CR 616.1 replacement selection from within the free functions that apply
// replacement and prevention effects (e.g. replacementZoneChange,
// applyDamageModifications). Those functions are reached through deep zone-change
// and damage call chains that do not carry the player agents, so the rules layer
// stashes this context on the game for the duration of an agent-driven turn (see
// runTurn) and retrieves it at the selection point.
type replacementChoiceContext struct {
	engine      *Engine
	agents      [game.NumPlayers]PlayerAgent
	log         *TurnLog
	drawLineage *drawReplacementLineage
}

// setReplacementChoiceContext stashes the choice context on the game so the
// replacement-selection chokepoint can prompt the affected player. The caller is
// responsible for clearing it (g.ClearChoiceContext) when the turn ends.
func (e *Engine) setReplacementChoiceContext(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	ctx := &replacementChoiceContext{engine: e, agents: agents, log: log}
	if previous, ok := replacementChoiceContextFor(g); ok {
		ctx.drawLineage = previous.drawLineage
	}
	g.SetChoiceContext(ctx)
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
// engine fell back to a deterministic default (no agent answered). A private
// request still shows the chooser its options but keeps them out of the log.
func (ctx *replacementChoiceContext) chooseReplacement(g *game.Game, chooser game.PlayerID, options []string, private bool) (int, bool) {
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
		Private:          private,
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
	return chooseReplacementDecisionPrivacy(g, chooser, options, false)
}

// chooseZoneChangeReplacementDecision is chooseReplacementDecision for a zone
// change. When the moving card's identity is hidden (CR 400.2, 708.2), the
// choice is private to the chooser and its recorded labels are withheld, since
// a label such as a self-replacement's text can name the card.
func chooseZoneChangeReplacementDecision(g *game.Game, event game.Event, options []string) game.ReplacementDecision {
	return chooseReplacementDecisionPrivacy(g, replacementEventPlayer(event), options, zoneChangeCardHidden(g, event))
}

// zoneChangeCardHidden reports whether the card a zone change moves is hidden
// information before it moves: it leaves a library or hand (CR 400.2) or it is
// face down (CR 406.3, 708.2).
func zoneChangeCardHidden(g *game.Game, event game.Event) bool {
	if event.FaceDown {
		return true
	}
	switch event.FromZone {
	case zone.Library, zone.Hand:
		return true
	case zone.Exile:
		card, ok := g.GetCardInstance(event.CardID)
		return ok && cardFaceDownInZone(g, card, zone.Exile)
	default:
		return false
	}
}

func chooseReplacementDecisionPrivacy(g *game.Game, chooser game.PlayerID, options []string, private bool) game.ReplacementDecision {
	if departedPlayer(g, chooser) && len(g.AlivePlayers()) > 0 {
		chooser = g.TurnOrder.NextActivePlayer(chooser)
	}
	chosen := 0
	usedFallback := true
	if ctx, ok := replacementChoiceContextFor(g); ok && len(options) > 1 {
		chosen, usedFallback = ctx.chooseReplacement(g, chooser, options, private)
	}
	recorded := append([]string(nil), options...)
	if private {
		for i := range recorded {
			recorded[i] = "Private option"
		}
	}
	decision := game.ReplacementDecision{
		Player:       chooser,
		Options:      recorded,
		Selected:     []int{chosen},
		UsedFallback: usedFallback,
	}
	g.ReplacementDecisions = append(g.ReplacementDecisions, decision)
	return decision
}
