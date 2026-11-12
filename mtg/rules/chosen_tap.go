package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// chosenTapRequest describes a non-targeted choice of untapped permanents to
// tap: "tap N untapped <permanents> you control". Candidates that are already
// tapped are ignored. The chooser selects between minChoices and maxChoices of
// the rest; defaultCount is the agentless selection size.
type chosenTapRequest struct {
	chooser      game.PlayerID
	candidates   []*game.Permanent
	minChoices   int
	maxChoices   int
	defaultCount int
	prompt       string
	subject      *game.Permanent
}

// chooseAndTapPermanents is the shared chosen-group tap operation used by
// resolving Tap instructions and by optional tap costs such as enlist (CR
// 702.154). The chooser is asked only when a legal selection exists: with
// fewer untapped candidates than minChoices, or none at all, nothing is tapped.
// The chosen permanents become tapped simultaneously. Tapping this way is not
// a {T} cost, so summoning sickness does not restrict it (CR 302.6); callers
// that need a stricter eligibility filter the candidates first.
func (e *Engine) chooseAndTapPermanents(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, request chosenTapRequest) []*game.Permanent {
	untapped := make([]*game.Permanent, 0, len(request.candidates))
	for _, permanent := range request.candidates {
		if !permanent.Tapped {
			untapped = append(untapped, permanent)
		}
	}
	maxChoices := min(request.maxChoices, len(untapped))
	if maxChoices <= 0 || len(untapped) < request.minChoices {
		return nil
	}
	options := make([]game.ChoiceOption, len(untapped))
	for i, permanent := range untapped {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: permanentChoiceLabel(g, permanent),
			Card:  permanentChoiceInfo(g, permanent),
		}
	}
	choice := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           request.chooser,
		Prompt:           request.prompt,
		Options:          options,
		MinChoices:       request.minChoices,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(max(request.minChoices, min(request.defaultCount, maxChoices))),
	}
	if request.subject != nil {
		choice.Subject = permanentChoiceInfo(g, request.subject)
	}
	selected := e.chooseChoice(g, agents, choice, log)
	chosen := make([]*game.Permanent, 0, len(selected))
	seen := make(map[int]bool, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(untapped) && !seen[idx] {
			seen[idx] = true
			chosen = append(chosen, untapped[idx])
		}
	}
	if len(chosen) < request.minChoices || !setPermanentsTappedSimultaneously(g, chosen, true) {
		return nil
	}
	return chosen
}

// handleChosenTap resolves a Tap whose Choose is set: the chooser picks which
// untapped members of the group to tap. The result succeeds only when the
// chosen tap was performed (an "if you do" gate), and its amount is the number
// of permanents tapped.
func (r *effectResolver) handleChosenTap(prim game.Tap) effectResolved {
	res := effectResolved{accepted: true}
	if prim.PublishLinked != "" {
		clearLinkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)))
	}
	chooser := stackObjectController(r.obj)
	if prim.Chooser.Kind() != game.PlayerReferenceNone {
		playerID, ok := r.resolvePlayer(prim.Chooser)
		if !ok {
			return res
		}
		chooser = playerID
	}
	candidates := r.groupPermanents(prim.Group)
	request := chosenTapRequest{
		chooser:    chooser,
		candidates: candidates,
		maxChoices: len(candidates),
		prompt:     "Choose untapped permanents to tap",
	}
	switch prim.Choose {
	case game.TapChoiceExactly:
		amount := r.quantity(prim.Amount)
		if amount <= 0 {
			return res
		}
		request.minChoices, request.maxChoices = amount, amount
	case game.TapChoiceUpTo:
		amount := r.quantity(prim.Amount)
		if amount <= 0 {
			return res
		}
		request.maxChoices = amount
	case game.TapChoiceAnyNumber, game.TapChoiceNone:
	}
	request.defaultCount = request.maxChoices
	tapped := r.engine.chooseAndTapPermanents(r.game, r.agents, r.log, request)
	res.amount = len(tapped)
	res.succeeded = len(tapped) > 0
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		for _, permanent := range tapped {
			rememberLinkedObject(r.game, key, permanentObjectBindingRef(permanent))
		}
	}
	return res
}
