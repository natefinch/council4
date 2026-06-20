package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

type proliferateTarget struct {
	player      game.PlayerID
	permanentID id.ID
	counters    []counter.Kind
}

func (e *Engine) resolveProliferate(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	applied := false
	controller := stackObjectController(obj)
	for _, target := range proliferateTargets(g) {
		if len(target.counters) == 0 {
			continue
		}
		selected := e.chooseChoice(g, agents, proliferateChoiceRequest(controller, target), log)
		if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(target.counters) {
			continue
		}
		if addProliferatedCounter(g, controller, target, target.counters[selected[0]]) {
			applied = true
		}
	}
	return applied
}

func proliferateTargets(g *game.Game) []proliferateTarget {
	var targets []proliferateTarget
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || permanent.Counters.IsEmpty() {
			continue
		}
		targets = append(targets, proliferateTarget{
			permanentID: permanent.ObjectID,
			counters:    sortedCounterKinds(permanent.Counters.All()),
		})
	}
	for playerID, player := range g.Players {
		if player.Eliminated {
			continue
		}
		var counters []counter.Kind
		if player.PoisonCounters > 0 {
			counters = append(counters, counter.Poison)
		}
		if player.EnergyCounters > 0 {
			counters = append(counters, counter.Energy)
		}
		if player.ExperienceCounters > 0 {
			counters = append(counters, counter.Experience)
		}
		if len(counters) > 0 {
			targets = append(targets, proliferateTarget{
				player:   game.PlayerID(playerID),
				counters: counters,
			})
		}
	}
	return targets
}

func sortedCounterKinds(counts map[counter.Kind]int) []counter.Kind {
	kinds := make([]counter.Kind, 0, len(counts))
	for kind, amount := range counts {
		if amount > 0 {
			kinds = append(kinds, kind)
		}
	}
	slices.SortFunc(kinds, func(a, b counter.Kind) int {
		if a.String() < b.String() {
			return -1
		}
		if a.String() > b.String() {
			return 1
		}
		return 0
	})
	return kinds
}

func proliferateChoiceRequest(player game.PlayerID, target proliferateTarget) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(target.counters))
	for i, kind := range target.counters {
		options = append(options, game.ChoiceOption{Index: i, Label: kind.String()})
	}
	var prompt string
	if target.permanentID != 0 {
		prompt = fmt.Sprintf("Proliferate permanent %d: choose counter kind.", target.permanentID)
	} else {
		prompt = fmt.Sprintf("Proliferate player %d: choose counter kind.", target.player)
	}
	defaultSelection := []int(nil)
	if len(options) > 0 {
		defaultSelection = []int{options[0].Index}
	}
	return game.ChoiceRequest{
		Kind:             game.ChoiceProliferate,
		Player:           player,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: defaultSelection,
	}
}

func addProliferatedCounter(g *game.Game, placementController game.PlayerID, target proliferateTarget, kind counter.Kind) bool {
	if target.permanentID != 0 {
		permanent, ok := permanentByObjectID(g, target.permanentID)
		if !ok {
			return false
		}
		return addCountersToPermanentControlledBy(g, placementController, permanent, kind, 1)
	}
	player, ok := playerByID(g, target.player)
	if !ok {
		return false
	}
	return addCountersToPlayerControlledBy(g, placementController, player, kind, 1)
}
