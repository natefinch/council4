package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestBlockingCreaturesCountsAllBlockers covers the "+N/+N for each creature
// blocking it" pump amount (Rabid Elephant, Gang of Elk): unlike Rampage's
// beyond-the-first count, it counts every creature blocking the pumped permanent,
// read from the current combat's block declarations at resolution (CR 509.1).
func TestBlockingCreaturesCountsAllBlockers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Rabid Elephant", 1, 3))
	obj := &game.StackObject{Controller: game.Player1, SourceID: attacker.ObjectID}
	dynamic := game.DynamicAmount{
		Kind:       game.DynamicAmountBlockingCreatures,
		Object:     game.SourcePermanentReference(),
		Multiplier: 2,
	}

	cases := []struct {
		blockers int
		want     int
	}{
		{blockers: 0, want: 0},
		{blockers: 1, want: 2},
		{blockers: 2, want: 4},
		{blockers: 3, want: 6},
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
			t.Errorf("+2/+2 for each of %d blockers = %d, want %d", tc.blockers, got, tc.want)
		}
	}
}

// TestBlockingCreaturesZeroOutsideCombat confirms the amount is zero when no
// combat is in progress, so a becomes-blocked trigger that somehow resolves
// outside combat adds nothing.
func TestBlockingCreaturesZeroOutsideCombat(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Rabid Elephant", 1, 3))
	obj := &game.StackObject{Controller: game.Player1, SourceID: attacker.ObjectID}
	dynamic := game.DynamicAmount{
		Kind:       game.DynamicAmountBlockingCreatures,
		Object:     game.SourcePermanentReference(),
		Multiplier: 2,
	}
	if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != 0 {
		t.Errorf("blocking-creatures amount outside combat = %d, want 0", got)
	}
}
