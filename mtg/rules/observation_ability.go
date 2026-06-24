package rules

import (
	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game/action"
)

// ScorableActivatedAbility returns the value-oriented summary of the ability the
// action activates, so an agent can score it by cost and effect instead of the
// engine's execution primitives. The boolean is false when the action is not an
// ability activation or the ability cannot be resolved (for example a hand- or
// graveyard-activated ability), in which case the caller should fall back to its
// default scoring.
func (o PlayerObservation) ScorableActivatedAbility(act action.Action) (eval.ScorableAbility, bool) {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return eval.ScorableAbility{}, false
	}
	_, body, ok := activatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex)
	if !ok {
		return eval.ScorableAbility{}, false
	}
	return eval.ScorableAbilityOf(body), true
}
