package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// handleTakeExtraTurn adds an extra turn for the resolved player directly after
// the current turn (CR 500.7). The grant gets a game-owned identity; the
// primitive's rule effects apply only during that turn and its end-step
// triggers trigger only at that turn's end step, never during another extra
// turn queued for the same player. Nothing is granted to a player who has left
// the game.
func handleTakeExtraTurn(r *effectResolver, prim game.TakeExtraTurn) effectResolved {
	res := effectResolved{accepted: true}
	player, ok := r.resolvePlayer(prim.Player)
	if !ok || !isPlayerAlive(r.game, player) {
		return res
	}
	turnID := r.game.IDGen.Next()
	r.game.Turn.ExtraTurns = append(r.game.Turn.ExtraTurns, game.ExtraTurn{ID: turnID, Player: player})
	start := len(r.game.RuleEffects)
	createRuleEffectTemplates(r.game, r.obj, opt.V[game.ObjectReference]{}, prim.RuleEffects, game.DurationDuringExtraTurn)
	for i := start; i < len(r.game.RuleEffects); i++ {
		r.game.RuleEffects[i].ExtraTurnID = turnID
	}
	for i := range prim.EndStepTriggers {
		def := prim.EndStepTriggers[i]
		def.Timing = game.DelayedAtBeginningOfNextEndStep
		if scheduleDelayedTrigger(r.game, r.obj, &def) {
			r.game.DelayedTriggers[len(r.game.DelayedTriggers)-1].ExtraTurnID = turnID
		}
	}
	res.succeeded = true
	return res
}

// expireExtraTurnScope removes the rule effects and delayed triggers bound to
// the extra turn turnID, once that turn has ended or will never be taken.
// Effects that apply "during that turn" end with it, and a trigger for "that
// turn's end step" never triggers if that turn or its end step was skipped.
func expireExtraTurnScope(g *game.Game, turnID id.ID) {
	if turnID == 0 {
		return
	}
	effects := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		if g.RuleEffects[i].ExtraTurnID != turnID {
			effects = append(effects, g.RuleEffects[i])
		}
	}
	g.RuleEffects = effects
	triggers := g.DelayedTriggers[:0]
	for i := range g.DelayedTriggers {
		if g.DelayedTriggers[i].ExtraTurnID != turnID {
			triggers = append(triggers, g.DelayedTriggers[i])
		}
	}
	g.DelayedTriggers = triggers
}
