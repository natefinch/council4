package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// putEnteredBlocking makes blocker, a creature just put onto the battlefield,
// block the attacking creature attackerID (CR 509.4). batchID groups the block
// events of one effect.
//
// The creature blocks only if attackerID is still an attacking creature and the
// blocker's controller is a defending player for it; otherwise it stays on the
// battlefield without blocking (CR 509.4a). Blocking requirements and
// restrictions don't apply (CR 509.4b). A blocker of one band member blocks the
// whole band (CR 702.22h-i). Each new blocking pair emits the attacker's
// becomes-blocked event, marked already blocked unless the attacker was
// unblocked, so "becomes blocked by a creature" triggers per new pair and bare
// "becomes blocked" only on the transition (CR 509.3c-d). The creature never
// "blocked", so no blocker event is emitted (CR 509.3a-b, 509.4). It reports
// whether any block was recorded.
func putEnteredBlocking(g *game.Game, blocker *game.Permanent, attackerID id.ID, batchID id.ID) bool {
	if g.Combat == nil || !activeBattlefieldPermanent(blocker) ||
		!permanentHasType(g, blocker, types.Creature) || permanentHasType(g, blocker, types.Battle) {
		return false
	}
	if !enteredBlockAttackerActive(g, attackerID) ||
		effectiveController(g, blocker) != attackerDefendingPlayer(g, attackerID) {
		return false
	}
	var blocks []game.BlockDeclaration
	for _, member := range attackingBandMembers(g.Combat, attackerID) {
		block := game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: member}
		if !enteredBlockAttackerActive(g, member) || slices.Contains(g.Combat.Blockers, block) {
			continue
		}
		blocks = append(blocks, block)
	}
	if len(blocks) == 0 {
		return false
	}
	if g.Combat.BlockedAttackers == nil {
		g.Combat.BlockedAttackers = make(map[id.ID]bool)
	}
	if g.Combat.BlockerOrder == nil {
		g.Combat.BlockerOrder = make(map[id.ID][]id.ID)
	}
	g.Combat.Blockers = append(g.Combat.Blockers, blocks...)
	recordCombatantController(g, blocker.ObjectID)
	newlyBlocked := make([]bool, len(blocks))
	for i, block := range blocks {
		newlyBlocked[i] = !g.Combat.BlockedAttackers[block.Blocking]
		g.Combat.BlockedAttackers[block.Blocking] = true
		g.Combat.BlockerOrder[block.Blocking] = append(g.Combat.BlockerOrder[block.Blocking], block.Blocker)
	}
	for i, block := range blocks {
		emitBlockEvents(g, block, nil, newlyBlocked[i], false, batchID)
	}
	return true
}

// enteredBlockAttackerActive reports whether attackerID is an attacking
// creature on the battlefield that an entering creature can block.
func enteredBlockAttackerActive(g *game.Game, attackerID id.ID) bool {
	if _, ok := attackDeclarationForAttacker(g, attackerID); !ok {
		return false
	}
	attacker, ok := permanentByObjectID(g, attackerID)
	return ok && activeBattlefieldPermanent(attacker)
}

// putEnteredPermanentsBlocking puts permanents that just entered the battlefield
// blocking the referenced attacking creature, as one effect's block batch
// (CR 509.4). A reference that no longer resolves to a permanent leaves them on
// the battlefield without blocking (CR 509.4a).
func (r *effectResolver) putEnteredPermanentsBlocking(permanents []*game.Permanent, ref game.ObjectReference) {
	if len(permanents) == 0 {
		return
	}
	resolved, ok := resolveObjectReference(r.game, r.obj, ref)
	if !ok || resolved.permanent == nil {
		return
	}
	batchID := r.game.IDGen.Next()
	for _, permanent := range permanents {
		putEnteredBlocking(r.game, permanent, resolved.permanent.ObjectID, batchID)
	}
}
