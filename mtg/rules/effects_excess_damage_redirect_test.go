package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestExcessDamageRedirectDealsOnlyLethalToCreature proves the Pigment Storm /
// Flame Spill self-replacement: a spell that deals 5 damage to an indestructible
// 4/4 with "Excess damage is dealt to that creature's controller instead." marks
// only the creature's lethal damage (4) on it and deals the remaining 1 to its
// controller, rather than marking the full 5 and additionally dealing the excess.
func TestExcessDamageRedirectDealsOnlyLethalToCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 4, game.Indestructible)
	beforeP2 := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount:          game.Fixed(5),
			Recipient:       game.AnyTargetDamageRecipient(0),
			ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
		},
	}}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 4 {
		t.Fatalf("creature marked damage = %d, want 4 (lethal only, excess redirected)", target.MarkedDamage)
	}
	if lost := beforeP2 - g.Players[game.Player2].Life; lost != 1 {
		t.Fatalf("controller life lost = %d, want 1 (5 dealt minus 4 lethal)", lost)
	}
}

// TestExcessDamageRedirectToControllerOnDeath proves that when the creature is
// destroyed, only the excess beyond its lethal damage reaches the controller.
func TestExcessDamageRedirectToControllerOnDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	beforeP2 := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount:          game.Fixed(5),
			Recipient:       game.AnyTargetDamageRecipient(0),
			ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
		},
	}}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if lost := beforeP2 - g.Players[game.Player2].Life; lost != 3 {
		t.Fatalf("controller life lost = %d, want 3 (5 dealt minus 2 lethal)", lost)
	}
}

// TestExcessDamageRedirectNoExcessSparesController proves that when the damage
// does not exceed the creature's lethal damage, the controller takes nothing.
func TestExcessDamageRedirectNoExcessSparesController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5, game.Indestructible)
	beforeP2 := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount:          game.Fixed(3),
			Recipient:       game.AnyTargetDamageRecipient(0),
			ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
		},
	}}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 3 {
		t.Fatalf("creature marked damage = %d, want 3", target.MarkedDamage)
	}
	if lost := beforeP2 - g.Players[game.Player2].Life; lost != 0 {
		t.Fatalf("controller life lost = %d, want 0 (no excess)", lost)
	}
}
