package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// handleKeepPermanents resolves the count-based keep family ("Each player
// chooses three permanents they control, then sacrifices the rest." — Razia's
// Purification). Each affected player, in APNAP order, chooses which of their
// own affected permanents to keep, knowing the earlier players' choices (CR
// 101.4b). Only after every choice is made are all unkept permanents sacrificed
// in one simultaneous batch (CR 701.21a), so death and leave triggers see a
// single event and last-known information is read consistently.
func handleKeepPermanents(r *effectResolver, prim game.KeepPermanents) effectResolved {
	res := effectResolved{accepted: true}
	keep := max(r.quantity(prim.Keep), 0)
	var players []game.PlayerID
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		players = playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup))
	} else if playerID, ok := r.resolvePlayer(prim.Player); ok {
		players = []game.PlayerID{playerID}
	}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)

	var toSacrifice []*game.Permanent
	var decisions []game.KeepDecision
	for _, playerID := range players {
		// A can't-be-sacrificed permanent stays in the pool so it may be kept,
		// but it is filtered out of the sacrifice set below.
		pool := playerControlledSelectionCandidates(r.game, resolver, source, playerID, prim.AffectedSelection)
		kept := r.engine.chooseKeptPermanents(r.game, playerID, keep, prim.UpTo, pool, decisions, r.agents, r.log)
		decision := game.KeepDecision{Player: playerID}
		keptIDs := make(map[id.ID]bool, len(kept))
		for _, permanent := range kept {
			keptIDs[permanent.ObjectID] = true
			decision.Kept = append(decision.Kept, permanent.ObjectID)
		}
		decisions = append(decisions, decision)
		for _, permanent := range pool {
			if keptIDs[permanent.ObjectID] || permanentCantBeSacrificed(r.game, permanent) {
				continue
			}
			toSacrifice = append(toSacrifice, permanent)
		}
	}
	res.amount = len(toSacrifice)
	res.succeeded = sacrificePermanentsSimultaneously(r.game, toSacrifice)
	r.rememberDepartedResultPermanents(toSacrifice)
	return res
}

// chooseKeptPermanents asks playerID which of pool to keep. Exactly keep
// permanents are kept unless upTo lets the player keep fewer; a player who must
// keep at least as many as they have keeps them all without being asked.
// earlier carries the public choices already made in the same APNAP effect.
func (e *Engine) chooseKeptPermanents(g *game.Game, playerID game.PlayerID, keep int, upTo bool, pool []*game.Permanent, earlier []game.KeepDecision, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []*game.Permanent {
	maxKeep := min(keep, len(pool))
	if maxKeep == 0 {
		return nil
	}
	if !upTo && maxKeep == len(pool) {
		return pool
	}
	minKeep := maxKeep
	if upTo {
		minKeep = 0
	}
	options := make([]game.ChoiceOption, len(pool))
	for i, permanent := range pool {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose permanents to keep; the rest are sacrificed",
		Options:          options,
		MinChoices:       minKeep,
		MaxChoices:       maxKeep,
		DefaultSelection: firstChoiceIndices(maxKeep),
		KeepDecisions:    earlier,
	}
	selected := e.chooseChoice(g, agents, request, log)
	kept := make([]*game.Permanent, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(pool) {
			kept = append(kept, pool[index])
		}
	}
	return kept
}
