package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
)

// faceHasGiftKeyword reports whether any lowered static ability on the face is a
// Gift keyword action (CR 702.171).
func faceHasGiftKeyword(result loweredFaceAbilities) bool {
	for i := range result.StaticAbilities {
		if game.BodyHasKeyword(&result.StaticAbilities[i].Body, game.Gift) {
			return true
		}
	}
	return false
}

// instructionGiftGated reports whether an instruction resolves in only one Gift
// branch, i.e. its effect condition tests the resolving spell's gift-promised
// state ("if the gift was promised" / "if the gift wasn't promised").
func instructionGiftGated(inst *game.Instruction) bool {
	if !inst.Condition.Exists {
		return false
	}
	condition := inst.Condition.Val.Condition
	return condition.Exists && condition.Val.GiftPromised
}

// giftTargetingDependsOnPromise reports whether a Gift face's set of required
// targets differs between its promised and non-promised branches. The Gift
// keyword action promises the gift as the spell is cast (CR 702.171); a spell's
// alternative or additional targets are required only when the gift is promised
// and are ignored otherwise. The executable backend announces a single target
// set shared by the plain cast and the gift cast, so it can only honor a Gift
// spell whose required targets are the same either way: every target spec must
// be referenced by an effect that resolves regardless of the promise.
//
// It flags the face when some target spec is referenced only by
// gift-promised-conditional effects (an alternative "instead" target, or an
// additional promised-only target), because the shared announcement would then
// force that conditional target on the branch that must ignore it, leaving the
// card uncastable in otherwise-legal situations. Such faces fail closed until
// the backend can gate target requirements on the cast branch. Referenced target
// indices are gathered with the shared target-index walker so the check tracks
// exactly the reference forms the runtime resolves.
func giftTargetingDependsOnPromise(result loweredFaceAbilities) bool {
	if !result.SpellAbility.Exists || !faceHasGiftKeyword(result) {
		return false
	}
	ungated := map[[2]int]bool{}
	collect := func(kind targetIndexKind, old int) (int, bool) {
		ungated[[2]int{int(kind), old}] = true
		return old, true
	}
	var totalSpecs int
	for _, mode := range result.SpellAbility.Val.Modes {
		totalSpecs += len(mode.Targets)
		for i := range mode.Sequence {
			inst := &mode.Sequence[i]
			if instructionGiftGated(inst) {
				continue
			}
			transformPrimitiveTargetIndices(inst.Primitive, collect)
		}
	}
	return len(ungated) < totalSpecs
}
