package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMeleeCountScalesWithDistinctOpponentsAttacked covers the Melee count
// (CR 702.121a, 508.6): "for each opponent you attacked this combat" is the
// number of distinct opponents the controller declared attackers attacking this
// combat, read from the combat's attack history at resolution.
func TestMeleeCountScalesWithDistinctOpponentsAttacked(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		targets []game.PlayerID
		want    int
	}{
		{name: "one opponent", targets: []game.PlayerID{game.Player2}, want: 1},
		{name: "two distinct opponents", targets: []game.PlayerID{game.Player2, game.Player3}, want: 2},
		{name: "three distinct opponents", targets: []game.PlayerID{game.Player2, game.Player3, game.Player4}, want: 3},
		{name: "same opponent counted once", targets: []game.PlayerID{game.Player2, game.Player2}, want: 1},
	}
	dynamic := game.DynamicAmount{Kind: game.DynamicAmountOpponentsAttackedThisCombat}
	for _, tc := range cases {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatPermanent(g, game.Player1, vanillaCreature("Adriana", 3, 3))
		obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}
		g.Combat = &game.CombatState{}
		for _, target := range tc.targets {
			g.Combat.DeclaredPlayerAttacks[game.Player1][target] = true
		}
		if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != tc.want {
			t.Errorf("%s: Melee count = %d, want %d", tc.name, got, tc.want)
		}
	}
}

// TestMeleeCountIgnoresOtherPlayersAttacks confirms only the controller's own
// attacks count: another player's attack history, and attackers currently
// listed in combat that the controller never declared, do not contribute.
func TestMeleeCountIgnoresOtherPlayersAttacks(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, vanillaCreature("Adriana", 3, 3))
	dynamic := game.DynamicAmount{Kind: game.DynamicAmountOpponentsAttackedThisCombat}
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}

	entered := addCombatPermanent(g, game.Player1, vanillaCreature("Soldier", 1, 1))
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: entered.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	}}
	g.Combat.DeclaredPlayerAttacks[game.Player1][game.Player2] = true
	g.Combat.DeclaredPlayerAttacks[game.Player2][game.Player3] = true
	if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != 1 {
		t.Errorf("Melee count = %d, want 1 (only Player1's declared attack on Player2)", got)
	}
	if got := dynamicAmountValue(g, obj, game.Player3, dynamic); got != 0 {
		t.Errorf("Melee count for a player who attacked no one = %d, want 0", got)
	}
}

// TestMeleeCountZeroOutsideCombat confirms the count is zero when no combat is
// in progress, so a Melee trigger that somehow resolves outside combat adds
// nothing.
func TestMeleeCountZeroOutsideCombat(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, vanillaCreature("Adriana", 3, 3))
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}
	dynamic := game.DynamicAmount{Kind: game.DynamicAmountOpponentsAttackedThisCombat}
	if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != 0 {
		t.Errorf("Melee count outside combat = %d, want 0", got)
	}
}
