package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestTargetConstructors(t *testing.T) {
	playerTarget := PlayerTarget(Player2)
	if playerTarget.Kind != TargetPlayer || playerTarget.PlayerID != Player2 {
		t.Fatalf("PlayerTarget() = %+v, want player %v", playerTarget, Player2)
	}
	if playerTarget.PermanentID != 0 || playerTarget.StackObjectID != 0 {
		t.Fatalf("PlayerTarget() set unrelated IDs: %+v", playerTarget)
	}

	permanentTarget := PermanentTarget(id.ID(42))
	if permanentTarget.Kind != TargetPermanent || permanentTarget.PermanentID != 42 {
		t.Fatalf("PermanentTarget() = %+v, want permanent 42", permanentTarget)
	}
	if permanentTarget.PlayerID != 0 || permanentTarget.StackObjectID != 0 {
		t.Fatalf("PermanentTarget() set unrelated IDs: %+v", permanentTarget)
	}

	stackTarget := StackObjectTarget(id.ID(99))
	if stackTarget.Kind != TargetStackObject || stackTarget.StackObjectID != 99 {
		t.Fatalf("StackObjectTarget() = %+v, want stack object 99", stackTarget)
	}
	if stackTarget.PlayerID != 0 || stackTarget.PermanentID != 0 {
		t.Fatalf("StackObjectTarget() set unrelated IDs: %+v", stackTarget)
	}
}
