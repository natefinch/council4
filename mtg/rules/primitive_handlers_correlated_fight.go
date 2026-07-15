package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// handleCorrelatedFight pairs two object-scoped linked groups by shared list
// position and fights each pair, backing "Each of those tokens fights a
// different one of those creatures" (Ezuri's Predation). The Subjects group is
// the tokens published by the preceding CreateToken (PublishLinked); the Objects
// group is the creatures counted for the token amount (PublishCountGroup). Both
// lists are read raw and in their original publish order — not through the
// group-membership helpers that filter out departed permanents — so index i of
// each list still refers to the same permanent that was recorded, preserving the
// one-to-one correspondence even when an earlier member has since left.
//
// Pairing caps at min(len(subjects), len(objects)): a token doubler makes the
// Subjects list longer than the Objects list (2N tokens for N creatures), and
// the surplus tokens simply do not fight, matching the official ruling that only
// the originally counted number of fights occur. Each pair is re-resolved to a
// live permanent by object ID at fight time, so a member that has left the
// battlefield or is no longer a creature is skipped (legality is checked as the
// fights happen, one pair at a time), while a member that merely changed
// controllers still fights because it keeps its object identity.
func handleCorrelatedFight(r *effectResolver, prim game.CorrelatedFight) effectResolved {
	subjectKey := linkedObjectSourceKey(r.game, r.obj, string(prim.Subjects))
	objectKey := linkedObjectSourceKey(r.game, r.obj, string(prim.Objects))
	subjects := linkedObjects(r.game, subjectKey)
	objects := linkedObjects(r.game, objectKey)
	res := effectResolved{accepted: true}
	pairs := min(len(subjects), len(objects))
	for i := range pairs {
		subject, subjectOK := permanentByObjectID(r.game, subjects[i].ObjectID)
		object, objectOK := permanentByObjectID(r.game, objects[i].ObjectID)
		if !subjectOK || !objectOK || subject.ObjectID == object.ObjectID ||
			!permanentHasType(r.game, subject, types.Creature) ||
			!permanentHasType(r.game, object, types.Creature) {
			continue
		}
		resolveFightPermanents(r.game, subject, object)
		res.succeeded = true
	}
	return res
}
