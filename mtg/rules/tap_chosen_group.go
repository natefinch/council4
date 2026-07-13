package rules

import "github.com/natefinch/council4/mtg/game"

// handleTapChosenGroup resolves a TapChosenGroup: it prompts the resolving
// controller to choose any number of the candidate group's permanents and taps
// each chosen permanent simultaneously, then publishes the number tapped under
// PublishCount so a later scaled effect reads it through
// DynamicAmountChosenNumber. Choosing zero publishes zero and reports the
// instruction as not succeeded, so a reflexive "If you do" payoff gated on this
// choice resolves to nothing.
//
// The candidate group is resolved as the instruction resolves, so it reflects
// the untapped permanents the controller then controls (CR 608.2). It anchors
// on the ability's controller rather than the source permanent, so tapping still
// happens when the source has already left the battlefield (Myr Battlesphere:
// "you can tap untapped Myr you control even if Myr Battlesphere is no longer on
// the battlefield").
func handleTapChosenGroup(r *effectResolver, prim game.TapChosenGroup) effectResolved {
	res := effectResolved{accepted: true}
	chosen := r.chooseTapGroupPermanents(prim)
	if len(chosen) > 0 {
		setPermanentsTappedSimultaneously(r.game, chosen, true)
	}
	count := len(chosen)
	rememberResolutionChoice(r.obj, string(prim.PublishCount), game.ResolutionChoiceResult{
		Kind:   game.ResolutionChoiceNumber,
		Number: count,
	})
	res.amount = count
	res.succeeded = count > 0
	return res
}

// chooseTapGroupPermanents prompts the resolving controller to choose any number
// of distinct permanents from the primitive's candidate group and returns the
// chosen permanents. It returns nil when the candidate set is empty. The choice
// admits zero selections so the enclosing "you may" declines to tap anything.
func (r *effectResolver) chooseTapGroupPermanents(prim game.TapChosenGroup) []*game.Permanent {
	candidates := r.groupPermanents(prim.ChooseFrom)
	if len(candidates) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(r.game, permanent),
			Card:  permanentChoiceInfo(r.game, permanent),
		}
	}
	prompt := prim.Prompt
	if prompt == "" {
		prompt = "Choose permanents to tap"
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     r.obj.Controller,
		Prompt:     prompt,
		Options:    options,
		MinChoices: 0,
		MaxChoices: len(candidates),
	}, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}
