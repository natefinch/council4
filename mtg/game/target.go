package game

import "github.com/natefinch/council4/mtg/game/id"

// TargetKind identifies what kind of game object a runtime target points at.
type TargetKind int

const (
	TargetPermanent TargetKind = iota
	TargetPlayer
	TargetStackObject
	TargetDeferred
)

// Target is a runtime targeting choice made while casting a spell or activating
// an ability. Only the ID field matching Kind is meaningful.
type Target struct {
	Kind          TargetKind
	PermanentID   id.ID
	PlayerID      PlayerID
	StackObjectID id.ID
}

// PermanentTarget creates a target pointing at a permanent.
func PermanentTarget(permanentID id.ID) Target {
	return Target{Kind: TargetPermanent, PermanentID: permanentID}
}

// PlayerTarget creates a target pointing at a player.
func PlayerTarget(playerID PlayerID) Target {
	return Target{Kind: TargetPlayer, PlayerID: playerID}
}

// StackObjectTarget creates a target pointing at an object on the stack.
func StackObjectTarget(stackObjectID id.ID) Target {
	return Target{Kind: TargetStackObject, StackObjectID: stackObjectID}
}

// DeferredTarget marks a target slot that will be chosen by a non-controller
// player during spell or ability announcement.
func DeferredTarget() Target {
	return Target{Kind: TargetDeferred}
}
