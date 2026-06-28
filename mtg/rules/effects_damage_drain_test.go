package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDamageDealtThisWayDrainGainsLifeFromSingleTarget covers the Corrupt
// pattern: a single-recipient damage publishes the amount it dealt, and a
// follow-on "...equal to the damage dealt this way." life gain reads it.
func TestDamageDealtThisWayDrainGainsLifeFromSingleTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	beforeP1 := g.Players[game.Player1].Life
	beforeP2 := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(5), Recipient: game.AnyTargetDamageRecipient(0)},
			PublishResult: game.ResultKey("damage-dealt-this-way"),
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectResult,
					ResultKey: game.ResultKey("damage-dealt-this-way"),
				}),
			},
		},
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := beforeP2 - g.Players[game.Player2].Life; got != 5 {
		t.Fatalf("Player2 life lost = %d, want 5", got)
	}
	if got := g.Players[game.Player1].Life - beforeP1; got != 5 {
		t.Fatalf("Player1 life gained = %d, want 5 from damage dealt this way", got)
	}
}

// TestExcessDamageDealtThisWayDrainGainsOnlyExcess covers the Razor Rings
// pattern: the life gain reads only the damage beyond what was needed to
// destroy the recipient.
func TestExcessDamageDealtThisWayDrainGainsOnlyExcess(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	beforeP1 := g.Players[game.Player1].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(5), Recipient: game.AnyTargetDamageRecipient(0)},
			PublishResult: game.ResultKey("damage-dealt-this-way"),
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectExcessDamage,
					ResultKey: game.ResultKey("damage-dealt-this-way"),
				}),
			},
		},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life - beforeP1; got != 3 {
		t.Fatalf("Player1 life gained = %d, want 3 from 5 damage minus 2 toughness", got)
	}
}

// TestDamageDealtThisWayDrainSumsGroupDamage covers the Creeping Bloodsucker
// pattern: "deals 1 damage to each opponent. You gain life equal to the damage
// dealt this way." The life gain reads the total dealt across every opponent,
// not just one recipient's share.
func TestDamageDealtThisWayDrainSumsGroupDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	beforeP1 := g.Players[game.Player1].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive: game.Damage{
				Amount:    game.Fixed(1),
				Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
			},
			PublishResult: game.ResultKey("damage-dealt-this-way"),
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectResult,
					ResultKey: game.ResultKey("damage-dealt-this-way"),
				}),
			},
		},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life - beforeP1; got != game.NumPlayers-1 {
		t.Fatalf("Player1 life gained = %d, want %d (one per opponent)", got, game.NumPlayers-1)
	}
}
