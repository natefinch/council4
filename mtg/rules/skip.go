package rules

import "github.com/natefinch/council4/mtg/game"

func scheduleSkipStep(g *game.Game, playerID game.PlayerID, step game.Step) {
	if step == game.StepNone {
		return
	}
	if g.SkippedSteps == nil {
		g.SkippedSteps = make(map[game.PlayerID]map[game.Step]int)
	}
	if g.SkippedSteps[playerID] == nil {
		g.SkippedSteps[playerID] = make(map[game.Step]int)
	}
	g.SkippedSteps[playerID][step]++
}

func consumeSkipStep(g *game.Game, playerID game.PlayerID, step game.Step) bool {
	if g.SkippedSteps[playerID] == nil || g.SkippedSteps[playerID][step] <= 0 {
		return false
	}
	g.SkippedSteps[playerID][step]--
	if g.SkippedSteps[playerID][step] == 0 {
		delete(g.SkippedSteps[playerID], step)
	}
	return true
}
