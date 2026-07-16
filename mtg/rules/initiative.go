package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// This file implements the initiative designation (CR 720), which mirrors the
// monarch: at most one living player has the initiative at a time. Taking the
// initiative — including via combat damage, and even by a player who already has
// it — and beginning the initiative-holder's upkeep each cause that player to
// venture into Undercity. The venture is queued (a triggered ability) so it is
// resolved where player choices are available.

// currentInitiative returns the player who currently holds the initiative
// designation, or an unset value when no player has it. At most one player has
// the initiative at a time.
func currentInitiative(g *game.Game) opt.V[game.PlayerID] {
	for i := range g.Players {
		if g.Players[i].HasInitiative {
			return opt.Val(game.PlayerID(i))
		}
	}
	return opt.V[game.PlayerID]{}
}

// livingInitiative returns the current initiative-holder only when that player is
// still in the game. HasInitiative is cleared when the holder leaves the game
// (passInitiativeOnElimination), but a consumer that grants an ongoing benefit to
// "the player with the initiative" must still ignore an eliminated holder, so
// this gate matches livingMonarch.
func livingInitiative(g *game.Game) opt.V[game.PlayerID] {
	if init := currentInitiative(g); init.Exists && g.Players[init.Val].IsAlive() {
		return init
	}
	return opt.V[game.PlayerID]{}
}

// setInitiative makes playerID take the initiative (CR 720). At most one player
// has the initiative at a time, so any prior holder loses it. Unlike the monarch,
// a player takes the initiative even if they already have it, so this always
// emits EventTookInitiative and always queues the follow-up venture into
// Undercity. It reports whether playerID is an active player able to take it.
func setInitiative(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || playerID >= game.NumPlayers {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	for i := range g.Players {
		g.Players[i].HasInitiative = false
	}
	player.HasInitiative = true
	emitEvent(g, game.Event{
		Kind:       game.EventTookInitiative,
		Controller: playerID,
		Player:     playerID,
	})
	queueInitiativeVenture(g, playerID)
	return true
}

// queueInitiativeVenture queues playerID's "venture into Undercity" so it is
// resolved the next time triggered abilities are gathered, where player choices
// are available.
func queueInitiativeVenture(g *game.Game, playerID game.PlayerID) {
	g.PendingInitiativeVentures = append(g.PendingInitiativeVentures, playerID)
}

// takeInitiativeByCombatDamage applies the initiative's combat-damage trigger (CR
// 720): whenever one or more creatures a player controls deal combat damage to
// the initiative-holder, that player takes the initiative. A creature dealing
// damage to its own controller, or to a player who does not have the initiative,
// leaves the initiative unchanged.
func takeInitiativeByCombatDamage(g *game.Game, sourceController, defendingPlayer game.PlayerID) {
	if sourceController == defendingPlayer {
		return
	}
	defender, ok := playerByID(g, defendingPlayer)
	if !ok || !defender.HasInitiative {
		return
	}
	setInitiative(g, sourceController)
}

// passInitiativeOnElimination passes the initiative when its holder leaves the
// game (CR 720): the player whose turn it is takes the initiative, or, if that is
// the player who left, the next player in turn order who is still in the game.
// The new holder ventures into Undercity as they take it.
func passInitiativeOnElimination(g *game.Game, eliminatedID game.PlayerID) {
	holder := currentInitiative(g)
	if !holder.Exists || holder.Val != eliminatedID {
		return
	}
	g.Players[eliminatedID].HasInitiative = false
	newHolder, ok := nextLivingInitiativeHolder(g, eliminatedID)
	if !ok {
		return
	}
	setInitiative(g, newHolder)
}

// nextLivingInitiativeHolder returns the player who should take the initiative
// when it passes on elimination: the active player if they are still in the game
// and did not leave, otherwise the next player in turn order who is still in the
// game.
func nextLivingInitiativeHolder(g *game.Game, eliminatedID game.PlayerID) (game.PlayerID, bool) {
	active := int(g.Turn.ActivePlayer)
	for offset := range game.NumPlayers {
		candidate := game.PlayerID((active + offset) % game.NumPlayers)
		if candidate == eliminatedID {
			continue
		}
		if player, ok := playerByID(g, candidate); ok && player.IsAlive() {
			return candidate, true
		}
	}
	return 0, false
}

// handleTakeInitiative resolves the TakeInitiative primitive, making the
// referenced player take the initiative.
func handleTakeInitiative(r *effectResolver, prim game.TakeInitiative) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = setInitiative(r.game, playerID)
	}
	return res
}
