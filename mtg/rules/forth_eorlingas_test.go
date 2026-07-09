package rules

import (
	"testing"

	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
)

// scheduleForthEorlingas casts the real Forth Eorlingas! card as Player1 and
// resolves it through the real spell-resolution path so its "whenever one or more
// creatures you control deal combat damage to one or more players this turn, you
// become the monarch" delayed trigger is scheduled. The spell's X is zero here
// (no tokens are created), which is irrelevant to the delayed trigger under test.
func scheduleForthEorlingas(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	addImplementationSpellToStack(g, game.Player1, cardf.ForthEorlingas(), nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1 (the OneOrMore become-monarch trigger)", len(g.DelayedTriggers))
	}
	pattern := g.DelayedTriggers[0].EventPattern
	if !pattern.Exists || !pattern.Val.OneOrMore || pattern.Val.Event != game.EventDamageDealt {
		t.Fatalf("delayed trigger pattern = %+v, want a OneOrMore EventDamageDealt pattern", pattern)
	}
}

// TestForthEorlingasFiresOncePerSimultaneousCombatDamageBatch proves the
// OneOrMore batch coalescing in drainReadyEventDelayedTriggers: when two
// creatures Player1 controls deal combat damage to a player in the same
// simultaneous batch (shared SimultaneousID, as batchCombatDamageEvents assigns),
// Forth Eorlingas' "whenever one or more creatures ... this turn" delayed trigger
// fires exactly once (CR 603.3e), not once per event. Without the coalesce it
// would put two become-monarch abilities on the stack.
func TestForthEorlingasFiresOncePerSimultaneousCombatDamageBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	log := TurnLog{}

	attacker1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attacker2 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	scheduleForthEorlingas(t, g, engine)

	// Advance the trigger cursor past the resolution so the drain that follows
	// only sees the combat-damage batch, exactly as it would after the spell
	// resolves in a main phase and combat happens later in the turn.
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log)
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &log)
	}

	// Two of Player1's creatures connect at the same time. batchCombatDamageEvents
	// stamps both EventDamageDealt events with a shared SimultaneousID, which is
	// what makes them a single "one or more creatures" batch.
	eventStart := len(g.Events)
	dealPlayerDamage(g, attacker1.ObjectID, attacker1.ObjectID, game.Player1, game.Player2, 2, true)
	dealPlayerDamage(g, attacker2.ObjectID, attacker2.ObjectID, game.Player1, game.Player2, 2, true)
	batchCombatDamageEvents(g, eventStart)

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Forth Eorlingas delayed trigger did not fire when creatures dealt combat damage")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (a OneOrMore trigger fires once per simultaneous batch, not per event)", g.Stack.Size())
	}

	monarchEventsBefore := countEvents(g, game.EventBecameMonarch)
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &log)
	}
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 did not become the monarch after their creatures dealt combat damage")
	}
	if got := countEvents(g, game.EventBecameMonarch) - monarchEventsBefore; got != 1 {
		t.Fatalf("EventBecameMonarch fired %d times, want exactly 1", got)
	}
}

// countEvents reports how many emitted events of the given kind the game has seen.
func countEvents(g *game.Game, kind game.EventKind) int {
	count := 0
	for i := range g.Events {
		if g.Events[i].Kind == kind {
			count++
		}
	}
	return count
}
