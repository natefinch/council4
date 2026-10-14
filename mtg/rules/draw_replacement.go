package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Lineage is invocation-local Engine state. Each derived event inherits applied
// effects (CR 614.5), but siblings never share subsequent applications.
type drawReplacementLineage struct {
	applied         map[id.ID]bool
	firstInDrawStep bool
	eventStart      int
	player          game.PlayerID
}

func (l *drawReplacementLineage) branch() *drawReplacementLineage {
	return &drawReplacementLineage{applied: maps.Clone(l.applied)}
}

// drawCardWithReplacements handles exactly one pending draw, completing every
// replacement before the next draw (CR 121.2, 614.11). Authored actions consume
// the draw independently of whether they produce a card.
func (e *Engine) drawCardWithReplacements(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog, firstInDrawStep bool) bool {
	lineage := &drawReplacementLineage{applied: make(map[id.ID]bool)}
	if ctx, ok := replacementChoiceContextFor(g); ok && ctx.drawLineage != nil {
		lineage = ctx.drawLineage.branch()
		firstInDrawStep = ctx.drawLineage.player == playerID && ctx.drawLineage.firstInDrawStep &&
			!playerDrewSince(g, playerID, ctx.drawLineage.eventStart)
	}
	return e.replacePendingDraw(g, playerID, agents, log, firstInDrawStep, lineage)
}

func (e *Engine) replacePendingDraw(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog, firstInDrawStep bool, lineage *drawReplacementLineage) bool {
	if !isPlayerAlive(g, playerID) || playerAtDrawLimit(g, playerID) {
		return false
	}
	if log == nil {
		log = &TurnLog{}
	}
	if g.Turn.Step == game.StepDraw && g.Turn.ActivePlayer == playerID {
		firstInDrawStep = firstActualDrawThisStep(g, playerID)
	}
	previous, _ := g.ChoiceContext()
	e.setReplacementChoiceContext(g, agents, log)
	ctx, _ := replacementChoiceContextFor(g)
	lineage.firstInDrawStep, lineage.eventStart = firstInDrawStep, len(g.Events)
	lineage.player = playerID
	ctx.drawLineage = lineage
	defer g.SetChoiceContext(previous)

	// Declining is not applying an effect. Do not reoffer it to this unchanged
	// event, but it remains available for draws created by another replacement.
	skipped := make(map[id.ID]bool)
	dredgeSkipped := false
	for {
		if playerAtDrawLimit(g, playerID) {
			return false
		}
		event := game.Event{
			Kind: game.EventCardDrawn, Controller: playerID, Player: playerID,
			FromZone: zone.Library, ToZone: zone.Hand, Amount: 1, FirstInDrawStep: firstInDrawStep,
		}
		replacements := matchingDrawReplacements(g, event, lineage.applied, skipped)
		self := slices.ContainsFunc(replacements, func(r game.ReplacementEffect) bool { return r.SelfReplacement })
		if self {
			replacements = slices.DeleteFunc(replacements, func(r game.ReplacementEffect) bool { return !r.SelfReplacement })
		}
		hasDredge := !self && !dredgeSkipped && len(eligibleDredgeCandidates(g, playerID)) > 0
		labels := replacementEffectLabels(replacements)
		if hasDredge {
			labels = append(labels, "Dredge")
		}
		if len(labels) == 0 {
			cardID, drew := e.drawCard(g, playerID, firstInDrawStep)
			log.addDraw(DrawLog{Player: playerID, CardID: cardID, Failed: !drew})
			return drew
		}
		chosen := 0
		if len(labels) > 1 {
			chosen = chooseReplacementDecision(g, playerID, labels).Selected[0]
		}
		if hasDredge && chosen == len(replacements) {
			if e.tryDredge(g, playerID, agents, log) {
				return false
			}
			dredgeSkipped = true
			continue
		}
		replacement := replacements[chosen]
		if replacement.Optional {
			request := mayChoiceRequest(playerID, "Apply replacement: "+replacement.Description+"?")
			request.DefaultSelection = []int{0}
			answer := e.chooseChoice(g, agents, request, log)
			if len(answer) != 1 || answer[0] != 1 {
				skipped[replacement.ID] = true
				continue
			}
		}
		lineage.applied[replacement.ID] = true
		if replacement.Program != (game.ProgramReference{}) {
			e.runReplacementProgram(g, replacement, event, agents, log)
			return playerDrewSince(g, playerID, lineage.eventStart)
		}
		if replacement.DrawCardDigLook > 0 {
			e.digCards(g, agents, log, nil, playerID, replacement.DrawCardDigLook, replacement.DrawCardDigTake, replacement.DrawCardDigRemainder, digFilter{}, nil)
			return false
		}
		if replacement.DrawFromEmptyLibraryWins {
			markPlayerWinsGame(g, playerID)
			return false
		}
		drew := false
		for range replacement.DrawCardMultiplier {
			// Each modified draw receives the parent lineage, not mutations from
			// its siblings. An undrawn/replaced first draw remains the first.
			drew = e.replacePendingDraw(g, playerID, agents, log, firstInDrawStep && !drew, lineage.branch()) || drew
		}
		return drew
	}
}

func playerDrewSince(g *game.Game, playerID game.PlayerID, start int) bool {
	return slices.ContainsFunc(g.Events[start:], func(event game.Event) bool {
		return event.Kind == game.EventCardDrawn && event.Player == playerID
	})
}

func firstActualDrawThisStep(g *game.Game, playerID game.PlayerID) bool {
	for index := len(g.Events) - 1; index >= 0; index-- {
		event := g.Events[index]
		if event.Kind == game.EventBeginningOfStep {
			break
		}
		if event.Kind == game.EventCardDrawn && event.Player == playerID {
			return false
		}
	}
	return true
}

func matchingDrawReplacements(g *game.Game, event game.Event, applied, skipped map[id.ID]bool) []game.ReplacementEffect {
	var matches []game.ReplacementEffect
	player, ok := playerByID(g, event.Player)
	if !ok {
		return nil
	}
	for _, replacement := range g.ReplacementEffects {
		if applied[replacement.ID] || skipped[replacement.ID] {
			continue
		}
		if replacement.Program == (game.ProgramReference{}) && replacement.DrawCardMultiplier <= 1 &&
			replacement.DrawCardDigLook <= 0 && !replacement.DrawFromEmptyLibraryWins {
			continue
		}
		if replacement.DrawCardExceptFirstInDrawStep && event.FirstInDrawStep ||
			replacement.DrawFromEmptyLibraryWins && player.Library.Size() != 0 {
			continue
		}
		source, _ := permanentByObjectID(g, replacement.SourceObjectID)
		if replacementEffectMatchesEventWithSource(g, &replacement, event, source, nil) {
			matches = append(matches, replacement)
		}
	}
	return matches
}

func (e *Engine) runReplacementProgram(g *game.Game, replacement game.ReplacementEffect, event game.Event, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	obj := &game.StackObject{
		SourceID: replacement.SourceObjectID, SourceCardID: replacement.SourceCardID,
		Controller:   replacementCurrentController(g, &replacement),
		TriggerEvent: event, HasTriggerEvent: true,
	}
	ctx := &programContext{
		engine: e, game: g, obj: obj, agents: agents, log: log, active: true,
		source: cardprogram.Source{
			CardID: replacement.SourceCardID, ObjectID: replacement.SourceObjectID,
			Controller: obj.Controller, EventPlayer: opt.Val(event.Player),
		},
	}
	defer func() { ctx.active = false }()
	if err := e.programs.Run(replacement.Program, ctx); err != nil {
		panic(UnsupportedError{Reason: err.Error()})
	}
}
