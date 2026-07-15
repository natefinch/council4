package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// activateNecropotencePayLife activates Player1's registered "Pay 1 life:" ability
// and resolves it off the stack, returning the id of the card exiled face down.
// It asserts the ability actually scheduled the delayed return so the timing tests
// below observe a real captured-card binding rather than an empty no-op. Player1
// must be the active player holding priority with life to spend and a non-empty
// library.
func activateNecropotencePayLife(t *testing.T, g *game.Game, engine *Engine, source *game.Permanent) id.ID {
	t.Helper()
	topID, ok := g.Players[game.Player1].Library.Top()
	if !ok {
		t.Fatal("library is empty; there is no top card to exile")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("activating the Pay 1 life ability failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(topID) {
		t.Fatalf("expected top card %v to be exiled face down at resolution", topID)
	}
	if len(g.DelayedTriggers) == 0 {
		t.Fatal("resolution did not schedule the delayed end-step return trigger")
	}
	return topID
}

// runControllerEndStep advances the game to the given player's end step and runs
// the real ending phase (which emits the beginning-of-end-step event, drains any
// ready delayed triggers onto the stack, and resolves them). Driving the actual
// phase runner rather than the delayed-trigger drain directly exercises the same
// end-step boundary that a full turn would.
func runControllerEndStep(engine *Engine, g *game.Game, active game.PlayerID) {
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = active
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
}

// TestNecropotenceDelayedReturnSameTurnEndStep proves the common case: a card
// exiled by an activation made before the end step (here in the precombat main
// phase) returns at that same turn's end step, because the delayed trigger already
// exists when the beginning-of-end-step event fires.
func TestNecropotenceDelayedReturnSameTurnEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Card", Types: []types.Card{types.Creature}}})

	exiled := activateNecropotencePayLife(t, g, engine, source)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("card was not returned to hand at the same turn's end step")
	}
	if g.Players[game.Player1].Exile.Contains(exiled) {
		t.Fatal("card remained in exile after the delayed return resolved")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed trigger count = %d, want 0 (the one-shot return was consumed)", len(g.DelayedTriggers))
	}
}

// TestNecropotenceDelayedReturnWaitsForControllersNextEndStep proves the timing
// nuance the ruling turns on: a card exiled by an activation made during the
// controller's own end step (after that end step's beginning event) is not
// returned that same end step, is not returned at an intervening opponent's end
// step (the trigger is keyed to the controller as active player), and returns at
// the controller's next end step.
func TestNecropotenceDelayedReturnWaitsForControllersNextEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Card", Types: []types.Card{types.Creature}}})

	// Player1's end step begins before Necropotence is activated: run it now with
	// nothing scheduled so the beginning-of-end-step event fires and its delayed
	// drain passes with no return pending.
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	// Player1 activates during that end step's priority window, after the beginning
	// event has already gone by.
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	g.Turn.PriorityPlayer = game.Player1
	exiled := activateNecropotencePayLife(t, g, engine, source)
	if g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("card returned during the very end step it was exiled in; it must wait for the next one")
	}

	// The opponent's end step must not fire the controller-keyed delayed return.
	runControllerEndStep(engine, g, game.Player2)
	if g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("opponent's end step fired the controller-keyed delayed return")
	}
	if !g.Players[game.Player1].Exile.Contains(exiled) {
		t.Fatal("card left exile before the controller's next end step")
	}

	// The controller's next end step returns it.
	runControllerEndStep(engine, g, game.Player1)
	if !g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("controller's next end step did not return the exiled card")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed trigger count = %d, want 0 (the one-shot return was consumed)", len(g.DelayedTriggers))
	}
}

// TestNecropotenceDelayedReturnAfterSourceLeaves proves the delayed return is
// independent of Necropotence: once scheduled it returns the card even if
// Necropotence has left the battlefield, because the trigger froze both its
// controller and the captured card at schedule time (CR 603.3d).
func TestNecropotenceDelayedReturnAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Card", Types: []types.Card{types.Creature}}})

	exiled := activateNecropotencePayLife(t, g, engine, source)
	removePermanentFromBattlefield(g, source.ObjectID)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("delayed return did not fire after Necropotence left the battlefield")
	}
}

// TestNecropotenceDelayedReturnNoOpsWhenCardLeftExile proves the return fails
// closed: if the exiled card has already left exile when the end step arrives, the
// trigger does nothing rather than pulling the card out of its new zone, because
// the move requires the captured card to still be in exile.
func TestNecropotenceDelayedReturnNoOpsWhenCardLeftExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Card", Types: []types.Card{types.Creature}}})

	exiled := activateNecropotencePayLife(t, g, engine, source)
	// Another effect moves the exiled card to the graveyard before the end step.
	g.Players[game.Player1].Exile.Remove(exiled)
	g.Players[game.Player1].Graveyard.Add(exiled)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if g.Players[game.Player1].Hand.Contains(exiled) {
		t.Fatal("card that had left exile was pulled into hand; the return must be exile-only")
	}
	if !g.Players[game.Player1].Graveyard.Contains(exiled) {
		t.Fatal("card that had left exile was disturbed; the return must no-op")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed trigger count = %d, want 0 (the one-shot return was still consumed)", len(g.DelayedTriggers))
	}
}

// TestNecropotenceDelayedReturnMultipleActivationsReturnOwnCards proves two
// activations in one turn stay isolated even though they publish under the same
// link key: each delayed trigger froze the card that activation exiled, so both
// distinct cards return rather than one being returned twice and the other leaking.
func TestNecropotenceDelayedReturnMultipleActivationsReturnOwnCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "First", Types: []types.Card{types.Creature}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Second", Types: []types.Card{types.Creature}}})

	firstExiled := activateNecropotencePayLife(t, g, engine, source)
	secondExiled := activateNecropotencePayLife(t, g, engine, source)
	if firstExiled == secondExiled {
		t.Fatal("two activations exiled the same card; the test cannot prove isolation")
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed trigger count = %d, want 2 (one per activation)", len(g.DelayedTriggers))
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(firstExiled) {
		t.Fatal("first activation's exiled card was not returned")
	}
	if !g.Players[game.Player1].Hand.Contains(secondExiled) {
		t.Fatal("second activation's exiled card was not returned (its trigger must capture its own card)")
	}
	if got := g.Players[game.Player1].Exile.Size(); got != 0 {
		t.Fatalf("exile size = %d, want 0 (both exiled cards returned)", got)
	}
}

// TestNecropotenceDelayedReturnGoesToOwnersHand proves the card returns to its
// owner's hand: the move sends the captured card to the card owner, which for a
// card exiled off the owner's own library is the activating player.
func TestNecropotenceDelayedReturnGoesToOwnersHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := setUpNecropotenceActivation(g, 20)
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Owned Card", Types: []types.Card{types.Creature}}})

	exiled := activateNecropotencePayLife(t, g, engine, source)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(topID) {
		t.Fatal("card did not return to its owner's hand")
	}
	if g.Players[game.Player2].Hand.Contains(exiled) {
		t.Fatal("card returned to the wrong player's hand")
	}
}
