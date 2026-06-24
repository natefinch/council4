package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCommitCrimeEmittedForOpponentPlayerTarget covers the core "commit a crime"
// event (CR 700.15): a spell that targets an opponent fires its controller's
// "Whenever you commit a crime" trigger.
func TestCommitCrimeEmittedForOpponentPlayerTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commit-a-crime trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want commit-a-crime trigger to draw one card", got)
	}
}

// TestCommitCrimeNotEmittedForOwnPlayerTarget confirms a spell that targets only
// its own controller is not a crime, so the trigger does not fire.
func TestCommitCrimeNotEmittedForOwnPlayerTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player1)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commit-a-crime trigger fired for a non-crime self-target")
	}
}

// TestCommitCrimeEmittedForOpponentControlledPermanent confirms that targeting a
// permanent an opponent controls commits a crime, and that an "Whenever an
// opponent commits a crime" trigger fires for the targeted player's controller.
func TestCommitCrimeEmittedForOpponentControlledPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerOpponent,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	victim := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Victim"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(victim.ObjectID)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent commit-a-crime trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want opponent commit-a-crime trigger to draw one card", got)
	}
}
