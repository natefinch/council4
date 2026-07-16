package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// eachPlayerChooseCandidates gathers every active battlefield permanent matching
// sel, the single shared candidate pool for an EachPlayerChooseDestroy effect.
// Unlike playerControlledSelectionCandidates it does not restrict candidates to
// one player's permanents: sel is evaluated relative to source (the ability's
// controller), so a controller-relative filter such as "an artifact or
// enchantment you don't control" offers the identical pool to every chooser.
func eachPlayerChooseCandidates(g *game.Game, resolver referenceResolver, source *game.Permanent, sel game.Selection) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		if !resolver.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		candidates = append(candidates, permanent)
	}
	return candidates
}

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

func permanentChoiceExtremumCandidates(g *game.Game, candidates []*game.Permanent, extremum game.PermanentChoiceExtremum) []*game.Permanent {
	if extremum == game.PermanentChoiceExtremumNone || len(candidates) < 2 {
		return candidates
	}
	value := func(permanent *game.Permanent) (int, bool) {
		switch extremum {
		case game.PermanentChoiceGreatestPower:
			return effectivePower(g, permanent), true
		case game.PermanentChoiceGreatestToughness:
			return effectiveToughness(g, permanent)
		case game.PermanentChoiceGreatestManaValue:
			def, ok := permanentCardDef(g, permanent)
			if !ok {
				return 0, false
			}
			return def.ManaValue(), true
		default:
			return 0, false
		}
	}
	bestSet := false
	best := 0
	values := make([]int, len(candidates))
	valid := make([]bool, len(candidates))
	for i, candidate := range candidates {
		values[i], valid[i] = value(candidate)
		if valid[i] && (!bestSet || values[i] > best) {
			best = values[i]
			bestSet = true
		}
	}
	if !bestSet {
		return nil
	}
	filtered := make([]*game.Permanent, 0, len(candidates))
	for i, candidate := range candidates {
		if valid[i] && values[i] == best {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
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
