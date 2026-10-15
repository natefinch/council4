package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// currentMonarch returns the player who currently holds the monarch designation
// (CR 725), or an unset value when no player is the monarch. At most one player
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

// setMonarch makes playerID the monarch (CR 725.3). At most one player is the
// monarch at a time, so any prior monarch loses the designation. It reports
// whether playerID is an active player able to take the crown. A player who was
// not already the monarch emits EventBecameMonarch so "whenever you/an opponent
// become(s) the monarch" triggers fire; a player who is already the monarch
// keeps the crown without re-triggering.
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

func captureInherentMonarchTriggers(g *game.Game, event game.Event) []game.EventTriggeredAbility {
	monarch := livingMonarch(g)
	if !monarch.Exists || event.Player != monarch.Val {
		return nil
	}
	var instruction game.Primitive
	text := ""
	switch event.Kind {
	case game.EventBeginningOfStep:
		if event.Step != game.StepEnd {
			return nil
		}
		instruction = game.Draw{Amount: game.Fixed(1), Player: game.EventPlayerReference()}
		text = "At the beginning of the monarch's end step, that player draws a card."
	case game.EventDamageDealt:
		if !event.CombatDamage || event.DamageRecipient != game.DamageRecipientPlayer || event.Amount <= 0 ||
			!damageSourceHasAllTypes(g, damageEvent{sourceID: event.SourceID, sourceObjectID: event.SourceObjectID}, []types.Card{types.Creature}) {
			return nil
		}
		instruction = game.BecomeMonarch{Player: game.ObjectControllerReference(game.EventPermanentReference())}
		text = "Whenever a creature deals combat damage to the monarch, its controller becomes the monarch."
	default:
		return nil
	}
	return []game.EventTriggeredAbility{{
		Controller: monarch.Val,
		Ability: &game.TriggeredAbility{
			Text:    text,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: instruction}}}.Ability(),
		},
	}}
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
