package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestSourcePowerDamageHitsOtherTarget verifies the Rabid Bite shape: the first
// target creature deals damage equal to its power to the second target, and the
// dealing creature itself is unharmed.
func TestSourcePowerDamageHitsOtherTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	dealer := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectPower,
				Multiplier: 1,
				Object:     game.TargetPermanentReference(0),
			}),
			Recipient:    game.AnyTargetDamageRecipient(1),
			DamageSource: opt.Val(game.TargetPermanentReference(0)),
		},
	}}, []game.Target{
		game.PermanentTarget(dealer.ObjectID),
		game.PermanentTarget(victim.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if victim.MarkedDamage != 4 {
		t.Fatalf("victim marked damage = %d, want 4 (dealer power)", victim.MarkedDamage)
	}
	if dealer.MarkedDamage != 0 {
		t.Fatalf("dealer marked damage = %d, want 0", dealer.MarkedDamage)
	}
}

// TestSourcePowerDamageHitsItself verifies the Justice Strike shape: the target
// creature deals damage to itself equal to its power.
func TestSourcePowerDamageHitsItself(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountObjectPower,
				Multiplier: 1,
				Object:     game.TargetPermanentReference(0),
			}),
			Recipient:    game.AnyTargetDamageRecipient(0),
			DamageSource: opt.Val(game.TargetPermanentReference(0)),
		},
	}}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 3 {
		t.Fatalf("self-damage marked = %d, want 3 (own power)", target.MarkedDamage)
	}
}

// TestEachOfTwoTargetsFixedDamage verifies the Furious Reprisal shape: a fixed
// amount of damage is dealt to each of two independently chosen targets.
func TestEachOfTwoTargetsFixedDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(1)}},
	}, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if first.MarkedDamage != 2 || second.MarkedDamage != 2 {
		t.Fatalf("each-of damage = %d/%d, want 2/2", first.MarkedDamage, second.MarkedDamage)
	}
}
