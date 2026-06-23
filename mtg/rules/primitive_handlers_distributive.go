package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// playerControlledSelectionCandidates gathers the active battlefield permanents
// controlled by playerID that satisfy sel, the candidate pool one player
// contributes to a distributive "for each player" effect. Passing source honors
// sel.ExcludeSource (the "other" qualifier) so a distributive clause never picks
// its own source permanent.
func playerControlledSelectionCandidates(g *game.Game, resolver referenceResolver, source *game.Permanent, playerID game.PlayerID, sel game.Selection) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || effectiveController(g, permanent) != playerID {
			continue
		}
		if !resolver.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		candidates = append(candidates, permanent)
	}
	return candidates
}

// chooseUpToOnePermanent has chooser pick up to one permanent from candidates,
// modeling the "up to one ... that player controls" cardinality of a
// distributive per-player effect. An empty pool chooses nothing; otherwise the
// chooser selects which one and may decline because the choice allows zero.
// prompt labels the choice request.
func (e *Engine) chooseUpToOnePermanent(g *game.Game, candidates []*game.Permanent, chooser game.PlayerID, prompt string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (*game.Permanent, bool) {
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
		MinChoices:       0,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}
	selected := e.chooseChoice(g, agents, request, log)
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			return candidates[idx], true
		}
	}
	return nil, false
}
