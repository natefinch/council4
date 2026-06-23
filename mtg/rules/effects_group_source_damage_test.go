package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGroupSourceDamageHitsEachController verifies that GroupSourceDamage has
// every creature deal its amount to the player who controls it ("Each creature
// deals 1 damage to its controller."): each controller loses life once per
// creature they control.
func TestGroupSourceDamageHitsEachController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	startP1 := g.Players[game.Player1].Life
	startP2 := g.Players[game.Player2].Life
	addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.GroupSourceDamage{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			}),
			Amount: game.Fixed(1),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != startP1-1 {
		t.Fatalf("Player1 life = %d, want %d (1 creature)", got, startP1-1)
	}
	if got := g.Players[game.Player2].Life; got != startP2-2 {
		t.Fatalf("Player2 life = %d, want %d (2 creatures)", got, startP2-2)
	}
}

// TestGroupSelfPowerDamageHitsEachMember verifies that GroupSelfPowerDamage has
// every creature in the group deal its own power to itself ("Each creature deals
// damage to itself equal to its power."): each creature marks damage equal to its
// power, computed per member.
func TestGroupSelfPowerDamageHitsEachMember(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	two := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	three := addCombatCreaturePermanentWithPower(g, game.Player2, 3)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.GroupSelfPowerDamage{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			}),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if two.MarkedDamage != 2 {
		t.Fatalf("power-2 creature marked damage = %d, want 2", two.MarkedDamage)
	}
	if three.MarkedDamage != 3 {
		t.Fatalf("power-3 creature marked damage = %d, want 3", three.MarkedDamage)
	}
}

// TestGroupSelfPowerDamageRespectsGroupFilter verifies that a filtered group
// ("each tapped creature") only damages members matching the selection: a tapped
// creature marks its own power while an untapped creature is untouched.
func TestGroupSelfPowerDamageRespectsGroupFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	tapped := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	tapped.Tapped = true
	untapped := addCombatCreaturePermanentWithPower(g, game.Player2, 5)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.GroupSelfPowerDamage{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Tapped:        game.TriTrue,
			}),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if tapped.MarkedDamage != 4 {
		t.Fatalf("tapped creature marked damage = %d, want 4", tapped.MarkedDamage)
	}
	if untapped.MarkedDamage != 0 {
		t.Fatalf("untapped creature marked damage = %d, want 0", untapped.MarkedDamage)
	}
}
