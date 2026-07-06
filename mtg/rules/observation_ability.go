package rules

import (
	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game"
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
	return eval.ScorableAbilityOfModes(body, payload.ChosenModes), true
}

// RepeatedFreeActivation reports whether act re-activates, this turn, an ability
// that costs nothing to activate — no mana after its announced X, and no
// additional cost such as tapping or sacrificing — and has already been activated
// this turn. Only a zero-cost ability can be activated without limit, and a
// repeat of one almost never changes anything the first activation did not, so an
// agent scores such a repeat at or below passing rather than re-activating it
// without end (equip {0}, a tapped-out "{X}" ability at X = 0). A cost that spends
// a bounded resource (mana, tapping, sacrificing) already limits repetition, so
// those are not flagged.
func (o PlayerObservation) RepeatedFreeActivation(act action.Action) bool {
	payload, ok := act.ActivateAbilityPayload()
	if !ok || payload.XValue != 0 {
		return false
	}
	use := game.ActivatedAbilityUse{SourceID: payload.SourceID, AbilityIndex: payload.AbilityIndex}
	if o.g.AbilityActivationsThisTurn[use] == 0 {
		return false
	}
	permanent, body, ok := activatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex)
	if !ok || permanent == nil {
		return false
	}
	activated, ok := body.(*game.ActivatedAbility)
	if !ok || len(game.BodyAdditionalCosts(activated)) != 0 {
		return false
	}
	sourceCard, _ := o.g.GetCardInstance(permanent.CardInstanceID)
	manaCost := manaCostPtr(effectiveActivatedAbilityCost(o.g, o.Player, sourceCard, activated))
	return manaCost == nil || manaCost.ManaValue() == 0
}
