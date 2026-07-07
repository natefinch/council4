package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// currentMonarch returns the player who currently holds the monarch designation
// (CR 720), or an unset value when no player is the monarch. At most one player
// is the monarch at a time.
func currentMonarch(g *game.Game) opt.V[game.PlayerID] {
	for i := range g.Players {
		if g.Players[i].IsMonarch {
			return opt.Val(game.PlayerID(i))
		}
	}
	return opt.V[game.PlayerID]{}
}

// livingMonarch returns the current monarch only when that player is still in
// the game. IsMonarch is never cleared when a player leaves (only setMonarch
// unsets it), so a departed monarch keeps the flag; consumers that grant an
// ongoing benefit to "the monarch" (e.g. Fealty to the Realm's continuous
// control effect) must ignore an eliminated monarch (CR 800.4a), matching the
// IsAlive gate used by anyPlayerIsMonarch.
func livingMonarch(g *game.Game) opt.V[game.PlayerID] {
	if monarch := currentMonarch(g); monarch.Exists && g.Players[monarch.Val].IsAlive() {
		return monarch
	}
	return opt.V[game.PlayerID]{}
}

// setMonarch makes playerID the monarch (CR 720.2). At most one player is the
// monarch at a time, so any prior monarch loses the designation. It reports
// whether playerID is an active player able to take the crown. A player who was
// not already the monarch emits EventBecameMonarch so "whenever you/an opponent
// become(s) the monarch" triggers fire (CR 720.2); a player who is already the
// monarch keeps the crown without re-triggering (CR 720.5).
func setMonarch(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || playerID >= game.NumPlayers {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated || player.CantBecomeMonarchThisTurn {
		return false
	}
	alreadyMonarch := player.IsMonarch
	for i := range g.Players {
		g.Players[i].IsMonarch = false
	}
	player.IsMonarch = true
	if !alreadyMonarch {
		emitEvent(g, game.Event{
			Kind:       game.EventBecameMonarch,
			Controller: playerID,
			Player:     playerID,
		})
	}
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

// handleCantBecomeMonarch resolves the CantBecomeMonarch primitive, blocking the
// referenced player from becoming the monarch for the rest of the turn.
func handleCantBecomeMonarch(r *effectResolver, prim game.CantBecomeMonarch) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	player.CantBecomeMonarchThisTurn = true
	res.succeeded = true
	return res
}
