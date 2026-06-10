package game

import "github.com/natefinch/council4/mtg/game/id"

// TargetKind identifies what kind of game object a runtime target points at.
type TargetKind int

// Target kind values identify what a runtime target points at.
const (
	TargetPermanent TargetKind = iota
	TargetPlayer
	TargetStackObject
	TargetDeferred
	TargetCard
)

// Target is a runtime targeting choice made while casting a spell or activating
// an ability. Only the ID field matching Kind is meaningful. Card targets also
// record the chosen zone incarnation once the target is announced.
type Target struct {
	Kind               TargetKind
	PermanentID        id.ID
	PlayerID           PlayerID
	StackObjectID      id.ID
	CardID             id.ID
	CardZoneVersion    uint64
	CardZoneVersionSet bool
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

// CardTarget creates a target pointing at a card in a non-battlefield zone.
func CardTarget(cardID id.ID) Target {
	return Target{Kind: TargetCard, CardID: cardID}
}

// CardTargetWithZoneVersion creates a target pointing at a specific card
// incarnation in a non-battlefield zone.
func CardTargetWithZoneVersion(cardID id.ID, zoneVersion uint64) Target {
	return Target{Kind: TargetCard, CardID: cardID, CardZoneVersion: zoneVersion, CardZoneVersionSet: true}
}

// DeferredTarget marks a target slot that will be chosen by a non-controller
// player during spell or ability announcement.
func DeferredTarget() Target {
	return Target{Kind: TargetDeferred}
}
