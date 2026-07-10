package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCounterTriggeringStackObject verifies "counter that spell or ability"
// resolves the stack object carried by the became-target trigger event without
// using a target slot.
func TestCounterTriggeringStackObject(t *testing.T) {
	t.Parallel()
	if err := game.ValidateInstructionSequence([]game.Instruction{{
		Primitive: game.CounterObject{Object: game.EventStackObjectReference()},
	}}, nil); err != nil {
		t.Fatalf("event-stack counter validation failed: %v", err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellCardID := addEffectSpellToStack(g, game.Player2, game.Draw{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
	}, nil)
	spell, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("triggering spell missing from stack")
	}

	source := addCombatPermanent(g, game.Player1, vanillaCreature("Glasskite", 2, 3))
	trigger := game.TriggeredAbility{
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.CounterObject{Object: game.EventStackObjectReference()},
		}}}.Ability(),
	}
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		InlineTrigger:   &trigger,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:          game.EventObjectBecameTarget,
			StackObjectID: spell.ID,
		},
	})

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 after trigger counters spell", got)
	}
	if !g.Players[game.Player2].Graveyard.Contains(spellCardID) {
		t.Fatal("countered spell was not put into its owner's graveyard")
	}
}
