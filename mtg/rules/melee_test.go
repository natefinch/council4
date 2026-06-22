package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMeleeCountScalesWithDistinctOpponentsAttacked covers the Melee count
// (CR 702.72): "for each opponent you attacked this combat" is the number of
// distinct opponents that creatures the controller controls are attacking this
// combat, read from the current combat's attack declarations at resolution.
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
		declarations := make([]game.AttackDeclaration, 0, len(tc.targets))
		for _, target := range tc.targets {
			attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Soldier", 1, 1))
			declarations = append(declarations, game.AttackDeclaration{
				Attacker: attacker.ObjectID,
				Target:   game.AttackTarget{Player: target},
			})
		}
		g.Combat = &game.CombatState{Attackers: declarations}
		if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != tc.want {
			t.Errorf("%s: Melee count = %d, want %d", tc.name, got, tc.want)
		}
	}
}

// TestMeleeCountIgnoresOpponentsAttacks confirms only the controller's own
// attackers count: an opponent attacking a third player does not contribute to
// the controller's Melee count.
func TestMeleeCountIgnoresOpponentsAttacks(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, vanillaCreature("Adriana", 3, 3))
	dynamic := game.DynamicAmount{Kind: game.DynamicAmountOpponentsAttackedThisCombat}
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}

	mine := addCombatPermanent(g, game.Player1, vanillaCreature("Soldier", 1, 1))
	theirs := addCombatPermanent(g, game.Player2, vanillaCreature("Raider", 2, 2))
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: mine.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: theirs.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}
	if got := dynamicAmountValue(g, obj, game.Player1, dynamic); got != 1 {
		t.Errorf("Melee count = %d, want 1 (only the controller's attack on Player2)", got)
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
