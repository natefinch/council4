package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// handleKeepOnePerType resolves the "keep one of each type" family ("Each
// opponent chooses a permanent they control of each permanent type and
// sacrifices the rest." — Liliana, Dreadhorde General's −9; "Each player chooses
// from among the permanents they control an artifact, a creature, an
// enchantment, and a land, then sacrifices the rest." — Cataclysm). Walking
// every affected player in APNAP order, the chooser keeps one permanent of each
// named type present in that player's affected pool, and the player then
// sacrifices every unkept permanent in the pool.
//
// All players' keep choices are gathered before any permanent leaves, so each
// player chooses against the full board, and the unkept permanents are then
// sacrificed simultaneously (CR 800-style single event batch), which fires each
// controller's death triggers together and reads last-known information
// consistently. A permanent that has several of the named types may be kept for
// several type slots, so the kept set is the union of the chosen permanents.
func handleKeepOnePerType(r *effectResolver, prim game.KeepOnePerType) effectResolved {
	res := effectResolved{accepted: true}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	players := playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.Players))

	var toSacrifice []*game.Permanent
	for _, playerID := range players {
		// The affected pool is every permanent the player controls that matches
		// AffectedSelection and is active (phased-out permanents are excluded by
		// playerControlledSelectionCandidates, which also reads effective controller
		// and type so control and type changes are honored). A can't-be-sacrificed
		// permanent stays in the pool so it remains eligible to be the kept one of
		// its type, but is filtered out of the sacrifice set below.
		pool := playerControlledSelectionCandidates(r.game, resolver, source, playerID, prim.AffectedSelection)
		if len(pool) == 0 {
			continue
		}
		chooser := playerID
		if prim.ControllerChoosesForAll {
			chooser = r.obj.Controller
		}
		kept := make(map[game.ObjectID]bool)
		for _, cardType := range prim.Types {
			candidates := keepCandidatesOfType(resolver, source, pool, cardType)
			if len(candidates) == 0 {
				continue
			}
			permanent, ok := r.engine.chooseOnePermanent(r.game, candidates, chooser, "Choose a permanent to keep", r.agents, r.log)
			if !ok {
				continue
			}
			kept[permanent.ObjectID] = true
		}
		for _, permanent := range pool {
			if kept[permanent.ObjectID] {
				continue
			}
			if permanentCantBeSacrificed(r.game, permanent) {
				continue
			}
			toSacrifice = append(toSacrifice, permanent)
		}
	}
	res.amount = len(toSacrifice)
	res.succeeded = sacrificePermanentsSimultaneously(r.game, toSacrifice)
	return res
}

// keepCandidatesOfType filters pool to the permanents that have cardType,
// evaluated through effective permanent values so a permanent whose types have
// changed, or a token, is matched by its current types.
func keepCandidatesOfType(resolver referenceResolver, source *game.Permanent, pool []*game.Permanent, cardType types.Card) []*game.Permanent {
	sel := game.Selection{RequiredTypesAny: []types.Card{cardType}}
	var candidates []*game.Permanent
	for _, permanent := range pool {
		if resolver.permanentMatchesGroupSelection(&sel, source, permanent) {
			candidates = append(candidates, permanent)
		}
	}
	return candidates
}
