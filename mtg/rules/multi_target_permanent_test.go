package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMultiTargetDestroyDestroysEachChosenTarget proves the multi-instruction
// destroy sequence the cardgen backend emits for plural and optional permanent
// targets ("Destroy two target creatures.", "Destroy up to two target
// creatures.") destroys every chosen target, and that an "up to" destroy with
// fewer chosen targets than instructions safely no-ops the unfilled slots (the
// Destroy primitive resolves nothing for an out-of-range target index) without
// touching untargeted permanents.
func TestMultiTargetDestroyDestroysEachChosenTarget(t *testing.T) {
	destroySequence := func(slots int) []game.Instruction {
		instructions := make([]game.Instruction, 0, slots)
		for i := range slots {
			instructions = append(instructions, game.Instruction{
				Primitive: game.Destroy{Object: game.TargetPermanentReference(i)},
			})
		}
		return instructions
	}

	t.Run("all chosen targets destroyed", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		first := addCreaturePermanent(g, game.Player2)
		second := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, destroySequence(2), []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		for _, target := range []*game.Permanent{first, second} {
			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("chosen target remained on battlefield")
			}
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
				t.Fatal("chosen target was not put into owner's graveyard")
			}
		}
	})

	t.Run("up to N with fewer chosen no-ops empty slots", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		chosen := addCreaturePermanent(g, game.Player2)
		untargeted := addCreaturePermanent(g, game.Player2)
		// Two Destroy instructions ("up to two") but only one chosen target.
		addInstructionSpellToStackForController(g, game.Player1, destroySequence(2), []game.Target{
			game.PermanentTarget(chosen.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, chosen.ObjectID); ok {
			t.Fatal("chosen target remained on battlefield")
		}
		if !g.Players[game.Player2].Graveyard.Contains(chosen.CardInstanceID) {
			t.Fatal("chosen target was not put into owner's graveyard")
		}
		if _, ok := permanentByObjectID(g, untargeted.ObjectID); !ok {
			t.Fatal("untargeted permanent was destroyed by an empty slot")
		}
	})
}

// TestMultiTargetBounceReturnsEachChosenTarget proves the multi-instruction
// bounce sequence the cardgen backend emits for plural battlefield bounce
// ("Return two target creatures to their owners' hands.", "Return up to two
// target creatures to their owners' hands.") returns every chosen target to its
// owner's hand, that an "up to" bounce with fewer chosen targets than
// instructions safely no-ops the unfilled slots without touching untargeted
// permanents, and that the single-slot optional bounce ("Return up to one target
// creature to its owner's hand.") returns nothing when its target is declined.
func TestMultiTargetBounceReturnsEachChosenTarget(t *testing.T) {
	bounceSequence := func(slots int) []game.Instruction {
		instructions := make([]game.Instruction, 0, slots)
		for i := range slots {
			instructions = append(instructions, game.Instruction{
				Primitive: game.Bounce{Object: game.TargetPermanentReference(i)},
			})
		}
		return instructions
	}

	t.Run("all chosen targets bounced", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		first := addCreaturePermanent(g, game.Player2)
		second := addCreaturePermanent(g, game.Player2)
		addInstructionSpellToStackForController(g, game.Player1, bounceSequence(2), []game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		for _, target := range []*game.Permanent{first, second} {
			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("chosen target remained on battlefield")
			}
			if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
				t.Fatal("chosen target was not returned to owner's hand")
			}
		}
	})

	t.Run("up to N with fewer chosen no-ops empty slots", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		chosen := addCreaturePermanent(g, game.Player2)
		untargeted := addCreaturePermanent(g, game.Player2)
		// Two Bounce instructions ("up to two") but only one chosen target.
		addInstructionSpellToStackForController(g, game.Player1, bounceSequence(2), []game.Target{
			game.PermanentTarget(chosen.ObjectID),
		})

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, chosen.ObjectID); ok {
			t.Fatal("chosen target remained on battlefield")
		}
		if !g.Players[game.Player2].Hand.Contains(chosen.CardInstanceID) {
			t.Fatal("chosen target was not returned to owner's hand")
		}
		if _, ok := permanentByObjectID(g, untargeted.ObjectID); !ok {
			t.Fatal("untargeted permanent was bounced by an empty slot")
		}
	})

	t.Run("up to one declined returns nothing", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		untargeted := addCreaturePermanent(g, game.Player2)
		// Single Bounce instruction ("up to one") with the optional target
		// declined: the lone slot has no chosen target and must no-op.
		addInstructionSpellToStackForController(g, game.Player1, bounceSequence(1), nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if _, ok := permanentByObjectID(g, untargeted.ObjectID); !ok {
			t.Fatal("permanent was bounced by a declined optional slot")
		}
		if g.Players[game.Player2].Hand.Contains(untargeted.CardInstanceID) {
			t.Fatal("permanent was returned to hand by a declined optional slot")
		}
	})
}
