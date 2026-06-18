package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMultiTargetExileExilesEachChosenTarget proves the multi-instruction exile
// sequence the cardgen backend emits for plural and optional permanent targets
// ("Exile two target creatures.", "Exile up to two target creatures.") exiles
// every chosen target, and that an "up to" exile with fewer chosen targets than
// instructions safely no-ops the unfilled slots (the Exile primitive resolves
// nothing for an out-of-range target index) without touching untargeted
// permanents.
func TestMultiTargetExileExilesEachChosenTarget(t *testing.T) {
	exileSequence := func(slots int) []game.Instruction {
		instructions := make([]game.Instruction, 0, slots)
		for i := range slots {
			instructions = append(instructions, game.Instruction{
				Primitive: game.Exile{Object: game.TargetPermanentReference(i)},
			})
		}
		return instructions
	}

	t.Run("all chosen targets exiled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		first := addCreaturePermanent(g, game.Player2)
		second := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, exileSequence(2), []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		for _, target := range []*game.Permanent{first, second} {
			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("chosen target remained on battlefield")
			}
			if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
				t.Fatal("chosen target was not exiled to owner's exile zone")
			}
		}
	})

	t.Run("up to N with fewer chosen no-ops empty slots", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		chosen := addCreaturePermanent(g, game.Player2)
		untargeted := addCreaturePermanent(g, game.Player2)
		// Two Exile instructions ("up to two") but only one chosen target.
		addInstructionSpellToStackForController(g, game.Player1, exileSequence(2), []game.Target{
			game.PermanentTarget(chosen.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, chosen.ObjectID); ok {
			t.Fatal("chosen target remained on battlefield")
		}
		if !g.Players[game.Player2].Exile.Contains(chosen.CardInstanceID) {
			t.Fatal("chosen target was not exiled")
		}
		if _, ok := permanentByObjectID(g, untargeted.ObjectID); !ok {
			t.Fatal("untargeted permanent was exiled by an empty slot")
		}
	})

	t.Run("up to N with none chosen exiles nothing", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		untargeted := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, exileSequence(2), nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, untargeted.ObjectID); !ok {
			t.Fatal("permanent was exiled despite no chosen targets")
		}
	})
}
