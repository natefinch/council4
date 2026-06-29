package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleChampionExile resolves the Champion keyword enters-the-battlefield
// action (CR 702.71): the source's controller exiles another permanent they
// control matching prim.Selection, remembered under prim.LinkedKey so the paired
// leaves-the-battlefield trigger returns it. When the controller controls no
// other matching permanent, the source is sacrificed instead so nothing is
// championed. prim.Selection.ExcludeSource models the "another" qualifier.
func handleChampionExile(r *effectResolver, prim game.ChampionExile) effectResolved {
	res := effectResolved{accepted: true}
	source, ok := sourcePermanent(r.game, r.obj)
	if !ok {
		return res
	}
	resolver := newReferenceResolver(r.game, r.obj)
	candidates := playerControlledSelectionCandidates(r.game, resolver, source, r.obj.Controller, prim.Selection)
	if len(candidates) == 0 {
		if sacrificePermanent(r.game, source) {
			res.succeeded = true
		}
		return res
	}
	permanent, chosen := r.engine.chooseOnePermanent(r.game, candidates, r.obj.Controller, "Choose a permanent to champion", r.agents, r.log)
	if !chosen {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	linkedRef := permanentLinkedObjectRef(permanent)
	if movePermanentToZone(r.game, permanent, zone.Exile) {
		rememberLinkedObject(r.game, key, linkedRef)
		res.succeeded = true
	}
	return res
}

// chooseOnePermanent has chooser pick exactly one permanent from a non-empty
// candidate pool, modeling the mandatory "exile another <permanent> you control"
// Champion payment that has no decline. prompt labels the choice request.
func (e *Engine) chooseOnePermanent(g *game.Game, candidates []*game.Permanent, chooser game.PlayerID, prompt string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
	if len(candidates) == 0 {
		return nil, false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           chooser,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}
	selected := e.chooseChoice(g, agents, request, log)
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			return candidates[idx], true
		}
	}
	return candidates[0], true
}
