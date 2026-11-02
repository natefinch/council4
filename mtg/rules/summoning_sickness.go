package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// permanentSummoningSick reports whether a permanent has not been continuously
// controlled by its current controller since that player's most recent turn
// began (CR 302.6). A control change not yet observed by
// observeControlForSummoningSickness already counts. Callers apply the
// creature and haste exceptions.
func permanentSummoningSick(g *game.Game, permanent *game.Permanent) bool {
	return permanent.SummoningSick || sicknessController(permanent) != effectiveController(g, permanent)
}

func sicknessController(permanent *game.Permanent) game.PlayerID {
	if permanent.SicknessController.Exists {
		return permanent.SicknessController.Val
	}
	return permanent.Controller
}

// observeControlForSummoningSickness records each permanent's current
// controller and marks it summoning sick when that controller differs from the
// one last recorded, so a permanent whose control changes away and back is
// still sick until its controller's next turn begins (CR 302.6). Phased-out
// permanents keep their record because phasing is not a control change.
func observeControlForSummoningSickness(g *game.Game) {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut {
			continue
		}
		markObservedController(permanent, effectiveController(g, permanent))
	}
}

func markObservedController(permanent *game.Permanent, controller game.PlayerID) {
	if controller != sicknessController(permanent) {
		permanent.SummoningSick = true
	}
	permanent.SicknessController = opt.Val(controller)
}
