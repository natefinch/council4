// Package agent contains AI player implementations.
package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// FirstLegal chooses the first legal action offered by the rules engine.
//
// This is intentionally simple: legal actions are ordered by the engine so
// productive actions appear before Pass. FirstLegal plays lands, casts spells,
// declares attacks, and declares blocks when those actions are first in the
// legal action list.
type FirstLegal struct{}

// ChooseAction implements rules.PlayerAgent.
func (FirstLegal) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}

// SimpleCaster plays lands first, then casts the first spell that does not only
// target itself, then falls back to the first legal action.
type SimpleCaster struct{}

// ChooseAction implements rules.PlayerAgent.
func (SimpleCaster) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if act.Kind == action.ActionPlayLand {
			return act
		}
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if ok && !targetsOnlySelf(obs.Player, cast.Targets) {
			return act
		}
	}
	for _, act := range legal {
		if act.Kind == action.ActionCastSpell {
			return act
		}
	}
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}

func targetsOnlySelf(player game.PlayerID, targets []game.Target) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		if target.Kind != game.TargetPlayer || target.PlayerID != player {
			return false
		}
	}
	return true
}
