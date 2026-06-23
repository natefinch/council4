package rules

import "github.com/natefinch/council4/mtg/game"

// setMonarch makes playerID the monarch (CR 720.2). At most one player is the
// monarch at a time, so any prior monarch loses the designation. It reports
// whether playerID is an active player able to take the crown.
func setMonarch(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || playerID >= game.NumPlayers {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	for i := range g.Players {
		g.Players[i].IsMonarch = false
	}
	player.IsMonarch = true
	return true
}

// stealMonarchByCombatDamage applies the monarch's combat-damage trigger (CR
// 720.6): whenever a creature deals combat damage to the monarch, that
// creature's controller becomes the monarch. A creature dealing damage to its
// own controller, or to a non-monarch player, leaves the monarchy unchanged.
func stealMonarchByCombatDamage(g *game.Game, sourceController, defendingPlayer game.PlayerID) {
	if sourceController == defendingPlayer {
		return
	}
	defender, ok := playerByID(g, defendingPlayer)
	if !ok || !defender.IsMonarch {
		return
	}
	setMonarch(g, sourceController)
}

// handleBecomeMonarch resolves the BecomeMonarch primitive, making the
// referenced player the monarch.
func handleBecomeMonarch(r *effectResolver, prim game.BecomeMonarch) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = setMonarch(r.game, playerID)
	}
	return res
}
