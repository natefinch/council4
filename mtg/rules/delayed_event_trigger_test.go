package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func scheduleEventDelayedTrigger(g *game.Game, controller game.PlayerID, oneShot bool) {
	def := &game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:      game.EventSpellCast,
			Controller: game.TriggerControllerYou,
		}),
		OneShot: oneShot,
		Window:  game.DelayedWindowThisTurn,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, &game.StackObject{Controller: controller}, def) {
		panic("scheduleDelayedTrigger returned false")
	}
}

// TestEventDelayedTriggerRepeatingFiresEachMatchingCastThisTurn verifies a
// "whenever you cast a spell this turn" event-based delayed trigger fires on each
// matching cast event, ignores an opponent's cast, and is removed at cleanup so
// it no longer fires next turn.
func TestEventDelayedTriggerRepeatingFiresEachMatchingCastThisTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	scheduleEventDelayedTrigger(g, game.Player1, false)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first matching cast did not fire the delayed trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	g.Stack.Pop()
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("repeating delayed trigger removed after one fire: %d", len(g.DelayedTriggers))
	}

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent's cast fired a controller-scoped delayed trigger")
	}

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second matching cast did not fire the repeating delayed trigger")
	}
	g.Stack.Pop()

	expireEventDelayedTriggers(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived cleanup: %d", len(g.DelayedTriggers))
	}
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delayed trigger fired after its window ended")
	}
}

// TestEventDelayedTriggerOneShotFiresOnceThenRemoved verifies a "the next time
// you cast a spell this turn" one-shot event-based delayed trigger fires on the
// first matching cast and is then removed.
func TestEventDelayedTriggerOneShotFiresOnceThenRemoved(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	scheduleEventDelayedTrigger(g, game.Player1, true)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-shot delayed trigger did not fire on the first matching cast")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("one-shot delayed trigger survived its fire: %d", len(g.DelayedTriggers))
	}
	g.Stack.Pop()

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-shot delayed trigger fired a second time")
	}
}
