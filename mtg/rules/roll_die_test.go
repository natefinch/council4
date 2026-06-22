package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestRollDieCreatesTokensEqualToResult verifies the "roll a d20. You create a
// number of Treasure tokens equal to the result." sequence (Ancient Copper
// Dragon): the RollDie instruction publishes its rolled value, which the
// following CreateToken reads via a previous-effect-result amount, so the number
// of Treasure tokens created equals the die roll. The engine RNG is seeded so
// the roll is deterministic; an identical stream computes the expected result.
func TestRollDieCreatesTokensEqualToResult(t *testing.T) {
	const seed1, seed2 = 42, 99
	expected := rand.New(rand.NewPCG(seed1, seed2)).IntN(20) + 1
	if expected < 1 || expected > 20 {
		t.Fatalf("expected roll = %d, want within [1,20]", expected)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(seed1, seed2)))
	treasure := &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}

	const resultKey = game.ResultKey("die-roll-result")
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive:     game.RollDie{Sides: 20},
			PublishResult: resultKey,
		},
		{Primitive: game.CreateToken{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountPreviousEffectResult,
				ResultKey: resultKey,
			}),
			Source: game.TokenDef(treasure),
		}},
	})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Treasure"); got != expected {
		t.Fatalf("Treasure tokens = %d, want %d (equal to the d20 result)", got, expected)
	}
}

// TestRollDieResultWithinDieRange resolves the die roll on its own and asserts
// the published result is a valid d20 face (1..20), proving the roll uses the
// engine RNG and the +1 offset (no zero result).
func TestRollDieResultWithinDieRange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(7, 13)))
	const resultKey = game.ResultKey("die-roll-result")
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.RollDie{Sides: 20}, PublishResult: resultKey},
	})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Stack.Peek() = false, want the pushed spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	rolled := obj.ResolvedAmounts[string(resultKey)]
	if rolled < 1 || rolled > 20 {
		t.Fatalf("rolled d20 result = %d, want within [1,20]", rolled)
	}
}
