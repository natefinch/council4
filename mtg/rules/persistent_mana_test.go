package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// combinationManaCountAbility builds an add-mana instruction that adds one mana
// per permanent matching selection, split freely among colors. It models the
// dynamic-amount combination output of Grand Warlord Radha ("add that much mana
// in any combination of {R} and/or {G}", counting attacking creatures) and
// Axebane Guardian ("Add X mana in any combination of colors", counting
// defenders), whose lowered amount is a count-of-matching-creatures dynamic.
func combinationManaCountAbility(colors []mana.Color, selection game.Selection) game.AddMana {
	return game.AddMana{
		CombinationColors: colors,
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountSelector,
			Multiplier: 1,
			Group:      game.BattlefieldGroup(selection),
		}),
	}
}

// TestDynamicCombinationManaAttackerCount proves the attacker-count dynamic
// amount that unlocks Grand Warlord Radha resolves to the number of attacking
// creatures the controller controls: opponents' attackers and the controller's
// non-attacking creatures are excluded.
func TestDynamicCombinationManaAttackerCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	attacker1 := addCombatCreaturePermanent(g, game.Player1)
	attacker2 := addCombatCreaturePermanent(g, game.Player1)
	addCombatCreaturePermanent(g, game.Player1) // controlled but not attacking
	opponentAttacker := addCombatCreaturePermanent(g, game.Player2)

	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: attacker1.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: attacker2.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: opponentAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}}

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, combinationManaCountAbility(
		[]mana.Color{mana.R, mana.G},
		game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
			CombatState:   game.CombatStateAttacking,
		},
	), &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Total(); got != 2 {
		t.Fatalf("total mana = %d, want 2 (one per attacking creature you control)", got)
	}
}

// TestDynamicCombinationManaDefenderCount proves the defender-count dynamic
// amount that unlocks Axebane Guardian resolves to the number of creatures with
// defender the controller controls: opponents' defenders and the controller's
// non-defender creatures are excluded.
func TestDynamicCombinationManaDefenderCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addCombatCreaturePermanent(g, game.Player1, game.Defender)
	addCombatCreaturePermanent(g, game.Player1, game.Defender)
	addCombatCreaturePermanent(g, game.Player1, game.Defender)
	addCombatCreaturePermanent(g, game.Player1)                // controlled but no defender
	addCombatCreaturePermanent(g, game.Player2, game.Defender) // opponent's defender

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, combinationManaCountAbility(
		[]mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
		game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
			Keyword:       game.Defender,
		},
	), &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Total(); got != 3 {
		t.Fatalf("total mana = %d, want 3 (one per defender you control)", got)
	}
}

// TestPersistentManaSurvivesStepAndPhaseBoundaries proves the "Until end of turn,
// you don't lose this mana as steps and phases end" rider (Grand Warlord Radha):
// mana added with PersistUntilEndOfTurn survives every step- and phase-ending
// mana empty for the rest of the turn, then empties normally once end-of-turn
// cleanup releases the reservation.
func TestPersistentManaSurvivesStepAndPhaseBoundaries(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.AddMana{
		ManaColor:             mana.R,
		Amount:                game.Fixed(3),
		PersistUntilEndOfTurn: true,
	}, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 3 {
		t.Fatalf("red mana after add = %d, want 3", got)
	}

	// Each step and phase ends by emptying every pool; persistent mana survives.
	for i := range 3 {
		emptyManaPools(g)
		if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 3 {
			t.Fatalf("red mana after step/phase boundary %d = %d, want 3 (persists until end of turn)", i, got)
		}
	}

	// End-of-turn cleanup releases the reservation, so the following empty removes
	// the mana like any other.
	clearPersistentMana(g)
	emptyManaPools(g)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("total mana after end-of-turn cleanup = %d, want 0", got)
	}
}

// TestNonPersistentManaEmptiesAtBoundary is the control for the persistent-mana
// case: ordinary added mana empties at the very next step- or phase-ending empty,
// confirming persistence is opt-in and does not change default emptying.
func TestNonPersistentManaEmptiesAtBoundary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.AddMana{
		ManaColor: mana.R,
		Amount:    game.Fixed(3),
	}, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 3 {
		t.Fatalf("red mana after add = %d, want 3", got)
	}
	emptyManaPools(g)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("total mana after boundary = %d, want 0 (ordinary mana empties)", got)
	}
}
