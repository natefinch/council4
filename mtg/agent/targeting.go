package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

// threatAwareTargetChoice picks the target option aimed at the most dangerous
// permanent: the highest-threat opponent permanent when any option targets one
// (removal and damage aimed at the scariest board), otherwise the highest-threat
// own permanent (a beneficial effect placed on the agent's best creature, since a
// "you control" target choice only offers the agent's own permanents). It reports
// ok=false when no option resolves to a single permanent, so the caller falls
// back to the default for player-, stack-, or multi-target options.
//
// It cannot see the effect's intent, so it assumes a permanent-targeting choice
// with any opponent option is harmful — the common case for engine-mediated
// target choices (triggered removal, damage, taps). This is still far better than
// taking the first option offered.
func threatAwareTargetChoice(obs rules.PlayerObservation, request game.ChoiceRequest) ([]int, bool) {
	bestOpponent, bestOpponentThreat := -1, 0.0
	bestOwn, bestOwnThreat := -1, 0.0
	for i := range request.Options {
		permanentID, ok := singlePermanentTarget(request.Options[i].Targets)
		if !ok {
			continue
		}
		permanent, ok := permanentByID(obs, permanentID)
		if !ok {
			continue
		}
		threat := permanentThreat(permanent)
		if permanent.Controller == obs.Player {
			if bestOwn < 0 || threat > bestOwnThreat {
				bestOwn, bestOwnThreat = request.Options[i].Index, threat
			}
			continue
		}
		if bestOpponent < 0 || threat > bestOpponentThreat {
			bestOpponent, bestOpponentThreat = request.Options[i].Index, threat
		}
	}
	if bestOpponent >= 0 {
		return []int{bestOpponent}, true
	}
	if bestOwn >= 0 {
		return []int{bestOwn}, true
	}
	return nil, false
}

// threatAwarePlayerChoice picks the option aimed at the most threatening
// opponent, for "choose an opponent to ..." decisions. It reports ok=false when
// no option targets an opponent, so the caller falls back to the default.
func threatAwarePlayerChoice(obs rules.PlayerObservation, request game.ChoiceRequest) ([]int, bool) {
	model := NewThreatModel(obs)
	best, bestThreat := -1, 0.0
	for i := range request.Options {
		playerID, ok := singlePlayerTarget(request.Options[i].Targets)
		if !ok || playerID == obs.Player {
			continue
		}
		if threat := model.PlayerThreat(playerID); best < 0 || threat > bestThreat {
			best, bestThreat = request.Options[i].Index, threat
		}
	}
	if best >= 0 {
		return []int{best}, true
	}
	return nil, false
}

// singlePermanentTarget returns the permanent an option targets when it targets
// exactly one permanent.
func singlePermanentTarget(targets []game.Target) (id.ID, bool) {
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return 0, false
	}
	return targets[0].PermanentID, true
}

// singlePlayerTarget returns the player an option targets when it targets exactly
// one player.
func singlePlayerTarget(targets []game.Target) (game.PlayerID, bool) {
	if len(targets) != 1 || targets[0].Kind != game.TargetPlayer {
		return 0, false
	}
	return targets[0].PlayerID, true
}
