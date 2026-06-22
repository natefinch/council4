package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// morbidInsteadDamageInstructions builds Brimstone Volley's fused spell
// sequence: a base 3 damage gated on no creature having died this turn and a
// larger 5 damage gated on a creature having died this turn (the Morbid
// "instead" replacement), both dealt to the spell's single any-target.
func morbidInsteadDamageInstructions() []game.Instruction {
	creatureDied := game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	}
	noCreatureDied := creatureDied
	noCreatureDied.Negate = true
	damage := func(amount int, condition game.Condition) game.Instruction {
		return game.Instruction{
			Primitive: game.Damage{
				Amount:    game.Fixed(amount),
				Recipient: game.AnyTargetDamageRecipient(0),
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
		}
	}
	return []game.Instruction{
		damage(3, noCreatureDied),
		damage(5, creatureDied),
	}
}

// TestMorbidInsteadDamageDealsBaseWhenNoCreatureDied proves Brimstone Volley's
// base 3 damage resolves (and its Morbid 5 alternative does not) when no
// creature has died this turn.
func TestMorbidInsteadDamageDealsBaseWhenNoCreatureDied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	start := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1,
		morbidInsteadDamageInstructions(),
		[]game.Target{game.PlayerTarget(game.Player2)},
	)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != start-3 {
		t.Fatalf("target player life = %d, want %d (3 base damage)", got, start-3)
	}
}

// TestMorbidInsteadDamageDealsMorbidWhenCreatureDied proves Brimstone Volley's
// Morbid 5 damage replaces the base 3 when a creature has died this turn.
func TestMorbidInsteadDamageDealsMorbidWhenCreatureDied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	emitCreatureDiedEvent(g)
	start := g.Players[game.Player2].Life

	addInstructionSpellToStackForController(g, game.Player1,
		morbidInsteadDamageInstructions(),
		[]game.Target{game.PlayerTarget(game.Player2)},
	)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != start-5 {
		t.Fatalf("target player life = %d, want %d (5 morbid damage)", got, start-5)
	}
}
