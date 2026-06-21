package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// rampageDelta extracts the dynamic power delta from the canonical Rampage N
// triggered ability so the test exercises exactly what lowering produces.
func rampageDelta(t *testing.T, n int) game.DynamicAmount {
	t.Helper()
	ability := game.RampageTriggeredAbility(n)
	modify, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("rampage content is not a ModifyPT: %#v", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	dynamic := modify.PowerDelta.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("rampage power delta is not dynamic: %#v", modify.PowerDelta)
	}
	return dynamic.Val
}

// TestRampageBonusScalesWithBlockersBeyondFirst covers Rampage N (CR 702.23):
// the source gets +N/+N for each creature blocking it beyond the first, read
// from the current combat's block declarations at resolution.
func TestRampageBonusScalesWithBlockersBeyondFirst(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Craw Giant", 6, 4))
	dynamic := rampageDelta(t, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: attacker.ObjectID}

	cases := []struct {
		blockers int
		want     int
	}{
		{blockers: 0, want: 0},
		{blockers: 1, want: 0},
		{blockers: 2, want: 2},
		{blockers: 3, want: 4},
	}
	for _, tc := range cases {
		g.Combat = &game.CombatState{
			Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		}
		for i := 0; i < tc.blockers; i++ {
			blocker := addCombatPermanent(g, game.Player2, vanillaCreature("Footsoldier", 1, 1))
			g.Combat.Blockers = append(g.Combat.Blockers, game.BlockDeclaration{
				Blocker:  blocker.ObjectID,
				Blocking: attacker.ObjectID,
			})
		}
		got := dynamicAmountValue(g, obj, game.Player1, dynamic)
		if got != tc.want {
			t.Errorf("Rampage 2 with %d blockers = %d, want %d", tc.blockers, got, tc.want)
		}
	}
}

// TestRampageBonusZeroOutsideCombat confirms the bonus is zero when no combat is
// in progress, so a Rampage trigger that somehow resolves outside combat adds
// nothing.
func TestRampageBonusZeroOutsideCombat(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Craw Giant", 6, 4))
	obj := &game.StackObject{Controller: game.Player1, SourceID: attacker.ObjectID}
	if got := dynamicAmountValue(g, obj, game.Player1, rampageDelta(t, 3)); got != 0 {
		t.Errorf("Rampage 3 outside combat = %d, want 0", got)
	}
}
