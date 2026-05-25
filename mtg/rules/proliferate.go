package rules

import (
	"fmt"
	"sort"

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
	if g == nil || obj == nil {
		return false
	}
	applied := false
	for _, target := range proliferateTargets(g) {
		if len(target.counters) == 0 {
			continue
		}
		selected := e.chooseChoice(g, agents, proliferateChoiceRequest(obj.Controller, target), log)
		if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(target.counters) {
			continue
		}
		if addProliferatedCounter(g, target, target.counters[selected[0]]) {
			applied = true
		}
	}
	return applied
}

func proliferateTargets(g *game.Game) []proliferateTarget {
	var targets []proliferateTarget
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.Counters.IsEmpty() {
			continue
		}
		targets = append(targets, proliferateTarget{
			permanentID: permanent.ObjectID,
			counters:    sortedCounterKinds(permanent.Counters.All()),
		})
	}
	for playerID, player := range g.Players {
		if player == nil || player.Eliminated {
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
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].String() < kinds[j].String()
	})
	return kinds
}

func proliferateChoiceRequest(player game.PlayerID, target proliferateTarget) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(target.counters))
	for i, kind := range target.counters {
		options = append(options, game.ChoiceOption{Index: i, Label: kind.String()})
	}
	prompt := "Proliferate: choose counter kind."
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

func addProliferatedCounter(g *game.Game, target proliferateTarget, kind counter.Kind) bool {
	if target.permanentID != 0 {
		permanent := permanentByObjectID(g, target.permanentID)
		if permanent == nil {
			return false
		}
		permanent.Counters.Add(kind, 1)
		return true
	}
	player := playerByID(g, target.player)
	if player == nil {
		return false
	}
	switch kind {
	case counter.Poison:
		player.PoisonCounters++
	case counter.Energy:
		player.EnergyCounters++
	case counter.Experience:
		player.ExperienceCounters++
	default:
		return false
	}
	return true
}
