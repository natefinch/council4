package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// regOrderedMultiBlock sets up one attacker blocked by several blockers in a
// fixed blocker order so deterministic damage assignment is reproducible.
func regOrderedMultiBlock(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) {
	declarations := make([]game.BlockDeclaration, 0, len(blockers))
	order := make([]id.ID, 0, len(blockers))
	for _, blocker := range blockers {
		declarations = append(declarations, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
		order = append(order, blocker.ObjectID)
	}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  declarations,
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: order,
		},
	}
}

// regDeterministicAssignment mirrors the engine's default lethal-in-order
// assignment: each blocker but the last receives up to lethal damage in order,
// and the final blocker absorbs all remaining damage.
func regDeterministicAssignment(power int, toughnesses []int) []int {
	marks := make([]int, len(toughnesses))
	remaining := power
	for i, toughness := range toughnesses {
		if remaining <= 0 {
			continue
		}
		assign := remaining
		if i < len(toughnesses)-1 {
			assign = min(remaining, toughness)
		}
		marks[i] = assign
		remaining -= assign
	}
	return marks
}

// TestRegMultiBlockLethalBeforeMovingOn asserts a blocked attacker must assign
// lethal damage to each blocker in order before assigning to the next.
func TestRegMultiBlockLethalBeforeMovingOn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	third := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	regOrderedMultiBlock(g, attacker, []*game.Permanent{first, second, third})

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	// 5 power: lethal 2 to first, lethal 2 to second, remaining 1 to last.
	if first.MarkedDamage != 2 || second.MarkedDamage != 2 || third.MarkedDamage != 1 {
		t.Fatalf("blocker damage = %d/%d/%d, want lethal-in-order 2/2/1",
			first.MarkedDamage, second.MarkedDamage, third.MarkedDamage)
	}
}

// TestRegMultiBlockAssignmentProperty fuzzes attacker power and blocker
// toughnesses and asserts the engine assigns damage exactly as the
// lethal-in-order reference model, never skipping ahead in the blocker order.
func TestRegMultiBlockAssignmentProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 5))
	for iteration := range 400 {
		power := 1 + rng.IntN(10)
		blockerCount := 2 + rng.IntN(3)
		toughnesses := make([]int, blockerCount)
		blockers := make([]*game.Permanent, blockerCount)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		attacker := addCombatCreaturePermanentWithPower(g, game.Player1, power)
		for i := range blockerCount {
			toughnesses[i] = 1 + rng.IntN(4)
			blockers[i] = addCombatCreaturePermanentWithPower(g, game.Player2, toughnesses[i])
		}
		regOrderedMultiBlock(g, attacker, blockers)

		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

		want := regDeterministicAssignment(power, toughnesses)
		sawUnderfed := false
		for i, blocker := range blockers {
			if blocker.MarkedDamage != want[i] {
				t.Fatalf("iteration %d: blocker %d damage = %d, want %d (power %d toughnesses %v)",
					iteration, i, blocker.MarkedDamage, want[i], power, toughnesses)
			}
			// Lethal-before-moving-on invariant: once a non-final blocker is
			// assigned less than lethal, no later blocker may be assigned damage.
			if sawUnderfed && blocker.MarkedDamage != 0 {
				t.Fatalf("iteration %d: blocker %d received %d after an earlier blocker was under-assigned",
					iteration, i, blocker.MarkedDamage)
			}
			if i < blockerCount-1 && blocker.MarkedDamage < toughnesses[i] {
				sawUnderfed = true
			}
		}
	}
}

// TestRegDoubleStrikeUnblockedDealsTwicePowerProperty asserts that a
// double-strike attacker deals its power in both the first-strike and normal
// combat damage steps.
func TestRegDoubleStrikeUnblockedDealsTwicePowerProperty(t *testing.T) {
	t.Parallel()
	for power := 2; power <= 6; power++ {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addCombatCreaturePermanentWithPower(g, game.Player1, power, game.DoubleStrike)
		engine := NewEngine(nil)
		log := TurnLog{}

		engine.runCombatPhase(g, allFirstLegalAgents(), &log)

		want := 40 - 2*power
		if g.Players[game.Player2].Life != want {
			t.Fatalf("power %d: defending life = %d, want %d (double strike)", power, g.Players[game.Player2].Life, want)
		}
		if len(log.CombatDamage) != 2 {
			t.Fatalf("power %d: combat damage logs = %d, want 2 passes", power, len(log.CombatDamage))
		}
	}
}

// TestRegFirstStrikeKillsBlockerBeforeItStrikesBackProperty asserts a
// first-strike attacker destroys an equally lethal blocker in the first-strike
// step, so the blocker never deals its normal combat damage back.
func TestRegFirstStrikeKillsBlockerBeforeItStrikesBackProperty(t *testing.T) {
	t.Parallel()
	for size := 2; size <= 5; size++ {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		attacker := addCombatCreaturePermanentWithPower(g, game.Player1, size, game.FirstStrike)
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, size)
		engine := NewEngine(nil)
		log := TurnLog{}

		engine.runCombatPhase(g, allFirstLegalAgents(), &log)

		if _, ok := permanentByObjectID(g, attacker.ObjectID); !ok {
			t.Fatalf("size %d: first-strike attacker died to a blocker it should have killed first", size)
		}
		if attacker.MarkedDamage != 0 {
			t.Fatalf("size %d: first-strike attacker marked damage = %d, want 0", size, attacker.MarkedDamage)
		}
		if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
			t.Fatalf("size %d: blocker survived lethal first-strike damage", size)
		}
	}
}
