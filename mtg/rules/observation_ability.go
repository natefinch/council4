package rules

import (
	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ScorableActivatedAbility returns the value-oriented summary of the ability the
// action activates, so an agent can score it by cost and effect instead of the
// engine's execution primitives. The boolean is false when the action is not an
// ability activation or the ability cannot be resolved to a scorable summary, in
// which case the caller should fall back to its default scoring.
func (o PlayerObservation) ScorableActivatedAbility(act action.Action) (eval.ScorableAbility, bool) {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return eval.ScorableAbility{}, false
	}
	if _, body, ok := activatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex); ok {
		return eval.ScorableAbilityOfModes(body, payload.ChosenModes), true
	}
	// Cycling is activated from the hand, not the battlefield, so it is not found
	// above. Score it on its merits: it draws a card but discards the card being
	// cycled, a cost the ability body does not carry because the engine applies the
	// discard specially (applyCyclingAbilityWithChoices). Without that discard cost
	// cycling reads as free card draw, and the agent cycles its whole hand away
	// instead of keeping cards to develop a game plan and win. Other hand-activated
	// abilities keep the caller's default score.
	if _, body, ok := handActivatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex); ok && game.BodyHasKeyword(&body, game.Cycling) {
		summary := eval.ScorableAbilityOfModes(&body, payload.ChosenModes)
		summary.Costs = append(summary.Costs, cost.Additional{Kind: cost.AdditionalDiscard, Amount: 1, Source: zone.Hand})
		return summary, true
	}
	return eval.ScorableAbility{}, false
}

// IsCyclingActivation reports whether act activates a cycling ability from the
// observing player's hand.
func (o PlayerObservation) IsCyclingActivation(act action.Action) bool {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return false
	}
	_, body, ok := handActivatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex)
	return ok && game.BodyHasKeyword(&body, game.Cycling)
}

// DiscardingTriggersOwnAbility reports whether the observing player discarding a
// card would trigger one of their own permanents' abilities — a "whenever you
// discard" payoff such as Captain Howler (pump a creature and draw on its combat
// damage), Brallin, or Glint-Horn Buccaneer (damage each opponent). An agent uses
// it to know that discarding (through cycling, looting, or a discard cost)
// advances a plan rather than idly churning cards, so the action's value must
// account for the payoff: without a payoff cycling is card-neutral, but with one
// it is a real play the search should keep and evaluate on the resulting board
// (CR 603.2).
func (o PlayerObservation) DiscardingTriggersOwnAbility() bool {
	for _, permanent := range o.g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || effectiveController(o.g, permanent) != o.Player {
			continue
		}
		for _, body := range permanentEffectiveAbilities(o.g, permanent) {
			triggered, ok := body.(*game.TriggeredAbility)
			if !ok {
				continue
			}
			pattern := triggered.Trigger.Pattern
			if pattern.Event == game.EventCardDiscarded && pattern.Player == game.TriggerPlayerYou {
				return true
			}
		}
	}
	return false
}

// IsManaAbilityActivation reports whether the action activates a mana ability —
// one that only adds mana (CR 605.1). An agent scores activating one standalone
// at or below passing: mana is spent through the payment system as it pays for a
// spell or ability, so activating a mana ability on its own merely floats mana
// that empties at end of step. It matters especially for a mana-neutral ability
// that pays for itself (Skyshroud Elf, "{1}: Add {R} or {W}"), which would
// otherwise be activated without end, spinning the priority loop.
func (o PlayerObservation) IsManaAbilityActivation(act action.Action) bool {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return false
	}
	_, body, ok := activatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex)
	if !ok {
		return false
	}
	_, isMana := body.(*game.ManaAbility)
	return isMana
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
