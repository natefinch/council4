package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
)

// ActivatedAbilityProfile is a coarse summary of an activated ability, exposed
// so an agent can decide whether activating it is worthwhile rather than
// activating any available ability. It is a heuristic, not a full evaluation.
type ActivatedAbilityProfile struct {
	// SpendsOwnResources reports that the ability's additional cost spends the
	// agent's other cards or permanents — sacrificing a permanent other than the
	// source, discarding cards, or exiling cards other than the source. It does
	// not flag a source that sacrifices or exiles itself (a fetchland cracking
	// for a land), tapping, or paying life, which are routine.
	SpendsOwnResources bool
}

// ActivatedAbilityProfile returns a coarse profile of the ability the action
// activates. The boolean is false when the action is not an ability activation
// or the ability cannot be resolved (for example a hand- or graveyard-activated
// ability), in which case the caller should fall back to its default scoring.
func (o PlayerObservation) ActivatedAbilityProfile(act action.Action) (ActivatedAbilityProfile, bool) {
	payload, ok := act.ActivateAbilityPayload()
	if !ok {
		return ActivatedAbilityProfile{}, false
	}
	_, body, ok := activatedAbilitySource(o.g, o.Player, payload.SourceID, payload.AbilityIndex)
	if !ok {
		return ActivatedAbilityProfile{}, false
	}
	profile := ActivatedAbilityProfile{}
	for _, additional := range game.BodyAdditionalCosts(body) {
		switch additional.Kind {
		case cost.AdditionalSacrifice, cost.AdditionalDiscard, cost.AdditionalExile:
			profile.SpendsOwnResources = true
		default:
		}
	}
	return profile, true
}
