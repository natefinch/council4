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

// TestRollDieCreatesCreatureTokensEqualToResult verifies the creature-token
// payoff (Ancient Gold Dragon): "roll a d20. You create a number of 1/1 blue
// Faerie Dragon creature tokens ... equal to the result." The number of Faerie
// Dragon tokens created equals the deterministic die roll.
func TestRollDieCreatesCreatureTokensEqualToResult(t *testing.T) {
	const seed1, seed2 = 5, 17
	expected := rand.New(rand.NewPCG(seed1, seed2)).IntN(20) + 1

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(seed1, seed2)))
	faerie := &game.CardDef{CardFace: game.CardFace{
		Name:     "Faerie Dragon",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Faerie, types.Dragon},
	}}

	const resultKey = game.ResultKey("die-roll-result")
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.RollDie{Sides: 20}, PublishResult: resultKey},
		{Primitive: game.CreateToken{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountPreviousEffectResult,
				ResultKey: resultKey,
			}),
			Source: game.TokenDef(faerie),
		}},
	})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Faerie Dragon"); got != expected {
		t.Fatalf("Faerie Dragon tokens = %d, want %d (equal to the d20 result)", got, expected)
	}
}

// TestRollDieDrawsCardsEqualToResult verifies the draw payoff (Ancient Silver
// Dragon): "roll a d20. Draw cards equal to the result." The controller draws a
// number of cards equal to the deterministic die roll.
func TestRollDieDrawsCardsEqualToResult(t *testing.T) {
	const seed1, seed2 = 3, 8
	expected := rand.New(rand.NewPCG(seed1, seed2)).IntN(20) + 1

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(seed1, seed2)))
	for range 20 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}

	const resultKey = game.ResultKey("die-roll-result")
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.RollDie{Sides: 20}, PublishResult: resultKey},
		{Primitive: game.Draw{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountPreviousEffectResult,
				ResultKey: resultKey,
			}),
			Player: game.ControllerReference(),
		}},
	})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != expected {
		t.Fatalf("hand size = %d, want %d (drew cards equal to the d20 result)", got, expected)
	}
}
